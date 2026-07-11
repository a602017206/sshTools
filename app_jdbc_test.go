package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"AHaSSHTools/internal/service"
)

func TestBuildJDBCServicesInjectsManagedGateway(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".sshtools")
	bundle, err := buildJDBCServices(root, []byte("embedded-agent-jar"), jdbcServiceDependencies{})
	if err != nil {
		t.Fatalf("build JDBC services failed: %v", err)
	}
	if bundle.gateway == nil || bundle.supervisor == nil {
		t.Fatalf("expected managed gateway and supervisor")
	}
	installedAgent := filepath.Join(root, "agent", "jdbc-agent.jar")
	content, err := os.ReadFile(installedAgent)
	if err != nil {
		t.Fatalf("agent artifact not installed: %v", err)
	}
	if string(content) != "embedded-agent-jar" {
		t.Fatalf("unexpected installed agent: %q", content)
	}
}

func TestBuildJDBCServicesRestoresPersistedRuntimeMode(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".sshtools")
	javaPath := filepath.Join(root, "jdk", "bin", "java")
	if err := os.MkdirAll(filepath.Dir(javaPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(javaPath, []byte("java"), 0o700); err != nil {
		t.Fatal(err)
	}
	starter := &serviceStartCountingAgentStarter{}
	bundle, err := buildJDBCServices(root, []byte("agent"), jdbcServiceDependencies{
		systemJavaPath: javaPath,
		runtimeMode:    "system",
		runtimePath:    javaPath,
		starter:        starter,
	})
	if err != nil {
		t.Fatalf("build services failed: %v", err)
	}
	selected, err := bundle.runtime.SelectRuntime()
	if err != nil {
		t.Fatal(err)
	}
	if selected.Kind != service.RuntimeKindSystem || selected.JavaPath != javaPath {
		t.Fatalf("runtime not restored: %+v", selected)
	}
	if starter.startCalls != 0 {
		t.Fatalf("startup restore should not start agent: %d", starter.startCalls)
	}
}

func TestBuildJDBCServicesKeepsInvalidPersistedSystemRuntimeMissing(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".sshtools")
	bundle, err := buildJDBCServices(root, []byte("agent"), jdbcServiceDependencies{
		systemJavaPath: "/usr/bin/java",
		runtimeMode:    "system",
		runtimePath:    filepath.Join(root, "missing-java"),
	})
	if err == nil {
		t.Fatal("expected invalid persisted system Java error")
	}
	selected, selectErr := bundle.runtime.SelectRuntime()
	if selectErr != nil {
		t.Fatal(selectErr)
	}
	if selected.Kind != service.RuntimeKindMissing {
		t.Fatalf("invalid persisted runtime silently fell back: %+v", selected)
	}
}

type serviceStartCountingAgentStarter struct {
	startCalls int
}

func (s *serviceStartCountingAgentStarter) StartAgent(context.Context, service.AgentProcessConfig) (*service.AgentProcessHandle, error) {
	s.startCalls++
	return &service.AgentProcessHandle{}, nil
}

func (s *serviceStartCountingAgentStarter) Stop() error { return nil }

func TestJDBCManagementAPIReturnsAgentAndRuntimeState(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".sshtools")
	paths := service.NewJDBCPaths(root)
	javaPath := filepath.Join(root, "system-java")
	if err := os.MkdirAll(filepath.Dir(javaPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(javaPath, []byte("java"), 0o700); err != nil {
		t.Fatal(err)
	}
	runtimeService := service.NewRuntimeService(paths, javaPath)
	runtimeService.UseSystemJava(true)
	starter := newBlockingAgentStarter()
	supervisor := service.NewJDBCAgentSupervisor(runtimeService, starter, failingAgentDialer{}, "/agent/jdbc-agent.jar")
	dialogs := &fakeJDBCFileDialogs{
		runtimeArchive: "/tmp/jre.zip",
		driverPackage:  "/tmp/driver.zip",
		javaExecutable: "/tmp/java",
	}
	app := &App{jdbcPaths: paths, jdbcRuntime: runtimeService, jdbcAgentSupervisor: supervisor, jdbcFileDialogs: dialogs}

	initial, err := app.GetJDBCAgentStatus()
	if err != nil {
		t.Fatalf("get initial status failed: %v", err)
	}
	if initial.State != service.JDBCAgentStateStopped || initial.RuntimeKind != service.RuntimeKindSystem || initial.LastError != "" {
		t.Fatalf("unexpected initial status: %+v", initial)
	}

	done := make(chan error, 1)
	go func() {
		_, err := supervisor.Client(context.Background())
		done <- err
	}()
	select {
	case <-starter.started:
	case <-time.After(time.Second):
		t.Fatal("agent did not enter starting state")
	}
	starting, err := app.GetJDBCAgentStatus()
	if err != nil {
		t.Fatalf("get starting status failed: %v", err)
	}
	if starting.State != service.JDBCAgentStateStarting {
		t.Fatalf("expected starting state, got %+v", starting)
	}
	close(starter.release)
	if err := <-done; err == nil {
		t.Fatal("expected dial failure")
	}
	failed, err := app.GetJDBCAgentStatus()
	if err != nil {
		t.Fatalf("get failed status failed: %v", err)
	}
	if failed.State != service.JDBCAgentStateFailed || failed.LastError == "" {
		t.Fatalf("expected failed status with last error, got %+v", failed)
	}

	if path, _ := app.SelectJDBCRuntimeArchive(); path != dialogs.runtimeArchive {
		t.Fatalf("unexpected runtime archive: %q", path)
	}
	if path, _ := app.SelectJDBCDriverPackage(); path != dialogs.driverPackage {
		t.Fatalf("unexpected driver package: %q", path)
	}
	if path, _ := app.SelectJDBCJavaExecutable(); path != dialogs.javaExecutable {
		t.Fatalf("unexpected java executable: %q", path)
	}
}

type blockingAgentStarter struct {
	started chan struct{}
	release chan struct{}
}

func newBlockingAgentStarter() *blockingAgentStarter {
	return &blockingAgentStarter{started: make(chan struct{}), release: make(chan struct{})}
}

func (s *blockingAgentStarter) StartAgent(context.Context, service.AgentProcessConfig) (*service.AgentProcessHandle, error) {
	close(s.started)
	<-s.release
	return &service.AgentProcessHandle{Port: 47001, Token: "token"}, nil
}

func (s *blockingAgentStarter) Stop() error { return nil }

type failingAgentDialer struct{}

func (failingAgentDialer) Dial(context.Context, string, int) (service.JdbcAgentClient, func() error, error) {
	return nil, nil, os.ErrNotExist
}

type fakeJDBCFileDialogs struct {
	runtimeArchive string
	driverPackage  string
	javaExecutable string
}

func (d *fakeJDBCFileDialogs) SelectRuntimeArchive(context.Context) (string, error) {
	return d.runtimeArchive, nil
}

func (d *fakeJDBCFileDialogs) SelectDriverPackage(context.Context) (string, error) {
	return d.driverPackage, nil
}

func (d *fakeJDBCFileDialogs) SelectJavaExecutable(context.Context) (string, error) {
	return d.javaExecutable, nil
}
