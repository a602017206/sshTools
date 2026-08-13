package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFallbackConfigManagerSavesInMemory(t *testing.T) {
	cm := NewFallbackConfigManager()

	conn := ConnectionConfig{
		ID:       "test-connection",
		Name:     "Test Connection",
		Host:     "127.0.0.1",
		Port:     22,
		User:     "tester",
		AuthType: "password",
		Type:     "ssh",
	}

	if err := cm.AddConnection(conn); err != nil {
		t.Fatalf("fallback AddConnection should save in memory without disk path: %v", err)
	}

	got, err := cm.GetConnection(conn.ID)
	if err != nil {
		t.Fatalf("expected fallback connection to be readable: %v", err)
	}
	if got.Name != conn.Name {
		t.Fatalf("expected connection name %q, got %q", conn.Name, got.Name)
	}

	if err := cm.UpdateSettings(map[string]interface{}{"theme": "light"}); err != nil {
		t.Fatalf("fallback UpdateSettings should save in memory without disk path: %v", err)
	}
	if gotTheme := cm.GetSettings().Theme; gotTheme != "light" {
		t.Fatalf("expected theme to update to light, got %q", gotTheme)
	}
}

func TestConfigManagerPersistsJDBCRuntimeSettings(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cm := newDiskTestConfigManager(configPath)
	if err := cm.UpdateJDBCRuntimeSettings("system", "/opt/jdk-21/bin/java"); err != nil {
		t.Fatal(err)
	}

	reloaded := newDiskTestConfigManager(configPath)
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	settings := reloaded.GetSettings()
	if settings.JDBCRuntimeMode != "system" || settings.JDBCSystemJavaPath != "/opt/jdk-21/bin/java" {
		t.Fatalf("unexpected JDBC runtime settings: %+v", settings)
	}
}

func TestConfigManagerLoadsLegacySettingsWithoutJDBCFields(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	legacy := AppConfig{Connections: []ConnectionConfig{}, Settings: DefaultSettings()}
	legacy.Settings.Theme = "light"
	data, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	cm := newDiskTestConfigManager(configPath)
	if err := cm.Load(); err != nil {
		t.Fatal(err)
	}
	settings := cm.GetSettings()
	if settings.JDBCRuntimeMode != "" || settings.JDBCSystemJavaPath != "" {
		t.Fatalf("legacy config should use empty JDBC settings: %+v", settings)
	}
	if settings.Theme != "light" {
		t.Fatalf("legacy setting changed: %q", settings.Theme)
	}
}

func TestConfigManagerRejectsInvalidJDBCRuntimeMode(t *testing.T) {
	cm := NewFallbackConfigManager()
	if err := cm.UpdateJDBCRuntimeSettings("other", ""); err == nil {
		t.Fatal("expected invalid mode error")
	}
	if err := cm.UpdateJDBCRuntimeSettings("system", ""); err == nil {
		t.Fatal("expected empty system Java path error")
	}
}

func TestUpdateSettingsPersistsCopilotFieldsWithoutAPIKey(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cm := newDiskTestConfigManager(configPath)
	if err := cm.UpdateSettings(map[string]interface{}{
		"copilot_provider": "openai_compatible",
		"copilot_base_url": "https://api.deepseek.com/v1",
		"copilot_model":    "deepseek-chat",
		"copilot_api_key":  "sk-secret",
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "sk-secret") || strings.Contains(string(raw), "copilot_api_key") {
		t.Fatalf("api key leaked into config.json: %s", raw)
	}
	reloaded := newDiskTestConfigManager(configPath)
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	s := reloaded.GetSettings()
	if s.CopilotBaseURL != "https://api.deepseek.com/v1" || s.CopilotModel != "deepseek-chat" {
		t.Fatalf("unexpected copilot settings: %+v", s)
	}
}

func newDiskTestConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		config: &AppConfig{
			Connections: []ConnectionConfig{},
			Settings:    DefaultSettings(),
		},
	}
}
