package service

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestDriverInstallImportsOfflinePackageAndValidatesChecksum(t *testing.T) {
	root := t.TempDir()
	jarBytes := []byte("fake-h2-driver")
	sum := sha256.Sum256(jarBytes)
	zipPath := filepath.Join(root, "driver-package.zip")
	createTestDriverPackage(t, zipPath, jarBytes, hex.EncodeToString(sum[:]))

	installer := NewDriverInstallService(NewJDBCPaths(filepath.Join(root, ".sshtools")))
	result, err := installer.ImportOfflinePackage(zipPath)
	if err != nil {
		t.Fatalf("import failed: %v", err)
	}
	if result.ProfileID != "h2-2.2.224" {
		t.Fatalf("unexpected profile id: %s", result.ProfileID)
	}
	if _, err := os.Stat(filepath.Join(result.InstallPath, "jars", "h2.jar")); err != nil {
		t.Fatalf("jar not installed: %v", err)
	}
}

func TestDriverInstallRollsBackOnChecksumMismatch(t *testing.T) {
	root := t.TempDir()
	jarBytes := []byte("fake-h2-driver")
	zipPath := filepath.Join(root, "driver-package.zip")
	createTestDriverPackage(t, zipPath, jarBytes, "bad-checksum")

	paths := NewJDBCPaths(filepath.Join(root, ".sshtools"))
	installer := NewDriverInstallService(paths)
	result, err := installer.ImportOfflinePackage(zipPath)
	if err == nil {
		t.Fatalf("expected checksum error, got result: %+v", result)
	}
	targetPath := filepath.Join(paths.DriversDir, "h2", "2.2.224")
	if _, statErr := os.Stat(targetPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected target directory rollback, stat error: %v", statErr)
	}
}

func createTestDriverPackage(t *testing.T, zipPath string, jarBytes []byte, sha string) {
	t.Helper()
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	zw := zip.NewWriter(file)
	defer zw.Close()

	files := map[string]string{
		"package.json":     `{"id":"h2","name":"H2","version":"2.2.224","driverClass":"org.h2.Driver","urlTemplate":"jdbc:h2:mem:{database}","defaultPort":0,"jre":">=17","jars":["h2.jar"]}`,
		"checksums.sha256": sha + "  jars/h2.jar\n",
	}
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(body))
	}
	w, err := zw.Create("jars/h2.jar")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write(jarBytes)
}
