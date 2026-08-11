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
