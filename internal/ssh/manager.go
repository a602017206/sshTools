package ssh

import (
	"fmt"
	"io"
	"sync"
)

// SessionManager manages multiple SSH sessions
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*ManagedSession
}

// ManagedSession represents a managed SSH session
type ManagedSession struct {
	ID       string
	Client   *Client
	Session  *Session
	Running  bool
	stopChan chan struct{}
}

// NewSessionManager creates a new session manager
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*ManagedSession),
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

	// Stop output reader
	close(managed.stopChan)

	// Close session and client
	if managed.Session != nil {
		managed.Session.Close()
	}
	if managed.Client != nil {
		managed.Client.Close()
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
