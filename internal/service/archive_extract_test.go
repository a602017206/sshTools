package service

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestArchiveExtractorRejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "traversal.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	part, err := writer.Create("../escape")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("escape"))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(root, "extracted")
	if err := ExtractArchive(archivePath, target); err == nil {
		t.Fatalf("expected path traversal error")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("failed extraction left target directory: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "escape")); !os.IsNotExist(err) {
		t.Fatalf("archive escaped target directory: %v", err)
	}
}

func TestArchiveExtractorExtractsTarGz(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "runtime.tar.gz")
	writeRuntimeTarGz(t, archivePath, []tar.Header{
		{Name: "jdk-test/bin/java", Mode: 0o755, Size: 5, Typeflag: tar.TypeReg},
	}, [][]byte{[]byte("java\n")})
	target := filepath.Join(root, "extracted")

	if err := ExtractArchive(archivePath, target); err != nil {
		t.Fatalf("extract tar.gz failed: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(target, "jdk-test", "bin", "java"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "java\n" {
		t.Fatalf("unexpected extracted content: %q", content)
	}
}

func TestArchiveExtractorRejectsSymbolicLink(t *testing.T) {
	root := t.TempDir()
	archivePath := filepath.Join(root, "symlink.tar.gz")
	writeRuntimeTarGz(t, archivePath, []tar.Header{
		{Name: "jdk-test/bin/java", Linkname: "/usr/bin/java", Typeflag: tar.TypeSymlink},
	}, [][]byte{nil})
	target := filepath.Join(root, "extracted")

	if err := ExtractArchive(archivePath, target); err == nil {
		t.Fatalf("expected symbolic link rejection")
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("failed extraction left target directory: %v", err)
	}
}

func writeRuntimeTarGz(t *testing.T, archivePath string, headers []tar.Header, bodies [][]byte) {
	t.Helper()
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	for i := range headers {
		header := headers[i]
		if err := tarWriter.WriteHeader(&header); err != nil {
			t.Fatal(err)
		}
		if len(bodies[i]) > 0 {
			if _, err := tarWriter.Write(bodies[i]); err != nil {
				t.Fatal(err)
			}
		}
	}
}
