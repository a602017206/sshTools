package copilot

import "strings"

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
