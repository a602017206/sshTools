//go:build linux

package service

import (
	"context"
	"os/exec"
	"time"
)

func readClipboardFilePaths() ([]string, error) {
	raw, err := linuxClipboardURIList()
	if err != nil || raw == "" {
		return nil, nil
	}
	return ParseFileURIList(raw), nil
}

func linuxClipboardURIList() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	commands := [][]string{
		{"wl-paste", "-t", "text/uri-list", "-n"},
		{"xclip", "-selection", "clipboard", "-t", "text/uri-list", "-o"},
	}
	for _, args := range commands {
		out, err := exec.CommandContext(ctx, args[0], args[1:]...).Output()
		if err == nil && len(out) > 0 {
			return string(out), nil
		}
	}
	return "", nil
}
