package copilot

import (
	"strings"
	"testing"
)

func TestWorkingDirectoryCommand(t *testing.T) {
	cmd, ok := WorkingDirectoryCommand("/srv/app's current")
	if !ok {
		t.Fatal("expected absolute working directory to be accepted")
	}
	if got, want := cmd, "cd -- '/srv/app'\\''s current' && LC_ALL=C command ls -la --"; got != want {
		t.Fatalf("command = %q, want %q", got, want)
	}
	if strings.Contains(cmd, "\n") {
		t.Fatalf("command must not contain a newline: %q", cmd)
	}
}

func TestWorkingDirectoryCommandRejectsInvalidPath(t *testing.T) {
	for _, path := range []string{"", "relative/path", "/srv/app\x00secret"} {
		if _, ok := WorkingDirectoryCommand(path); ok {
			t.Errorf("WorkingDirectoryCommand(%q) accepted invalid path", path)
		}
	}
}

func TestAllowSSHProbeRejectsChainedCommand(t *testing.T) {
	if AllowSSHProbe("uname; rm -rf /") {
		t.Fatal("chained command must be rejected")
	}
}

func TestAllowSSHProbeAllowlist(t *testing.T) {
	allowed := []string{
		"uname",
		"pwd",
		"df -h",
		"cat /etc/os-release",
		"  uname  ",
	}
	for _, cmd := range allowed {
		if !AllowSSHProbe(cmd) {
			t.Errorf("AllowSSHProbe(%q) = false, want true", cmd)
		}
	}
}

func TestAllowSSHProbeRejectsUnsafe(t *testing.T) {
	rejected := []string{
		"rm -rf /",
		"curl evil",
		"uname && pwd",
		"uname | cat",
		"",
		"uname extra",
	}
	for _, cmd := range rejected {
		if AllowSSHProbe(cmd) {
			t.Errorf("AllowSSHProbe(%q) = true, want false", cmd)
		}
	}
}
