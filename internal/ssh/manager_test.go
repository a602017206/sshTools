package ssh

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestInitialLocalWorkingDirectory(t *testing.T) {
	want, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

	if got := initialLocalWorkingDirectory(); got != want {
		t.Fatalf("initial cwd = %q, want %q", got, want)
	}
}

func TestExecuteCommandRejectsLocalSession(t *testing.T) {
	sm := NewSessionManager()
	sm.sessions["local-agent"] = &ManagedSession{
		ID: "local-agent", Type: SessionTypeLocal, Running: true, stopChan: make(chan struct{}),
	}

	_, _, err := sm.ExecuteCommand("local-agent", "pwd", time.Second)
	if err == nil || !strings.Contains(err.Error(), "仅支持 SSH 会话") {
		t.Fatalf("ExecuteCommand() error = %v, want local-session rejection", err)
	}
}

func TestExecuteCommandAllowsSSHSessionWhenTypeUnset(t *testing.T) {
	managed := &ManagedSession{
		ID:      "remote",
		Running: true,
		Session: &Session{},
	}

	got, err := remoteCommandSession(managed)
	if err != nil {
		t.Fatalf("untyped SSH session was rejected: %v", err)
	}
	if got != managed.Session {
		t.Fatal("expected the SSH session to be usable for remote commands")
	}
}

func TestCreateSessionMarksRemoteSessionsAsSSH(t *testing.T) {
	got := newManagedSSHSession("session-1", nil, &Session{})
	if got.Type != SessionTypeSSH {
		t.Fatalf("Type = %q, want %q", got.Type, SessionTypeSSH)
	}
}

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

func TestSplitShellCommandsPreservesQuotedAndEscapedSeparators(t *testing.T) {
	input := `cd '/srv;logs'; cd /tmp\;cache && cd "~/work|draft" || pwd`
	want := []string{"cd '/srv;logs'", " cd /tmp;cache ", " cd \"~/work|draft\" ", " pwd"}
	if got := splitShellCommands(input); !equalStringSlices(got, want) {
		t.Fatalf("splitShellCommands(%q) = %#v, want %#v", input, got, want)
	}
}

func TestSplitShellTokensPreservesQuotedAndEscapedValues(t *testing.T) {
	input := `cd -- "/srv/with space" escaped\ value 'single quoted'`
	want := []string{"cd", "--", "/srv/with space", "escaped value", "single quoted"}
	if got := splitShellTokens(input); !equalStringSlices(got, want) {
		t.Fatalf("splitShellTokens(%q) = %#v, want %#v", input, got, want)
	}
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
