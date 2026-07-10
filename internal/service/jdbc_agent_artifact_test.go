package service

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAgentArtifactInstallerWritesJarAtomically(t *testing.T) {
	paths := NewJDBCPaths(filepath.Join(t.TempDir(), ".sshtools"))
	installer := NewAgentArtifactInstaller(paths)
	jar := []byte("test-jdbc-agent-jar")

	installedPath, err := installer.Install(jar)
	if err != nil {
		t.Fatalf("install agent artifact failed: %v", err)
	}
	expectedPath := filepath.Join(paths.AgentDir, "jdbc-agent.jar")
	if installedPath != expectedPath {
		t.Fatalf("expected %s, got %s", expectedPath, installedPath)
	}
	installed, err := os.ReadFile(installedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(installed, jar) {
		t.Fatalf("unexpected artifact content: %q", installed)
	}
	info, err := os.Stat(installedPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o022 != 0 {
		t.Fatalf("artifact must not be group or world writable: %o", info.Mode().Perm())
	}

	fixedTime := time.Unix(1_700_000_000, 0)
	if err := os.Chtimes(installedPath, fixedTime, fixedTime); err != nil {
		t.Fatal(err)
	}
	installedAgain, err := installer.Install(jar)
	if err != nil {
		t.Fatalf("second install failed: %v", err)
	}
	if installedAgain != installedPath {
		t.Fatalf("second install returned a different path: %s", installedAgain)
	}
	info, err = os.Stat(installedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !info.ModTime().Equal(fixedTime) {
		t.Fatalf("identical artifact was rewritten: %s", info.ModTime())
	}
}
