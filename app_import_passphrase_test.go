package main

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/store"
)

func TestImportConnectionsWithPassphrase(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	configManager, err := config.NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager() error = %v", err)
	}

	credentialStore := store.NewCredentialStore()
	connectionService := service.NewConnectionService(configManager, credentialStore)
	app := &App{connectionService: connectionService}

	conn := config.ConnectionConfig{
		ID:       "conn-1",
		Name:     "Imported",
		Host:     "10.0.0.5",
		Port:     22,
		User:     "root",
		AuthType: "password",
	}

	passphrase := "test-passphrase"
	salt := make([]byte, 16)
	for i := range salt {
		salt[i] = byte(i + 1)
	}
	key := derivePassphraseKey(passphrase, salt)
	ciphertext, err := encryptWithKey("secret", key)
	if err != nil {
		t.Fatalf("encryptWithKey() error = %v", err)
	}

	exportData := ExportData{
		Version:     "1.0",
		ExportedAt:  "",
		Connections: []config.ConnectionConfig{conn},
		Passwords: map[string]string{
			conn.ID: passphrasePrefix + ciphertext,
		},
		PasswordEncryption: &PasswordEncryption{
			Mode: "passphrase",
			Salt: base64.StdEncoding.EncodeToString(salt),
			KDF:  passphraseKDF,
		},
	}

	jsonData, err := json.Marshal(exportData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	count, err := app.ImportConnectionsWithPassphrase(string(jsonData), passphrase)
	if err != nil {
		t.Fatalf("ImportConnectionsWithPassphrase() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("expected import count 1, got %d", count)
	}

	password, err := connectionService.GetPassword(conn.ID)
	if err != nil {
		t.Fatalf("GetPassword() error = %v", err)
	}
	if password != "secret" {
		t.Fatalf("expected password to match, got %q", password)
	}
}
