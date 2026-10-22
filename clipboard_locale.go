package main

import (
	"os"
	"runtime"
	"strings"
)

// Wails 在 macOS 上用 pbcopy/pbpaste 读写剪贴板。从 Finder/Dock 启动时进程没有 LANG，
// pbcopy 遇到中文等非 ASCII 文本会写入空剪贴板，pbpaste 会读出乱码。
func ensureUTF8Locale() {
	if runtime.GOOS != "darwin" {
		return
	}
	if key, value, ok := utf8LocaleOverride(os.Getenv); ok {
		_ = os.Setenv(key, value)
	}
}

func utf8LocaleOverride(getenv func(string) string) (string, string, bool) {
	effective := getenv("LC_ALL")
	if effective == "" {
		effective = getenv("LC_CTYPE")
	}
	if effective == "" {
		effective = getenv("LANG")
	}
	normalized := strings.ToLower(strings.ReplaceAll(effective, "-", ""))
	if strings.Contains(normalized, "utf8") {
		return "", "", false
	}
	if getenv("LC_ALL") != "" {
		return "LC_ALL", "en_US.UTF-8", true
	}
	return "LC_CTYPE", "UTF-8", true
}
