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
