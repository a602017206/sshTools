package service

import (
	"fmt"
	"path/filepath"
)

// 本文件提供同步版本的单文件上传/下载，供移动端 API 在请求内完成传输
// 并直接返回结果；桌面端原有的异步 + 进度回调模型（UploadFile 等）不受影响。

// UploadFileSync 把本地单个文件同步上传到远端目录，返回远端文件路径。
func (s *SFTPService) UploadFileSync(sessionID, localPath, remoteDir string) (string, error) {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}

	remotePath := joinRemoteUploadPath(remoteDir, filepath.Base(localPath))
	if err := sftpClient.UploadFile(localPath, remotePath, nil); err != nil {
		return "", err
	}
	return remotePath, nil
}

// DownloadFileSync 把远端单个文件同步下载到本地目录，返回本地文件路径。
func (s *SFTPService) DownloadFileSync(sessionID, remotePath, localDir string) (string, error) {
	sftpClient, err := s.sessionManager.GetOrCreateSFTPClient(sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get SFTP client: %w", err)
	}

	localPath := filepath.Join(localDir, filepath.Base(remotePath))
	if err := sftpClient.DownloadFile(remotePath, localPath, nil); err != nil {
		return "", err
	}
	return localPath, nil
}
