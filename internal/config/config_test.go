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
	if len(s.CopilotProviders) != 1 {
		t.Fatalf("expected legacy copilot settings to migrate into one provider, got %d", len(s.CopilotProviders))
	}
	if s.CopilotProviders[0].BaseURL != "https://api.deepseek.com/v1" {
		t.Fatalf("migrated base url = %q", s.CopilotProviders[0].BaseURL)
	}
	if len(s.CopilotProviders[0].Models) != 1 || s.CopilotProviders[0].Models[0].ModelID != "deepseek-chat" {
		t.Fatalf("unexpected migrated models: %+v", s.CopilotProviders[0].Models)
	}
	if s.CopilotActiveModelID == "" {
		t.Fatal("expected migrated active model id")
	}
}

func TestUpdateSettingsPersistsCopilotProvidersWithoutAPIKey(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cm := newDiskTestConfigManager(configPath)
	if err := cm.UpdateSettings(map[string]interface{}{
		"copilot_active_model_id": "m1",
		"copilot_providers": []interface{}{
			map[string]interface{}{
				"id":       "p1",
				"kind":     "deepseek",
				"name":     "DeepSeek",
				"base_url": "https://api.deepseek.com/v1",
				"api_key":  "sk-should-not-persist",
				"models": []interface{}{
					map[string]interface{}{"id": "m1", "name": "Chat", "model_id": "deepseek-chat"},
				},
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "sk-should-not-persist") || strings.Contains(string(raw), "api_key") {
		t.Fatalf("provider api key leaked into config.json: %s", raw)
	}
	reloaded := newDiskTestConfigManager(configPath)
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	s := reloaded.GetSettings()
	if len(s.CopilotProviders) != 1 || s.CopilotProviders[0].ID != "p1" {
		t.Fatalf("unexpected providers: %+v", s.CopilotProviders)
	}
	if s.CopilotActiveModelID != "m1" || s.CopilotModel != "deepseek-chat" {
		t.Fatalf("expected mirrored active model, got %+v", s)
	}
}

func TestUpdateSettingsClearsProvidersWithoutRevivingLegacyFields(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cm := newDiskTestConfigManager(configPath)
	if err := cm.UpdateSettings(map[string]interface{}{
		"copilot_providers": []interface{}{
			map[string]interface{}{
				"id":       "p1",
				"kind":     "custom",
				"name":     "默认",
				"base_url": "https://api.deepseek.com/v1",
				"models": []interface{}{
					map[string]interface{}{"id": "m1", "name": "Chat", "model_id": "deepseek-chat"},
				},
			},
		},
		"copilot_active_model_id": "m1",
	}); err != nil {
		t.Fatal(err)
	}
	if err := cm.UpdateSettings(map[string]interface{}{
		"copilot_providers": []interface{}{},
	}); err != nil {
		t.Fatal(err)
	}
	s := cm.GetSettings()
	if len(s.CopilotProviders) != 0 || s.CopilotBaseURL != "" || s.CopilotModel != "" {
		t.Fatalf("expected cleared copilot settings, got %+v", s)
	}
	reloaded := newDiskTestConfigManager(configPath)
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	s = reloaded.GetSettings()
	if len(s.CopilotProviders) != 0 {
		t.Fatalf("cleared providers should not remigrate on load: %+v", s.CopilotProviders)
	}
}

func TestMigrateLegacyCopilotSettingsIsIdempotent(t *testing.T) {
	settings := AppSettings{
		CopilotBaseURL: "https://api.deepseek.com/v1",
		CopilotModel:   "deepseek-chat",
	}
	first, changed := MigrateLegacyCopilotSettings(settings)
	if !changed || len(first.CopilotProviders) != 1 {
		t.Fatalf("expected first migration, got changed=%v providers=%+v", changed, first.CopilotProviders)
	}
	second, changed := MigrateLegacyCopilotSettings(first)
	if changed {
		t.Fatal("second migration should be a no-op")
	}
	if first.CopilotProviders[0].ID != second.CopilotProviders[0].ID {
		t.Fatal("migration must keep stable ids")
	}
}

func TestDefaultSettingsIncludeSessionLogAndCommandSuggest(t *testing.T) {
	settings := DefaultSettings()
	if !settings.SessionLogEnabled {
		t.Fatal("expected SessionLogEnabled default true")
	}
	if settings.SessionLogRetentionDays != 30 {
		t.Fatalf("expected SessionLogRetentionDays 30, got %d", settings.SessionLogRetentionDays)
	}
	if !settings.SessionLogRedactEnabled {
		t.Fatal("expected SessionLogRedactEnabled default true")
	}
	if !settings.CommandSuggestEnabled {
		t.Fatal("expected CommandSuggestEnabled default true")
	}
	if settings.CommandSuggestLimit != 8 {
		t.Fatalf("expected CommandSuggestLimit 8, got %d", settings.CommandSuggestLimit)
	}
}

func TestDefaultFileManagerDirectoryTrackingEnabled(t *testing.T) {
	settings := DefaultFileManagerSettings()
	if !settings.DirectoryTracking {
		t.Fatal("expected DirectoryTracking default true")
	}
	if settings.Favorites == nil {
		t.Fatal("expected empty favorites list")
	}
}

func TestUpdateSettingsPersistsFileManagerFavorites(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cm := newDiskTestConfigManager(configPath)
	if err := cm.UpdateSettings(map[string]interface{}{
		"connection_id": "conn-1",
		"file_manager_settings": map[string]interface{}{
			"history":   []interface{}{"/tmp"},
			"favorites": []interface{}{"/opt/nginx"},
		},
	}); err != nil {
		t.Fatal(err)
	}

	reloaded := newDiskTestConfigManager(configPath)
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	settings := reloaded.GetFileManagerSettings("conn-1")
	if len(settings.History) != 1 || settings.History[0] != "/tmp" {
		t.Fatalf("history = %#v", settings.History)
	}
	if len(settings.Favorites) != 1 || settings.Favorites[0] != "/opt/nginx" {
		t.Fatalf("favorites = %#v", settings.Favorites)
	}
}

func TestUpdateSettingsPersistsSessionLogFields(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cm := newDiskTestConfigManager(configPath)
	if err := cm.UpdateSettings(map[string]interface{}{
		"session_log_enabled":        false,
		"session_log_retention_days": float64(7),
		"session_log_redact_enabled": false,
		"command_suggest_enabled":    false,
		"command_suggest_limit":      float64(12),
	}); err != nil {
		t.Fatal(err)
	}

	reloaded := newDiskTestConfigManager(configPath)
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	s := reloaded.GetSettings()
	if s.SessionLogEnabled {
		t.Fatal("expected SessionLogEnabled false")
	}
	if s.SessionLogRetentionDays != 7 {
		t.Fatalf("expected SessionLogRetentionDays 7, got %d", s.SessionLogRetentionDays)
	}
	if s.SessionLogRedactEnabled {
		t.Fatal("expected SessionLogRedactEnabled false")
	}
	if s.CommandSuggestEnabled {
		t.Fatal("expected CommandSuggestEnabled false")
	}
	if s.CommandSuggestLimit != 12 {
		t.Fatalf("expected CommandSuggestLimit 12, got %d", s.CommandSuggestLimit)
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
