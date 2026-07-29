package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"AHaSSHTools/internal/config"
)

func TestDriverCatalogReturnsDefaultPortsForInitialJDBCTypes(t *testing.T) {
	cases := map[string]int{
		"mysql":      3306,
		"postgresql": 5432,
		"sqlite":     0,
		"oracle":     1521,
		"sqlserver":  1433,
		"dm":         5236,
		"kingbase":   54321,
		"opengauss":  5432,
	}
	for dbType, want := range cases {
		if got := config.GetDefaultPort(dbType); got != want {
			t.Fatalf("%s default port: got %d want %d", dbType, got, want)
		}
	}
}

func TestDriverCatalogBootstrapsBuiltinManifest(t *testing.T) {
	root := t.TempDir()
	manifestPath := filepath.Join(root, "config", "jdbc-drivers.json")
	catalog := NewDriverCatalogService(manifestPath, filepath.Join(root, "drivers"))

	drivers, err := catalog.ListDriversWithInstallStatus()
	if err != nil {
		t.Fatalf("list builtin drivers failed: %v", err)
	}
	want := []string{"mysql", "postgresql", "sqlite", "oracle", "sqlserver", "dm", "kingbase", "opengauss"}
	if len(drivers) != len(want) {
		t.Fatalf("builtin driver count: got %d want %d", len(drivers), len(want))
	}
	for i, driverID := range want {
		if drivers[i].ID != driverID {
			t.Fatalf("builtin driver %d: got %q want %q", i, drivers[i].ID, driverID)
		}
	}
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("builtin manifest was not persisted: %v", err)
	}
}

func TestDriverCatalogLoadsManifestAndSelectsRecommendedProfile(t *testing.T) {
	root := t.TempDir()
	manifestPath := filepath.Join(root, "manifest.json")
	err := os.WriteFile(manifestPath, []byte(`{
	  "version": 1,
	  "drivers": [{
	    "id": "oracle",
	    "name": "Oracle",
	    "recommendedVersion": "23.5",
	    "profiles": [{
	      "id": "oracle-23.5",
	      "version": "23.5",
	      "driverClass": "oracle.jdbc.OracleDriver",
	      "urlTemplate": "jdbc:oracle:thin:@//{host}:{port}/{database}",
	      "defaultPort": 1521,
	      "jre": ">=17",
	      "jars": [{"name": "ojdbc11.jar", "sha256": "abc"}]
	    }]
	  }]
	}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}

	catalog := NewDriverCatalogService(manifestPath, "")
	driver, profile, err := catalog.GetRecommendedProfile("oracle")
	if err != nil {
		t.Fatalf("expected profile, got error: %v", err)
	}
	if driver.Name != "Oracle" {
		t.Fatalf("expected Oracle, got %q", driver.Name)
	}
	if profile.DriverClass != "oracle.jdbc.OracleDriver" {
		t.Fatalf("unexpected driver class: %s", profile.DriverClass)
	}
	if profile.DefaultPort != 1521 {
		t.Fatalf("unexpected default port: %d", profile.DefaultPort)
	}
}

func TestDriverCatalogProvidesVerifiedOnlineProfiles(t *testing.T) {
	root := t.TempDir()
	catalog := NewDriverCatalogService(filepath.Join(root, "manifest.json"), filepath.Join(root, "drivers"))
	if _, err := catalog.ListDriversWithInstallStatus(); err != nil {
		t.Fatalf("bootstrap manifest failed: %v", err)
	}

	for _, wanted := range []struct {
		driverID string
		version  string
		class    string
	}{
		{driverID: "oracle", version: "23.26.2.0.0", class: "oracle.jdbc.OracleDriver"},
		{driverID: "dm", version: "8.1.5.45", class: "dm.jdbc.driver.DmDriver"},
		{driverID: "kingbase", version: "8.6.1", class: "com.kingbase8.Driver"},
		{driverID: "kingbase", version: "9.0.1", class: "com.kingbase8.Driver"},
	} {
		_, profile, err := catalog.GetProfile(wanted.driverID, wanted.version)
		if err != nil {
			t.Fatalf("get profile %s %s failed: %v", wanted.driverID, wanted.version, err)
		}
		if profile.DriverClass != wanted.class {
			t.Fatalf("profile %s driver class = %q, want %q", profile.ID, profile.DriverClass, wanted.class)
		}
		for _, jar := range profile.Jars {
			if !strings.HasPrefix(jar.URL, "https://repo.maven.apache.org/") {
				t.Fatalf("profile %s jar URL = %q", profile.ID, jar.URL)
			}
			if len(jar.SHA256) != 64 {
				t.Fatalf("profile %s jar SHA-256 = %q", profile.ID, jar.SHA256)
			}
		}
	}
}

func TestDriverCatalogMigratesOutdatedBuiltinProfiles(t *testing.T) {
	root := t.TempDir()
	manifestPath := filepath.Join(root, "manifest.json")
	historicalManifest := `{
	  "version": 2,
	  "drivers": [
	    {"id":"oracle","name":"Oracle","recommendedVersion":"23","profiles":[{"id":"oracle-23","version":"23","driverClass":"oracle.jdbc.OracleDriver","urlTemplate":"jdbc:oracle:thin:@//{host}:{port}/{database}","defaultPort":1521,"jre":">=17","jars":[{"name":"ojdbc11.jar","sha256":""}]}]},
	    {"id":"dm","name":"达梦数据库","recommendedVersion":"8","profiles":[{"id":"dm-8","version":"8","driverClass":"dm.jdbc.driver.DmDriver","urlTemplate":"jdbc:dm://{host}:{port}/{database}","defaultPort":5236,"jre":">=17","jars":[{"name":"DmJdbcDriver18.jar","sha256":""}]}]},
	    {"id":"kingbase","name":"人大金仓","recommendedVersion":"8","profiles":[{"id":"kingbase-8","version":"8","driverClass":"com.kingbase8.Driver","urlTemplate":"jdbc:kingbase8://{host}:{port}/{database}","defaultPort":54321,"jre":">=17","jars":[{"name":"kingbase8.jar","sha256":""}]}]},
	    {"id":"private","name":"私有驱动","recommendedVersion":"1.0","profiles":[{"id":"private-1.0","version":"1.0","driverClass":"example.Driver","urlTemplate":"jdbc:private:{database}","jars":[{"name":"private.jar","sha256":"abc"}]}]}
	  ]
	}`
	if err := os.WriteFile(manifestPath, []byte(historicalManifest), 0o600); err != nil {
		t.Fatal(err)
	}

	catalog := NewDriverCatalogService(manifestPath, filepath.Join(root, "drivers"))
	if _, err := catalog.LoadManifest(); err != nil {
		t.Fatalf("load migrated manifest failed: %v", err)
	}
	_, dm, err := catalog.GetProfile("dm", "8.1.5.45")
	if err != nil {
		t.Fatal(err)
	}
	_, kingbaseV8, err := catalog.GetProfile("kingbase", "8.6.1")
	if err != nil {
		t.Fatal(err)
	}
	_, kingbaseV9, err := catalog.GetProfile("kingbase", "9.0.1")
	if err != nil {
		t.Fatal(err)
	}
	_, custom, err := catalog.GetProfile("private", "1.0")
	if err != nil {
		t.Fatal(err)
	}
	_, oracle, err := catalog.GetProfile("oracle", "23.26.2.0.0")
	if err != nil {
		t.Fatal(err)
	}
	if dm.Jars[0].URL == "" || kingbaseV8.Jars[0].URL == "" || kingbaseV9.Jars[0].URL == "" || oracle.Jars[0].URL == "" {
		t.Fatal("migration did not add online installation sources")
	}
	if custom.ID != "private-1.0" {
		t.Fatalf("custom profile = %q", custom.ID)
	}
}

func TestDriverManagerListsDriversWithInstallStatus(t *testing.T) {
	root := t.TempDir()
	manifestPath := filepath.Join(root, "manifest.json")
	err := os.WriteFile(manifestPath, []byte(`{
	  "version": 1,
	  "drivers": [
	    {
	      "id": "oracle",
	      "name": "Oracle",
	      "recommendedVersion": "23.5",
	      "profiles": [{
	        "id": "oracle-23.5",
	        "version": "23.5",
	        "driverClass": "oracle.jdbc.OracleDriver",
	        "urlTemplate": "jdbc:oracle:thin:@//{host}:{port}/{database}",
	        "defaultPort": 1521,
	        "jre": ">=17",
	        "jars": [{"name": "ojdbc11.jar", "sha256": "abc"}]
	      }]
	    },
	    {
	      "id": "h2",
	      "name": "H2",
	      "recommendedVersion": "2.2.224",
	      "profiles": [{
	        "id": "h2-2.2.224",
	        "version": "2.2.224",
	        "driverClass": "org.h2.Driver",
	        "urlTemplate": "jdbc:h2:mem:{database}",
	        "defaultPort": 0,
	        "jre": ">=17",
	        "jars": [{"name": "h2.jar", "sha256": "def"}]
	      }]
	    }
	  ]
	}`), 0o600)
	if err != nil {
		t.Fatal(err)
	}
	installedRoot := filepath.Join(root, "drivers")
	if err := os.MkdirAll(filepath.Join(installedRoot, "h2", "2.2.224", "jars"), 0o700); err != nil {
		t.Fatal(err)
	}

	catalog := NewDriverCatalogService(manifestPath, installedRoot)
	drivers, err := catalog.ListDriversWithInstallStatus()
	if err != nil {
		t.Fatalf("list drivers failed: %v", err)
	}
	status := map[string]bool{}
	for _, driver := range drivers {
		status[driver.ID] = driver.Installed
	}
	if status["oracle"] {
		t.Fatalf("oracle should not be installed")
	}
	if !status["h2"] {
		t.Fatalf("h2 should be installed")
	}
}
