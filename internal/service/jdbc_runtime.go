package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
}

func NewRuntimeService(paths JDBCPaths, systemJavaPath string) *RuntimeService {
	return &RuntimeService{paths: paths, systemJavaPath: systemJavaPath}
}

func (s *RuntimeService) UseSystemJava(enabled bool) {
	s.useSystemJava = enabled
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
	return nil, fmt.Errorf("导入 JDBC JRE 运行时暂未实现: %s", archivePath)
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
