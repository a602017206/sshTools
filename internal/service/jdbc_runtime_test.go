package service

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRuntimeServicePrefersManagedRuntimeThenSystemRuntime(t *testing.T) {
	root := t.TempDir()
	paths := NewJDBCPaths(filepath.Join(root, ".sshtools"))
	managedJava := filepath.Join(paths.RuntimesDir, "jre-21-test", "bin", "java")
	if err := os.MkdirAll(filepath.Dir(managedJava), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(managedJava, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	runtimeSvc := NewRuntimeService(paths, "/usr/bin/java")
	selected, err := runtimeSvc.SelectRuntime()
	if err != nil {
		t.Fatalf("select runtime failed: %v", err)
	}
	if selected.Kind != RuntimeKindManaged {
		t.Fatalf("expected managed runtime, got %s", selected.Kind)
	}
}

func TestRuntimeServiceImportsJREArchive(t *testing.T) {
	root := t.TempDir()
	paths := NewJDBCPaths(filepath.Join(root, ".sshtools"))
	archivePath := filepath.Join(root, "jdk-test.zip")
	createRuntimeZip(t, archivePath, map[string]runtimeZipEntry{
		"jdk-test/bin/java": {body: "#!/bin/sh\n", mode: 0o755},
	})

	selected, err := NewRuntimeService(paths, "").ImportRuntimeArchive(archivePath)
	if err != nil {
		t.Fatalf("import runtime failed: %v", err)
	}
	expectedJava := filepath.Join(paths.RuntimesDir, "jre-test-"+runtime.GOOS+"-"+runtime.GOARCH, "bin", "java")
	if selected.Kind != RuntimeKindManaged || selected.JavaPath != expectedJava || selected.Version != "test" {
		t.Fatalf("unexpected runtime selection: %+v", selected)
	}
	info, err := os.Stat(expectedJava)
	if err != nil {
		t.Fatalf("installed java missing: %v", err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatalf("installed java is not executable: %o", info.Mode().Perm())
	}
}

func TestRuntimeServiceRollsBackInvalidArchive(t *testing.T) {
	root := t.TempDir()
	paths := NewJDBCPaths(filepath.Join(root, ".sshtools"))
	archivePath := filepath.Join(root, "invalid-jre.zip")
	createRuntimeZip(t, archivePath, map[string]runtimeZipEntry{
		"jdk-invalid/README.txt": {body: "missing java", mode: 0o644},
	})

	if _, err := NewRuntimeService(paths, "").ImportRuntimeArchive(archivePath); err == nil {
		t.Fatalf("expected invalid runtime archive error")
	}
	entries, err := os.ReadDir(paths.RuntimesDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("invalid archive left runtime files: %v", entries)
	}
}

type runtimeZipEntry struct {
	body string
	mode os.FileMode
}

func createRuntimeZip(t *testing.T, archivePath string, entries map[string]runtimeZipEntry) {
	t.Helper()
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	defer writer.Close()
	for name, entry := range entries {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.SetMode(entry.mode)
		part, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(entry.body)); err != nil {
			t.Fatal(err)
		}
	}
}
