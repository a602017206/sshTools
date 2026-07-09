package service

import (
	"context"
	"strings"
	"testing"
)

func TestAgentProcessManagerStartsAgentWithLocalPortAndToken(t *testing.T) {
	runner := &fakeCommandRunner{}
	manager := NewAgentProcessManager(runner, AgentProcessConfig{JavaPath: "/bin/java", AgentJar: "/tmp/jdbc-agent.jar"})
	handle, err := manager.Start(context.Background())
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if handle.Port == 0 || handle.Token == "" {
		t.Fatalf("expected port and token")
	}
	if !strings.Contains(strings.Join(runner.args, " "), "--token") {
		t.Fatalf("token argument missing")
	}
}

type fakeCommandRunner struct {
	name string
	args []string
}

func (r *fakeCommandRunner) Start(ctx context.Context, name string, args ...string) (AgentProcess, error) {
	r.name = name
	r.args = args
	return fakeAgentProcess{}, nil
}

type fakeAgentProcess struct{}

func (fakeAgentProcess) Stop() error {
	return nil
}
