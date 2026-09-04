package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSessionLogAppendListSearchPurge(t *testing.T) {
	root := t.TempDir()
	svc := NewSessionLogService(root)
	if err := svc.Append("c1", "s1", []byte("hello password=secret\n"), true); err != nil {
		t.Fatal(err)
	}
	list, err := svc.List("c1")
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}
	if strings.Contains(mustRead(t, list[0].Path), "secret") {
		t.Fatal("expected redaction on disk")
	}
	hits, err := svc.Search("c1", "hello", 10)
	if err != nil || len(hits) == 0 {
		t.Fatalf("hits=%v err=%v", hits, err)
	}
	old := filepath.Join(root, "c1", "2000-01-01T00-00-00_old.log")
	_ = os.MkdirAll(filepath.Dir(old), 0o755)
	_ = os.WriteFile(old, []byte("old\n"), 0o644)
	n, err := svc.PurgeExpired(30)
	if err != nil || n < 1 {
		t.Fatalf("purge n=%d err=%v", n, err)
	}
}

func TestSessionLogAppendSameSession(t *testing.T) {
	root := t.TempDir()
	svc := NewSessionLogService(root)
	if err := svc.Append("c1", "s1", []byte("line1\n"), false); err != nil {
		t.Fatal(err)
	}
	if err := svc.Append("c1", "s1", []byte("line2\n"), false); err != nil {
		t.Fatal(err)
	}
	list, err := svc.List("c1")
	if err != nil || len(list) != 1 {
		t.Fatalf("expected one log file, list=%v err=%v", list, err)
	}
	content := mustRead(t, list[0].Path)
	if !strings.Contains(content, "line1") || !strings.Contains(content, "line2") {
		t.Fatalf("expected both lines in same file, got %q", content)
	}
}

func TestSessionLogExportDelete(t *testing.T) {
	root := t.TempDir()
	svc := NewSessionLogService(root)
	if err := svc.Append("c1", "s1", []byte("export me\n"), false); err != nil {
		t.Fatal(err)
	}
	list, err := svc.List("c1")
	if err != nil || len(list) != 1 {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "exported.log")
	if err := svc.Export(list[0].ID, dest); err != nil {
		t.Fatal(err)
	}
	if mustRead(t, dest) != "export me\n" {
		t.Fatal("export content mismatch")
	}
	if err := svc.Delete(list[0].ID); err != nil {
		t.Fatal(err)
	}
	list, err = svc.List("c1")
	if err != nil || len(list) != 0 {
		t.Fatalf("expected empty list after delete, got %v err=%v", list, err)
	}
}

func TestSessionLogSearchLimit(t *testing.T) {
	root := t.TempDir()
	svc := NewSessionLogService(root)
	if err := svc.Append("c1", "s1", []byte("match\nmatch\nmatch\n"), false); err != nil {
		t.Fatal(err)
	}
	hits, err := svc.Search("c1", "match", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(hits))
	}
}

func TestSessionLogListEmptyConnection(t *testing.T) {
	root := t.TempDir()
	svc := NewSessionLogService(root)
	list, err := svc.List("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %v", list)
	}
}

func TestSessionLogCloseSession(t *testing.T) {
	root := t.TempDir()
	svc := NewSessionLogService(root)
	if err := svc.Append("c1", "s1", []byte("line1\n"), false); err != nil {
		t.Fatal(err)
	}
	key := svc.writerKey("c1", "s1")
	svc.mu.Lock()
	_, open := svc.writers[key]
	svc.mu.Unlock()
	if !open {
		t.Fatal("expected open writer before CloseSession")
	}

	svc.CloseSession("c1", "s1")
	svc.CloseSession("c1", "missing") // ignore missing

	svc.mu.Lock()
	_, stillOpen := svc.writers[key]
	svc.mu.Unlock()
	if stillOpen {
		t.Fatal("expected writer removed after CloseSession")
	}

	if err := svc.Append("c1", "s1", []byte("line2\n"), false); err != nil {
		t.Fatal(err)
	}
}

func TestSessionLogSearchEmptyQuery(t *testing.T) {
	root := t.TempDir()
	svc := NewSessionLogService(root)
	if err := svc.Append("c1", "s1", []byte("hello\n"), false); err != nil {
		t.Fatal(err)
	}
	hits, err := svc.Search("c1", "   ", 10)
	if err != nil {
		t.Fatal(err)
	}
	if hits != nil {
		t.Fatalf("expected nil hits for blank query, got %+v", hits)
	}
}

func TestSessionLogFilePermissions(t *testing.T) {
	root := t.TempDir()
	svc := NewSessionLogService(root)
	if err := svc.Append("c1", "s1", []byte("secret\n"), false); err != nil {
		t.Fatal(err)
	}
	dirInfo, err := os.Stat(filepath.Join(root, "c1"))
	if err != nil {
		t.Fatal(err)
	}
	if dirInfo.Mode().Perm() != 0o700 {
		t.Fatalf("expected dir mode 0700, got %o", dirInfo.Mode().Perm())
	}
	list, err := svc.List("c1")
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}
	fileInfo, err := os.Stat(list[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	if fileInfo.Mode().Perm() != 0o600 {
		t.Fatalf("expected file mode 0600, got %o", fileInfo.Mode().Perm())
	}
}
