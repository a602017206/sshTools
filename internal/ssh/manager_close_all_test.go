package ssh

import "testing"

func TestSessionManagerCloseAllSessions(t *testing.T) {
	sm := NewSessionManager()
	if _, err := sm.CreateLocalSession("local-a", "zsh"); err != nil {
		t.Skipf("local shell unavailable in this environment: %v", err)
	}
	if _, err := sm.CreateLocalSession("local-b", "zsh"); err != nil {
		_ = sm.CloseSession("local-a")
		t.Skipf("local shell unavailable in this environment: %v", err)
	}

	if err := sm.CloseAllSessions(); err != nil {
		t.Fatalf("CloseAllSessions: %v", err)
	}
	if got := sm.ListSessions(); len(got) != 0 {
		t.Fatalf("expected no sessions, got %v", got)
	}
}
