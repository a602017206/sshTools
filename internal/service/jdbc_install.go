package service

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"AHaSSHTools/internal/config"
)

type DriverInstallService struct {
	paths   JDBCPaths
	fetcher ArtifactFetcher
}

type DriverInstallResult struct {
	DriverID    string
	ProfileID   string
	Version     string
	InstallPath string
}

type offlineDriverPackage struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	DriverClass string   `json:"driverClass"`
	URLTemplate string   `json:"urlTemplate"`
	DefaultPort int      `json:"defaultPort"`
	JRE         string   `json:"jre"`
	Jars        []string `json:"jars"`
}

func NewDriverInstallService(paths JDBCPaths) *DriverInstallService {
	return &DriverInstallService{
		paths:   paths,
		fetcher: NewArtifactDownloader(ArtifactDownloadOptions{}),
	}
}

func (s *DriverInstallService) ConfigureArtifactFetcher(fetcher ArtifactFetcher) {
	s.fetcher = fetcher
}

func (s *DriverInstallService) InstallProfile(ctx context.Context, driver config.JDBCDriver, profile config.JDBCDriverProfile) (*DriverInstallResult, error) {
	if err := s.validateInstallProfile(driver, profile); err != nil {
		return nil, err
	}
	temporaryDir, err := s.createInstallDirectory(driver.ID)
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(temporaryDir)
		}
	}()

	if err := s.downloadProfileJars(ctx, temporaryDir, profile.Jars); err != nil {
		return nil, err
	}
	metadata := offlinePackageFromProfile(driver, profile)
	if err := writeDriverPackageMetadata(temporaryDir, metadata); err != nil {
		return nil, err
	}

	targetDir := filepath.Join(s.paths.DriversDir, driver.ID, profile.Version)
	if err := replaceDriverVersionDirectory(temporaryDir, targetDir); err != nil {
		return nil, fmt.Errorf("提交 JDBC 驱动安装失败: %w", err)
	}
	committed = true
	return &DriverInstallResult{DriverID: driver.ID, ProfileID: profile.ID, Version: profile.Version, InstallPath: targetDir}, nil
}

func (s *DriverInstallService) validateInstallProfile(driver config.JDBCDriver, profile config.JDBCDriverProfile) error {
	if driver.ID == "" || profile.ID == "" || profile.Version == "" || len(profile.Jars) == 0 {
		return &JDBCError{Code: JDBCErrorDriverInvalid, Message: "JDBC 驱动 profile 不完整"}
	}
	if s.fetcher == nil {
		return &JDBCError{Code: JDBCErrorDriverInvalid, Message: "JDBC 驱动下载器未配置"}
	}
	for _, jar := range profile.Jars {
		if jar.URL == "" {
			return &JDBCError{Code: JDBCErrorDriverMissing, Message: fmt.Sprintf("%s %s 没有可用的官方在线下载地址，请离线导入驱动包", driver.Name, profile.Version)}
		}
		if jar.Name == "" || filepath.Base(jar.Name) != jar.Name {
			return &JDBCError{Code: JDBCErrorDriverInvalid, Message: "JDBC 驱动 jar 名称无效"}
		}
	}
	return nil
}

func (s *DriverInstallService) createInstallDirectory(driverID string) (string, error) {
	driverDir := filepath.Join(s.paths.DriversDir, driverID)
	if err := os.MkdirAll(driverDir, 0o700); err != nil {
		return "", fmt.Errorf("创建 JDBC 驱动目录失败: %w", err)
	}
	temporaryDir, err := os.MkdirTemp(driverDir, ".install-*")
	if err != nil {
		return "", fmt.Errorf("创建 JDBC 驱动临时目录失败: %w", err)
	}
	return temporaryDir, nil
}

func (s *DriverInstallService) downloadProfileJars(ctx context.Context, temporaryDir string, jars []config.JDBCJar) error {
	for _, jar := range jars {
		target := filepath.Join(temporaryDir, "jars", jar.Name)
		if err := s.fetcher.Download(ctx, jar.URL, jar.SHA256, target); err != nil {
			return &JDBCError{Code: JDBCErrorDriverInvalid, Message: fmt.Sprintf("下载 JDBC 驱动 %s 失败", jar.Name), Err: err}
		}
	}
	return nil
}

func offlinePackageFromProfile(driver config.JDBCDriver, profile config.JDBCDriverProfile) *offlineDriverPackage {
	metadata := &offlineDriverPackage{
		ID: driver.ID, Name: driver.Name, Version: profile.Version,
		DriverClass: profile.DriverClass, URLTemplate: profile.URLTemplate,
		DefaultPort: profile.DefaultPort, JRE: profile.JRERequirement,
		Jars: make([]string, 0, len(profile.Jars)),
	}
	for _, jar := range profile.Jars {
		metadata.Jars = append(metadata.Jars, jar.Name)
	}
	return metadata
}

func replaceDriverVersionDirectory(sourceDir, targetDir string) error {
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return os.Rename(sourceDir, targetDir)
	} else if err != nil {
		return fmt.Errorf("检查现有版本目录失败: %w", err)
	}

	parent := filepath.Dir(targetDir)
	backupDir, err := os.MkdirTemp(parent, ".driver-backup-*")
	if err != nil {
		return fmt.Errorf("创建版本备份路径失败: %w", err)
	}
	if err := os.Remove(backupDir); err != nil {
		return fmt.Errorf("准备版本备份路径失败: %w", err)
	}
	if err := os.Rename(targetDir, backupDir); err != nil {
		return fmt.Errorf("备份现有版本失败: %w", err)
	}
	if err := os.Rename(sourceDir, targetDir); err != nil {
		if restoreErr := os.Rename(backupDir, targetDir); restoreErr != nil {
			return fmt.Errorf("提交新版本失败: %v；恢复旧版本失败: %w", err, restoreErr)
		}
		return fmt.Errorf("提交新版本失败，已恢复旧版本: %w", err)
	}
	_ = os.RemoveAll(backupDir)
	return nil
}

func (s *DriverInstallService) ImportOfflinePackage(zipPath string) (*DriverInstallResult, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("打开 JDBC 离线驱动包失败: %w", err)
	}
	defer reader.Close()

	pkg, err := readOfflineDriverPackage(reader.File)
	if err != nil {
		return nil, err
	}
	checksums, err := readDriverChecksums(reader.File)
	if err != nil {
		return nil, err
	}
	files := mapZipFiles(reader.File)
	if err := validateDriverChecksums(files, checksums); err != nil {
		return nil, err
	}

	targetDir := filepath.Join(s.paths.DriversDir, pkg.ID, pkg.Version)
	if err := os.RemoveAll(targetDir); err != nil {
		return nil, fmt.Errorf("清理 JDBC 驱动安装目录失败: %w", err)
	}
	if err := os.MkdirAll(targetDir, 0o700); err != nil {
		return nil, fmt.Errorf("创建 JDBC 驱动安装目录失败: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(targetDir)
		}
	}()

	for _, jar := range pkg.Jars {
		name := filepath.ToSlash(filepath.Join("jars", jar))
		file, ok := files[name]
		if !ok {
			return nil, fmt.Errorf("JDBC 驱动包缺少 jar: %s", name)
		}
		if err := extractZipFile(file, filepath.Join(targetDir, name)); err != nil {
			return nil, err
		}
	}
	if err := writeDriverPackageMetadata(targetDir, pkg); err != nil {
		return nil, err
	}

	committed = true
	return &DriverInstallResult{
		DriverID:    pkg.ID,
		ProfileID:   pkg.ID + "-" + pkg.Version,
		Version:     pkg.Version,
		InstallPath: targetDir,
	}, nil
}

func readOfflineDriverPackage(files []*zip.File) (*offlineDriverPackage, error) {
	data, err := readZipFile(files, "package.json")
	if err != nil {
		return nil, err
	}
	var pkg offlineDriverPackage
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("解析 JDBC 驱动包元数据失败: %w", err)
	}
	if pkg.ID == "" || pkg.Version == "" {
		return nil, fmt.Errorf("JDBC 驱动包缺少 id 或 version")
	}
	return &pkg, nil
}

func readDriverChecksums(files []*zip.File) (map[string]string, error) {
	data, err := readZipFile(files, "checksums.sha256")
	if err != nil {
		return nil, err
	}
	checksums := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("JDBC 驱动 checksum 格式无效: %s", line)
		}
		checksums[filepath.ToSlash(parts[1])] = parts[0]
	}
	return checksums, nil
}

func validateDriverChecksums(files map[string]*zip.File, checksums map[string]string) error {
	for name, expected := range checksums {
		file, ok := files[name]
		if !ok {
			return fmt.Errorf("JDBC 驱动包缺少 checksum 文件: %s", name)
		}
		actual, err := zipFileSHA256(file)
		if err != nil {
			return err
		}
		if !strings.EqualFold(actual, expected) {
			return fmt.Errorf("JDBC 驱动文件 checksum 不匹配: %s", name)
		}
	}
	return nil
}

func writeDriverPackageMetadata(targetDir string, pkg *offlineDriverPackage) error {
	data, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return fmt.Errorf("生成 JDBC 驱动元数据失败: %w", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "driver.json"), data, 0o600); err != nil {
		return fmt.Errorf("写入 JDBC 驱动元数据失败: %w", err)
	}
	return nil
}

func mapZipFiles(files []*zip.File) map[string]*zip.File {
	mapped := make(map[string]*zip.File, len(files))
	for _, file := range files {
		mapped[filepath.ToSlash(file.Name)] = file
	}
	return mapped
}

func readZipFile(files []*zip.File, name string) ([]byte, error) {
	for _, file := range files {
		if filepath.ToSlash(file.Name) != name {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("读取 JDBC 驱动包文件失败: %w", err)
		}
		defer rc.Close()
		data, err := io.ReadAll(rc)
		if err != nil {
			return nil, fmt.Errorf("读取 JDBC 驱动包文件失败: %w", err)
		}
		return data, nil
	}
	return nil, fmt.Errorf("JDBC 驱动包缺少文件: %s", name)
}

func zipFileSHA256(file *zip.File) (string, error) {
	rc, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("读取 JDBC 驱动 jar 失败: %w", err)
	}
	defer rc.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, rc); err != nil {
		return "", fmt.Errorf("计算 JDBC 驱动 checksum 失败: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func extractZipFile(file *zip.File, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return fmt.Errorf("创建 JDBC 驱动目录失败: %w", err)
	}
	rc, err := file.Open()
	if err != nil {
		return fmt.Errorf("读取 JDBC 驱动包文件失败: %w", err)
	}
	defer rc.Close()

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("写入 JDBC 驱动文件失败: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, rc); err != nil {
		return fmt.Errorf("写入 JDBC 驱动文件失败: %w", err)
	}
	return nil
}
