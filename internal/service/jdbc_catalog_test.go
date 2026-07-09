package service

import (
	"os"
	"path/filepath"
	"testing"
)

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
