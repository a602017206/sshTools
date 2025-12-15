package main

import (
	"context"
	"fmt"

	"sshTools/internal/config"
	"sshTools/internal/ssh"
	"sshTools/internal/store"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	configManager   *config.ConfigManager
	credentialStore *store.CredentialStore
	sessionManager  *ssh.SessionManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		sessionManager: ssh.NewSessionManager(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize configuration manager
	cm, err := config.NewConfigManager()
	if err != nil {
		fmt.Printf("Failed to initialize config manager: %v\n", err)
	}
	a.configManager = cm

	// Initialize credential store
	a.credentialStore = store.NewCredentialStore()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// GetConnections returns all saved connections
func (a *App) GetConnections() []config.ConnectionConfig {
	if a.configManager == nil {
		return []config.ConnectionConfig{}
	}
	return a.configManager.GetConfig().Connections
}

// AddConnection adds a new SSH connection
func (a *App) AddConnection(conn config.ConnectionConfig) error {
	if a.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return a.configManager.AddConnection(conn)
}

// RemoveConnection removes an SSH connection
func (a *App) RemoveConnection(id string) error {
	if a.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return a.configManager.RemoveConnection(id)
}

// TestConnection tests an SSH connection
func (a *App) TestConnection(host string, port int, user, password string) error {
	client, err := ssh.NewClient(&ssh.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	client.Close()
	return nil
}

// ConnectSSH creates and starts an SSH session
func (a *App) ConnectSSH(sessionID, host string, port int, user, password string, cols, rows int) error {
	// Create session
	_, err := a.sessionManager.CreateSession(sessionID, &ssh.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Start shell with output handler
	err = a.sessionManager.StartShell(sessionID, cols, rows, func(data []byte) {
		// Emit output to frontend
		runtime.EventsEmit(a.ctx, "ssh:output:"+sessionID, string(data))
	})
	if err != nil {
		a.sessionManager.CloseSession(sessionID)
		return fmt.Errorf("failed to start shell: %w", err)
	}

	fmt.Printf("SSH session started: %s (%s@%s:%d)\n", sessionID, user, host, port)
	return nil
}

// SendSSHData sends data to an SSH session
func (a *App) SendSSHData(sessionID string, data string) error {
	return a.sessionManager.WriteToSession(sessionID, []byte(data))
}

// ResizeSSH resizes the terminal for an SSH session
func (a *App) ResizeSSH(sessionID string, cols, rows int) error {
	return a.sessionManager.ResizeSession(sessionID, cols, rows)
}

// CloseSSH closes an SSH session
func (a *App) CloseSSH(sessionID string) error {
	err := a.sessionManager.CloseSession(sessionID)
	if err != nil {
		return err
	}
	fmt.Printf("SSH session closed: %s\n", sessionID)
	return nil
}

// ListSSHSessions returns all active session IDs
func (a *App) ListSSHSessions() []string {
	return a.sessionManager.ListSessions()
}
