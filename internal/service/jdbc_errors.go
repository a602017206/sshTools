package service

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	JDBCErrorRuntimeMissing   = "RUNTIME_MISSING"
	JDBCErrorDriverMissing    = "DRIVER_MISSING"
	JDBCErrorDriverInvalid    = "DRIVER_INVALID"
	JDBCErrorAgentUnavailable = "AGENT_UNAVAILABLE"
	JDBCErrorDBConnectFailed  = "DB_CONNECT_FAILED"
	JDBCErrorConnectionLost   = "CONNECTION_LOST"
	JDBCErrorQueryTimeout     = "QUERY_TIMEOUT"
	JDBCErrorQueryFailed      = "QUERY_FAILED"
)

type JDBCError struct {
	Code    string
	Message string
	Err     error
}

func (e *JDBCError) Error() string {
	message := e.Message
	if message == "" && e.Err != nil {
		message = e.Err.Error()
	}
	if message == "" {
		return e.Code
	}
	return fmt.Sprintf("[%s] %s", e.Code, message)
}

func (e *JDBCError) Unwrap() error {
	return e.Err
}

type JDBCLogPaths struct {
	Agent          string
	DriverInstall  string
	RuntimeInstall string
}

func NewJDBCLogPaths(paths JDBCPaths) JDBCLogPaths {
	return JDBCLogPaths{
		Agent:          filepath.Join(paths.LogsDir, "jdbc-agent.log"),
		DriverInstall:  filepath.Join(paths.LogsDir, "driver-install.log"),
		RuntimeInstall: filepath.Join(paths.LogsDir, "runtime-install.log"),
	}
}

func MapJDBCAgentError(message string) error {
	return newJDBCError(message, nil)
}

func newJDBCError(message string, cause error) *JDBCError {
	normalized := strings.ToLower(message)
	code := JDBCErrorDBConnectFailed
	friendly := message

	switch {
	case containsAny(normalized,
		"runtime_missing", "runtime not found", "java runtime not found",
		"java executable not found", "no java runtime"):
		code = JDBCErrorRuntimeMissing
	case containsAny(normalized,
		"driver_invalid", "invalid driver", "checksum mismatch",
		"checksum 校验失败", "classnotfoundexception", "class not found"):
		code = JDBCErrorDriverInvalid
	case containsAny(normalized,
		"driver_missing", "driver not found", "no suitable driver",
		"driver profile resolver not configured", "驱动未安装"):
		code = JDBCErrorDriverMissing
	case containsAny(normalized,
		"agent_unavailable", "agent unavailable", "agent client not configured",
		"connection refused", "code = unavailable", "transport is closing"):
		code = JDBCErrorAgentUnavailable
	case containsAny(normalized, staleConnectionPatterns...):
		code = JDBCErrorConnectionLost
	case containsAny(normalized,
		"deadlineexceeded", "context deadline exceeded", "client.timeout exceeded",
		"i/o timeout", "query timeout"):
		code = JDBCErrorQueryTimeout
		friendly = "查询执行超时：数据库响应过慢，或同一连接上并发请求导致阻塞。请稍后重试；若持续超时，可先关闭其他表页再打开。"
	case containsAny(normalized,
		"ora-", "sql command not properly ended", "00933",
		"syntax error", "sqlstate"):
		code = JDBCErrorQueryFailed
	}

	return &JDBCError{Code: code, Message: friendly, Err: cause}
}

// staleConnectionPatterns 匹配数据库服务端已断开（如空闲超时）后驱动抛出的错误，
// 覆盖 Oracle、PostgreSQL/人大金仓/openGauss、MySQL、SQL Server、达梦等常见驱动。
var staleConnectionPatterns = []string{
	"ora-17008",
	"ora-17027",
	"ora-03113",
	"ora-03114",
	"ora-12571",
	"closed connection",
	"connection closed",
	"connection is closed",
	"connection has been closed",
	"connection has already been closed",
	"connection was closed",
	"已关闭连接",
	"连接已关闭",
	"连接已断开",
	"流已被关闭",
	"stream has already been closed",
	"socket closed",
	"broken pipe",
	"connection reset",
	"no more data to read from socket",
	"communications link failure",
	"an i/o error occurred while sending to the backend",
	"terminating connection",
	"网络通信异常",
	"session not found",
}

func isJDBCSessionStale(err error) bool {
	if err == nil {
		return false
	}
	return containsAny(strings.ToLower(err.Error()), staleConnectionPatterns...)
}

func containsAny(message string, values ...string) bool {
	for _, value := range values {
		if strings.Contains(message, value) {
			return true
		}
	}
	return false
}
