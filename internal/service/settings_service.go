package service

import (
	"fmt"

	"AHaSSHTools/internal/config"
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

// GetFileManagerSettings returns file manager settings for a specific connection
func (s *SettingsService) GetFileManagerSettings(connectionId string) config.FileManagerSettings {
	if s.configManager == nil {
		return config.DefaultFileManagerSettings()
	}
	return s.configManager.GetFileManagerSettings(connectionId)
}

// UpdateFileManagerSettings updates file manager settings for a specific connection
func (s *SettingsService) UpdateFileManagerSettings(connectionId string, settings map[string]interface{}) error {
	if s.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	updates := map[string]interface{}{
		"connection_id":         connectionId,
		"file_manager_settings": settings,
	}
	return s.configManager.UpdateSettings(updates)
}
