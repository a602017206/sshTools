package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/ssh"
	"AHaSSHTools/internal/store"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

// ConnectSSH creates and starts an SSH session
// authType: "password" or "key"
// authValue: password for password auth, or key file path for key auth
// passphrase: passphrase for encrypted keys (optional)
func (a *App) ConnectSSH(sessionID, host string, port int, user, authType, authValue, passphrase string, cols, rows int) error {
	// Use service with Wails-specific output callback
	err := a.sessionService.ConnectSSH(sessionID, host, port, user, authType, authValue, passphrase, cols, rows, func(data []byte) {
		// Emit output to frontend as base64 to preserve binary data for ZMODEM
		// ZMODEM protocol requires raw binary data, not UTF-8 encoded strings
		encoded := base64.StdEncoding.EncodeToString(data)
		runtime.EventsEmit(a.ctx, "ssh:output:"+sessionID, encoded)
	})

	if err == nil {
		fmt.Printf("SSH session started: %s (%s@%s:%d)\n", sessionID, user, host, port)
	}
	return err
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

// UploadFile uploads a single file
func (a *App) UploadFile(sessionID string, localPath string, remotePath string) (string, error) {
	// Use service with Wails-specific progress callback
	return a.sftpService.UploadFile(sessionID, localPath, remotePath, func(progress ssh.TransferProgress) {
		// Emit event to frontend
		runtime.EventsEmit(a.ctx, "sftp:progress:"+progress.TransferID, progress)
	})
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
