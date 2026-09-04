package service

import (
	"net/url"
	"strings"
)

// ReadClipboardFilePaths returns local file/folder paths currently on the OS clipboard.
func ReadClipboardFilePaths() ([]string, error) {
	paths, err := readClipboardFilePaths()
	if err != nil {
		return nil, err
	}
	return NormalizeClipboardFilePaths(paths), nil
}

// NormalizeClipboardFilePaths trims, drops empties, and de-duplicates clipboard paths.
func NormalizeClipboardFilePaths(paths []string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0, len(paths))
	for _, raw := range paths {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	return out
}

// ParseFileURIList parses a text/uri-list clipboard payload into local paths.
func ParseFileURIList(raw string) []string {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	paths := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if path, ok := fileURLToPath(line); ok {
			paths = append(paths, path)
		}
	}
	return paths
}

func fileURLToPath(raw string) (string, bool) {
	if !strings.HasPrefix(strings.ToLower(raw), "file:") {
		return "", false
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", false
	}
	path := parsed.Path
	if parsed.Opaque != "" && path == "" {
		path = parsed.Opaque
	}
	if path == "" {
		return "", false
	}
	if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
		path = strings.TrimPrefix(path, "/")
	}
	return path, true
}

func splitClipboardLines(raw string) []string {
	raw = strings.ReplaceAll(raw, "\r\n", "\n")
	raw = strings.ReplaceAll(raw, "\x00", "\n")
	return strings.Split(raw, "\n")
}
