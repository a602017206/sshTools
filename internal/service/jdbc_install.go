package service

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type DriverInstallService struct {
	paths JDBCPaths
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
	return &DriverInstallService{paths: paths}
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
