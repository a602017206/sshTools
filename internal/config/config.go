package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConnectionConfig represents a saved connection (SSH, Database, Docker)
type ConnectionConfig struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	User     string            `json:"user"`
	AuthType string            `json:"auth_type"` // "password" or "key"
	KeyPath  string            `json:"key_path,omitempty"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Type     string            `json:"type,omitempty"` // "ssh", "database", "docker"
}

// AppConfig represents application configuration
type AppConfig struct {
	Connections []ConnectionConfig `json:"connections"`
	Settings    AppSettings        `json:"settings"`
}

// AppSettings represents application settings
type AppSettings struct {
	Theme         string `json:"theme"` // "light" or "dark"
	FontFamily    string `json:"font_family"`
	FontSize      int    `json:"font_size"`
	TerminalTheme string `json:"terminal_theme"`
	SidebarWidth  int    `json:"sidebar_width"` // Sidebar width in pixels

	// Monitor panel settings
	MonitorCollapsed       bool `json:"monitor_collapsed"`
	MonitorWidth           int  `json:"monitor_width"`
	MonitorRefreshInterval int  `json:"monitor_refresh_interval"` // seconds

	// File manager settings
	FileManagerCollapsed     bool                           `json:"file_manager_collapsed"`
	FileManagerWidth         int                            `json:"file_manager_width"`
	FileManagerShowHidden    bool                           `json:"file_manager_show_hidden"`
	FileManagerSortBy        string                         `json:"file_manager_sort_by"`
	FileManagerSortOrder     string                         `json:"file_manager_sort_order"`
	FileManagerPerConnection map[string]FileManagerSettings `json:"file_manager_per_connection,omitempty"`
}

// FileManagerSettings stores file manager configuration per connection
type FileManagerSettings struct {
	DirectoryTracking bool     `json:"directory_tracking"` // Enable terminal → FM sync
	HistoryEnabled    bool     `json:"history_enabled"`    // Enable history tracking
	HistoryLimit      int      `json:"history_limit"`      // Max history entries
	History           []string `json:"history"`            // History path list
}

// DefaultSettings returns default application settings
func DefaultSettings() AppSettings {
	return AppSettings{
		Theme:                  "dark",
		FontFamily:             "monospace",
		FontSize:               14,
		TerminalTheme:          "default",
		SidebarWidth:           300,
		MonitorCollapsed:       true,
		MonitorWidth:           350,
		MonitorRefreshInterval: 2,
		FileManagerCollapsed:   true,
		FileManagerWidth:       400,
		FileManagerShowHidden:  false,
		FileManagerSortBy:      "name",
		FileManagerSortOrder:   "asc",
	}
}

// DefaultFileManagerSettings returns default file manager settings for a connection
func DefaultFileManagerSettings() FileManagerSettings {
	return FileManagerSettings{
		DirectoryTracking: false,
		HistoryEnabled:    true,
		HistoryLimit:      5,
		History:           []string{},
	}
}

// ConfigManager manages application configuration
type ConfigManager struct {
	configPath string
	config     *AppConfig
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".ahasshtools")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.json")

	cm := &ConfigManager{
		configPath: configPath,
		config: &AppConfig{
			Connections: []ConnectionConfig{},
			Settings:    DefaultSettings(),
		},
	}

	// Load existing config if available
	if _, err := os.Stat(configPath); err == nil {
		if err := cm.Load(); err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
	}

	return cm, nil
}

// Load loads configuration from disk
func (cm *ConfigManager) Load() error {
	data, err := os.ReadFile(cm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, cm.config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	return nil
}

// Save saves configuration to disk
func (cm *ConfigManager) Save() error {
	data, err := json.MarshalIndent(cm.config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *AppConfig {
	return cm.config
}

// AddConnection adds a new connection configuration
func (cm *ConfigManager) AddConnection(conn ConnectionConfig) error {
	cm.config.Connections = append(cm.config.Connections, conn)
	return cm.Save()
}

// GetConnection retrieves a single connection configuration by ID
func (cm *ConfigManager) GetConnection(id string) (ConnectionConfig, error) {
	for _, conn := range cm.config.Connections {
		if conn.ID == id {
			return conn, nil
		}
	}
	return ConnectionConfig{}, fmt.Errorf("connection not found: %s", id)
}

// RemoveConnection removes a connection by ID
func (cm *ConfigManager) RemoveConnection(id string) error {
	for i, conn := range cm.config.Connections {
		if conn.ID == id {
			cm.config.Connections = append(cm.config.Connections[:i], cm.config.Connections[i+1:]...)
			return cm.Save()
		}
	}
	return fmt.Errorf("connection not found: %s", id)
}

// UpdateConnection updates an existing connection
func (cm *ConfigManager) UpdateConnection(conn ConnectionConfig) error {
	for i, c := range cm.config.Connections {
		if c.ID == conn.ID {
			cm.config.Connections[i] = conn
			return cm.Save()
		}
	}
	return fmt.Errorf("connection not found: %s", conn.ID)
}

// GetSettings returns the current application settings
func (cm *ConfigManager) GetSettings() AppSettings {
	return cm.config.Settings
}

// UpdateSettings updates application settings (partial update)
func (cm *ConfigManager) UpdateSettings(updates map[string]interface{}) error {
	if theme, ok := updates["theme"].(string); ok {
		cm.config.Settings.Theme = theme
	}
	if sidebarWidth, ok := updates["sidebar_width"].(float64); ok {
		cm.config.Settings.SidebarWidth = int(sidebarWidth)
	}
	if fontFamily, ok := updates["font_family"].(string); ok {
		cm.config.Settings.FontFamily = fontFamily
	}
	if fontSize, ok := updates["font_size"].(float64); ok {
		cm.config.Settings.FontSize = int(fontSize)
	}
	if terminalTheme, ok := updates["terminal_theme"].(string); ok {
		cm.config.Settings.TerminalTheme = terminalTheme
	}

	// Monitor panel settings
	if monitorCollapsed, ok := updates["monitor_collapsed"].(bool); ok {
		cm.config.Settings.MonitorCollapsed = monitorCollapsed
	}
	if monitorWidth, ok := updates["monitor_width"].(float64); ok {
		cm.config.Settings.MonitorWidth = int(monitorWidth)
	}
	if monitorRefreshInterval, ok := updates["monitor_refresh_interval"].(float64); ok {
		cm.config.Settings.MonitorRefreshInterval = int(monitorRefreshInterval)
	}

	// File manager settings
	if fileManagerCollapsed, ok := updates["file_manager_collapsed"].(bool); ok {
		cm.config.Settings.FileManagerCollapsed = fileManagerCollapsed
	}
	if fileManagerWidth, ok := updates["file_manager_width"].(float64); ok {
		cm.config.Settings.FileManagerWidth = int(fileManagerWidth)
	}
	if fileManagerShowHidden, ok := updates["file_manager_show_hidden"].(bool); ok {
		cm.config.Settings.FileManagerShowHidden = fileManagerShowHidden
	}
	if fileManagerSortBy, ok := updates["file_manager_sort_by"].(string); ok {
		cm.config.Settings.FileManagerSortBy = fileManagerSortBy
	}
	if fileManagerSortOrder, ok := updates["file_manager_sort_order"].(string); ok {
		cm.config.Settings.FileManagerSortOrder = fileManagerSortOrder
	}

	// File manager per-connection settings
	if connID, ok := updates["connection_id"].(string); ok {
		if fmSettings, ok := updates["file_manager_settings"].(map[string]interface{}); ok {
			if cm.config.Settings.FileManagerPerConnection == nil {
				cm.config.Settings.FileManagerPerConnection = make(map[string]FileManagerSettings)
			}

			settings := cm.config.Settings.FileManagerPerConnection[connID]

			if directoryTracking, ok := fmSettings["directory_tracking"].(bool); ok {
				settings.DirectoryTracking = directoryTracking
			}
			if historyEnabled, ok := fmSettings["history_enabled"].(bool); ok {
				settings.HistoryEnabled = historyEnabled
			}
			if historyLimit, ok := fmSettings["history_limit"].(float64); ok {
				settings.HistoryLimit = int(historyLimit)
			}
			if history, ok := fmSettings["history"].([]interface{}); ok {
				historyList := []string{}
				for _, h := range history {
					if path, ok := h.(string); ok {
						historyList = append(historyList, path)
					}
				}
				settings.History = historyList
			}

			cm.config.Settings.FileManagerPerConnection[connID] = settings
		}
	}

	return cm.Save()
}

// GetFileManagerSettings returns file manager settings for a specific connection
func (cm *ConfigManager) GetFileManagerSettings(connectionId string) FileManagerSettings {
	if cm.config.Settings.FileManagerPerConnection != nil {
		if settings, exists := cm.config.Settings.FileManagerPerConnection[connectionId]; exists {
			return settings
		}
	}
	return DefaultFileManagerSettings()
}
