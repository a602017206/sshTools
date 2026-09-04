package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"

	"AHaSSHTools/internal/ssh"
)

// ProgressCallback is a callback function for transfer progress
type ProgressCallback func(progress ssh.TransferProgress)

// SFTPService handles SFTP file operations
type SFTPService struct {
	sessionManager  *ssh.SessionManager
	transferManager *ssh.TransferManager
	uploadGates     sync.Map
}

// NewSFTPService creates a new SFTP service
func NewSFTPService(sm *ssh.SessionManager, tm *ssh.TransferManager) *SFTPService {
	return &SFTPService{
		sessionManager:  sm,
		transferManager: tm,
	}
}

func (s *SFTPService) lockSessionUpload(sessionID string) func() {
	value, _ := s.uploadGates.LoadOrStore(sessionID, &sync.Mutex{})
	mu := value.(*sync.Mutex)
	mu.Lock()
	return mu.Unlock
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

// GetCurrentPath returns current working directory from the SSH session
// This queries the actual terminal session to get the real current directory (not SFTP's cached path)
func (s *SFTPService) GetCurrentPath(sessionID string) (string, error) {
	return s.sessionManager.GetCurrentWorkingDirectory(sessionID)
}

// UpdateCurrentPath updates the tracked current directory for a session
func (s *SFTPService) UpdateCurrentPath(sessionID, path string) error {
	return s.sessionManager.UpdateCurrentWorkingDirectory(sessionID, path)
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

	localFilename := filepath.Base(localPath)
	remoteFilePath := joinRemoteUploadPath(remotePath, localFilename)
	return s.startFileUpload(sessionID, sftpClient, localPath, remoteFilePath, localFilename, progressCallback)
}

func (s *SFTPService) startFileUpload(sessionID string, sftpClient *ssh.SFTPClient, localPath, remoteFilePath, displayName string, progressCallback ProgressCallback) (string, error) {
	transfer, err := s.transferManager.StartTransfer(sessionID, "upload", []string{localPath})
	if err != nil {
		return "", fmt.Errorf("failed to start transfer: %w", err)
	}

	go func() {
		unlock := s.lockSessionUpload(sessionID)
		defer unlock()
		progressCb := func(progress ssh.TransferProgress) {
			progress.TransferID = transfer.ID
			progress.SessionID = sessionID
			progress.Filename = displayName

			s.transferManager.UpdateProgress(transfer.ID, progress)

			if progressCallback != nil {
				progressCallback(progress)
			}
		}

		err := sftpClient.UploadFile(localPath, remoteFilePath, progressCb)
		if err != nil {
			errorProgress := ssh.TransferProgress{
				TransferID: transfer.ID,
				SessionID:  sessionID,
				Filename:   displayName,
				Status:     "failed",
				Error:      err.Error(),
			}
			s.transferManager.UpdateProgress(transfer.ID, errorProgress)
			if progressCallback != nil {
				progressCallback(errorProgress)
			}
		}

		s.transferManager.CleanupTransfer(transfer.ID)
	}()

	return transfer.ID, nil
}

// UploadFiles uploads files and directories. Directories are expanded into a
// relative tree under remotePath (the folder name is preserved).
func (s *SFTPService) UploadFiles(sessionID string, localPaths []string, remotePath string, progressCallback ProgressCallback) ([]string, error) {
	items, err := ExpandLocalUploadPaths(localPaths)
	if err != nil {
		return nil, err
	}
	return s.UploadItems(sessionID, remotePath, items, progressCallback)
}

// UploadItems uploads an already expanded local tree, using each item's RelPath.
// Directories and files share one transfer and run one at a time on the existing
// SFTP channel so a large folder cannot fan out thousands of goroutines or stall sshd.
func (s *SFTPService) UploadItems(sessionID, remotePath string, items []LocalUploadItem, progressCallback ProgressCallback) ([]string, error) {
	if len(items) == 0 {
		return []string{}, nil
	}

	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SFTP client: %w", err)
	}

	_, files := partitionLocalUploadItems(items)
	paths := make([]string, 0, len(files))
	for _, file := range files {
		paths = append(paths, file.LocalPath)
	}
	transfer, err := s.transferManager.StartTransfer(sessionID, "upload", paths)
	if err != nil {
		return nil, fmt.Errorf("failed to start transfer: %w", err)
	}

	go s.runItemUpload(sessionID, remotePath, items, sftpClient, transfer, progressCallback)
	return []string{transfer.ID}, nil
}

func (s *SFTPService) runItemUpload(
	sessionID, remotePath string,
	items []LocalUploadItem,
	sftpClient *ssh.SFTPClient,
	transfer *ssh.TransferContext,
	progressCallback ProgressCallback,
) {
	defer s.transferManager.CleanupTransfer(transfer.ID)
	unlock := s.lockSessionUpload(sessionID)
	defer unlock()

	_, files := partitionLocalUploadItems(items)
	total := len(files)
	done := 0
	emit := func(filename, status, errMsg string) {
		percentage := 100.0
		if total > 0 {
			percentage = float64(done) / float64(total) * 100
		}
		progress := ssh.TransferProgress{
			TransferID: transfer.ID,
			SessionID:  sessionID,
			Filename:   filename,
			Percentage: percentage,
			Status:     status,
			Error:      errMsg,
		}
		s.transferManager.UpdateProgress(transfer.ID, progress)
		if progressCallback != nil {
			progressCallback(progress)
		}
	}

	err := runFolderUpload(folderUploadRunner{
		mkdir: sftpClient.EnsureDirectory,
		upload: func(localPath, remotePath string) error {
			rel := remotePath
			if total > 0 {
				rel = fmt.Sprintf("%s (%d/%d)", filepath.ToSlash(filepath.Base(remotePath)), done+1, total)
			}
			emit(rel, "running", "")
			if err := sftpClient.UploadFile(localPath, remotePath, nil); err != nil {
				return err
			}
			done++
			emit(rel, "running", "")
			return nil
		},
		cancelled: transfer.IsCancelled,
	}, remotePath, items)

	switch {
	case errors.Is(err, errFolderUploadCancelled):
		emit("已取消", "cancelled", err.Error())
	case err != nil:
		emit("上传失败", "failed", err.Error())
	default:
		emit("上传完成", "completed", "")
	}
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

		s.transferManager.CleanupTransfer(transfer.ID)
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

// CreateFile creates an empty remote file.
func (s *SFTPService) CreateFile(sessionID string, path string) error {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.CreateFile(path)
}

// CopyFile copies a remote file to another remote path.
func (s *SFTPService) CopyFile(sessionID string, srcPath string, dstPath string) error {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.CopyFile(srcPath, dstPath)
}

// ChmodFile updates remote file permissions.
func (s *SFTPService) ChmodFile(sessionID string, path string, mode string) error {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.ChmodFile(path, mode)
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

// SearchDirectories searches for directories matching the query recursively
func (s *SFTPService) SearchDirectories(sessionID string, searchPath string, query string, maxDepth int, maxResults int) ([]ssh.SearchResult, error) {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SFTP client: %w", err)
	}

	return sftpClient.SearchDirectories(searchPath, query, maxDepth, maxResults)
}
