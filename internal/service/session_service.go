package service

import (
	"fmt"

	"AHaSSHTools/internal/ssh"
)

// OutputCallback is a callback function for SSH output
type OutputCallback func(data []byte)

// SessionService handles SSH session operations
type SessionService struct {
	sessionManager *ssh.SessionManager
}

// NewSessionService creates a new session service
func NewSessionService(sm *ssh.SessionManager) *SessionService {
	return &SessionService{
		sessionManager: sm,
	}
}

// ConnectSSH creates and starts an SSH session
// authType: "password" or "key"
// authValue: password for password auth, or key file path for key auth
// passphrase: passphrase for encrypted keys (optional)
// outputCallback: callback function for SSH output data
func (s *SessionService) ConnectSSH(sessionID, host string, port int, user, authType, authValue, passphrase string, cols, rows int, outputCallback OutputCallback) error {
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

	// Create session
	_, err := s.sessionManager.CreateSession(sessionID, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Start shell with output handler
	err = s.sessionManager.StartShell(sessionID, cols, rows, outputCallback)
	if err != nil {
		s.sessionManager.CloseSession(sessionID)
		return fmt.Errorf("failed to start shell: %w", err)
	}

	return nil
}

// SendData sends data to an SSH session
func (s *SessionService) SendData(sessionID string, data string) error {
	return s.sessionManager.WriteToSession(sessionID, []byte(data))
}

// SendDataBytes sends raw bytes to an SSH session
func (s *SessionService) SendDataBytes(sessionID string, data []byte) error {
	return s.sessionManager.WriteToSession(sessionID, data)
}

// SendLocalData sends data to a local shell session
func (s *SessionService) SendLocalData(sessionID string, data string) error {
	return s.sessionManager.WriteToLocalSession(sessionID, []byte(data))
}

// SendLocalDataBytes sends raw bytes to a local shell session
func (s *SessionService) SendLocalDataBytes(sessionID string, data []byte) error {
	return s.sessionManager.WriteToLocalSession(sessionID, data)
}

// ResizeLocalTerminal resizes terminal for a local shell session
func (s *SessionService) ResizeLocalTerminal(sessionID string, cols, rows int) error {
	return s.sessionManager.ResizeLocalSession(sessionID, cols, rows)
}

// ResizeTerminal resizes the terminal for an SSH session
func (s *SessionService) ResizeTerminal(sessionID string, cols, rows int) error {
	return s.sessionManager.ResizeSession(sessionID, cols, rows)
}

// CloseSession closes an SSH session
func (s *SessionService) CloseSession(sessionID string) error {
	return s.sessionManager.CloseSession(sessionID)
}

// ListSessions returns all active session IDs
func (s *SessionService) ListSessions() []string {
	return s.sessionManager.ListSessions()
}

// GetSFTPClient gets or creates an SFTP client for a session
func (s *SessionService) GetSFTPClient(sessionID string) (*ssh.SFTPClient, error) {
	return s.sessionManager.GetOrCreateSFTPClient(sessionID)
}

// ConnectLocalShell creates and starts a local shell session
func (s *SessionService) ConnectLocalShell(sessionID string, shellType string, cols, rows int, outputCallback OutputCallback) error {
	_, err := s.sessionManager.CreateLocalSession(sessionID, shellType)
	if err != nil {
		return fmt.Errorf("failed to create local session: %w", err)
	}

	err = s.sessionManager.StartLocalShell(sessionID, cols, rows, outputCallback)
	if err != nil {
		s.sessionManager.CloseSession(sessionID)
		return fmt.Errorf("failed to start local shell: %w", err)
	}

	return nil
}
