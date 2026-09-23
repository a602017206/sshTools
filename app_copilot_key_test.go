package main

import (
	"testing"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service/copilot"
	"AHaSSHTools/internal/store"
)

// newAppWithTempCredentialStore builds a CredentialStore rooted at a temp HOME
// so tests never touch the user's real ~/.ahasshtools/credentials.enc.
func newAppWithTempCredentialStore(t *testing.T) *App {
	t.Helper()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	return &App{credentialStore: store.NewCredentialStore()}
}

func TestSetCopilotAPIKeyEmptyClearsAndReportsFalse(t *testing.T) {
	app := newAppWithTempCredentialStore(t)

	// 先存一个非空 key，确认 Has 为 true。
	if err := app.SetCopilotAPIKey("sk-real"); err != nil {
		t.Fatalf("set real key: %v", err)
	}
	if !app.HasCopilotAPIKey() {
		t.Fatal("HasCopilotAPIKey should be true after storing a real key")
	}

	// 传空串应清除，Has 变为 false。
	if err := app.SetCopilotAPIKey(""); err != nil {
		t.Fatalf("set empty key: %v", err)
	}
	if app.HasCopilotAPIKey() {
		t.Fatal("HasCopilotAPIKey must be false after SetCopilotAPIKey(\"\")")
	}

	// 纯空白串同样视为空。
	if err := app.SetCopilotAPIKey("   "); err != nil {
		t.Fatalf("set whitespace key: %v", err)
	}
	if app.HasCopilotAPIKey() {
		t.Fatal("HasCopilotAPIKey must be false after SetCopilotAPIKey(whitespace)")
	}
}

func TestSetCopilotAPIKeyNilStoreReturnsError(t *testing.T) {
	app := &App{}
	if err := app.SetCopilotAPIKey("sk-x"); err == nil {
		t.Fatal("expected error when credential store is nil")
	}
	if app.HasCopilotAPIKey() {
		t.Fatal("HasCopilotAPIKey must be false when credential store is nil")
	}
}

func TestClearCopilotAPIKeyRemovesKey(t *testing.T) {
	app := newAppWithTempCredentialStore(t)
	if err := app.SetCopilotAPIKey("sk-real"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := app.ClearCopilotAPIKey(); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if app.HasCopilotAPIKey() {
		t.Fatal("HasCopilotAPIKey must be false after ClearCopilotAPIKey")
	}
}

func TestCopilotProviderAPIKeyIsIsolatedFromLegacyKey(t *testing.T) {
	app := newAppWithTempCredentialStore(t)
	if err := app.SetCopilotAPIKey("sk-legacy"); err != nil {
		t.Fatalf("legacy set: %v", err)
	}
	if app.HasCopilotProviderAPIKey("p1") {
		t.Fatal("provider key should not exist before it is stored")
	}
	if err := app.SetCopilotProviderAPIKey("p1", "sk-provider"); err != nil {
		t.Fatalf("provider set: %v", err)
	}
	if !app.HasCopilotProviderAPIKey("p1") {
		t.Fatal("expected provider key")
	}
	if got := app.copilotAPIKey("p1"); got != "sk-provider" {
		t.Fatalf("expected provider key to win, got %q", got)
	}
	if err := app.ClearCopilotProviderAPIKey("p1"); err != nil {
		t.Fatalf("clear provider: %v", err)
	}
	if app.HasCopilotProviderAPIKey("p1") {
		t.Fatal("provider key should be cleared")
	}
	if got := app.copilotAPIKey("p1"); got != "sk-legacy" {
		t.Fatalf("expected legacy fallback after provider key cleared, got %q", got)
	}
}

func TestMigrateCopilotAPIKeyCopiesLegacyKeyOntoSingleProvider(t *testing.T) {
	app := newAppWithTempCredentialStore(t)
	if err := app.SetCopilotAPIKey("sk-legacy"); err != nil {
		t.Fatalf("legacy set: %v", err)
	}
	app.migrateCopilotAPIKey(config.AppSettings{
		CopilotProviders: []config.CopilotProvider{{ID: "p1"}},
	})
	if !app.HasCopilotProviderAPIKey("p1") {
		t.Fatal("expected legacy key to be copied onto the migrated provider")
	}
}

// 编译期保证我们引用了 copilot 包的常量，避免未使用导入。
var _ = copilot.APIKeyCredentialID
