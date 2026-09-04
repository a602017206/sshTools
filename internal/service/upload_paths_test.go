package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandLocalUploadPathsKeepsFilesAndWalksFolders(t *testing.T) {
	root := t.TempDir()
	filePath := filepath.Join(root, "readme.txt")
	if err := os.WriteFile(filePath, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	folder := filepath.Join(root, "project")
	nested := filepath.Join(folder, "src")
	empty := filepath.Join(folder, "empty")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatal(err)
	}
	nestedFile := filepath.Join(nested, "main.go")
	if err := os.WriteFile(nestedFile, []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := ExpandLocalUploadPaths([]string{filePath, folder})
	if err != nil {
		t.Fatalf("expand failed: %v", err)
	}

	got := map[string]bool{}
	var files, dirs int
	for _, item := range items {
		got[item.RelPath] = item.IsDir
		if item.IsDir {
			dirs++
		} else {
			files++
		}
	}

	if isDir, ok := got["readme.txt"]; !ok || isDir {
		t.Fatalf("expected file readme.txt, got %v ok=%v", isDir, ok)
	}
	if isDir, ok := got["project"]; !ok || !isDir {
		t.Fatalf("expected directory project, got %v ok=%v", isDir, ok)
	}
	if isDir, ok := got["project/src"]; !ok || !isDir {
		t.Fatalf("expected directory project/src, got %v ok=%v", isDir, ok)
	}
	if isDir, ok := got["project/empty"]; !ok || !isDir {
		t.Fatalf("expected empty directory project/empty, got %v ok=%v", isDir, ok)
	}
	if isDir, ok := got["project/src/main.go"]; !ok || isDir {
		t.Fatalf("expected file project/src/main.go, got %v ok=%v", isDir, ok)
	}
	if files != 2 || dirs != 3 {
		t.Fatalf("files=%d dirs=%d items=%+v", files, dirs, items)
	}
}

func TestExpandLocalUploadPathsSkipsSymlinks(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.txt")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link.txt")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlink not supported")
	}

	items, err := ExpandLocalUploadPaths([]string{link})
	if err != nil {
		t.Fatalf("expand failed: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("symlinks should be skipped, got %+v", items)
	}
}

func TestJoinRemoteUploadPathKeepsUnixNestedRel(t *testing.T) {
	if got := joinRemoteUploadPath("/opt/app", "project/src/main.go"); got != "/opt/app/project/src/main.go" {
		t.Fatalf("got %q", got)
	}
	if got := joinRemoteUploadPath("/", "project/empty"); got != "/project/empty" {
		t.Fatalf("got %q", got)
	}
	if got := joinRemoteUploadPath("/opt/", "readme.txt"); got != "/opt/readme.txt" {
		t.Fatalf("got %q", got)
	}
}
