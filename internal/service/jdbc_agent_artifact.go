package service

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
)

type AgentArtifactInstaller struct {
	paths JDBCPaths
}

func NewAgentArtifactInstaller(paths JDBCPaths) *AgentArtifactInstaller {
	return &AgentArtifactInstaller{paths: paths}
}

func (s *AgentArtifactInstaller) Install(jar []byte) (string, error) {
	if len(jar) == 0 {
		return "", fmt.Errorf("JDBC agent jar 不能为空")
	}
	if err := os.MkdirAll(s.paths.AgentDir, 0o700); err != nil {
		return "", fmt.Errorf("创建 JDBC agent 目录失败: %w", err)
	}

	target := filepath.Join(s.paths.AgentDir, "jdbc-agent.jar")
	if current, err := os.ReadFile(target); err == nil {
		currentSum := sha256.Sum256(current)
		incomingSum := sha256.Sum256(jar)
		if bytes.Equal(currentSum[:], incomingSum[:]) {
			return target, nil
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("读取现有 JDBC agent jar 失败: %w", err)
	}

	temporary, err := os.CreateTemp(s.paths.AgentDir, ".jdbc-agent-*.tmp")
	if err != nil {
		return "", fmt.Errorf("创建 JDBC agent 临时文件失败: %w", err)
	}
	temporaryPath := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(temporaryPath)
		}
	}()

	if _, err := temporary.Write(jar); err != nil {
		return "", fmt.Errorf("写入 JDBC agent 临时文件失败: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return "", fmt.Errorf("同步 JDBC agent 临时文件失败: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		return "", fmt.Errorf("设置 JDBC agent 文件权限失败: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return "", fmt.Errorf("关闭 JDBC agent 临时文件失败: %w", err)
	}
	if err := os.Rename(temporaryPath, target); err != nil {
		return "", fmt.Errorf("提交 JDBC agent jar 失败: %w", err)
	}
	committed = true
	return target, nil
}
