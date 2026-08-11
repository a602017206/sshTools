package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMimeForBackgroundExt(t *testing.T) {
	if _, err := mimeForBackgroundExt(".png"); err != nil {
		t.Fatalf("png should be allowed: %v", err)
	}
	if _, err := mimeForBackgroundExt(".BMP"); err == nil {
		t.Fatal("bmp should be rejected")
	}
}

func TestNormalizeBackgroundFitAndOpacity(t *testing.T) {
	if got := normalizeBackgroundFit("CONTAIN"); got != "contain" {
		t.Fatalf("fit = %q", got)
	}
	if got := normalizeBackgroundOpacity(140); got != 100 {
		t.Fatalf("opacity = %d", got)
	}
	if got := normalizeBackgroundOpacity(-3); got != 0 {
		t.Fatalf("opacity = %d", got)
	}
}

func TestEncodeBackgroundDataURL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wallpaper.png")
	// Minimal PNG header bytes are enough for encode path; content is opaque to encoder.
	payload := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x01, 0x02}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatal(err)
	}
	dataURL, err := encodeBackgroundDataURL(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(dataURL, "data:image/png;base64,") {
		t.Fatalf("unexpected data url prefix: %s", dataURL[:min(40, len(dataURL))])
	}
}

func TestCopyBackgroundImageUsesUniqueName(t *testing.T) {
	srcDir := t.TempDir()
	src := filepath.Join(srcDir, "photo.jpg")
	if err := os.WriteFile(src, []byte("fake-jpeg-bytes"), 0o600); err != nil {
		t.Fatal(err)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)

	dst1, err := copyBackgroundImage(src)
	if err != nil {
		t.Fatal(err)
	}
	dst2, err := copyBackgroundImage(src)
	if err != nil {
		t.Fatal(err)
	}
	if dst1 == dst2 {
		t.Fatalf("expected unique destinations, both %q", dst1)
	}
	if !strings.Contains(dst1, filepath.Join(".ahasshtools", "backgrounds")) {
		t.Fatalf("unexpected destination: %s", dst1)
	}
	if _, err := os.Stat(dst1); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dst2); err != nil {
		t.Fatal(err)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
