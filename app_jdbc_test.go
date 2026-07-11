package main

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/service/jdbcproto"
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

func TestSetJDBCRuntimeModePersistsAndRestartsAgent(t *testing.T) {
	app, settings, starter := newRuntimeActivationTestApp(t, activationAgentDialer{})
	javaPath := writeTestJava(t, filepath.Join(t.TempDir(), "bin", "java"))

	result, err := app.SetJDBCRuntimeMode("system", javaPath)
	if err != nil {
		t.Fatal(err)
	}
	if settings.mode != "system" || settings.path != javaPath || settings.calls != 1 {
		t.Fatalf("settings not persisted: %+v", settings)
	}
	if starter.startCalls != 1 {
		t.Fatalf("expected one agent start, got %d", starter.startCalls)
	}
	if result.Runtime.Kind != service.RuntimeKindSystem || result.Agent.State != service.JDBCAgentStateRunning {
		t.Fatalf("unexpected activation result: %+v", result)
	}
}

func TestSetJDBCRuntimeModeRollsBackWhenPersistenceFails(t *testing.T) {
	app, settings, starter := newRuntimeActivationTestApp(t, activationAgentDialer{})
	settings.err = errors.New("save failed")
	before := app.jdbcRuntime.Snapshot()
	javaPath := writeTestJava(t, filepath.Join(t.TempDir(), "bin", "java"))

	if _, err := app.SetJDBCRuntimeMode("system", javaPath); err == nil {
		t.Fatal("expected persistence error")
	}
	if got := app.jdbcRuntime.Snapshot(); got != before {
		t.Fatalf("runtime not rolled back: got %+v want %+v", got, before)
	}
	if starter.startCalls != 0 {
		t.Fatalf("agent started after persistence failure: %d", starter.startCalls)
	}
}

func TestSetJDBCRuntimeModeKeepsSelectionWhenRestartFails(t *testing.T) {
	app, settings, _ := newRuntimeActivationTestApp(t, failingAgentDialer{})
	javaPath := writeTestJava(t, filepath.Join(t.TempDir(), "bin", "java"))

	result, err := app.SetJDBCRuntimeMode("system", javaPath)
	if err == nil {
		t.Fatal("expected restart error")
	}
	if settings.mode != "system" || app.jdbcRuntime.Snapshot().Mode != "system" {
		t.Fatalf("new selection was rolled back: settings=%+v snapshot=%+v", settings, app.jdbcRuntime.Snapshot())
	}
	if result.Agent.State != service.JDBCAgentStateFailed || result.Agent.LastError == "" {
		t.Fatalf("expected failed agent status: %+v", result)
	}
}

func TestManagedRuntimeInstallAndImportActivateManagedMode(t *testing.T) {
	app, settings, starter := newRuntimeActivationTestApp(t, activationAgentDialer{})
	archivePath := createAppTestRuntimeArchive(t)
	provider := &appTestRuntimeProvider{pkg: service.ManagedRuntimePackage{
		Version: "21.0.8", Name: "runtime.zip", URL: "https://example.invalid/runtime.zip", SHA256: "checksum",
	}}
	app.jdbcRuntime.ConfigureManagedInstaller(provider, &appTestArtifactFetcher{source: archivePath})

	if _, err := app.InstallJDBCManagedRuntime(); err != nil {
		t.Fatalf("managed install activation failed: %v", err)
	}
	if settings.mode != "managed" || starter.startCalls != 1 {
		t.Fatalf("managed install not activated: settings=%+v starts=%d", settings, starter.startCalls)
	}
	if err := app.jdbcAgentSupervisor.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := app.ImportJDBCRuntimeArchive(archivePath); err != nil {
		t.Fatalf("runtime import activation failed: %v", err)
	}
	if settings.calls != 2 || starter.startCalls != 2 {
		t.Fatalf("runtime import not activated: saves=%d starts=%d", settings.calls, starter.startCalls)
	}
}

type fakeJDBCRuntimeSettingsStore struct {
	mode  string
	path  string
	calls int
	err   error
}

func (s *fakeJDBCRuntimeSettingsStore) UpdateJDBCRuntimeSettings(mode, path string) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	s.mode = mode
	s.path = path
	return nil
}

type activationAgentStarter struct {
	startCalls int
	stopCalls  int
}

func (s *activationAgentStarter) StartAgent(context.Context, service.AgentProcessConfig) (*service.AgentProcessHandle, error) {
	s.startCalls++
	return &service.AgentProcessHandle{Port: 47000 + s.startCalls, Token: "token"}, nil
}

func (s *activationAgentStarter) Stop() error {
	s.stopCalls++
	return nil
}

type activationAgentDialer struct{}

func (activationAgentDialer) Dial(context.Context, string, int) (service.JdbcAgentClient, func() error, error) {
	return activationJdbcClient{}, func() error { return nil }, nil
}

type activationJdbcClient struct{}

func (activationJdbcClient) OpenSession(context.Context, *jdbcproto.OpenSessionRequest) (*jdbcproto.OpenSessionResponse, error) {
	return &jdbcproto.OpenSessionResponse{}, nil
}
func (activationJdbcClient) ExecuteQuery(context.Context, *jdbcproto.ExecuteQueryRequest) (*jdbcproto.QueryResult, error) {
	return &jdbcproto.QueryResult{}, nil
}
func (activationJdbcClient) ListTables(context.Context, *jdbcproto.ListTablesRequest) (*jdbcproto.ListTablesResponse, error) {
	return &jdbcproto.ListTablesResponse{}, nil
}
func (activationJdbcClient) ListColumns(context.Context, *jdbcproto.ListColumnsRequest) (*jdbcproto.ListColumnsResponse, error) {
	return &jdbcproto.ListColumnsResponse{}, nil
}
func (activationJdbcClient) CloseSession(context.Context, *jdbcproto.CloseSessionRequest) (*jdbcproto.CloseSessionResponse, error) {
	return &jdbcproto.CloseSessionResponse{}, nil
}

func newRuntimeActivationTestApp(t *testing.T, dialer service.JDBCAgentDialer) (*App, *fakeJDBCRuntimeSettingsStore, *activationAgentStarter) {
	t.Helper()
	paths := service.NewJDBCPaths(filepath.Join(t.TempDir(), ".sshtools"))
	runtimeService := service.NewRuntimeService(paths, "")
	starter := &activationAgentStarter{}
	supervisor := service.NewJDBCAgentSupervisor(runtimeService, starter, dialer, "/agent/jdbc-agent.jar")
	settings := &fakeJDBCRuntimeSettingsStore{}
	return &App{
		jdbcPaths: paths, jdbcRuntime: runtimeService, jdbcAgentSupervisor: supervisor, jdbcRuntimeSettings: settings,
	}, settings, starter
}

func writeTestJava(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("java"), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

type appTestRuntimeProvider struct{ pkg service.ManagedRuntimePackage }

func (p *appTestRuntimeProvider) Latest(context.Context, int) (service.ManagedRuntimePackage, error) {
	return p.pkg, nil
}

type appTestArtifactFetcher struct{ source string }

func (f *appTestArtifactFetcher) Download(_ context.Context, _, _ string, target string) error {
	source, err := os.Open(f.source)
	if err != nil {
		return err
	}
	defer source.Close()
	destination, err := os.Create(target)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func createAppTestRuntimeArchive(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "runtime.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	header := &zip.FileHeader{Name: "jdk-21.0.8/bin/java", Method: zip.Deflate}
	header.SetMode(0o755)
	part, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("java")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
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
