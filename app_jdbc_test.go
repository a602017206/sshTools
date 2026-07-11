package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildJDBCServicesInjectsManagedGateway(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".sshtools")
	bundle, err := buildJDBCServices(root, []byte("embedded-agent-jar"), jdbcServiceDependencies{})
	if err != nil {
		t.Fatalf("build JDBC services failed: %v", err)
	}
	if bundle.gateway == nil || bundle.supervisor == nil {
		t.Fatalf("expected managed gateway and supervisor")
	}
	installedAgent := filepath.Join(root, "agent", "jdbc-agent.jar")
	content, err := os.ReadFile(installedAgent)
	if err != nil {
		t.Fatalf("agent artifact not installed: %v", err)
	}
	if string(content) != "embedded-agent-jar" {
		t.Fatalf("unexpected installed agent: %q", content)
	}
}
