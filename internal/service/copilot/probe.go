package copilot

import (
	"path/filepath"
	"strings"
)

var sshProbeAllowlist = []string{
	"uname",
	"pwd",
	"df -h",
	"cat /etc/os-release",
}

// AllowSSHProbe reports whether cmd is an exact allowlisted read-only probe.
func AllowSSHProbe(cmd string) bool {
	cmd = strings.TrimSpace(cmd)
	for _, allowed := range sshProbeAllowlist {
		if cmd == allowed {
			return true
		}
	}
	return false
}

// WorkingDirectoryCommand returns the fixed, read-only directory listing command
// for an absolute working directory. The directory is shell-quoted so the model
// cannot turn the workspace context into extra shell syntax.
func WorkingDirectoryCommand(workingDir string) (string, bool) {
	workingDir = strings.TrimSpace(workingDir)
	if workingDir == "" || !filepath.IsAbs(workingDir) || strings.ContainsRune(workingDir, '\x00') {
		return "", false
	}
	return "cd -- " + shellQuote(workingDir) + " && LC_ALL=C command ls -la --", true
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
