package ssh

import "testing"

func TestCloseExitedSessionCleansUpAndNotifiesOnce(t *testing.T) {
	sm := NewSessionManager()
	sessionID := "session-exited"
	sm.sessions[sessionID] = &ManagedSession{
		ID:       sessionID,
		Type:     SessionTypeSSH,
		Running:  true,
		stopChan: make(chan struct{}),
	}

	shouldNotify, err := sm.closeSession(sessionID, false)
	if err != nil {
		t.Fatalf("closeSession returned error: %v", err)
	}
	if !shouldNotify {
		t.Fatal("expected remote exit cleanup to notify caller")
	}

	if _, err := sm.GetSession(sessionID); err == nil {
		t.Fatal("expected exited session to be removed")
	}

	shouldNotify, err = sm.closeSession(sessionID, false)
	if err != nil {
		t.Fatalf("second closeSession returned error: %v", err)
	}
	if shouldNotify {
		t.Fatal("expected repeated cleanup to skip notification")
	}
}

func TestClientRequestedCloseDoesNotNotifyRemoteExit(t *testing.T) {
	sm := NewSessionManager()
	sessionID := "session-client-close"
	sm.sessions[sessionID] = &ManagedSession{
		ID:       sessionID,
		Type:     SessionTypeSSH,
		Running:  true,
		stopChan: make(chan struct{}),
	}

	shouldNotify, err := sm.closeSession(sessionID, true)
	if err != nil {
		t.Fatalf("closeSession returned error: %v", err)
	}
	if shouldNotify {
		t.Fatal("expected client-requested close to skip remote-exit notification")
	}
}
