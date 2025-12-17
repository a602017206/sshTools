package service

import (
	"fmt"
	"path/filepath"

	"sshTools/internal/ssh"
)

// ProgressCallback is a callback function for transfer progress
type ProgressCallback func(progress ssh.TransferProgress)

// SFTPService handles SFTP file operations
type SFTPService struct {
	sessionManager  *ssh.SessionManager
	transferManager *ssh.TransferManager
}

// NewSFTPService creates a new SFTP service
func NewSFTPService(sm *ssh.SessionManager, tm *ssh.TransferManager) *SFTPService {
	return &SFTPService{
		sessionManager:  sm,
		transferManager: tm,
	}
}

// ListFiles lists files in a directory
func (s *SFTPService) ListFiles(sessionID string, path string) ([]ssh.FileInfo, error) {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.ListDirectory(path)
}

// ChangeDirectory changes the current working directory for a session
func (s *SFTPService) ChangeDirectory(sessionID string, path string) error {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.ChangeDirectory(path)
}

// GetCurrentPath returns the current working directory
func (s *SFTPService) GetCurrentPath(sessionID string) (string, error) {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.GetCurrentPath(), nil
}

// GetFileInfo gets information about a file
func (s *SFTPService) GetFileInfo(sessionID string, path string) (*ssh.FileInfo, error) {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.GetFileInfo(path)
}

// UploadFile uploads a single file
// Returns transferID for progress tracking
func (s *SFTPService) UploadFile(sessionID string, localPath string, remotePath string, progressCallback ProgressCallback) (string, error) {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}

	// Extract filename from local path and append to remote directory
	localFilename := filepath.Base(localPath)
	remoteFilePath := filepath.ToSlash(filepath.Join(remotePath, localFilename))

	// Create transfer context
	transfer, err := s.transferManager.StartTransfer(sessionID, "upload", []string{localPath})
	if err != nil {
		return "", fmt.Errorf("failed to start transfer: %w", err)
	}

	// Start upload in goroutine
	go func() {
		// Progress callback wrapper
		progressCb := func(progress ssh.TransferProgress) {
			progress.TransferID = transfer.ID
			progress.SessionID = sessionID
			progress.Filename = localFilename

			// Update transfer manager
			s.transferManager.UpdateProgress(transfer.ID, progress)

			// Call external callback
			if progressCallback != nil {
				progressCallback(progress)
			}
		}

		// Perform upload
		err := sftpClient.UploadFile(localPath, remoteFilePath, progressCb)
		if err != nil {
			// Report error
			errorProgress := ssh.TransferProgress{
				TransferID: transfer.ID,
				SessionID:  sessionID,
				Filename:   localFilename,
				Status:     "failed",
				Error:      err.Error(),
			}
			s.transferManager.UpdateProgress(transfer.ID, errorProgress)
			if progressCallback != nil {
				progressCallback(errorProgress)
			}
		}

		// Cleanup after completion
		go func() {
			<-transfer.Context().Done()
			s.transferManager.CleanupTransfer(transfer.ID)
		}()
	}()

	return transfer.ID, nil
}

// UploadFiles uploads multiple files
func (s *SFTPService) UploadFiles(sessionID string, localPaths []string, remotePath string, progressCallback ProgressCallback) ([]string, error) {
	transferIDs := make([]string, 0, len(localPaths))

	for _, localPath := range localPaths {
		transferID, err := s.UploadFile(sessionID, localPath, remotePath, progressCallback)
		if err != nil {
			return transferIDs, err
		}
		transferIDs = append(transferIDs, transferID)
	}

	return transferIDs, nil
}

// DownloadFile downloads a single file
// Returns transferID for progress tracking
func (s *SFTPService) DownloadFile(sessionID string, remotePath string, localPath string, progressCallback ProgressCallback) (string, error) {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}

	// Extract filename from remote path and append to local directory
	remoteFilename := filepath.Base(remotePath)
	localFilePath := filepath.Join(localPath, remoteFilename)

	// Create transfer context
	transfer, err := s.transferManager.StartTransfer(sessionID, "download", []string{remotePath})
	if err != nil {
		return "", fmt.Errorf("failed to start transfer: %w", err)
	}

	// Start download in goroutine
	go func() {
		// Progress callback wrapper
		progressCb := func(progress ssh.TransferProgress) {
			progress.TransferID = transfer.ID
			progress.SessionID = sessionID
			progress.Filename = remoteFilename

			// Update transfer manager
			s.transferManager.UpdateProgress(transfer.ID, progress)

			// Call external callback
			if progressCallback != nil {
				progressCallback(progress)
			}
		}

		// Perform download
		err := sftpClient.DownloadFile(remotePath, localFilePath, progressCb)
		if err != nil {
			// Report error
			errorProgress := ssh.TransferProgress{
				TransferID: transfer.ID,
				SessionID:  sessionID,
				Filename:   remoteFilename,
				Status:     "failed",
				Error:      err.Error(),
			}
			s.transferManager.UpdateProgress(transfer.ID, errorProgress)
			if progressCallback != nil {
				progressCallback(errorProgress)
			}
		}

		// Cleanup after completion
		go func() {
			<-transfer.Context().Done()
			s.transferManager.CleanupTransfer(transfer.ID)
		}()
	}()

	return transfer.ID, nil
}

// DownloadFiles downloads multiple files
func (s *SFTPService) DownloadFiles(sessionID string, remotePaths []string, localPath string, progressCallback ProgressCallback) ([]string, error) {
	transferIDs := make([]string, 0, len(remotePaths))

	for _, remotePath := range remotePaths {
		transferID, err := s.DownloadFile(sessionID, remotePath, localPath, progressCallback)
		if err != nil {
			return transferIDs, err
		}
		transferIDs = append(transferIDs, transferID)
	}

	return transferIDs, nil
}

// DeleteFile deletes a single file or directory
func (s *SFTPService) DeleteFile(sessionID string, path string) error {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
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

// DeleteFiles deletes multiple files or directories
func (s *SFTPService) DeleteFiles(sessionID string, paths []string) error {
	for _, path := range paths {
		if err := s.DeleteFile(sessionID, path); err != nil {
			return err
		}
	}
	return nil
}

// RenameFile renames a file or directory
func (s *SFTPService) RenameFile(sessionID string, oldPath string, newPath string) error {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.RenameFile(oldPath, newPath)
}

// CreateDirectory creates a new directory
func (s *SFTPService) CreateDirectory(sessionID string, path string) error {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.CreateDirectory(path)
}

// CancelTransfer cancels a file transfer
func (s *SFTPService) CancelTransfer(transferID string) error {
	return s.transferManager.CancelTransfer(transferID)
}

// GetTransferStatus gets the status of a transfer
func (s *SFTPService) GetTransferStatus(transferID string) (*ssh.TransferProgress, error) {
	progress, err := s.transferManager.GetProgress(transferID)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}
