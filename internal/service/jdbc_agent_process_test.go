package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestAgentProcessManagerHealthDetectsExitedProcess(t *testing.T) {
	process := &fakeAgentProcess{alive: true}
	manager := NewAgentProcessManager(&fakeCommandRunner{process: process}, AgentProcessConfig{JavaPath: "/bin/java", AgentJar: "/tmp/jdbc-agent.jar"})
	if _, err := manager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	process.alive = false
	if err := manager.Health(context.Background()); err == nil {
		t.Fatal("expected exited process health failure")
	}
}

func TestAgentProcessManagerPassesConfiguredLogPath(t *testing.T) {
	runner := &fakeCommandRunner{}
	manager := NewAgentProcessManager(runner, AgentProcessConfig{
		JavaPath: "/bin/java",
		AgentJar: "/tmp/jdbc-agent.jar",
		LogPath:  "/tmp/jdbc-agent.log",
	})
	if _, err := manager.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if runner.logPath != "/tmp/jdbc-agent.log" {
		t.Fatalf("log path not passed to command runner: %q", runner.logPath)
	}
}

func TestExecAgentCommandRunnerWritesLifecycleLog(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "logs", "jdbc-agent.log")
	process, err := (execAgentCommandRunner{}).StartWithOutput(context.Background(), "/usr/bin/true", logPath)
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for process.Alive() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "JDBC agent 启动") {
		t.Fatalf("lifecycle log missing start entry: %q", data)
	}
}

type fakeCommandRunner struct {
	name    string
	args    []string
	process *fakeAgentProcess
	logPath string
}

func (r *fakeCommandRunner) StartWithOutput(ctx context.Context, name, logPath string, args ...string) (AgentProcess, error) {
	r.logPath = logPath
	return r.Start(ctx, name, args...)
}

func (r *fakeCommandRunner) Start(ctx context.Context, name string, args ...string) (AgentProcess, error) {
	r.name = name
	r.args = args
	if r.process == nil {
		r.process = &fakeAgentProcess{alive: true}
	}
	return r.process, nil
}

type fakeAgentProcess struct {
	alive bool
}

func (*fakeAgentProcess) Stop() error {
	return nil
}

func (p *fakeAgentProcess) Alive() bool {
	return p.alive
}
