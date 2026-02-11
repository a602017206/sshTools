package main

import (
	"encoding/json"
	"testing"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/store"
)

func TestExportConnectionsByIDsWithPassphrase(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	configManager, err := config.NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}

	credentialStore := store.NewCredentialStore()
	connectionService := service.NewConnectionService(configManager, credentialStore)
	app := &App{connectionService: connectionService}

	connA := config.ConnectionConfig{
		ID:       "conn-a",
		Name:     "A",
		Host:     "10.0.0.1",
		Port:     22,
		User:     "root",
		AuthType: "password",
	}
	connB := config.ConnectionConfig{
		ID:       "conn-b",
		Name:     "B",
		Host:     "10.0.0.2",
		Port:     22,
		User:     "root",
		AuthType: "password",
	}

	if err := connectionService.AddConnection(connA); err != nil {
		t.Fatalf("AddConnection(connA) error = %v", err)
	}
	if err := connectionService.AddConnection(connB); err != nil {
		t.Fatalf("AddConnection(connB) error = %v", err)
	}
	if err := connectionService.SavePassword(connA.ID, "secret-a"); err != nil {
		t.Fatalf("SavePassword(connA) error = %v", err)
	}
	if err := connectionService.SavePassword(connB.ID, "secret-b"); err != nil {
		t.Fatalf("SavePassword(connB) error = %v", err)
	}

	jsonData, err := app.ExportConnectionsByIDsWithPassphrase([]string{connB.ID}, "passphrase")
	if err != nil {
		t.Fatalf("ExportConnectionsByIDsWithPassphrase() error = %v", err)
	}

	var exportData ExportData
	if err := json.Unmarshal([]byte(jsonData), &exportData); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if exportData.PasswordEncryption == nil || exportData.PasswordEncryption.Mode != "passphrase" {
		t.Fatalf("expected passphrase password_encryption metadata")
	}

	if exportData.PasswordEncryption.Salt == "" {
		t.Fatalf("expected passphrase salt to be set")
	}

	password, ok := exportData.Passwords[connB.ID]
	if !ok || password == "" {
		t.Fatalf("expected password for %s", connB.ID)
	}
	if len(password) < 5 || password[:5] != "penc:" {
		t.Fatalf("expected password to use penc: prefix")
	}
}
