package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"time"
	"unicode/utf8"

	"github.com/pkg/sftp"
)

// MaxRemoteTextBytes is the largest file the online editor will read or write.
// Keep it aligned with frontend MAX_REMOTE_TEXT_BYTES.
const MaxRemoteTextBytes int64 = 1 << 20

// RemoteTextFile is a UTF-8 text file loaded for the online editor.
type RemoteTextFile struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Size    int64  `json:"size"`
}

// ValidateRemoteText accepts empty and UTF-8 text up to MaxRemoteTextBytes.
func ValidateRemoteText(content []byte) (string, error) {
	if int64(len(content)) > MaxRemoteTextBytes {
		return "", remoteTextTooLargeError()
	}
	if bytes.IndexByte(content, 0) >= 0 {
		return "", fmt.Errorf("文件包含二进制内容，无法在线编辑")
	}
	if !utf8.Valid(content) {
		return "", fmt.Errorf("文件不是 UTF-8 文本，无法在线编辑")
	}
	return string(content), nil
}

func remoteTextTooLargeError() error {
	return fmt.Errorf("文件超过 1MB，请下载后本地编辑")
}

// ReadTextFile loads a remote UTF-8 text file for the editor.
func (sc *SFTPClient) ReadTextFile(filePath string) (*RemoteTextFile, error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	filePath = normalizePath(filePath)
	info, err := sc.client.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件信息失败: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("不能编辑目录")
	}
	if info.Size() > MaxRemoteTextBytes {
		return nil, remoteTextTooLargeError()
	}

	file, err := sc.client.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, MaxRemoteTextBytes+1))
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}
	text, err := ValidateRemoteText(data)
	if err != nil {
		return nil, err
	}
	return &RemoteTextFile{
		Path:    filePath,
		Name:    path.Base(filePath),
		Content: text,
		Size:    int64(len(data)),
	}, nil
}

// WriteTextFile writes UTF-8 text back to an existing remote file.
// Regular files are replaced with posix-rename when the server supports it.
// Symlinks are written through, so the link itself is kept.
func (sc *SFTPClient) WriteTextFile(filePath, content string) error {
	data := []byte(content)
	if _, err := ValidateRemoteText(data); err != nil {
		return err
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	filePath = normalizePath(filePath)
	info, err := sc.client.Lstat(filePath)
	if err != nil {
		return fmt.Errorf("读取文件信息失败: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("不能编辑目录")
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return writeRemoteTextInPlace(sc.client, filePath, data)
	}
	return replaceRemoteTextFile(sc.client, filePath, data, info.Mode().Perm())
}

func replaceRemoteTextFile(client *sftp.Client, filePath string, data []byte, mode os.FileMode) error {
	tmp := path.Join(path.Dir(filePath), fmt.Sprintf(".%s.sshtools-%d.tmp", path.Base(filePath), time.Now().UnixNano()))
	file, err := client.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	writeErr := writeAll(file, data)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		_ = client.Remove(tmp)
		if writeErr != nil {
			return fmt.Errorf("写入文件失败: %w", writeErr)
		}
		return fmt.Errorf("关闭临时文件失败: %w", closeErr)
	}
	_ = client.Chmod(tmp, mode)
	if err := client.PosixRename(tmp, filePath); err != nil {
		if writeErr := writeRemoteTextInPlace(client, filePath, data); writeErr != nil {
			return fmt.Errorf("写回文件失败，临时文件保留在 %s: %w", tmp, writeErr)
		}
		_ = client.Remove(tmp)
	}
	return nil
}

func writeRemoteTextInPlace(client *sftp.Client, filePath string, data []byte) error {
	file, err := client.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()
	if err := writeAll(file, data); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

func writeAll(writer io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := writer.Write(data)
		if n > 0 {
			data = data[n:]
		}
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}
