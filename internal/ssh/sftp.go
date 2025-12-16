package ssh

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/pkg/sftp"
)

// FileInfo represents file metadata for frontend
type FileInfo struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	Mode       string    `json:"mode"`
	ModTime    time.Time `json:"mod_time"`
	IsDir      bool      `json:"is_dir"`
	IsSymlink  bool      `json:"is_symlink"`
	LinkTarget string    `json:"link_target,omitempty"`
}

// TransferProgress represents transfer progress for frontend
type TransferProgress struct {
	TransferID string  `json:"transfer_id"`
	SessionID  string  `json:"session_id"`
	Filename   string  `json:"filename"`
	BytesSent  int64   `json:"bytes_sent"`
	TotalBytes int64   `json:"total_bytes"`
	Percentage float64 `json:"percentage"`
	Speed      int64   `json:"speed"` // bytes per second
	Status     string  `json:"status"`
	Error      string  `json:"error,omitempty"`
}

// SFTPClient wraps pkg/sftp client with metadata
type SFTPClient struct {
	client      *sftp.Client
	sshClient   *Client
	currentPath string
	mu          sync.Mutex
}

// NewSFTPClient creates a new SFTP client from an existing SSH client
func NewSFTPClient(sshClient *Client) (*SFTPClient, error) {
	if sshClient == nil || sshClient.client == nil {
		return nil, fmt.Errorf("invalid SSH client")
	}

	// Create SFTP client from SSH connection
	sftpClient, err := sftp.NewClient(sshClient.client)
	if err != nil {
		return nil, fmt.Errorf("failed to create SFTP client: %w", err)
	}

	// Get home directory as initial path
	homeDir, err := sftpClient.Getwd()
	if err != nil {
		homeDir = "/"
	}

	return &SFTPClient{
		client:      sftpClient,
		sshClient:   sshClient,
		currentPath: homeDir,
	}, nil
}

// Close closes the SFTP client
func (sc *SFTPClient) Close() error {
	if sc.client != nil {
		return sc.client.Close()
	}
	return nil
}

// ListDirectory lists files in a directory
func (sc *SFTPClient) ListDirectory(dirPath string) ([]FileInfo, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Normalize path
	if dirPath == "" {
		dirPath = sc.currentPath
	}
	dirPath = normalizePath(dirPath)

	files, err := sc.client.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	result := make([]FileInfo, 0, len(files))
	for _, file := range files {
		filePath := path.Join(dirPath, file.Name())

		fileInfo := FileInfo{
			Name:      file.Name(),
			Path:      filePath,
			Size:      file.Size(),
			Mode:      file.Mode().String(),
			ModTime:   file.ModTime(),
			IsDir:     file.IsDir(),
			IsSymlink: file.Mode()&os.ModeSymlink != 0,
		}

		// Read symlink target if it's a symlink
		if fileInfo.IsSymlink {
			target, err := sc.client.ReadLink(filePath)
			if err == nil {
				fileInfo.LinkTarget = target
			}
		}

		result = append(result, fileInfo)
	}

	return result, nil
}

// GetFileInfo gets information about a specific file
func (sc *SFTPClient) GetFileInfo(filePath string) (*FileInfo, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	filePath = normalizePath(filePath)

	stat, err := sc.client.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileInfo := &FileInfo{
		Name:      stat.Name(),
		Path:      filePath,
		Size:      stat.Size(),
		Mode:      stat.Mode().String(),
		ModTime:   stat.ModTime(),
		IsDir:     stat.IsDir(),
		IsSymlink: stat.Mode()&os.ModeSymlink != 0,
	}

	// Read symlink target if it's a symlink
	if fileInfo.IsSymlink {
		target, err := sc.client.ReadLink(filePath)
		if err == nil {
			fileInfo.LinkTarget = target
		}
	}

	return fileInfo, nil
}

// UploadFile uploads a file from local to remote with progress tracking
func (sc *SFTPClient) UploadFile(localPath, remotePath string, progressCb func(TransferProgress)) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	remotePath = normalizePath(remotePath)

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get file size
	stat, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}
	totalBytes := stat.Size()

	// Create remote file
	remoteFile, err := sc.client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// Stream file with progress tracking
	return sc.streamWithProgress(localFile, remoteFile, totalBytes, progressCb)
}

// DownloadFile downloads a file from remote to local with progress tracking
func (sc *SFTPClient) DownloadFile(remotePath, localPath string, progressCb func(TransferProgress)) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	remotePath = normalizePath(remotePath)

	// Open remote file
	remoteFile, err := sc.client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("failed to open remote file: %w", err)
	}
	defer remoteFile.Close()

	// Get file size
	stat, err := remoteFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat remote file: %w", err)
	}
	totalBytes := stat.Size()

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	// Stream file with progress tracking
	return sc.streamWithProgress(remoteFile, localFile, totalBytes, progressCb)
}

// streamWithProgress streams data from reader to writer with progress tracking
func (sc *SFTPClient) streamWithProgress(reader io.Reader, writer io.Writer, totalBytes int64, progressCb func(TransferProgress)) error {
	const bufferSize = 64 * 1024 // 64KB buffer
	buffer := make([]byte, bufferSize)

	var bytesTransferred int64
	var lastProgress int64
	var lastTime time.Time = time.Now()
	var speed int64

	// Throttle progress updates: emit every 1MB or 5% (whichever is smaller)
	progressThreshold := min(1024*1024, int64(float64(totalBytes)*0.05))
	if progressThreshold < bufferSize {
		progressThreshold = bufferSize
	}

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			_, writeErr := writer.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("write error: %w", writeErr)
			}

			bytesTransferred += int64(n)

			// Emit progress if threshold reached
			if bytesTransferred-lastProgress >= progressThreshold || bytesTransferred == totalBytes {
				now := time.Now()
				elapsed := now.Sub(lastTime).Seconds()
				if elapsed > 0 {
					speed = int64(float64(bytesTransferred-lastProgress) / elapsed)
				}

				if progressCb != nil {
					percentage := float64(bytesTransferred) / float64(totalBytes) * 100
					progressCb(TransferProgress{
						BytesSent:  bytesTransferred,
						TotalBytes: totalBytes,
						Percentage: percentage,
						Speed:      speed,
						Status:     "running",
					})
				}

				lastProgress = bytesTransferred
				lastTime = now
			}
		}

		if err == io.EOF {
			// Final progress update
			if progressCb != nil {
				progressCb(TransferProgress{
					BytesSent:  bytesTransferred,
					TotalBytes: totalBytes,
					Percentage: 100.0,
					Speed:      speed,
					Status:     "completed",
				})
			}
			break
		}

		if err != nil {
			return fmt.Errorf("read error: %w", err)
		}
	}

	return nil
}

// DeleteFile deletes a file
func (sc *SFTPClient) DeleteFile(filePath string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	filePath = normalizePath(filePath)

	err := sc.client.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// DeleteDirectory deletes a directory recursively
func (sc *SFTPClient) DeleteDirectory(dirPath string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	dirPath = normalizePath(dirPath)

	// Recursive delete
	return sc.deleteRecursive(dirPath)
}

// deleteRecursive recursively deletes a directory
func (sc *SFTPClient) deleteRecursive(dirPath string) error {
	// List directory contents
	files, err := sc.client.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Delete all contents first
	for _, file := range files {
		filePath := path.Join(dirPath, file.Name())

		if file.IsDir() {
			// Recursively delete subdirectory
			if err := sc.deleteRecursive(filePath); err != nil {
				return err
			}
		} else {
			// Delete file
			if err := sc.client.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete file %s: %w", filePath, err)
			}
		}
	}

	// Delete the directory itself
	if err := sc.client.RemoveDirectory(dirPath); err != nil {
		return fmt.Errorf("failed to delete directory %s: %w", dirPath, err)
	}

	return nil
}

// RenameFile renames a file or directory
func (sc *SFTPClient) RenameFile(oldPath, newPath string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	oldPath = normalizePath(oldPath)
	newPath = normalizePath(newPath)

	err := sc.client.Rename(oldPath, newPath)
	if err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	return nil
}

// CreateDirectory creates a directory
func (sc *SFTPClient) CreateDirectory(dirPath string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	dirPath = normalizePath(dirPath)

	err := sc.client.Mkdir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

// ChangeDirectory changes the current working directory
func (sc *SFTPClient) ChangeDirectory(dirPath string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	dirPath = normalizePath(dirPath)

	// Verify directory exists
	stat, err := sc.client.Stat(dirPath)
	if err != nil {
		return fmt.Errorf("directory does not exist: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("not a directory: %s", dirPath)
	}

	sc.currentPath = dirPath
	return nil
}

// GetCurrentPath returns the current working directory
func (sc *SFTPClient) GetCurrentPath() string {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	return sc.currentPath
}

// normalizePath normalizes a file path to Unix-style
func normalizePath(p string) string {
	// Convert backslashes to forward slashes
	p = filepath.ToSlash(p)

	// Clean the path
	p = path.Clean(p)

	// Ensure absolute path
	if !path.IsAbs(p) {
		p = "/" + p
	}

	return p
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
