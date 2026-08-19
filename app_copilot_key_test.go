package main

import (
	"testing"

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

// 编译期保证我们引用了 copilot 包的常量，避免未使用导入。
var _ = copilot.APIKeyCredentialID
