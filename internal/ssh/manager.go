package ssh

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// SessionManager manages multiple SSH sessions
type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*ManagedSession
	sftpClients map[string]*SFTPClient
}

// SessionType represents the type of session (SSH or local)
type SessionType string

const (
	SessionTypeSSH   SessionType = "ssh"
	SessionTypeLocal SessionType = "local"
)

// ManagedSession represents a managed SSH session
type ManagedSession struct {
	ID       string
	Client   *Client
	Session  *Session
	Local    *LocalSession
	Type     SessionType
	Running  bool
	stopChan chan struct{}
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions:    make(map[string]*ManagedSession),
		sftpClients: make(map[string]*SFTPClient),
	}
}

// CreateSession creates a new SSH session
func (sm *SessionManager) CreateSession(sessionID string, cfg *Config) (*ManagedSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if session already exists
	if _, exists := sm.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session already exists: %s", sessionID)
	}

	// Create SSH client
	client, err := NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	// Connect to SSH server
	if err := client.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	managed := &ManagedSession{
		ID:       sessionID,
		Client:   client,
		Session:  session,
		Running:  false,
		stopChan: make(chan struct{}),
	}

	sm.sessions[sessionID] = managed
	return managed, nil
}

// StartShell starts a shell for the session
func (sm *SessionManager) StartShell(sessionID string, cols, rows int, onOutput func([]byte)) error {
	sm.mu.Lock()
	managed, exists := sm.sessions[sessionID]
	if !exists {
		sm.mu.Unlock()
		return fmt.Errorf("session not found: %s", sessionID)
	}
	sm.mu.Unlock()

	// Request PTY
	if err := managed.Session.RequestPTY("xterm-256color", rows, cols); err != nil {
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	// Start shell
	if err := managed.Session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	managed.Running = true

	// Start reading output
	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-managed.stopChan:
				return
			default:
				n, err := managed.Session.Read(buf)
				if err != nil {
					if err != io.EOF {
						fmt.Printf("Read error: %v\n", err)
					}
					return
				}
				if n > 0 && onOutput != nil {
					onOutput(buf[:n])
				}
			}
		}
	}()

	return nil
}

// WriteToSession writes data to a session
func (sm *SessionManager) WriteToSession(sessionID string, data []byte) error {
	sm.mu.RLock()
	managed, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	_, err := managed.Session.Write(data)
	return err
}

// ResizeSession resizes the terminal
func (sm *SessionManager) ResizeSession(sessionID string, cols, rows int) error {
	sm.mu.RLock()
	managed, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return managed.Session.Resize(rows, cols)
}

// CloseSession closes and removes a session
func (sm *SessionManager) CloseSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	managed, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	close(managed.stopChan)

	if sftpClient, exists := sm.sftpClients[sessionID]; exists {
		sftpClient.Close()
		delete(sm.sftpClients, sessionID)
	}

	if managed.Type == SessionTypeLocal {
		if managed.Local != nil {
			managed.Local.Close()
		}
	} else {
		if managed.Session != nil {
			managed.Session.Close()
		}
		if managed.Client != nil {
			managed.Client.Close()
		}
	}

	delete(sm.sessions, sessionID)
	return nil
}

// GetSession returns a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*ManagedSession, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	managed, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return managed, nil
}

// ListSessions returns all session IDs
func (sm *SessionManager) ListSessions() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ids := make([]string, 0, len(sm.sessions))
	for id := range sm.sessions {
		ids = append(ids, id)
	}
	return ids
}

// ExecuteCommand executes a command on an existing session's connection
func (sm *SessionManager) ExecuteCommand(sessionID string, cmd string, timeout time.Duration) (string, string, error) {
	sm.mu.RLock()
	managed, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return "", "", fmt.Errorf("session not found: %s", sessionID)
	}

	if !managed.Running {
		return "", "", fmt.Errorf("session not running: %s", sessionID)
	}

	return managed.Session.ExecuteCommand(cmd, timeout)
}

// GetOrCreateSFTPClient gets or creates an SFTP client for a session
func (sm *SessionManager) GetOrCreateSFTPClient(sessionID string) (*SFTPClient, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if SFTP client already exists
	if sftpClient, exists := sm.sftpClients[sessionID]; exists {
		return sftpClient, nil
	}

	// Get session
	managed, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Create SFTP client
	sftpClient, err := NewSFTPClient(managed.Client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	// Store SFTP client
	sm.sftpClients[sessionID] = sftpClient

	return sftpClient, nil
}

// CloseSFTPClient closes an SFTP client for a session
func (sm *SessionManager) CloseSFTPClient(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sftpClient, exists := sm.sftpClients[sessionID]
	if !exists {
		return nil // Already closed or never created
	}

	err := sftpClient.Close()
	delete(sm.sftpClients, sessionID)

	return err
}

// CreateLocalSession creates a new local shell session
func (sm *SessionManager) CreateLocalSession(sessionID string, shellType string) (*ManagedSession, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session already exists: %s", sessionID)
	}

	localSession, err := NewLocalSession(shellType)
	if err != nil {
		return nil, fmt.Errorf("failed to create local session: %w", err)
	}

	managed := &ManagedSession{
		ID:       sessionID,
		Client:   nil,
		Session:  nil,
		Local:    localSession,
		Type:     SessionTypeLocal,
		Running:  false,
		stopChan: make(chan struct{}),
	}

	sm.sessions[sessionID] = managed
	return managed, nil
}

// StartLocalShell starts a local shell session
func (sm *SessionManager) StartLocalShell(sessionID string, cols, rows int, onOutput func([]byte)) error {
	sm.mu.Lock()
	managed, exists := sm.sessions[sessionID]
	if !exists {
		sm.mu.Unlock()
		return fmt.Errorf("session not found: %s", sessionID)
	}
	sm.mu.Unlock()

	if err := managed.Local.Resize(cols, rows); err != nil {
		return fmt.Errorf("failed to resize PTY: %w", err)
	}

	managed.Running = true

	go func() {
		buf := make([]byte, 1024)
		for {
			select {
			case <-managed.stopChan:
				return
			default:
				n, err := managed.Local.Read(buf)
				if err != nil {
					if err != io.EOF {
						fmt.Printf("Read error: %v\n", err)
					}
					return
				}
				if n > 0 && onOutput != nil {
					onOutput(buf[:n])
				}
			}
		}
	}()

	return nil
}

// WriteToLocalSession writes data to a local session
func (sm *SessionManager) WriteToLocalSession(sessionID string, data []byte) error {
	sm.mu.RLock()
	managed, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if managed.Type != SessionTypeLocal {
		return fmt.Errorf("session is not a local session: %s", sessionID)
	}

	_, err := managed.Local.Write(data)
	return err
}

// ResizeLocalSession resizes local terminal
func (sm *SessionManager) ResizeLocalSession(sessionID string, cols, rows int) error {
	sm.mu.RLock()
	managed, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if managed.Type != SessionTypeLocal {
		return fmt.Errorf("session is not a local session: %s", sessionID)
	}

	return managed.Local.Resize(cols, rows)
}

// CloseLocalSession closes a local session
func (sm *SessionManager) CloseLocalSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	managed, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	close(managed.stopChan)

	if managed.Local != nil {
		managed.Local.Close()
	}

	delete(sm.sessions, sessionID)
	return nil
}
