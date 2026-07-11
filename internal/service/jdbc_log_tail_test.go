package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJDBCLogTailReturnsLastBytes(t *testing.T) {
	paths := NewJDBCPaths(t.TempDir())
	logPath := NewJDBCLogPaths(paths).Agent
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		t.Fatal(err)
	}
	content := strings.Repeat("a", 2048) + "END"
	if err := os.WriteFile(logPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	tail, err := NewJDBCLogTailService(paths).Read(1024)
	if err != nil {
		t.Fatal(err)
	}
	if !tail.Truncated || !strings.HasSuffix(tail.Content, "END") || len(tail.Content) != 1024 {
		t.Fatalf("unexpected log tail: %+v", tail)
	}
	if tail.Size != int64(len(content)) {
		t.Fatalf("size=%d want=%d", tail.Size, len(content))
	}
}

func TestJDBCLogTailHandlesMissingAndInvalidFiles(t *testing.T) {
	paths := NewJDBCPaths(t.TempDir())
	service := NewJDBCLogTailService(paths)
	tail, err := service.Read(1024)
	if err != nil {
		t.Fatal(err)
	}
	if tail.Content != "" || tail.Size != 0 || tail.Truncated {
		t.Fatalf("missing log should be empty: %+v", tail)
	}

	logPath := NewJDBCLogPaths(paths).Agent
	if err := os.MkdirAll(logPath, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Read(1024); err == nil {
		t.Fatal("expected directory log rejection")
	}

	paths = NewJDBCPaths(t.TempDir())
	service = NewJDBCLogTailService(paths)
	logPath = NewJDBCLogPaths(paths).Agent
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "outside.log")
	if err := os.WriteFile(target, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, logPath); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Read(1024); err == nil {
		t.Fatal("expected symbolic link rejection")
	}
}

func TestJDBCLogTailClampsRequestedSize(t *testing.T) {
	paths := NewJDBCPaths(t.TempDir())
	logPath := NewJDBCLogPaths(paths).Agent
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath, []byte(strings.Repeat("x", 2<<20)), 0o600); err != nil {
		t.Fatal(err)
	}
	service := NewJDBCLogTailService(paths)

	defaultTail, err := service.Read(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(defaultTail.Content) != 64<<10 {
		t.Fatalf("default tail bytes=%d", len(defaultTail.Content))
	}
	minimumTail, err := service.Read(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(minimumTail.Content) != 1<<10 {
		t.Fatalf("minimum tail bytes=%d", len(minimumTail.Content))
	}
	maximumTail, err := service.Read(2 << 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(maximumTail.Content) != 1<<20 {
		t.Fatalf("maximum tail bytes=%d", len(maximumTail.Content))
	}
}

func TestJDBCLogTailReplacesInvalidUTF8(t *testing.T) {
	paths := NewJDBCPaths(t.TempDir())
	logPath := NewJDBCLogPaths(paths).Agent
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath, []byte{0xff, 'o', 'k'}, 0o600); err != nil {
		t.Fatal(err)
	}
	tail, err := NewJDBCLogTailService(paths).Read(1024)
	if err != nil {
		t.Fatal(err)
	}
	if tail.Content != "�ok" {
		t.Fatalf("invalid UTF-8 not replaced: %q", tail.Content)
	}
}
