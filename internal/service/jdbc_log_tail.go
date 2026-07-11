package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	defaultJDBCLogTailBytes int64 = 64 << 10
	minimumJDBCLogTailBytes int64 = 1 << 10
	maximumJDBCLogTailBytes int64 = 1 << 20
)

type JDBCLogTailService struct {
	logPath string
}

func NewJDBCLogTailService(paths JDBCPaths) *JDBCLogTailService {
	return &JDBCLogTailService{logPath: NewJDBCLogPaths(paths).Agent}
}

func (s *JDBCLogTailService) Read(maxBytes int64) (JDBCLogTail, error) {
	limit := clampJDBCLogTailBytes(maxBytes)
	info, err := os.Lstat(s.logPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return JDBCLogTail{}, nil
		}
		return JDBCLogTail{}, fmt.Errorf("读取 JDBC Agent 日志信息失败: %w", err)
	}
	if !info.Mode().IsRegular() {
		return JDBCLogTail{}, fmt.Errorf("JDBC Agent 日志不是普通文件")
	}

	file, err := os.Open(s.logPath)
	if err != nil {
		return JDBCLogTail{}, fmt.Errorf("打开 JDBC Agent 日志失败: %w", err)
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil {
		return JDBCLogTail{}, fmt.Errorf("确认 JDBC Agent 日志文件失败: %w", err)
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		return JDBCLogTail{}, fmt.Errorf("JDBC Agent 日志在读取前发生变化")
	}

	size := openedInfo.Size()
	readSize := size
	if readSize > limit {
		readSize = limit
	}
	if _, err := file.Seek(size-readSize, io.SeekStart); err != nil {
		return JDBCLogTail{}, fmt.Errorf("定位 JDBC Agent 日志尾部失败: %w", err)
	}
	content, err := io.ReadAll(io.LimitReader(file, readSize))
	if err != nil {
		return JDBCLogTail{}, fmt.Errorf("读取 JDBC Agent 日志尾部失败: %w", err)
	}
	return JDBCLogTail{
		Content:   strings.ToValidUTF8(string(content), "�"),
		Truncated: size > readSize,
		Size:      size,
	}, nil
}

func clampJDBCLogTailBytes(maxBytes int64) int64 {
	if maxBytes == 0 {
		return defaultJDBCLogTailBytes
	}
	if maxBytes < minimumJDBCLogTailBytes {
		return minimumJDBCLogTailBytes
	}
	if maxBytes > maximumJDBCLogTailBytes {
		return maximumJDBCLogTailBytes
	}
	return maxBytes
}
