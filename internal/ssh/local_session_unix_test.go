//go:build !windows

package ssh

import (
	"testing"
	"time"
)

func TestLocalShellStaysAvailableAfterStartup(t *testing.T) {
	manager := NewSessionManager()
	closed := make(chan error, 1)

	if _, err := manager.CreateLocalSession("local-startup", ""); err != nil {
		t.Fatalf("create local session: %v", err)
	}
	t.Cleanup(func() {
		_ = manager.CloseSession("local-startup")
	})
	if err := manager.StartLocalShell("local-startup", 80, 24, nil, func(err error) {
		closed <- err
	}); err != nil {
		t.Fatalf("start local shell: %v", err)
	}

	select {
	case err := <-closed:
		t.Fatalf("local shell closed immediately: %v", err)
	case <-time.After(300 * time.Millisecond):
	}

	if _, err := manager.GetSession("local-startup"); err != nil {
		t.Fatalf("local shell was removed after startup: %v", err)
	}
}
