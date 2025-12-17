package service

import (
	"fmt"

	"sshTools/internal/config"
)

// SettingsService handles application settings operations
type SettingsService struct {
	configManager *config.ConfigManager
}

// NewSettingsService creates a new settings service
func NewSettingsService(cm *config.ConfigManager) *SettingsService {
	return &SettingsService{
		configManager: cm,
	}
}

// GetSettings returns application settings
func (s *SettingsService) GetSettings() config.AppSettings {
	if s.configManager == nil {
		return config.DefaultSettings()
	}
	return s.configManager.GetSettings()
}

// UpdateSettings updates application settings
func (s *SettingsService) UpdateSettings(updates map[string]interface{}) error {
	if s.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return s.configManager.UpdateSettings(updates)
}
