package service

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AHaSSHTools/internal/config"
)

func TestDriverInstallDownloadsProfileJarsAtomically(t *testing.T) {
	firstJar := []byte("first-driver-jar")
	secondJar := []byte("second-driver-jar")
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/first.jar":
			_, _ = writer.Write(firstJar)
		case "/second.jar":
			_, _ = writer.Write(secondJar)
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	root := t.TempDir()
	paths := NewJDBCPaths(filepath.Join(root, ".sshtools"))
	installer := NewDriverInstallService(paths)
	installer.ConfigureArtifactFetcher(NewArtifactDownloader(ArtifactDownloadOptions{
		Client: server.Client(), AllowHTTP: true,
	}))
	driver, profile := onlineTestDriver(server.URL, firstJar, secondJar)

	result, err := installer.InstallProfile(context.Background(), driver, profile)
	if err != nil {
		t.Fatalf("install profile failed: %v", err)
	}
	for _, name := range []string{"first.jar", "second.jar"} {
		if _, err := os.Stat(filepath.Join(result.InstallPath, "jars", name)); err != nil {
			t.Fatalf("installed jar %s missing: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(result.InstallPath, "driver.json")); err != nil {
		t.Fatalf("driver metadata missing: %v", err)
	}
	markerPath := filepath.Join(result.InstallPath, "stale-marker")
	if err := os.WriteFile(markerPath, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}
	reinstalled, err := installer.InstallProfile(context.Background(), driver, profile)
	if err != nil {
		t.Fatalf("reinstall profile failed: %v", err)
	}
	if reinstalled.InstallPath != result.InstallPath {
		t.Fatalf("reinstall path changed: %q", reinstalled.InstallPath)
	}
	if _, err := os.Stat(markerPath); !os.IsNotExist(err) {
		t.Fatalf("reinstall retained stale version content: %v", err)
	}

	badDriver, badProfile := onlineTestDriver(server.URL, firstJar, secondJar)
	badDriver.ID = "broken"
	badProfile.ID = "broken-1.0"
	badProfile.Jars[1].SHA256 = strings.Repeat("0", 64)
	if _, err := installer.InstallProfile(context.Background(), badDriver, badProfile); err == nil {
		t.Fatal("expected checksum failure")
	}
	badTarget := filepath.Join(paths.DriversDir, badDriver.ID, badProfile.Version)
	if _, err := os.Stat(badTarget); !os.IsNotExist(err) {
		t.Fatalf("failed install left target directory: %v", err)
	}
}

func TestDriverInstallRejectsProfileWithoutOfficialURLs(t *testing.T) {
	installer := NewDriverInstallService(NewJDBCPaths(filepath.Join(t.TempDir(), ".sshtools")))
	driver := config.JDBCDriver{ID: "oracle", Name: "Oracle"}
	profile := config.JDBCDriverProfile{
		ID: "oracle-23", Version: "23", Jars: []config.JDBCJar{{Name: "ojdbc11.jar"}},
	}

	_, err := installer.InstallProfile(context.Background(), driver, profile)
	var jdbcErr *JDBCError
	if !errors.As(err, &jdbcErr) || jdbcErr.Code != JDBCErrorDriverMissing {
		t.Fatalf("expected DRIVER_MISSING, got %v", err)
	}
	if !strings.Contains(err.Error(), "离线导入") {
		t.Fatalf("expected offline import guidance, got %v", err)
	}
}

func onlineTestDriver(baseURL string, firstJar, secondJar []byte) (config.JDBCDriver, config.JDBCDriverProfile) {
	checksum := func(data []byte) string {
		sum := sha256.Sum256(data)
		return hex.EncodeToString(sum[:])
	}
	profile := config.JDBCDriverProfile{
		ID: "test-1.0", Version: "1.0", DriverClass: "example.Driver",
		URLTemplate: "jdbc:test://{host}:{port}/{database}", DefaultPort: 1234, JRERequirement: ">=17",
		Jars: []config.JDBCJar{
			{Name: "first.jar", URL: baseURL + "/first.jar", SHA256: checksum(firstJar)},
			{Name: "second.jar", URL: baseURL + "/second.jar", SHA256: checksum(secondJar)},
		},
	}
	return config.JDBCDriver{ID: "test", Name: "Test", RecommendedVersion: "1.0", Profiles: []config.JDBCDriverProfile{profile}}, profile
}

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
