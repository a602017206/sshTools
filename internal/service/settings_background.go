package service

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxBackgroundImageBytes = 8 << 20 // 8 MiB

// BackgroundImageResult is returned after selecting a wallpaper.
type BackgroundImageResult struct {
	Path    string `json:"path"`
	DataURL string `json:"data_url"`
	Fit     string `json:"fit"`
	Opacity int    `json:"opacity"`
}

func backgroundsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("获取用户目录失败: %w", err)
	}
	dir := filepath.Join(homeDir, ".ahasshtools", "backgrounds")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("创建背景图目录失败: %w", err)
	}
	return dir, nil
}

func normalizeBackgroundFit(fit string) string {
	switch strings.ToLower(strings.TrimSpace(fit)) {
	case "contain":
		return "contain"
	default:
		return "cover"
	}
}

func normalizeBackgroundOpacity(opacity int) int {
	if opacity < 0 {
		return 0
	}
	if opacity > 100 {
		return 100
	}
	return opacity
}

func mimeForBackgroundExt(ext string) (string, error) {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg", nil
	case ".png":
		return "image/png", nil
	case ".webp":
		return "image/webp", nil
	case ".gif":
		return "image/gif", nil
	default:
		return "", fmt.Errorf("不支持的图片格式: %s（请使用 jpg/png/webp/gif）", ext)
	}
}

func encodeBackgroundDataURL(path string) (string, error) {
	ext := filepath.Ext(path)
	mime, err := mimeForBackgroundExt(ext)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("读取背景图失败: %w", err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("背景图文件为空")
	}
	if len(data) > maxBackgroundImageBytes {
		return "", fmt.Errorf("背景图过大（最大 8MB）")
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}

func copyBackgroundImage(srcPath string) (string, error) {
	ext := strings.ToLower(filepath.Ext(srcPath))
	if _, err := mimeForBackgroundExt(ext); err != nil {
		return "", err
	}

	info, err := os.Stat(srcPath)
	if err != nil {
		return "", fmt.Errorf("无法访问所选文件: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("请选择图片文件，而不是文件夹")
	}
	if info.Size() <= 0 {
		return "", fmt.Errorf("背景图文件为空")
	}
	if info.Size() > maxBackgroundImageBytes {
		return "", fmt.Errorf("背景图过大（最大 8MB）")
	}

	dir, err := backgroundsDir()
	if err != nil {
		return "", err
	}
	// Unique name so selecting a new image does not overwrite the saved wallpaper
	// until the user confirms in settings (cancel can keep the previous path).
	dstPath := filepath.Join(dir, fmt.Sprintf("wallpaper_%d%s", time.Now().UnixNano(), ext))

	src, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("打开背景图失败: %w", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", fmt.Errorf("写入背景图失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("复制背景图失败: %w", err)
	}
	return dstPath, nil
}

// InstallBackgroundImage copies a local image into the app data dir.
// Settings are not updated here; the frontend persists path/options on Save.
func (s *SettingsService) InstallBackgroundImage(srcPath string) (*BackgroundImageResult, error) {
	if s.configManager == nil {
		return nil, fmt.Errorf("config manager not initialized")
	}
	srcPath = strings.TrimSpace(srcPath)
	if srcPath == "" {
		return nil, fmt.Errorf("未选择背景图")
	}

	dstPath, err := copyBackgroundImage(srcPath)
	if err != nil {
		return nil, err
	}
	dataURL, err := encodeBackgroundDataURL(dstPath)
	if err != nil {
		_ = os.Remove(dstPath)
		return nil, err
	}

	current := s.configManager.GetSettings()
	fit := normalizeBackgroundFit(current.BackgroundImageFit)
	opacity := normalizeBackgroundOpacity(current.BackgroundImageOpacity)
	if opacity == 0 {
		opacity = 35
	}

	return &BackgroundImageResult{
		Path:    dstPath,
		DataURL: dataURL,
		Fit:     fit,
		Opacity: opacity,
	}, nil
}

// ClearBackgroundImage removes the installed wallpaper and clears settings.
func (s *SettingsService) ClearBackgroundImage() error {
	if s.configManager == nil {
		return fmt.Errorf("config manager not initialized")
	}
	current := s.configManager.GetSettings()
	path := strings.TrimSpace(current.BackgroundImagePath)
	if path != "" {
		_ = os.Remove(path)
	}
	return s.configManager.UpdateSettings(map[string]interface{}{
		"background_image_enabled": false,
		"background_image_path":    "",
	})
}

// GetBackgroundImageDataURL returns a data URL for the configured wallpaper.
func (s *SettingsService) GetBackgroundImageDataURL() (string, error) {
	if s.configManager == nil {
		return "", fmt.Errorf("config manager not initialized")
	}
	settings := s.configManager.GetSettings()
	if !settings.BackgroundImageEnabled {
		return "", nil
	}
	path := strings.TrimSpace(settings.BackgroundImagePath)
	if path == "" {
		return "", nil
	}
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("背景图不存在，请重新选择: %w", err)
	}
	return encodeBackgroundDataURL(path)
}
