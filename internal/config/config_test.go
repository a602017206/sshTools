package config

import "testing"

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
