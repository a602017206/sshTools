package service

import (
	"fmt"

	"sshTools/internal/config"
	"sshTools/internal/ssh"
	"sshTools/internal/store"
)

// ConnectionService handles SSH connection management operations
type ConnectionService struct {
	configManager   *config.ConfigManager
	credentialStore *store.CredentialStore
}

// NewConnectionService creates a new connection service
func NewConnectionService(cm *config.ConfigManager, cs *store.CredentialStore) *ConnectionService {
	return &ConnectionService{
		configManager:   cm,
		credentialStore: cs,
	}
}

// GetConnections returns all saved connections
func (s *ConnectionService) GetConnections() ([]config.ConnectionConfig, error) {
	if s.configManager == nil {
		return []config.ConnectionConfig{}, nil
	}
	return s.configManager.GetConfig().Connections, nil
}

// AddConnection adds a new SSH connection
func (s *ConnectionService) AddConnection(conn config.ConnectionConfig) error {
	if s.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return s.configManager.AddConnection(conn)
}

// UpdateConnection updates an existing SSH connection
func (s *ConnectionService) UpdateConnection(conn config.ConnectionConfig) error {
	if s.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return s.configManager.UpdateConnection(conn)
}

// RemoveConnection removes an SSH connection
func (s *ConnectionService) RemoveConnection(id string) error {
	if s.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return s.configManager.RemoveConnection(id)
}

// TestConnection tests an SSH connection
// authType: "password" or "key"
// authValue: password for password auth, or key file path for key auth
// passphrase: passphrase for encrypted keys (optional)
func (s *ConnectionService) TestConnection(host string, port int, user, authType, authValue, passphrase string) error {
	sshConfig := &ssh.Config{
		Host: host,
		Port: port,
		User: user,
	}

	if authType == "key" {
		sshConfig.KeyPath = authValue
		sshConfig.Passphrase = passphrase
	} else {
		sshConfig.Password = authValue
	}

	client, err := ssh.NewClient(sshConfig)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	client.Close()
	return nil
}

// SavePassword saves a password for a connection (encrypted)
func (s *ConnectionService) SavePassword(connectionID, password string) error {
	if s.credentialStore == nil {
		return fmt.Errorf("credential store not initialized")
	}
	return s.credentialStore.Store(connectionID, password)
}

// GetPassword retrieves a saved password for a connection
func (s *ConnectionService) GetPassword(connectionID string) (string, error) {
	if s.credentialStore == nil {
		return "", fmt.Errorf("credential store not initialized")
	}
	return s.credentialStore.Get(connectionID)
}

// HasPassword checks if a password is saved for a connection
func (s *ConnectionService) HasPassword(connectionID string) bool {
	if s.credentialStore == nil {
		return false
	}
	return s.credentialStore.Has(connectionID)
}

// DeletePassword removes a saved password for a connection
func (s *ConnectionService) DeletePassword(connectionID string) error {
	if s.credentialStore == nil {
		return fmt.Errorf("credential store not initialized")
	}
	return s.credentialStore.Delete(connectionID)
}
