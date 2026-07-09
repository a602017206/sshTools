package service

import "path/filepath"

type JDBCPaths struct {
	RootDir     string
	DriversDir  string
	RuntimesDir string
	AgentDir    string
	LogsDir     string
	Manifest    string
}

func NewJDBCPaths(root string) JDBCPaths {
	return JDBCPaths{
		RootDir:     root,
		DriversDir:  filepath.Join(root, "drivers"),
		RuntimesDir: filepath.Join(root, "runtimes"),
		AgentDir:    filepath.Join(root, "agent"),
		LogsDir:     filepath.Join(root, "logs"),
		Manifest:    filepath.Join(root, "drivers", "manifest.json"),
	}
}
