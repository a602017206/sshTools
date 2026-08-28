package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultArtifactMaxBytes int64 = 1 << 30

type ArtifactDownloadOptions struct {
	Client    *http.Client
	AllowHTTP bool
	MaxBytes  int64
}

type ArtifactDownloader struct {
	client    *http.Client
	allowHTTP bool
	maxBytes  int64
}

func NewArtifactDownloader(options ArtifactDownloadOptions) *ArtifactDownloader {
	client := options.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Minute}
	}
	maxBytes := options.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultArtifactMaxBytes
	}
	return &ArtifactDownloader{client: client, allowHTTP: options.AllowHTTP, maxBytes: maxBytes}
}

func (d *ArtifactDownloader) Download(ctx context.Context, sourceURL, expectedSHA256, target string) error {
	parsedURL, err := d.parseSourceURL(sourceURL)
	if err != nil {
		return err
	}
	if !isSHA256(expectedSHA256) {
		return fmt.Errorf("SHA-256 格式无效")
	}
	response, err := d.downloadResponse(ctx, parsedURL)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if err := d.validateResponse(response); err != nil {
		return err
	}
	return d.writeVerifiedArtifact(response.Body, expectedSHA256, target)
}

func (d *ArtifactDownloader) parseSourceURL(sourceURL string) (*url.URL, error) {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("解析下载地址失败: %w", err)
	}
	if !d.isAllowedURL(parsedURL) {
		return nil, fmt.Errorf("下载地址必须使用 HTTPS")
	}
	return parsedURL, nil
}

func isSHA256(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func (d *ArtifactDownloader) downloadResponse(ctx context.Context, sourceURL *url.URL) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建下载请求失败: %w", err)
	}
	request.Header.Set("User-Agent", "AHaSSHTools-JDBC/1")
	response, err := d.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("下载文件失败: %w", err)
	}
	return response, nil
}

func (d *ArtifactDownloader) validateResponse(response *http.Response) error {
	if response.Request == nil || !d.isAllowedURL(response.Request.URL) {
		return fmt.Errorf("下载重定向到不安全地址")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("下载文件失败，HTTP 状态: %s", response.Status)
	}
	if response.ContentLength > d.maxBytes {
		return fmt.Errorf("下载文件超过大小限制")
	}
	return nil
}

func (d *ArtifactDownloader) isAllowedURL(parsedURL *url.URL) bool {
	return parsedURL.Scheme == "https" || (d.allowHTTP && parsedURL.Scheme == "http")
}

func (d *ArtifactDownloader) writeVerifiedArtifact(body io.Reader, expectedSHA256, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return fmt.Errorf("创建下载目录失败: %w", err)
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".artifact-*.tmp")
	if err != nil {
		return fmt.Errorf("创建下载临时文件失败: %w", err)
	}
	temporaryPath := temporary.Name()
	committed := false
	defer func() {
		_ = temporary.Close()
		if !committed {
			_ = os.Remove(temporaryPath)
		}
	}()

	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(temporary, hash), io.LimitReader(body, d.maxBytes+1))
	if err != nil {
		return fmt.Errorf("写入下载文件失败: %w", err)
	}
	if written > d.maxBytes {
		return fmt.Errorf("下载文件超过大小限制")
	}
	if !strings.EqualFold(hex.EncodeToString(hash.Sum(nil)), expectedSHA256) {
		return fmt.Errorf("下载文件 SHA-256 不匹配")
	}
	if err := temporary.Sync(); err != nil {
		return fmt.Errorf("同步下载文件失败: %w", err)
	}
	if err := temporary.Chmod(0o600); err != nil {
		return fmt.Errorf("设置下载文件权限失败: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("关闭下载文件失败: %w", err)
	}
	if err := os.Rename(temporaryPath, target); err != nil {
		return fmt.Errorf("提交下载文件失败: %w", err)
	}
	committed = true
	return nil
}
