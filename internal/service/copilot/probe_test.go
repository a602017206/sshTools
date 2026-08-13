package copilot

import "testing"

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
