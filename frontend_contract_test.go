package main

import (
	"os"
	"strings"
	"testing"
)

func TestJDBCDriverManagerIncludesPollingAndLogViewer(t *testing.T) {
	data, err := os.ReadFile("frontend/src/components/JDBCDriverManager.svelte")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	required := []string{
		"onDestroy",
		"statusPollInterval = 2000",
		"setInterval",
		"clearInterval",
		"GetJDBCAgentLogTail",
		"refreshAgentLog",
		"copyAgentLog",
		"closeAgentLog",
		"jdbc-manager__log-dialog",
	}
	for _, value := range required {
		if !strings.Contains(source, value) {
			t.Errorf("JDBC manager missing %q", value)
		}
	}

	pollStart := strings.Index(source, "async function pollJDBCStatus")
	if pollStart < 0 {
		t.Fatal("JDBC manager missing pollJDBCStatus")
	}
	pollEnd := strings.Index(source[pollStart:], "\n  }")
	if pollEnd < 0 {
		t.Fatal("cannot locate pollJDBCStatus end")
	}
	pollBody := source[pollStart : pollStart+pollEnd]
	if strings.Contains(pollBody, "ListJDBCDrivers") {
		t.Fatal("status polling must not refresh the driver catalog")
	}
}

func TestJDBCDriverManagerAlwaysExposesRuntimeSwitching(t *testing.T) {
	data, err := os.ReadFile("frontend/src/components/JDBCDriverManager.svelte")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	headerEnd := strings.Index(source, "{#if errorMessage}")
	if headerEnd < 0 {
		t.Fatal("JDBC manager missing error section")
	}
	header := source[:headerEnd]
	for _, handler := range []string{"useManagedRuntime", "importRuntimeArchive", "chooseJavaRuntime"} {
		if !strings.Contains(header, "on:click={"+handler+"}") {
			t.Errorf("normal runtime status missing %s action", handler)
		}
	}
}

func TestJDBCDriverManagerAlwaysExposesAgentLog(t *testing.T) {
	data, err := os.ReadFile("frontend/src/components/JDBCDriverManager.svelte")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	headerEnd := strings.Index(source, "{#if errorMessage}")
	if headerEnd < 0 {
		t.Fatal("JDBC manager missing error section")
	}
	if !strings.Contains(source[:headerEnd], "on:click={openAgentLog}") {
		t.Fatal("normal agent status missing log action")
	}
}

func TestAddAssetDialogPersistsSelectedJDBCProfile(t *testing.T) {
	data, err := os.ReadFile("frontend/src/components/AddAssetDialog.svelte")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	for _, required := range []string{
		"driverProfileID: ''",
		"selectedJDBCDriver?.profiles",
		"driver_profile_id: formData.driverProfileID || undefined",
		"metadata?.driver_profile_id",
	} {
		if !strings.Contains(source, required) {
			t.Errorf("database connection form missing %q", required)
		}
	}
}
