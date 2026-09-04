package service

import (
	"errors"
	"testing"
)

func TestJDBCErrorMapsRuntimeMissingToActionableCode(t *testing.T) {
	err := MapJDBCAgentError("runtime not found")
	var jdbcErr *JDBCError
	if !errors.As(err, &jdbcErr) {
		t.Fatalf("expected JDBCError")
	}
	if jdbcErr.Code != "RUNTIME_MISSING" {
		t.Fatalf("unexpected code: %s", jdbcErr.Code)
	}
}

func TestJDBCErrorMapsDeadlineExceededToQueryTimeout(t *testing.T) {
	err := MapJDBCAgentError("rpc error: code = DeadlineExceeded desc = context deadline exceeded")
	var jdbcErr *JDBCError
	if !errors.As(err, &jdbcErr) {
		t.Fatalf("expected JDBCError")
	}
	if jdbcErr.Code != JDBCErrorQueryTimeout {
		t.Fatalf("unexpected code: %s", jdbcErr.Code)
	}
	if jdbcErr.Message == "" || !containsAny(jdbcErr.Message, "超时") {
		t.Fatalf("expected Chinese timeout message, got %q", jdbcErr.Message)
	}
}

func TestJDBCErrorMapsOracleSyntaxToQueryFailed(t *testing.T) {
	err := MapJDBCAgentError("rpc error: code = Unknown desc = ORA-00933: SQL 命令未正确结束")
	var jdbcErr *JDBCError
	if !errors.As(err, &jdbcErr) {
		t.Fatalf("expected JDBCError")
	}
	if jdbcErr.Code != JDBCErrorQueryFailed {
		t.Fatalf("unexpected code: %s", jdbcErr.Code)
	}
}

func TestJDBCClosedOracleConnectionIsStaleSession(t *testing.T) {
	err := MapJDBCAgentError("rpc error: code = Unknown desc = ORA-17008: 已关闭连接")
	if !isJDBCSessionStale(err) {
		t.Fatalf("ORA-17008 should be treated as a stale JDBC session, got %v", err)
	}
	streamErr := MapJDBCAgentError("rpc error: code = Unknown desc = ORA-17027: 流已被关闭")
	if !isJDBCSessionStale(streamErr) {
		t.Fatalf("ORA-17027 should be treated as a stale JDBC session, got %v", streamErr)
	}
	syntaxErr := MapJDBCAgentError("rpc error: code = Unknown desc = ORA-00933: SQL 命令未正确结束")
	if isJDBCSessionStale(syntaxErr) {
		t.Fatal("SQL syntax errors must not trigger session reconnect")
	}
}
