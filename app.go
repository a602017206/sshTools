package main

import (
	"context"
	"fmt"
	"path/filepath"

	"sshTools/internal/config"
	"sshTools/internal/ssh"
	"sshTools/internal/store"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx              context.Context
	configManager    *config.ConfigManager
	credentialStore  *store.CredentialStore
	sessionManager   *ssh.SessionManager
	monitorCollector *ssh.MonitorCollector
	transferManager  *ssh.TransferManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		sessionManager:  ssh.NewSessionManager(),
		transferManager: ssh.NewTransferManager(),
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

// UpdateConnection updates an existing SSH connection
func (a *App) UpdateConnection(conn config.ConnectionConfig) error {
	if a.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return a.configManager.UpdateConnection(conn)
}

// RemoveConnection removes an SSH connection
func (a *App) RemoveConnection(id string) error {
	if a.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return a.configManager.RemoveConnection(id)
}

// TestConnection tests an SSH connection
// authType: "password" or "key"
// authValue: password for password auth, or key file path for key auth
// passphrase: passphrase for encrypted keys (optional)
func (a *App) TestConnection(host string, port int, user, authType, authValue, passphrase string) error {
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

// ConnectSSH creates and starts an SSH session
// authType: "password" or "key"
// authValue: password for password auth, or key file path for key auth
// passphrase: passphrase for encrypted keys (optional)
func (a *App) ConnectSSH(sessionID, host string, port int, user, authType, authValue, passphrase string, cols, rows int) error {
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
	_, err := a.sessionManager.CreateSession(sessionID, sshConfig)
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
	if a.credentialStore == nil {
		return fmt.Errorf("credential store not initialized")
	}
	return a.credentialStore.Store(connectionID, password)
}

// GetPassword retrieves a saved password for a connection
func (a *App) GetPassword(connectionID string) (string, error) {
	if a.credentialStore == nil {
		return "", fmt.Errorf("credential store not initialized")
	}
	return a.credentialStore.Get(connectionID)
}

// HasPassword checks if a password is saved for a connection
func (a *App) HasPassword(connectionID string) bool {
	if a.credentialStore == nil {
		return false
	}
	return a.credentialStore.Has(connectionID)
}

// DeletePassword removes a saved password for a connection
func (a *App) DeletePassword(connectionID string) error {
	if a.credentialStore == nil {
		return fmt.Errorf("credential store not initialized")
	}
	return a.credentialStore.Delete(connectionID)
}

// GetSettings returns application settings
func (a *App) GetSettings() config.AppSettings {
	if a.configManager == nil {
		return config.DefaultSettings()
	}
	return a.configManager.GetSettings()
}

// UpdateSettings updates application settings
func (a *App) UpdateSettings(updates map[string]interface{}) error {
	if a.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	return a.configManager.UpdateSettings(updates)
}

// GetMonitoringData retrieves monitoring data for a session
func (a *App) GetMonitoringData(sessionID string) (*ssh.MonitoringData, error) {
	if a.monitorCollector == nil {
		a.monitorCollector = ssh.NewMonitorCollector(a.sessionManager)
	}
	return a.monitorCollector.CollectMetrics(sessionID)
}

// ListFiles lists files in a directory
func (a *App) ListFiles(sessionID string, path string) ([]ssh.FileInfo, error) {
	sftpClient, err := a.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.ListDirectory(path)
}

// ChangeDirectory changes the current working directory for a session
func (a *App) ChangeDirectory(sessionID string, path string) error {
	sftpClient, err := a.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.ChangeDirectory(path)
}

// GetCurrentPath returns the current working directory
func (a *App) GetCurrentPath(sessionID string) (string, error) {
	sftpClient, err := a.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.GetCurrentPath(), nil
}

// UploadFile uploads a single file
func (a *App) UploadFile(sessionID string, localPath string, remotePath string) (string, error) {
	sftpClient, err := a.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}

	// Extract filename from local path and append to remote directory
	localFilename := filepath.Base(localPath)
	remoteFilePath := filepath.ToSlash(filepath.Join(remotePath, localFilename))

	// Create transfer context
	transfer, err := a.transferManager.StartTransfer(sessionID, "upload", []string{localPath})
	if err != nil {
		return "", fmt.Errorf("failed to start transfer: %w", err)
	}

	// Start upload in goroutine
	go func() {
		// Progress callback
		progressCb := func(progress ssh.TransferProgress) {
			progress.TransferID = transfer.ID
			progress.SessionID = sessionID
			progress.Filename = localFilename

			// Update transfer manager
			a.transferManager.UpdateProgress(transfer.ID, progress)

			// Emit event to frontend
			runtime.EventsEmit(a.ctx, "sftp:progress:"+transfer.ID, progress)
		}

		// Perform upload
		err := sftpClient.UploadFile(localPath, remoteFilePath, progressCb)
		if err != nil {
			// Emit error
			errorProgress := ssh.TransferProgress{
				TransferID: transfer.ID,
				SessionID:  sessionID,
				Filename:   localFilename,
				Status:     "failed",
				Error:      err.Error(),
			}
			a.transferManager.UpdateProgress(transfer.ID, errorProgress)
			runtime.EventsEmit(a.ctx, "sftp:progress:"+transfer.ID, errorProgress)
		}

		// Cleanup after some time
		go func() {
			<-transfer.Context().Done()
			a.transferManager.CleanupTransfer(transfer.ID)
		}()
	}()

	return transfer.ID, nil
}

// UploadFiles uploads multiple files
func (a *App) UploadFiles(sessionID string, localPaths []string, remotePath string) ([]string, error) {
	transferIDs := make([]string, 0, len(localPaths))

	for _, localPath := range localPaths {
		transferID, err := a.UploadFile(sessionID, localPath, remotePath)
		if err != nil {
			return transferIDs, err
		}
		transferIDs = append(transferIDs, transferID)
	}

	return transferIDs, nil
}

// DownloadFile downloads a single file
func (a *App) DownloadFile(sessionID string, remotePath string, localPath string) (string, error) {
	sftpClient, err := a.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}

	// Extract filename from remote path and append to local directory
	remoteFilename := filepath.Base(remotePath)
	localFilePath := filepath.Join(localPath, remoteFilename)

	// Create transfer context
	transfer, err := a.transferManager.StartTransfer(sessionID, "download", []string{remotePath})
	if err != nil {
		return "", fmt.Errorf("failed to start transfer: %w", err)
	}

	// Start download in goroutine
	go func() {
		// Progress callback
		progressCb := func(progress ssh.TransferProgress) {
			progress.TransferID = transfer.ID
			progress.SessionID = sessionID
			progress.Filename = remoteFilename

			// Update transfer manager
			a.transferManager.UpdateProgress(transfer.ID, progress)

			// Emit event to frontend
			runtime.EventsEmit(a.ctx, "sftp:progress:"+transfer.ID, progress)
		}

		// Perform download
		err := sftpClient.DownloadFile(remotePath, localFilePath, progressCb)
		if err != nil {
			// Emit error
			errorProgress := ssh.TransferProgress{
				TransferID: transfer.ID,
				SessionID:  sessionID,
				Filename:   remoteFilename,
				Status:     "failed",
				Error:      err.Error(),
			}
			a.transferManager.UpdateProgress(transfer.ID, errorProgress)
			runtime.EventsEmit(a.ctx, "sftp:progress:"+transfer.ID, errorProgress)
		}

		// Cleanup after some time
		go func() {
			<-transfer.Context().Done()
			a.transferManager.CleanupTransfer(transfer.ID)
		}()
	}()

	return transfer.ID, nil
}

// DownloadFiles downloads multiple files
func (a *App) DownloadFiles(sessionID string, remotePaths []string, localPath string) ([]string, error) {
	transferIDs := make([]string, 0, len(remotePaths))

	for _, remotePath := range remotePaths {
		transferID, err := a.DownloadFile(sessionID, remotePath, localPath)
		if err != nil {
			return transferIDs, err
		}
		transferIDs = append(transferIDs, transferID)
	}

	return transferIDs, nil
}

// DeleteFile deletes a single file
func (a *App) DeleteFile(sessionID string, path string) error {
	sftpClient, err := a.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	// Check if it's a directory
	fileInfo, err := sftpClient.GetFileInfo(path)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	if fileInfo.IsDir {
		return sftpClient.DeleteDirectory(path)
	}

	return sftpClient.DeleteFile(path)
}

// DeleteFiles deletes multiple files
func (a *App) DeleteFiles(sessionID string, paths []string) error {
	for _, path := range paths {
		if err := a.DeleteFile(sessionID, path); err != nil {
			return err
		}
	}
	return nil
}

// RenameFile renames a file or directory
func (a *App) RenameFile(sessionID string, oldPath string, newPath string) error {
	sftpClient, err := a.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.RenameFile(oldPath, newPath)
}

// CreateDirectory creates a new directory
func (a *App) CreateDirectory(sessionID string, path string) error {
	sftpClient, err := a.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.CreateDirectory(path)
}

// GetFileInfo gets information about a file
func (a *App) GetFileInfo(sessionID string, path string) (*ssh.FileInfo, error) {
	sftpClient, err := a.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.GetFileInfo(path)
}

// CancelTransfer cancels a file transfer
func (a *App) CancelTransfer(transferID string) error {
	return a.transferManager.CancelTransfer(transferID)
}

// GetTransferStatus gets the status of a transfer
func (a *App) GetTransferStatus(transferID string) (*ssh.TransferProgress, error) {
	progress, err := a.transferManager.GetProgress(transferID)
	if err != nil {
		return nil, err
	}
	return &progress, nil
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
