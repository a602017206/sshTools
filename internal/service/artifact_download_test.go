package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArtifactDownloaderValidatesChecksumAndCommitsAtomically(t *testing.T) {
	body := []byte("managed-runtime-archive")
	sum := sha256.Sum256(body)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/octet-stream")
		_, _ = writer.Write(body)
	}))
	defer server.Close()

	downloader := NewArtifactDownloader(ArtifactDownloadOptions{
		Client:    server.Client(),
		AllowHTTP: true,
		MaxBytes:  1024,
	})
	target := filepath.Join(t.TempDir(), "runtime.zip")
	if err := downloader.Download(context.Background(), server.URL, hex.EncodeToString(sum[:]), target); err != nil {
		t.Fatalf("download failed: %v", err)
	}
	downloaded, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(downloaded) != string(body) {
		t.Fatalf("unexpected downloaded body: %q", downloaded)
	}

	invalidTarget := filepath.Join(t.TempDir(), "invalid.zip")
	if err := downloader.Download(context.Background(), server.URL, strings.Repeat("0", 64), invalidTarget); err == nil {
		t.Fatalf("expected checksum mismatch")
	}
	if _, err := os.Stat(invalidTarget); !os.IsNotExist(err) {
		t.Fatalf("checksum failure committed target: %v", err)
	}
}
