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
