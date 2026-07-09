package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuntimeServicePrefersManagedRuntimeThenSystemRuntime(t *testing.T) {
	root := t.TempDir()
	paths := NewJDBCPaths(filepath.Join(root, ".sshtools"))
	managedJava := filepath.Join(paths.RuntimesDir, "jre-21-test", "bin", "java")
	if err := os.MkdirAll(filepath.Dir(managedJava), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(managedJava, []byte("#!/bin/sh\n"), 0o700); err != nil {
		t.Fatal(err)
	}

	runtimeSvc := NewRuntimeService(paths, "/usr/bin/java")
	selected, err := runtimeSvc.SelectRuntime()
	if err != nil {
		t.Fatalf("select runtime failed: %v", err)
	}
	if selected.Kind != RuntimeKindManaged {
		t.Fatalf("expected managed runtime, got %s", selected.Kind)
	}
}
