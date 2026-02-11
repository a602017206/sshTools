package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/ssh"
	"AHaSSHTools/internal/store"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/argon2"
)

var Version = "dev"

var cwdRegex = regexp.MustCompile(`\033\]0;CWD:([^\007]+)\007`)

// App struct
type App struct {
	ctx context.Context

	// Services
	connectionService *service.ConnectionService
	sessionService    *service.SessionService
	sftpService       *service.SFTPService
	monitorService    *service.MonitorService
	settingsService   *service.SettingsService
	devToolsService   *service.DevToolsService
	configManager     *config.ConfigManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize configuration manager
	configManager, err := config.NewConfigManager()
	if err != nil {
		fmt.Printf("Failed to initialize config manager: %v\n", err)
	}
	a.configManager = configManager

	// Initialize credential store
	credentialStore := store.NewCredentialStore()

	// Initialize managers
	sessionManager := ssh.NewSessionManager()
	transferManager := ssh.NewTransferManager()

	// Initialize services
	a.connectionService = service.NewConnectionService(configManager, credentialStore)
	a.sessionService = service.NewSessionService(sessionManager)
	a.sftpService = service.NewSFTPService(sessionManager, transferManager)
	a.monitorService = service.NewMonitorService(sessionManager)
	a.settingsService = service.NewSettingsService(configManager)
	a.devToolsService = service.NewDevToolsService()
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetVersion() string {
	return Version
}

// GetConnections returns all saved connections
func (a *App) GetConnections() []config.ConnectionConfig {
	conns, _ := a.connectionService.GetConnections()
	return conns
}

// GetConnection retrieves a single connection by ID
func (a *App) GetConnection(id string) (config.ConnectionConfig, error) {
	return a.connectionService.GetConnection(id)
}

// AddConnection adds a new SSH connection
func (a *App) AddConnection(conn config.ConnectionConfig) error {
	return a.connectionService.AddConnection(conn)
}

// UpdateConnection updates an existing SSH connection
func (a *App) UpdateConnection(conn config.ConnectionConfig) error {
	return a.connectionService.UpdateConnection(conn)
}

// RemoveConnection removes an SSH connection
func (a *App) RemoveConnection(id string) error {
	return a.connectionService.RemoveConnection(id)
}

// TestConnection tests an SSH connection
// authType: "password" or "key"
// authValue: password for password auth, or key file path for key auth
// passphrase: passphrase for encrypted keys (optional)
func (a *App) TestConnection(host string, port int, user, authType, authValue, passphrase string) error {
	return a.connectionService.TestConnection(host, port, user, authType, authValue, passphrase)
}

func (a *App) ConnectSSH(sessionID, host string, port int, user, authType, authValue, passphrase string, cols, rows int) error {
	err := a.sessionService.ConnectSSH(sessionID, host, port, user, authType, authValue, passphrase, cols, rows, func(data []byte) {
		cwd := a.parseCWDFromOutput(sessionID, data)
		if cwd != "" {
			runtime.EventsEmit(a.ctx, "ssh:cwd:"+sessionID, cwd)
		}

		encoded := base64.StdEncoding.EncodeToString(data)
		runtime.EventsEmit(a.ctx, "ssh:output:"+sessionID, encoded)
	})

	if err == nil {
		fmt.Printf("SSH session started: %s (%s@%s:%d)\n", sessionID, user, host, port)
		a.setupCWDTracking(sessionID)
	}
	return err
}

func (a *App) parseCWDFromOutput(sessionID string, data []byte) string {
	matches := cwdRegex.FindSubmatch(data)
	if len(matches) >= 2 {
		cwd := string(matches[1])
		if err := a.sftpService.UpdateCurrentPath(sessionID, cwd); err == nil {
			return cwd
		}
	}
	return ""
}

func (a *App) setupCWDTracking(sessionID string) {
	go func() {
		time.Sleep(500 * time.Millisecond)
		promptCmd := `export PROMPT_COMMAND='echo -ne "\033]0;CWD:$(pwd)\007"'` + "\n"
		if err := a.sessionService.SendData(sessionID, promptCmd); err != nil {
			fmt.Printf("Failed to setup CWD tracking for session %s: %v\n", sessionID, err)
		}
	}()
}

// SendSSHData sends data to an SSH session
func (a *App) SendSSHData(sessionID string, data string) error {
	return a.sessionService.SendData(sessionID, data)
}

// SendSSHDataBinary sends base64-encoded binary data to an SSH session
func (a *App) SendSSHDataBinary(sessionID string, data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return a.sessionService.SendDataBytes(sessionID, decoded)
}

// ResizeSSH resizes the terminal for an SSH session
func (a *App) ResizeSSH(sessionID string, cols, rows int) error {
	return a.sessionService.ResizeTerminal(sessionID, cols, rows)
}

// CloseSSH closes an SSH session
func (a *App) CloseSSH(sessionID string) error {
	err := a.sessionService.CloseSession(sessionID)
	if err != nil {
		return err
	}
	fmt.Printf("SSH session closed: %s\n", sessionID)
	return nil
}

// ConnectLocalShell creates and starts a local shell session
func (a *App) ConnectLocalShell(sessionID string, shellType string, cols, rows int) error {
	err := a.sessionService.ConnectLocalShell(sessionID, shellType, cols, rows, func(data []byte) {
		// Encode binary data as base64 to preserve ZMODEM protocol bytes
		encoded := base64.StdEncoding.EncodeToString(data)
		runtime.EventsEmit(a.ctx, "local:output:"+sessionID, encoded)
	})

	if err == nil {
		fmt.Printf("Local shell session started: %s\n", sessionID)
	}
	return err
}

// SendLocalShellData sends data to a local shell session
func (a *App) SendLocalShellData(sessionID string, data string) error {
	return a.sessionService.SendLocalData(sessionID, data)
}

// SendLocalShellDataBinary sends base64-encoded binary data to a local shell session
func (a *App) SendLocalShellDataBinary(sessionID string, data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return a.sessionService.SendLocalDataBytes(sessionID, decoded)
}

// SaveBinaryFile saves base64-encoded file contents to disk
func (a *App) SaveBinaryFile(filename string, data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存文件",
		DefaultFilename: filename,
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}

	if err := os.WriteFile(path, decoded, 0o644); err != nil {
		return "", err
	}

	return path, nil
}

// ResizeLocalShell resizes a local shell session
func (a *App) ResizeLocalShell(sessionID string, cols, rows int) error {
	return a.sessionService.ResizeLocalTerminal(sessionID, cols, rows)
}

// ListSSHSessions returns all active session IDs
func (a *App) ListSSHSessions() []string {
	return a.sessionService.ListSessions()
}

// ShowMessageDialog shows an information message dialog
func (a *App) ShowMessageDialog(title, message string) {
	_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
}

// ShowErrorDialog shows an error message dialog
func (a *App) ShowErrorDialog(title, message string) {
	_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   title,
		Message: message,
	})
}

// ShowQuestionDialog shows a question dialog and returns true if user confirms
func (a *App) ShowQuestionDialog(title, message string) (bool, error) {
	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"是", "否"},
		DefaultButton: "是",
		CancelButton:  "否",
	})

	if err != nil {
		return false, err
	}

	return result == "是", nil
}

// SelectSSHKeyFile opens a file picker dialog for selecting SSH private key files
func (a *App) SelectSSHKeyFile() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 SSH 私钥文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "SSH 私钥 (id_rsa, id_ed25519, id_ecdsa)",
				Pattern:     "id_rsa;id_ed25519;id_ecdsa;*.pem;*.key",
			},
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return "", err
	}

	return filePath, nil
}

// SavePassword saves a password for a connection (encrypted)
func (a *App) SavePassword(connectionID, password string) error {
	return a.connectionService.SavePassword(connectionID, password)
}

// GetPassword retrieves a saved password for a connection
func (a *App) GetPassword(connectionID string) (string, error) {
	return a.connectionService.GetPassword(connectionID)
}

// HasPassword checks if a password is saved for a connection
func (a *App) HasPassword(connectionID string) bool {
	return a.connectionService.HasPassword(connectionID)
}

// DeletePassword removes a saved password for a connection
func (a *App) DeletePassword(connectionID string) error {
	return a.connectionService.DeletePassword(connectionID)
}

// GetSettings returns application settings
func (a *App) GetSettings() config.AppSettings {
	return a.settingsService.GetSettings()
}

// UpdateSettings updates application settings
func (a *App) UpdateSettings(updates map[string]interface{}) error {
	return a.settingsService.UpdateSettings(updates)
}

// GetMonitoringData retrieves monitoring data for a session
func (a *App) GetMonitoringData(sessionID string) (*ssh.MonitoringData, error) {
	return a.monitorService.GetMonitoringData(sessionID)
}

// ListFiles lists files in a directory
func (a *App) ListFiles(sessionID string, path string) ([]ssh.FileInfo, error) {
	return a.sftpService.ListFiles(sessionID, path)
}

// ChangeDirectory changes the current working directory for a session
func (a *App) ChangeDirectory(sessionID string, path string) error {
	return a.sftpService.ChangeDirectory(sessionID, path)
}

// GetCurrentPath returns the current working directory
func (a *App) GetCurrentPath(sessionID string) (string, error) {
	return a.sftpService.GetCurrentPath(sessionID)
}

// UpdateCurrentPath updates the tracked working directory for a session
func (a *App) UpdateCurrentPath(sessionID string, path string) error {
	return a.sftpService.UpdateCurrentPath(sessionID, path)
}

// GetFileManagerSettings returns file manager settings for a specific connection
func (a *App) GetFileManagerSettings(connectionId string) (config.FileManagerSettings, error) {
	return a.settingsService.GetFileManagerSettings(connectionId), nil
}

// UpdateFileManagerSettings updates file manager settings for a specific connection
func (a *App) UpdateFileManagerSettings(connectionId string, settings map[string]interface{}) error {
	updates := map[string]interface{}{
		"connection_id":         connectionId,
		"file_manager_settings": settings,
	}
	return a.configManager.UpdateSettings(updates)
}

// UploadFiles uploads multiple files
func (a *App) UploadFiles(sessionID string, localPaths []string, remotePath string) ([]string, error) {
	return a.sftpService.UploadFiles(sessionID, localPaths, remotePath, func(progress ssh.TransferProgress) {
		// Emit event to frontend
		runtime.EventsEmit(a.ctx, "sftp:progress:"+progress.TransferID, progress)
	})
}

// DownloadFile downloads a single file
func (a *App) DownloadFile(sessionID string, remotePath string, localPath string) (string, error) {
	// Use service with Wails-specific progress callback
	return a.sftpService.DownloadFile(sessionID, remotePath, localPath, func(progress ssh.TransferProgress) {
		// Emit event to frontend
		runtime.EventsEmit(a.ctx, "sftp:progress:"+progress.TransferID, progress)
	})
}

// DownloadFiles downloads multiple files
func (a *App) DownloadFiles(sessionID string, remotePaths []string, localPath string) ([]string, error) {
	return a.sftpService.DownloadFiles(sessionID, remotePaths, localPath, func(progress ssh.TransferProgress) {
		// Emit event to frontend
		runtime.EventsEmit(a.ctx, "sftp:progress:"+progress.TransferID, progress)
	})
}

// DeleteFile deletes a single file or directory
func (a *App) DeleteFile(sessionID string, path string) error {
	return a.sftpService.DeleteFile(sessionID, path)
}

// DeleteFiles deletes multiple files or directories
func (a *App) DeleteFiles(sessionID string, paths []string) error {
	return a.sftpService.DeleteFiles(sessionID, paths)
}

// RenameFile renames a file or directory
func (a *App) RenameFile(sessionID string, oldPath string, newPath string) error {
	return a.sftpService.RenameFile(sessionID, oldPath, newPath)
}

// CreateDirectory creates a new directory
func (a *App) CreateDirectory(sessionID string, path string) error {
	return a.sftpService.CreateDirectory(sessionID, path)
}

// GetFileInfo gets information about a file
func (a *App) GetFileInfo(sessionID string, path string) (*ssh.FileInfo, error) {
	return a.sftpService.GetFileInfo(sessionID, path)
}

// SearchDirectories searches for directories matching the query recursively
func (a *App) SearchDirectories(sessionID string, searchPath string, query string, maxDepth int, maxResults int) ([]ssh.SearchResult, error) {
	return a.sftpService.SearchDirectories(sessionID, searchPath, query, maxDepth, maxResults)
}

// CancelTransfer cancels a file transfer
func (a *App) CancelTransfer(transferID string) error {
	return a.sftpService.CancelTransfer(transferID)
}

// GetTransferStatus gets the status of a transfer
func (a *App) GetTransferStatus(transferID string) (*ssh.TransferProgress, error) {
	return a.sftpService.GetTransferStatus(transferID)
}

// SelectUploadFiles opens a file picker for selecting files to upload
func (a *App) SelectUploadFiles() ([]string, error) {
	filePaths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要上传的文件",
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}

// SelectDownloadDirectory opens a directory picker for selecting download destination
func (a *App) SelectDownloadDirectory() (string, error) {
	dirPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择下载目录",
	})

	if err != nil {
		return "", err
	}

	return dirPath, nil
}

// ============================================================================
// DevTools Methods - 开发工具集相关方法
// ============================================================================

// FormatJSON 格式化JSON字符串
func (a *App) FormatJSON(input string) (string, error) {
	return a.devToolsService.FormatJSON(input)
}

// ValidateJSON 验证JSON字符串
func (a *App) ValidateJSON(input string) (service.JSONValidationResult, error) {
	return a.devToolsService.ValidateJSON(input)
}

// MinifyJSON 压缩JSON
func (a *App) MinifyJSON(input string) (string, error) {
	return a.devToolsService.MinifyJSON(input)
}

// EscapeJSON 转义JSON字符串
func (a *App) EscapeJSON(input string) (string, error) {
	return a.devToolsService.EscapeJSON(input)
}

// EncodeBase64 将字符串编码为 Base64
func (a *App) EncodeBase64(input string) (string, error) {
	return a.devToolsService.EncodeBase64(input)
}

// DecodeBase64 将 Base64 字符串解码
func (a *App) DecodeBase64(input string) (string, error) {
	return a.devToolsService.DecodeBase64(input)
}

// CalculateHash 计算字符串的哈希值
func (a *App) CalculateHash(input, algorithm string) (string, error) {
	return a.devToolsService.CalculateHash(input, algorithm)
}

// EncryptText 对文本进行加密，返回 Base64 密文
func (a *App) EncryptText(input, algorithm, keyHex, ivHex string) (string, error) {
	return a.devToolsService.EncryptText(input, algorithm, keyHex, ivHex)
}

// DecryptText 对 Base64 密文进行解密
func (a *App) DecryptText(input, algorithm, keyHex, ivHex string) (string, error) {
	return a.devToolsService.DecryptText(input, algorithm, keyHex, ivHex)
}

// TimestampToDateTime 将 Unix 时间戳转换为日期时间字符串
func (a *App) TimestampToDateTime(timestamp int64, format string) (string, error) {
	return a.devToolsService.TimestampToDateTime(timestamp, format)
}

// TimestampToDateTimeMs 将 Unix 毫秒时间戳转换为日期时间字符串
func (a *App) TimestampToDateTimeMs(timestampMs int64, format string) (string, error) {
	return a.devToolsService.TimestampToDateTimeMs(timestampMs, format)
}

// DateTimeToTimestamp 将日期时间字符串转换为 Unix 时间戳
func (a *App) DateTimeToTimestamp(datetime, format string) (int64, error) {
	return a.devToolsService.DateTimeToTimestamp(datetime, format)
}

// DateTimeToTimestampMs 将日期时间字符串转换为 Unix 毫秒时间戳
func (a *App) DateTimeToTimestampMs(datetime, format string) (int64, error) {
	return a.devToolsService.DateTimeToTimestampMs(datetime, format)
}

// GetCurrentTimestamp 获取当前 Unix 时间戳
func (a *App) GetCurrentTimestamp() int64 {
	return a.devToolsService.GetCurrentTimestamp()
}

// GetCurrentTimestampMs 获取当前 Unix 毫秒时间戳
func (a *App) GetCurrentTimestampMs() int64 {
	return a.devToolsService.GetCurrentTimestampMs()
}

// GenerateUUIDv4 生成 UUID v4
func (a *App) GenerateUUIDv4() (string, error) {
	return a.devToolsService.GenerateUUIDv4()
}

// URLEncode 对字符串进行 URL 编码
func (a *App) URLEncode(input, mode string) (service.URLEncodeResult, error) {
	return a.devToolsService.URLEncode(input, mode)
}

// URLDecode 对 URL 编码的字符串进行解码
func (a *App) URLDecode(input, mode string) (service.URLDecodeResult, error) {
	return a.devToolsService.URLDecode(input, mode)
}

// ParseURL 解析 URL 返回各个组成部分
func (a *App) ParseURL(input string) (map[string]interface{}, error) {
	return a.devToolsService.ParseURL(input)
}

// ShowAboutDialog 显示关于对话框
func (a *App) ShowAboutDialog() {
	runtime.EventsEmit(a.ctx, "app:show-about")
}

// ============================================================================
// Connection Export/Import Methods
// ============================================================================

const (
	passphrasePrefix = "penc:"
	passphraseKDF    = "argon2id"
)

// PasswordEncryption describes how passwords are encrypted in export data
type PasswordEncryption struct {
	Mode string `json:"mode"`
	Salt string `json:"salt"`
	KDF  string `json:"kdf"`
}

// ExportData represents exported connection data
type ExportData struct {
	Version            string                    `json:"version"`
	ExportedAt         string                    `json:"exported_at"`
	Connections        []config.ConnectionConfig `json:"connections"`
	Passwords          map[string]string         `json:"passwords,omitempty"`
	PasswordEncryption *PasswordEncryption       `json:"password_encryption,omitempty"`
}

// ExportConnections exports all connections with passwords to JSON
func (a *App) ExportConnections(encryptPasswords bool) (string, error) {
	conns, err := a.connectionService.GetConnections()
	if err != nil {
		return "", err
	}

	exportData := ExportData{
		Version:     "1.0",
		ExportedAt:  time.Now().UTC().Format(time.RFC3339),
		Connections: conns,
		Passwords:   make(map[string]string),
	}

	for _, conn := range conns {
		if encryptPasswords {
			password, err := a.connectionService.GetEncryptedPassword(conn.ID)
			if err == nil && password != "" {
				exportData.Passwords[conn.ID] = "enc:" + password
			}
			continue
		}

		password, err := a.connectionService.GetPassword(conn.ID)
		if err == nil && password != "" {
			exportData.Passwords[conn.ID] = password
		}
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export data: %w", err)
	}

	return string(data), nil
}

// ExportConnectionsByIDs exports selected connections to JSON
func (a *App) ExportConnectionsByIDs(connectionIDs []string, encryptPasswords bool) (string, error) {
	if len(connectionIDs) == 0 {
		return "", fmt.Errorf("no connections selected")
	}

	conns, err := a.connectionService.GetConnections()
	if err != nil {
		return "", err
	}

	selected := make(map[string]struct{}, len(connectionIDs))
	for _, id := range connectionIDs {
		if id != "" {
			selected[id] = struct{}{}
		}
	}

	filtered := make([]config.ConnectionConfig, 0, len(selected))
	for _, conn := range conns {
		if _, ok := selected[conn.ID]; ok {
			filtered = append(filtered, conn)
		}
	}

	exportData := ExportData{
		Version:     "1.0",
		ExportedAt:  time.Now().UTC().Format(time.RFC3339),
		Connections: filtered,
		Passwords:   make(map[string]string),
	}

	for _, conn := range filtered {
		if encryptPasswords {
			password, err := a.connectionService.GetEncryptedPassword(conn.ID)
			if err == nil && password != "" {
				exportData.Passwords[conn.ID] = "enc:" + password
			}
			continue
		}

		password, err := a.connectionService.GetPassword(conn.ID)
		if err == nil && password != "" {
			exportData.Passwords[conn.ID] = password
		}
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export data: %w", err)
	}

	return string(data), nil
}

// ExportConnectionsByIDsWithPassphrase exports selected connections using passphrase encryption
func (a *App) ExportConnectionsByIDsWithPassphrase(connectionIDs []string, passphrase string) (string, error) {
	if len(connectionIDs) == 0 {
		return "", fmt.Errorf("no connections selected")
	}
	if strings.TrimSpace(passphrase) == "" {
		return "", fmt.Errorf("passphrase required")
	}

	conns, err := a.connectionService.GetConnections()
	if err != nil {
		return "", err
	}

	selected := make(map[string]struct{}, len(connectionIDs))
	for _, id := range connectionIDs {
		if id != "" {
			selected[id] = struct{}{}
		}
	}

	filtered := make([]config.ConnectionConfig, 0, len(selected))
	for _, conn := range conns {
		if _, ok := selected[conn.ID]; ok {
			filtered = append(filtered, conn)
		}
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	key := derivePassphraseKey(passphrase, salt)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	exportData := ExportData{
		Version:     "1.0",
		ExportedAt:  time.Now().UTC().Format(time.RFC3339),
		Connections: filtered,
		Passwords:   make(map[string]string),
		PasswordEncryption: &PasswordEncryption{
			Mode: "passphrase",
			Salt: encodedSalt,
			KDF:  passphraseKDF,
		},
	}

	for _, conn := range filtered {
		password, err := a.connectionService.GetPassword(conn.ID)
		if err != nil || password == "" {
			continue
		}
		ciphertext, err := encryptWithKey(password, key)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt password: %w", err)
		}
		exportData.Passwords[conn.ID] = passphrasePrefix + ciphertext
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export data: %w", err)
	}

	return string(data), nil
}

// ImportConnections imports connections from JSON data
func (a *App) ImportConnections(jsonData string) (int, error) {
	var exportData ExportData

	if err := json.Unmarshal([]byte(jsonData), &exportData); err != nil {
		return 0, fmt.Errorf("failed to parse import data: %w", err)
	}

	if exportData.PasswordEncryption != nil && exportData.PasswordEncryption.Mode == "passphrase" {
		return 0, fmt.Errorf("passphrase required")
	}

	importedCount := 0

	for _, conn := range exportData.Connections {
		originalID := conn.ID
		existingConns, _ := a.connectionService.GetConnections()
		idExists := false
		for _, existing := range existingConns {
			if existing.ID == conn.ID {
				idExists = true
				break
			}
		}

		if idExists {
			conn.ID = generateNewID()
		}

		if err := a.connectionService.AddConnection(conn); err != nil {
			fmt.Printf("Failed to add connection %s: %v\n", conn.ID, err)
			continue
		}

		if password, hasPassword := exportData.Passwords[originalID]; hasPassword {
			if strings.HasPrefix(password, "enc:") {
				encrypted := strings.TrimPrefix(password, "enc:")
				if err := a.connectionService.StoreEncryptedPassword(conn.ID, encrypted); err != nil {
					fmt.Printf("Failed to store encrypted password for %s: %v\n", conn.ID, err)
				}
				continue
			}

			if isEncryptedPassword(password) {
				if err := a.connectionService.StoreEncryptedPassword(conn.ID, password); err != nil {
					fmt.Printf("Failed to store encrypted password for %s: %v\n", conn.ID, err)
				}
				continue
			}

			if err := a.connectionService.SavePassword(conn.ID, password); err != nil {
				fmt.Printf("Failed to save password for %s: %v\n", conn.ID, err)
			}
		}

		importedCount++
	}

	return importedCount, nil
}

// ImportConnectionsWithPassphrase imports passphrase-encrypted connections
func (a *App) ImportConnectionsWithPassphrase(jsonData, passphrase string) (int, error) {
	if strings.TrimSpace(passphrase) == "" {
		return 0, fmt.Errorf("passphrase required")
	}

	var exportData ExportData
	if err := json.Unmarshal([]byte(jsonData), &exportData); err != nil {
		return 0, fmt.Errorf("failed to parse import data: %w", err)
	}

	if exportData.PasswordEncryption == nil || exportData.PasswordEncryption.Mode != "passphrase" {
		return a.ImportConnections(jsonData)
	}

	salt, err := base64.StdEncoding.DecodeString(exportData.PasswordEncryption.Salt)
	if err != nil {
		return 0, fmt.Errorf("invalid passphrase salt")
	}

	key := derivePassphraseKey(passphrase, salt)

	importedCount := 0
	for _, conn := range exportData.Connections {
		originalID := conn.ID
		existingConns, _ := a.connectionService.GetConnections()
		idExists := false
		for _, existing := range existingConns {
			if existing.ID == conn.ID {
				idExists = true
				break
			}
		}

		if idExists {
			conn.ID = generateNewID()
		}

		if err := a.connectionService.AddConnection(conn); err != nil {
			fmt.Printf("Failed to add connection %s: %v\n", conn.ID, err)
			continue
		}

		if password, hasPassword := exportData.Passwords[originalID]; hasPassword {
			if strings.HasPrefix(password, passphrasePrefix) {
				ciphertext := strings.TrimPrefix(password, passphrasePrefix)
				plaintext, err := decryptWithKey(ciphertext, key)
				if err != nil {
					return 0, fmt.Errorf("invalid passphrase")
				}
				if err := a.connectionService.SavePassword(conn.ID, plaintext); err != nil {
					fmt.Printf("Failed to save password for %s: %v\n", conn.ID, err)
				}
			} else if err := a.connectionService.SavePassword(conn.ID, password); err != nil {
				fmt.Printf("Failed to save password for %s: %v\n", conn.ID, err)
			}
		}

		importedCount++
	}

	return importedCount, nil
}

// ImportConnectionsFromFile imports connections from a JSON file
func (a *App) ImportConnectionsFromFile(filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read import file: %w", err)
	}

	return a.ImportConnections(string(data))
}

// ImportConnectionsFromFileWithPassphrase imports passphrase-encrypted connections from file
func (a *App) ImportConnectionsFromFileWithPassphrase(filePath, passphrase string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read import file: %w", err)
	}

	return a.ImportConnectionsWithPassphrase(string(data), passphrase)
}

// SelectImportFile opens a file picker for selecting connection import file
func (a *App) SelectImportFile() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择导入文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON 文件 (*.json)",
				Pattern:     "*.json",
			},
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// isEncryptedPassword checks if a password is in AES-GCM encrypted format
func isEncryptedPassword(password string) bool {
	decoded, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return false
	}

	return len(decoded) >= 29
}

func generateNewID() string {
	return fmt.Sprintf("conn-%d", time.Now().UnixNano())
}

func derivePassphraseKey(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, 3, 64*1024, 4, 32)
}

func encryptWithKey(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptWithKey(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
