package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"unicode"
)

type RuntimeKind string

const (
	RuntimeKindManaged RuntimeKind = "managed"
	RuntimeKindSystem  RuntimeKind = "system"
	RuntimeKindMissing RuntimeKind = "missing"
)

type RuntimeSelection struct {
	Kind     RuntimeKind
	JavaPath string
	Version  string
}

type RuntimeService struct {
	paths          JDBCPaths
	systemJavaPath string
	useSystemJava  bool
	provider       ManagedRuntimeProvider
	fetcher        ArtifactFetcher
}

func NewRuntimeService(paths JDBCPaths, systemJavaPath string) *RuntimeService {
	return &RuntimeService{paths: paths, systemJavaPath: systemJavaPath}
}

func (s *RuntimeService) UseSystemJava(enabled bool) {
	s.useSystemJava = enabled
}

func (s *RuntimeService) ConfigureManagedInstaller(provider ManagedRuntimeProvider, fetcher ArtifactFetcher) {
	s.provider = provider
	s.fetcher = fetcher
}

func (s *RuntimeService) InstallManagedRuntime(ctx context.Context) (*RuntimeSelection, error) {
	if s.provider == nil || s.fetcher == nil {
		return nil, fmt.Errorf("托管 JDBC JRE 安装器未配置")
	}
	pkg, err := s.provider.Latest(ctx, 21)
	if err != nil {
		return nil, err
	}
	archiveName := filepath.Base(pkg.Name)
	if archiveName == "." || archiveName == "" || archiveName != pkg.Name {
		return nil, fmt.Errorf("托管 JDBC JRE 包名无效: %s", pkg.Name)
	}
	if err := os.MkdirAll(s.paths.RuntimesDir, 0o700); err != nil {
		return nil, fmt.Errorf("创建 JDBC 运行时目录失败: %w", err)
	}
	downloadDir, err := os.MkdirTemp(s.paths.RuntimesDir, ".runtime-download-*")
	if err != nil {
		return nil, fmt.Errorf("创建 JDBC JRE 下载目录失败: %w", err)
	}
	defer os.RemoveAll(downloadDir)
	archivePath := filepath.Join(downloadDir, archiveName)
	if err := s.fetcher.Download(ctx, pkg.URL, pkg.SHA256, archivePath); err != nil {
		return nil, fmt.Errorf("下载托管 JDBC JRE 失败: %w", err)
	}
	return s.ImportRuntimeArchive(archivePath)
}

func (s *RuntimeService) SelectRuntime() (*RuntimeSelection, error) {
	if s.useSystemJava && fileExists(s.systemJavaPath) {
		return &RuntimeSelection{Kind: RuntimeKindSystem, JavaPath: s.systemJavaPath}, nil
	}

	managed, err := s.latestManagedRuntime()
	if err != nil {
		return nil, err
	}
	if managed != "" {
		return &RuntimeSelection{Kind: RuntimeKindManaged, JavaPath: managed, Version: filepath.Base(filepath.Dir(filepath.Dir(managed)))}, nil
	}
	if fileExists(s.systemJavaPath) {
		return &RuntimeSelection{Kind: RuntimeKindSystem, JavaPath: s.systemJavaPath}, nil
	}
	return &RuntimeSelection{Kind: RuntimeKindMissing}, nil
}

func (s *RuntimeService) ImportRuntimeArchive(archivePath string) (*RuntimeSelection, error) {
	if err := os.MkdirAll(s.paths.RuntimesDir, 0o700); err != nil {
		return nil, fmt.Errorf("创建 JDBC 运行时目录失败: %w", err)
	}
	temporaryRoot, err := os.MkdirTemp(s.paths.RuntimesDir, ".runtime-import-*")
	if err != nil {
		return nil, fmt.Errorf("创建 JDBC 运行时临时目录失败: %w", err)
	}
	defer os.RemoveAll(temporaryRoot)
	unpacked := filepath.Join(temporaryRoot, "unpacked")
	if err := ExtractArchive(archivePath, unpacked); err != nil {
		return nil, fmt.Errorf("解压 JDBC JRE 归档失败: %w", err)
	}

	javaPath, archiveRoot, err := findRuntimeJava(unpacked)
	if err != nil {
		return nil, err
	}
	version := normalizeRuntimeVersion(archiveRoot, archivePath)
	targetName := fmt.Sprintf("jre-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
	targetDir := filepath.Join(s.paths.RuntimesDir, targetName)
	targetJava := filepath.Join(targetDir, "bin", "java")
	if fileExists(targetJava) {
		return &RuntimeSelection{Kind: RuntimeKindManaged, JavaPath: targetJava, Version: version}, nil
	}
	if _, err := os.Stat(targetDir); err == nil {
		return nil, fmt.Errorf("JDBC 运行时目标目录已存在但无效: %s", targetDir)
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("检查 JDBC 运行时目标目录失败: %w", err)
	}

	info, err := os.Stat(javaPath)
	if err != nil {
		return nil, fmt.Errorf("读取 Java 可执行文件失败: %w", err)
	}
	if err := os.Chmod(javaPath, info.Mode().Perm()|0o100); err != nil {
		return nil, fmt.Errorf("设置 Java 可执行权限失败: %w", err)
	}
	runtimeRoot := filepath.Dir(filepath.Dir(javaPath))
	if err := os.Rename(runtimeRoot, targetDir); err != nil {
		return nil, fmt.Errorf("提交 JDBC 运行时失败: %w", err)
	}
	return &RuntimeSelection{Kind: RuntimeKindManaged, JavaPath: targetJava, Version: version}, nil
}

func findRuntimeJava(unpacked string) (javaPath, archiveRoot string, err error) {
	var matches []string
	err = filepath.WalkDir(unpacked, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "java" || filepath.Base(filepath.Dir(path)) != "bin" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		return "", "", fmt.Errorf("扫描 JDBC JRE 归档失败: %w", err)
	}
	if len(matches) != 1 {
		return "", "", fmt.Errorf("JDBC JRE 归档必须包含唯一 bin/java，实际找到 %d 个", len(matches))
	}
	relative, err := filepath.Rel(unpacked, matches[0])
	if err != nil {
		return "", "", fmt.Errorf("解析 JDBC JRE 归档根目录失败: %w", err)
	}
	parts := strings.Split(filepath.ToSlash(relative), "/")
	root := ""
	if len(parts) > 2 {
		root = parts[0]
	}
	return matches[0], root, nil
}

func normalizeRuntimeVersion(archiveRoot, archivePath string) string {
	version := strings.TrimSpace(archiveRoot)
	if version == "" {
		version = filepath.Base(archivePath)
		for _, suffix := range []string{".tar.gz", ".tgz", ".zip"} {
			version = strings.TrimSuffix(strings.ToLower(version), suffix)
		}
	}
	lower := strings.ToLower(version)
	for _, prefix := range []string{"jdk-", "jre-"} {
		if strings.HasPrefix(lower, prefix) {
			version = version[len(prefix):]
			break
		}
	}
	var normalized strings.Builder
	previousDash := false
	for _, char := range version {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || char == '.' || char == '_' {
			normalized.WriteRune(char)
			previousDash = false
		} else if !previousDash {
			normalized.WriteByte('-')
			previousDash = true
		}
	}
	result := strings.Trim(normalized.String(), "-")
	if result == "" {
		return "unknown"
	}
	return result
}

func (s *RuntimeService) latestManagedRuntime() (string, error) {
	entries, err := os.ReadDir(s.paths.RuntimesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("读取 JDBC 运行时目录失败: %w", err)
	}

	var candidates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		javaPath := filepath.Join(s.paths.RuntimesDir, entry.Name(), "bin", "java")
		if fileExists(javaPath) {
			candidates = append(candidates, javaPath)
		}
	}
	if len(candidates) == 0 {
		return "", nil
	}
	sort.Strings(candidates)
	return candidates[len(candidates)-1], nil
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
