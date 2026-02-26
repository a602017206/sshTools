package ssh

import (
	"fmt"
	"io"
	"path"
	"strings"
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
	cwdMu    sync.Mutex
	cwd      string
	prevCwd  string
	homeCwd  string
	inputBuf string
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
	if err != nil {
		return err
	}

	sm.trackCwdFromInput(managed, data)
	return nil
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

// GetCurrentWorkingDirectory gets the current working directory from the SSH session
// This executes 'pwd' command on the SSH session to get the actual current directory
func (sm *SessionManager) GetCurrentWorkingDirectory(sessionID string) (string, error) {
	sm.mu.RLock()
	managed, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	if !managed.Running {
		return "", fmt.Errorf("session not running: %s", sessionID)
	}

	managed.cwdMu.Lock()
	defer managed.cwdMu.Unlock()

	if managed.cwd != "" {
		return managed.cwd, nil
	}

	stdout, _, err := managed.Session.ExecuteCommand("pwd", 2*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	currentPath := strings.TrimSpace(stdout)
	if currentPath == "" {
		currentPath = "/"
	}

	managed.homeCwd = currentPath
	managed.cwd = currentPath

	if sftpClient, sftpExists := sm.sftpClients[sessionID]; sftpExists {
		sftpClient.SetCurrentPath(currentPath)
	}

	return currentPath, nil
}

// UpdateCurrentWorkingDirectory updates the tracked cwd for a session
func (sm *SessionManager) UpdateCurrentWorkingDirectory(sessionID, cwd string) error {
	sm.mu.RLock()
	managed, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if !managed.Running {
		return fmt.Errorf("session not running: %s", sessionID)
	}

	normalized := normalizeCwdPath(cwd)
	if normalized == "" {
		return fmt.Errorf("invalid cwd")
	}

	managed.cwdMu.Lock()
	if managed.homeCwd == "" {
		managed.homeCwd = normalized
	}
	managed.prevCwd = managed.cwd
	managed.cwd = normalized
	managed.cwdMu.Unlock()

	sm.syncSftpPath(sessionID, normalized)
	return nil
}

func (sm *SessionManager) trackCwdFromInput(managed *ManagedSession, data []byte) {
	if managed == nil || len(data) == 0 {
		return
	}

	var nextPath string
	var pathChanged bool

	managed.cwdMu.Lock()
	managed.inputBuf += string(data)
	lines, remainder := splitInputLines(managed.inputBuf)
	managed.inputBuf = remainder

	for _, line := range lines {
		if applyCdCommand(managed, line) {
			pathChanged = true
			nextPath = managed.cwd
		}
	}
	managed.cwdMu.Unlock()

	if pathChanged && nextPath != "" {
		sm.syncSftpPath(managed.ID, nextPath)
	}
}

func splitInputLines(input string) ([]string, string) {
	if input == "" {
		return nil, ""
	}

	lines := []string{}
	remaining := input
	for {
		idx := strings.IndexAny(remaining, "\r\n")
		if idx == -1 {
			break
		}

		line := remaining[:idx]
		lines = append(lines, line)

		skip := idx + 1
		if idx < len(remaining) && remaining[idx] == '\r' {
			if skip < len(remaining) && remaining[skip] == '\n' {
				skip++
			}
		}
		remaining = remaining[skip:]
	}

	return lines, remaining
}

func applyCdCommand(managed *ManagedSession, line string) bool {
	if managed == nil {
		return false
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}

	commands := splitShellCommands(trimmed)
	changed := false
	for _, command := range commands {
		if applyCdCommandTokens(managed, command) {
			changed = true
		}
	}

	return changed
}

func applyCdCommandTokens(managed *ManagedSession, command string) bool {
	command = strings.TrimSpace(command)
	if command == "" {
		return false
	}

	tokens := splitShellTokens(command)
	if len(tokens) == 0 || tokens[0] != "cd" {
		return false
	}

	idx := 1
	for idx < len(tokens) {
		if tokens[idx] == "-" || tokens[idx] == "--" {
			break
		}
		if strings.HasPrefix(tokens[idx], "-") {
			idx++
			continue
		}
		break
	}

	if idx < len(tokens) && tokens[idx] == "--" {
		idx++
	}

	target := ""
	if idx < len(tokens) {
		target = tokens[idx]
	}

	if strings.Contains(target, "$") {
		return false
	}

	current := managed.cwd
	if current == "" {
		current = managed.homeCwd
	}
	if current == "" {
		current = "/"
	}

	var next string
	if target == "" || target == "~" {
		next = managed.homeCwd
		if next == "" {
			next = current
		}
	} else if target == "-" {
		if managed.prevCwd != "" {
			next = managed.prevCwd
		} else {
			next = current
		}
	} else if strings.HasPrefix(target, "~") {
		if managed.homeCwd == "" {
			return false
		}
		if target == "~" {
			next = managed.homeCwd
		} else if strings.HasPrefix(target, "~/") {
			next = path.Join(managed.homeCwd, strings.TrimPrefix(target, "~/"))
		} else {
			return false
		}
	} else if path.IsAbs(target) {
		next = target
	} else {
		next = path.Join(current, target)
	}

	next = path.Clean(next)
	if next == managed.cwd {
		return false
	}

	managed.prevCwd = managed.cwd
	managed.cwd = next
	return true
}

func splitShellCommands(input string) []string {
	if input == "" {
		return nil
	}

	commands := []string{}
	var buf strings.Builder
	var inSingle bool
	var inDouble bool
	var escaped bool

	for i := 0; i < len(input); i++ {
		ch := input[i]
		if escaped {
			buf.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' && !inSingle {
			escaped = true
			continue
		}

		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			buf.WriteByte(ch)
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			buf.WriteByte(ch)
			continue
		}

		if !inSingle && !inDouble {
			if ch == ';' {
				commands = append(commands, buf.String())
				buf.Reset()
				continue
			}
			if ch == '&' && i+1 < len(input) && input[i+1] == '&' {
				commands = append(commands, buf.String())
				buf.Reset()
				i++
				continue
			}
			if ch == '|' && i+1 < len(input) && input[i+1] == '|' {
				commands = append(commands, buf.String())
				buf.Reset()
				i++
				continue
			}
		}

		buf.WriteByte(ch)
	}

	if buf.Len() > 0 {
		commands = append(commands, buf.String())
	}

	return commands
}

func splitShellTokens(input string) []string {
	if input == "" {
		return nil
	}

	var tokens []string
	var buf strings.Builder
	var inSingle bool
	var inDouble bool
	var escaped bool

	flush := func() {
		if buf.Len() > 0 {
			tokens = append(tokens, buf.String())
			buf.Reset()
		}
	}

	for i := 0; i < len(input); i++ {
		ch := input[i]
		if escaped {
			buf.WriteByte(ch)
			escaped = false
			continue
		}

		if ch == '\\' && !inSingle {
			escaped = true
			continue
		}

		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}

		if !inSingle && !inDouble && (ch == ' ' || ch == '\t') {
			flush()
			continue
		}

		buf.WriteByte(ch)
	}

	flush()
	return tokens
}

func (sm *SessionManager) syncSftpPath(sessionID, nextPath string) {
	sm.mu.RLock()
	sftpClient, sftpExists := sm.sftpClients[sessionID]
	sm.mu.RUnlock()
	if sftpExists {
		sftpClient.SetCurrentPath(nextPath)
	}
}

func normalizeCwdPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	cleaned := path.Clean(trimmed)
	if !path.IsAbs(cleaned) {
		cleaned = "/" + cleaned
	}
	return cleaned
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

	// For Windows local shells (pipe mode), send an initial newline to trigger the prompt
	// This is necessary because PowerShell/CMD in pipe mode don't output prompt automatically
	if managed.Type == SessionTypeLocal {
		go func() {
			// Small delay to ensure the output goroutine is ready
			time.Sleep(100 * time.Millisecond)
			// Send a newline to trigger the prompt
			if _, err := managed.Local.Write([]byte("\r\n")); err != nil {
				fmt.Printf("Failed to send initial newline: %v\n", err)
			}
		}()
	}

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
