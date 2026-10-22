package ssh

import (
	"errors"
	"testing"
	"time"

	cryptossh "golang.org/x/crypto/ssh"
)

func listSessionsWithoutBlocking(t *testing.T, sm *SessionManager, wait time.Duration) []string {
	t.Helper()
	done := make(chan []string, 1)
	go func() { done <- sm.ListSessions() }()
	select {
	case ids := <-done:
		return ids
	case <-time.After(wait):
		t.Fatal("ListSessions blocked; manager lock likely held during I/O")
		return nil
	}
}

func TestCreateSessionDoesNotBlockManagerWhileDialing(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	prevDial := sshDial
	sshDial = func(network, addr string, config *cryptossh.ClientConfig) (*cryptossh.Client, error) {
		close(entered)
		<-release
		return nil, errors.New("dial aborted by test")
	}
	t.Cleanup(func() {
		sshDial = prevDial
		select {
		case <-release:
		default:
			close(release)
		}
	})

	sm := NewSessionManager()
	sm.sessions["existing"] = &ManagedSession{
		ID:       "existing",
		Type:     SessionTypeSSH,
		stopChan: make(chan struct{}),
	}

	createDone := make(chan error, 1)
	go func() {
		_, createErr := sm.CreateSession("hanging", &Config{
			Host:     "127.0.0.1",
			Port:     22,
			User:     "test",
			Password: "test",
			Timeout:  time.Second,
		})
		createDone <- createErr
	}()

	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("CreateSession did not reach dial")
	}

	ids := listSessionsWithoutBlocking(t, sm, 300*time.Millisecond)
	found := false
	for _, id := range ids {
		if id == "existing" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ListSessions = %v, want to include existing", ids)
	}

	close(release)
	select {
	case err := <-createDone:
		if err == nil {
			t.Fatal("expected CreateSession to fail after dial abort")
		}
	case <-time.After(time.Second):
		t.Fatal("CreateSession did not finish after dial release")
	}
}

func TestCloseSessionDoesNotHoldLockDuringResourceClose(t *testing.T) {
	sm := NewSessionManager()

	started := make(chan struct{})
	unblock := make(chan struct{})
	sm.sessions["slow"] = &ManagedSession{
		ID:       "slow",
		Type:     SessionTypeSSH,
		stopChan: make(chan struct{}),
		resourceCloseHook: func() {
			close(started)
			<-unblock
		},
	}

	closeDone := make(chan error, 1)
	go func() {
		closeDone <- sm.CloseSession("slow")
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("CloseSession did not reach resource close hook")
	}

	if ids := listSessionsWithoutBlocking(t, sm, 300*time.Millisecond); len(ids) != 0 {
		close(unblock)
		t.Fatalf("ListSessions = %v, want empty after map removal", ids)
	}

	close(unblock)
	select {
	case err := <-closeDone:
		if err != nil {
			t.Fatalf("CloseSession: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("CloseSession did not finish")
	}
}
