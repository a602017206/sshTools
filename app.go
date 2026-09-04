package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/service/copilot"
	"AHaSSHTools/internal/ssh"
	"AHaSSHTools/internal/store"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/crypto/argon2"
)

var Version = "dev"

var cwdRegex = regexp.MustCompile(`\033\]0;CWD:([^\007]+)\007`)

var sessionLogAppendWarnOnce sync.Once

const (
	sftpProgressEventPrefix               = "sftp:progress:"
	jdbcAgentSupervisorUnavailableMessage = "JDBC agent supervisor 未初始化"
)

// App struct
type App struct {
	ctx context.Context

	// Services
	connectionService     *service.ConnectionService
	sessionService        *service.SessionService
	sftpService           *service.SFTPService
	monitorService        *service.MonitorService
	settingsService       *service.SettingsService
	devToolsService       *service.DevToolsService
	databaseService       *service.DatabaseService
	nativeDatabaseService *service.NativeDatabaseService
	jdbcPaths             service.JDBCPaths
	jdbcCatalog           *service.DriverCatalogService
	jdbcInstaller         *service.DriverInstallService
	jdbcRuntime           *service.RuntimeService
	jdbcAgentSupervisor   *service.JDBCAgentSupervisor
	jdbcGateway           *service.ManagedJDBCGateway
	jdbcFileDialogs       jdbcFileDialogs
	jdbcRuntimeSettings   jdbcRuntimeSettingsStore
	configManager         *config.ConfigManager
	credentialStore       *store.CredentialStore
	copilotService        *copilot.Service
	copilotMetaMu         sync.Mutex
	copilotMeta           map[string]copilotSessionMeta
	sqlFileMu             sync.Mutex
	sqlFileCancel         map[string]context.CancelFunc
	sessionCharsetMu      sync.Mutex
	sessionCharset        map[string]string
	sessionLogService     *service.SessionLogService
	commandHistoryService *service.CommandHistoryService
	sessionConnectionMu   sync.Mutex
	sessionConnection     map[string]string
}

type copilotSessionMeta struct {
	host string
	user string
}

type jdbcRuntimeSettingsStore interface {
	UpdateJDBCRuntimeSettings(mode, javaPath string) error
}

type jdbcFileDialogs interface {
	SelectRuntimeArchive(context.Context) (string, error)
	SelectDriverPackage(context.Context) (string, error)
	SelectJavaExecutable(context.Context) (string, error)
}

type wailsJDBCFileDialogs struct{}

func (wailsJDBCFileDialogs) SelectRuntimeArchive(ctx context.Context) (string, error) {
	return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title:   "选择 JDBC JRE 归档",
		Filters: []runtime.FileFilter{{DisplayName: "JRE 归档 (*.zip;*.tar.gz;*.tgz)", Pattern: "*.zip;*.tar.gz;*.tgz"}},
	})
}

func (wailsJDBCFileDialogs) SelectDriverPackage(ctx context.Context) (string, error) {
	return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title:   "选择 JDBC 离线驱动包",
		Filters: []runtime.FileFilter{{DisplayName: "JDBC 驱动包 (*.zip)", Pattern: "*.zip"}},
	})
}

func (wailsJDBCFileDialogs) SelectJavaExecutable(ctx context.Context) (string, error) {
	return runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{Title: "选择 Java 可执行文件"})
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		jdbcFileDialogs:   wailsJDBCFileDialogs{},
		copilotMeta:       make(map[string]copilotSessionMeta),
		sessionConnection: make(map[string]string),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize configuration manager
	configManager, err := config.NewConfigManager()
	if err != nil {
		fmt.Printf("Failed to initialize config manager: %v\n", err)
		configManager = config.NewFallbackConfigManager()
	}
	a.configManager = configManager
	a.jdbcRuntimeSettings = configManager

	// Initialize credential store
	credentialStore := store.NewCredentialStore()
	a.credentialStore = credentialStore

	// Initialize managers
	sessionManager := ssh.NewSessionManager()
	transferManager := ssh.NewTransferManager()

	// Initialize services
	a.connectionService = service.NewConnectionService(configManager, credentialStore)
	a.sessionService = service.NewSessionService(sessionManager)
	a.sftpService = service.NewSFTPService(sessionManager, transferManager)
	a.monitorService = service.NewMonitorService(sessionManager)
	a.settingsService = service.NewSettingsService(configManager)
	a.devToolsService = service.NewDevToolsService()

	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		homeDir = "."
	}
	a.sessionLogService = service.NewSessionLogService(filepath.Join(homeDir, ".ahasshtools", "session-logs"))
	a.commandHistoryService = service.NewCommandHistoryService(filepath.Join(homeDir, ".ahasshtools", "commands"))
	if a.sessionConnection == nil {
		a.sessionConnection = make(map[string]string)
	}

	agentJar, readAgentErr := assets.ReadFile("frontend/build/jdbc-agent.jar")
	if readAgentErr != nil {
		fmt.Printf("Failed to read embedded JDBC agent: %v\n", readAgentErr)
	}
	if err := a.initJDBCServices(agentJar); err != nil {
		fmt.Printf("Failed to initialize JDBC services: %v\n", err)
	}
	a.databaseService = service.NewDatabaseServiceWithGateway(a.configManager, a.jdbcGateway)
	a.nativeDatabaseService = service.NewNativeDatabaseService(map[service.NativeDatabaseType]service.NativeDatabaseProvider{
		service.NativeDatabaseTypeRedis:         service.NewDefaultRedisNativeProvider(),
		service.NativeDatabaseTypeMongoDB:       service.NewDefaultMongoNativeProvider(),
		service.NativeDatabaseTypeElasticsearch: service.NewDefaultElasticsearchNativeProvider(),
		service.NativeDatabaseTypeMemcached:     service.NewDefaultMemcachedNativeProvider(),
		service.NativeDatabaseTypeCassandra:     service.NewDefaultCassandraNativeProvider(),
		service.NativeDatabaseTypeCouchbase:     service.NewDefaultCouchbaseNativeProvider(),
		service.NativeDatabaseTypeInfluxDB:      service.NewDefaultInfluxDBNativeProvider(),
		service.NativeDatabaseTypeNeo4j:         service.NewDefaultNeo4jNativeProvider(),
		service.NativeDatabaseTypeKafka:         service.NewDefaultKafkaNativeProvider(),
		service.NativeDatabaseTypeRocketMQ:      service.NewDefaultRocketMQNativeProvider(),
		service.NativeDatabaseTypeRabbitMQ:      service.NewDefaultRabbitMQNativeProvider(),
	})
	if a.copilotMeta == nil {
		a.copilotMeta = make(map[string]copilotSessionMeta)
	}
	a.copilotService = copilot.NewService(nil, a.databaseService, a.sessionService).WithNative(nativeCopilotReader{svc: a.nativeDatabaseService})

	if _, purgeErr := a.PurgeExpiredSessionLogs(); purgeErr != nil {
		fmt.Printf("Failed to purge expired session logs: %v\n", purgeErr)
	}
	go a.runSessionLogPurgeTicker()
}

func (a *App) initJDBCServices(agentJar []byte) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	settings := config.DefaultSettings()
	if a.configManager != nil {
		settings = a.configManager.GetSettings()
	}
	bundle, buildErr := buildJDBCServices(filepath.Join(homeDir, ".sshtools"), agentJar, jdbcServiceDependencies{
		runtimeMode: settings.JDBCRuntimeMode,
		runtimePath: settings.JDBCSystemJavaPath,
	})
	a.jdbcPaths = bundle.paths
	a.jdbcCatalog = bundle.catalog
	a.jdbcInstaller = bundle.installer
	a.jdbcRuntime = bundle.runtime
	a.jdbcAgentSupervisor = bundle.supervisor
	a.jdbcGateway = bundle.gateway
	return buildErr
}

type jdbcServiceDependencies struct {
	systemJavaPath string
	runtimeMode    string
	runtimePath    string
	starter        service.JDBCAgentStarter
	dialer         service.JDBCAgentDialer
}

type jdbcServiceBundle struct {
	paths      service.JDBCPaths
	catalog    *service.DriverCatalogService
	installer  *service.DriverInstallService
	runtime    *service.RuntimeService
	supervisor *service.JDBCAgentSupervisor
	gateway    *service.ManagedJDBCGateway
}

func buildJDBCServices(root string, agentJar []byte, deps jdbcServiceDependencies) (*jdbcServiceBundle, error) {
	paths := service.NewJDBCPaths(root)
	agentPath := filepath.Join(paths.AgentDir, "jdbc-agent.jar")
	var artifactErr error
	if len(agentJar) == 0 {
		artifactErr = fmt.Errorf("嵌入的 JDBC agent jar 为空")
	} else {
		agentPath, artifactErr = service.NewAgentArtifactInstaller(paths).Install(agentJar)
	}

	catalog := service.NewDriverCatalogService(paths.Manifest, paths.DriversDir)
	installer := service.NewDriverInstallService(paths)
	systemJavaPath := deps.systemJavaPath
	if systemJavaPath == "" {
		systemJavaPath = "/usr/bin/java"
	}
	runtimeService := service.NewRuntimeService(paths, systemJavaPath)
	runtimeService.ConfigureManagedInstaller(
		service.NewAdoptiumRuntimeProvider(nil, ""),
		service.NewArtifactDownloader(service.ArtifactDownloadOptions{}),
	)
	runtimeErr := runtimeService.RestoreMode(deps.runtimeMode, deps.runtimePath)
	starter := deps.starter
	if starter == nil {
		starter = service.NewAgentProcessManager(nil, service.AgentProcessConfig{})
	}
	supervisor := service.NewJDBCAgentSupervisor(runtimeService, starter, deps.dialer, agentPath, service.NewJDBCLogPaths(paths).Agent)
	gateway := service.NewManagedJDBCGateway(supervisor)
	gateway.SetProfileResolver(func(ctx context.Context, cfg config.DatabaseConfig) (config.JDBCDriverProfile, error) {
		driver, profile, err := catalog.GetProfile(cfg.DBType, cfg.DriverProfileID)
		if err != nil {
			return config.JDBCDriverProfile{}, err
		}
		resolved := *profile
		resolved.InstallPath = filepath.Join(paths.DriversDir, driver.ID, profile.Version)
		return resolved, nil
	})

	return &jdbcServiceBundle{
		paths:      paths,
		catalog:    catalog,
		installer:  installer,
		runtime:    runtimeService,
		supervisor: supervisor,
		gateway:    gateway,
	}, errors.Join(artifactErr, runtimeErr)
}

func (a *App) shutdown(context.Context) {
	// 先关闭业务连接，再停 JDBC agent / 本地子进程，避免残留连接与孤儿进程。
	if a.databaseService != nil {
		if err := a.databaseService.CloseAllSessions(); err != nil {
			fmt.Printf("Failed to close database sessions: %v\n", err)
		}
	}
	if a.nativeDatabaseService != nil {
		if err := a.nativeDatabaseService.CloseAll(); err != nil {
			fmt.Printf("Failed to close native database sessions: %v\n", err)
		}
	}
	if a.sessionService != nil {
		if err := a.sessionService.CloseAllSessions(); err != nil {
			fmt.Printf("Failed to close shell sessions: %v\n", err)
		}
	}
	if a.jdbcAgentSupervisor != nil {
		if err := a.jdbcAgentSupervisor.Close(); err != nil {
			fmt.Printf("Failed to close JDBC agent: %v\n", err)
		}
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) GetVersion() string {
	return Version
}

// GetConnections returns all saved connections
func (a *App) GetConnections() []config.ConnectionConfig {
	conns, _ := a.connectionService.GetConnections()
	return conns
}

// GetConnection retrieves a single connection by ID
func (a *App) GetConnection(id string) (config.ConnectionConfig, error) {
	return a.connectionService.GetConnection(id)
}

// AddConnection adds a new SSH connection
func (a *App) AddConnection(conn config.ConnectionConfig) error {
	return a.connectionService.AddConnection(conn)
}

// UpdateConnection updates an existing SSH connection
func (a *App) UpdateConnection(conn config.ConnectionConfig) error {
	return a.connectionService.UpdateConnection(conn)
}

// RemoveConnection removes an SSH connection
func (a *App) RemoveConnection(id string) error {
	return a.connectionService.RemoveConnection(id)
}

// TestConnection tests an SSH connection
// authType: "password" or "key"
// authValue: password for password auth, or key file path for key auth
// passphrase: passphrase for encrypted keys (optional)
func (a *App) TestConnection(host string, port int, user, authType, authValue, passphrase string) error {
	return a.connectionService.TestConnection(host, port, user, authType, authValue, passphrase)
}

func (a *App) SetSessionCharset(sessionID, charset string) {
	if a == nil || sessionID == "" {
		return
	}
	a.sessionCharsetMu.Lock()
	defer a.sessionCharsetMu.Unlock()
	if a.sessionCharset == nil {
		a.sessionCharset = map[string]string{}
	}
	a.sessionCharset[sessionID] = ssh.NormalizeCharset(charset)
}

func (a *App) sessionCharsetFor(sessionID string) string {
	if a == nil {
		return "utf-8"
	}
	a.sessionCharsetMu.Lock()
	defer a.sessionCharsetMu.Unlock()
	if charset := a.sessionCharset[sessionID]; charset != "" {
		return charset
	}
	return "utf-8"
}

func (a *App) clearSessionCharset(sessionID string) {
	if a == nil || sessionID == "" {
		return
	}
	a.sessionCharsetMu.Lock()
	defer a.sessionCharsetMu.Unlock()
	delete(a.sessionCharset, sessionID)
}

func (a *App) ConnectSSH(sessionID, host string, port int, user, authType, authValue, passphrase string, cols, rows int) error {
	err := a.sessionService.ConnectSSH(sessionID, host, port, user, authType, authValue, passphrase, cols, rows, func(data []byte) {
		cwd := a.parseCWDFromOutput(sessionID, data)
		if cwd != "" {
			runtime.EventsEmit(a.ctx, "ssh:cwd:"+sessionID, cwd)
		}

		encoded := base64.StdEncoding.EncodeToString(data)
		runtime.EventsEmit(a.ctx, "ssh:output:"+sessionID, encoded)
		a.appendSessionLog(sessionID, data)
	}, func(err error) {
		payload := map[string]string{"reason": "closed"}
		if err != nil && err != io.EOF {
			payload["error"] = err.Error()
		}
		runtime.EventsEmit(a.ctx, "ssh:closed:"+sessionID, payload)
	})

	if err == nil {
		fmt.Printf("SSH session started: %s (%s@%s:%d)\n", sessionID, user, host, port)
		a.rememberCopilotSession(sessionID, host, user)
		a.setupCWDTracking(sessionID)
	}
	return err
}

func (a *App) appendSessionLog(sessionID string, data []byte) {
	if a.sessionLogService == nil || a.settingsService == nil {
		return
	}
	settings := a.settingsService.GetSettings()
	if !settings.SessionLogEnabled {
		return
	}
	a.sessionConnectionMu.Lock()
	connID := a.sessionConnection[sessionID]
	a.sessionConnectionMu.Unlock()
	if connID == "" {
		return
	}
	if err := a.sessionLogService.Append(connID, sessionID, data, settings.SessionLogRedactEnabled); err != nil {
		sessionLogAppendWarnOnce.Do(func() {
			fmt.Printf("session log append failed (subsequent errors suppressed): %v\n", err)
		})
	}
}

func (a *App) runSessionLogPurgeTicker() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		if _, err := a.PurgeExpiredSessionLogs(); err != nil {
			fmt.Printf("Failed to purge expired session logs: %v\n", err)
		}
	}
}

// BindSessionConnection associates a terminal session with a saved connection for logging/history.
func (a *App) BindSessionConnection(sessionID, connectionID string) {
	if sessionID == "" {
		return
	}
	a.sessionConnectionMu.Lock()
	defer a.sessionConnectionMu.Unlock()
	if a.sessionConnection == nil {
		a.sessionConnection = make(map[string]string)
	}
	if connectionID == "" {
		delete(a.sessionConnection, sessionID)
		return
	}
	a.sessionConnection[sessionID] = connectionID
}

// ListSessionLogs returns session logs for a connection, newest first.
func (a *App) ListSessionLogs(connectionID string) ([]service.SessionLogInfo, error) {
	if a.sessionLogService == nil {
		return nil, nil
	}
	return a.sessionLogService.List(connectionID)
}

// SearchSessionLogs searches session logs for a connection.
func (a *App) SearchSessionLogs(connectionID, query string, limit int) ([]service.SessionLogHit, error) {
	if a.sessionLogService == nil {
		return nil, nil
	}
	return a.sessionLogService.Search(connectionID, query, limit)
}

// ExportSessionLog opens a save dialog and exports the session log to the chosen path.
func (a *App) ExportSessionLog(logID string) (string, error) {
	if a.sessionLogService == nil {
		return "", fmt.Errorf("session log service not initialized")
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "导出会话日志",
		DefaultFilename: filepath.Base(logID),
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := a.sessionLogService.Export(logID, path); err != nil {
		return "", err
	}
	return path, nil
}

// DeleteSessionLog deletes a session log by id.
func (a *App) DeleteSessionLog(logID string) error {
	if a.sessionLogService == nil {
		return nil
	}
	return a.sessionLogService.Delete(logID)
}

// PurgeExpiredSessionLogs removes session logs older than the configured retention days.
func (a *App) PurgeExpiredSessionLogs() (int, error) {
	if a.sessionLogService == nil {
		return 0, nil
	}
	retentionDays := 30
	if a.settingsService != nil {
		settings := a.settingsService.GetSettings()
		retentionDays = settings.SessionLogRetentionDays
	}
	return a.sessionLogService.PurgeExpired(retentionDays)
}

// RecordCommand records a submitted command for per-connection suggestions.
func (a *App) RecordCommand(connectionID, command string) error {
	if a.commandHistoryService == nil {
		return nil
	}
	if a.settingsService != nil {
		settings := a.settingsService.GetSettings()
		if !settings.CommandSuggestEnabled {
			return nil
		}
	}
	return a.commandHistoryService.Record(connectionID, command)
}

// SuggestCommands returns command suggestions for a connection and prefix.
func (a *App) SuggestCommands(connectionID, prefix string, limit int) ([]service.CommandHistoryEntry, error) {
	if a.commandHistoryService == nil {
		return nil, nil
	}
	if a.settingsService != nil {
		settings := a.settingsService.GetSettings()
		if !settings.CommandSuggestEnabled {
			return nil, nil
		}
		if limit <= 0 {
			limit = settings.CommandSuggestLimit
		}
	} else if limit <= 0 {
		limit = config.DefaultSettings().CommandSuggestLimit
	}
	return a.commandHistoryService.Suggest(connectionID, prefix, limit)
}

func (a *App) parseCWDFromOutput(sessionID string, data []byte) string {
	matches := cwdRegex.FindSubmatch(data)
	if len(matches) >= 2 {
		cwd := string(matches[1])
		if err := a.sftpService.UpdateCurrentPath(sessionID, cwd); err == nil {
			return cwd
		}
	}
	return ""
}

func (a *App) setupCWDTracking(sessionID string) {
	go func() {
		time.Sleep(500 * time.Millisecond)
		promptCmd := `export PROMPT_COMMAND='echo -ne "\033]0;CWD:$(pwd)\007"'` + "\n"
		if err := a.sessionService.SendData(sessionID, promptCmd); err != nil {
			fmt.Printf("Failed to setup CWD tracking for session %s: %v\n", sessionID, err)
		}
	}()
}

// SendSSHData sends data to an SSH session
func (a *App) SendSSHData(sessionID string, data string) error {
	encoded, err := ssh.EncodeFromUTF8(a.sessionCharsetFor(sessionID), data)
	if err != nil {
		return err
	}
	return a.sessionService.SendDataBytes(sessionID, encoded)
}

// SendSSHDataBinary sends base64-encoded binary data to an SSH session
func (a *App) SendSSHDataBinary(sessionID string, data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return a.sessionService.SendDataBytes(sessionID, decoded)
}

// ResizeSSH resizes the terminal for an SSH session
func (a *App) ResizeSSH(sessionID string, cols, rows int) error {
	return a.sessionService.ResizeTerminal(sessionID, cols, rows)
}

// CloseSSH closes an SSH session
func (a *App) CloseSSH(sessionID string) error {
	if a.copilotService != nil {
		a.copilotService.Cancel(sessionID)
	}
	a.clearCopilotSession(sessionID)
	a.clearSessionCharset(sessionID)
	a.sessionConnectionMu.Lock()
	connID := a.sessionConnection[sessionID]
	delete(a.sessionConnection, sessionID)
	a.sessionConnectionMu.Unlock()
	if a.sessionLogService != nil && connID != "" {
		a.sessionLogService.CloseSession(connID, sessionID)
	}
	err := a.sessionService.CloseSession(sessionID)
	if err != nil {
		return err
	}
	fmt.Printf("SSH session closed: %s\n", sessionID)
	return nil
}

// ConnectLocalShell creates and starts a local shell session
func (a *App) ConnectLocalShell(sessionID string, shellType string, cols, rows int) error {
	err := a.sessionService.ConnectLocalShell(sessionID, shellType, cols, rows, func(data []byte) {
		// Encode binary data as base64 to preserve ZMODEM protocol bytes
		encoded := base64.StdEncoding.EncodeToString(data)
		runtime.EventsEmit(a.ctx, "local:output:"+sessionID, encoded)
	}, func(err error) {
		payload := map[string]string{"reason": "closed"}
		if err != nil && err != io.EOF {
			payload["error"] = err.Error()
		}
		runtime.EventsEmit(a.ctx, "local:closed:"+sessionID, payload)
	})

	if err == nil {
		fmt.Printf("Local shell session started: %s\n", sessionID)
	}
	return err
}

// SendLocalShellData sends data to a local shell session
func (a *App) SendLocalShellData(sessionID string, data string) error {
	return a.sessionService.SendLocalData(sessionID, data)
}

// SendLocalShellDataBinary sends base64-encoded binary data to a local shell session
func (a *App) SendLocalShellDataBinary(sessionID string, data string) error {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return err
	}
	return a.sessionService.SendLocalDataBytes(sessionID, decoded)
}

// SaveBinaryFile saves base64-encoded file contents to disk
func (a *App) SaveBinaryFile(filename string, data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "保存文件",
		DefaultFilename: filename,
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil
	}

	if err := os.WriteFile(path, decoded, 0o644); err != nil {
		return "", err
	}

	return path, nil
}

// ResizeLocalShell resizes a local shell session
func (a *App) ResizeLocalShell(sessionID string, cols, rows int) error {
	return a.sessionService.ResizeLocalTerminal(sessionID, cols, rows)
}

// ListSSHSessions returns all active session IDs
func (a *App) ListSSHSessions() []string {
	return a.sessionService.ListSessions()
}

// ShowMessageDialog shows an information message dialog
func (a *App) ShowMessageDialog(title, message string) {
	_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.InfoDialog,
		Title:   title,
		Message: message,
	})
}

// ShowErrorDialog shows an error message dialog
func (a *App) ShowErrorDialog(title, message string) {
	_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Title:   title,
		Message: message,
	})
}

// ShowQuestionDialog shows a question dialog and returns true if user confirms
func (a *App) ShowQuestionDialog(title, message string) (bool, error) {
	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         title,
		Message:       message,
		Buttons:       []string{"是", "否"},
		DefaultButton: "是",
		CancelButton:  "否",
	})

	if err != nil {
		return false, err
	}

	return result == "是", nil
}

// SelectSSHKeyFile opens a file picker dialog for selecting SSH private key files
func (a *App) SelectSSHKeyFile() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 SSH 私钥文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "SSH 私钥 (id_rsa, id_ed25519, id_ecdsa)",
				Pattern:     "id_rsa;id_ed25519;id_ecdsa;*.pem;*.key",
			},
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	})

	if err != nil {
		return "", err
	}

	return filePath, nil
}

// SavePassword saves a password for a connection (encrypted)
func (a *App) SavePassword(connectionID, password string) error {
	return a.connectionService.SavePassword(connectionID, password)
}

// GetPassword retrieves a saved password for a connection
func (a *App) GetPassword(connectionID string) (string, error) {
	return a.connectionService.GetPassword(connectionID)
}

// HasPassword checks if a password is saved for a connection
func (a *App) HasPassword(connectionID string) bool {
	return a.connectionService.HasPassword(connectionID)
}

// DeletePassword removes a saved password for a connection
func (a *App) DeletePassword(connectionID string) error {
	return a.connectionService.DeletePassword(connectionID)
}

// GetSettings returns application settings
func (a *App) GetSettings() config.AppSettings {
	return a.settingsService.GetSettings()
}

// UpdateSettings updates application settings
func (a *App) UpdateSettings(updates map[string]interface{}) error {
	return a.settingsService.UpdateSettings(updates)
}

func (a *App) CopilotChat(req copilot.ChatRequest) (*copilot.ChatResponse, error) {
	if a.copilotService == nil {
		return nil, fmt.Errorf("copilot 未初始化")
	}

	var settings config.AppSettings
	if a.settingsService != nil {
		settings = a.settingsService.GetSettings()
	}

	apiKey := ""
	if a.credentialStore != nil {
		if stored, err := a.credentialStore.Get(copilot.APIKeyCredentialID); err == nil {
			apiKey = stored
		}
	}
	if err := copilot.ValidateConfig(settings.CopilotBaseURL, apiKey); err != nil {
		return nil, err
	}

	a.fillCopilotRequest(&req)
	req.Model = settings.CopilotModel

	provider := copilot.NewOpenAICompatible(settings.CopilotBaseURL, apiKey, &http.Client{Timeout: 60 * time.Second})
	parent := context.Background()
	if a.ctx != nil {
		parent = a.ctx
	}
	ctx, cancel := context.WithTimeout(parent, 60*time.Second)
	defer cancel()
	return a.copilotService.WithProvider(provider).WithRuntimeSettings(copilot.RuntimeSettings{MaxToolRounds: settings.CopilotMaxToolRounds, MaxToolResultChars: settings.CopilotMaxToolResultChars}).Chat(ctx, req)
}

// CopilotCancel stops an in-flight chat for the session.
func (a *App) CopilotCancel(sessionID string) {
	if a.copilotService == nil || strings.TrimSpace(sessionID) == "" {
		return
	}
	a.copilotService.Cancel(sessionID)
}

func (a *App) CopilotClassify(kind, content string) copilot.Result {
	return copilot.Classify(kind, content)
}

func (a *App) HasCopilotAPIKey() bool {
	if a.credentialStore == nil {
		return false
	}
	return a.credentialStore.Has(copilot.APIKeyCredentialID)
}

func (a *App) SetCopilotAPIKey(apiKey string) error {
	if a.credentialStore == nil {
		return fmt.Errorf("凭据存储未初始化")
	}
	if strings.TrimSpace(apiKey) == "" {
		// 空串视为清除：避免 HasCopilotAPIKey 误报 true，且与 ValidateConfig 行为一致。
		return a.credentialStore.Delete(copilot.APIKeyCredentialID)
	}
	return a.credentialStore.Store(copilot.APIKeyCredentialID, apiKey)
}

func (a *App) ClearCopilotAPIKey() error {
	if a.credentialStore == nil {
		return fmt.Errorf("凭据存储未初始化")
	}
	return a.credentialStore.Delete(copilot.APIKeyCredentialID)
}

func fillIfEmpty(dst *string, src string) {
	if dst == nil || strings.TrimSpace(*dst) != "" || src == "" {
		return
	}
	*dst = src
}

func (a *App) fillCopilotRequest(req *copilot.ChatRequest) {
	if req == nil {
		return
	}
	if a.databaseService != nil {
		if sess, err := a.databaseService.GetSession(req.SessionID); err == nil && sess != nil {
			fillIfEmpty(&req.Host, sess.Config.Host)
			fillIfEmpty(&req.User, sess.Config.User)
			fillIfEmpty(&req.DBType, sess.Config.DBType)
			fillIfEmpty(&req.Database, sess.Config.Database)
		}
	}
	if a.nativeDatabaseService != nil && (req.Host == "" || req.DBType == "") {
		if cfg, ok := a.nativeDatabaseService.SessionConfig(req.SessionID); ok {
			fillIfEmpty(&req.Host, cfg.Host)
			fillIfEmpty(&req.User, cfg.User)
			fillIfEmpty(&req.DBType, string(cfg.Type))
			fillIfEmpty(&req.Database, cfg.Database)
		}
	}
	if (req.Host == "" || req.User == "") && a.connectionService != nil {
		if conn, err := a.connectionService.GetConnection(req.SessionID); err == nil {
			fillIfEmpty(&req.Host, conn.Host)
			fillIfEmpty(&req.User, conn.User)
		}
	}
	if meta, ok := a.lookupCopilotSession(req.SessionID); ok {
		if req.Host == "" {
			req.Host = meta.host
		}
		if req.User == "" {
			req.User = meta.user
		}
	}
	if req.WorkingDir == "" && a.sftpService != nil {
		if cwd, err := a.sftpService.GetCurrentPath(req.SessionID); err == nil {
			req.WorkingDir = cwd
		}
	}
}

func (a *App) rememberCopilotSession(sessionID, host, user string) {
	if a == nil || sessionID == "" {
		return
	}
	a.copilotMetaMu.Lock()
	defer a.copilotMetaMu.Unlock()
	if a.copilotMeta == nil {
		a.copilotMeta = make(map[string]copilotSessionMeta)
	}
	a.copilotMeta[sessionID] = copilotSessionMeta{host: host, user: user}
}

func (a *App) lookupCopilotSession(sessionID string) (copilotSessionMeta, bool) {
	if a == nil {
		return copilotSessionMeta{}, false
	}
	a.copilotMetaMu.Lock()
	defer a.copilotMetaMu.Unlock()
	if a.copilotMeta == nil {
		return copilotSessionMeta{}, false
	}
	meta, ok := a.copilotMeta[sessionID]
	return meta, ok
}

func (a *App) clearCopilotSession(sessionID string) {
	if a == nil {
		return
	}
	a.copilotMetaMu.Lock()
	defer a.copilotMetaMu.Unlock()
	delete(a.copilotMeta, sessionID)
}

// SelectBackgroundImage opens a file dialog and installs the chosen wallpaper.
func (a *App) SelectBackgroundImage() (*service.BackgroundImageResult, error) {
	if a.ctx == nil {
		return nil, fmt.Errorf("应用上下文未初始化")
	}
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择背景图片",
		Filters: []runtime.FileFilter{
			{DisplayName: "图片", Pattern: "*.jpg;*.jpeg;*.png;*.webp;*.gif"},
		},
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(filePath) == "" {
		return nil, nil
	}
	return a.settingsService.InstallBackgroundImage(filePath)
}

// ClearBackgroundImage removes the custom wallpaper.
func (a *App) ClearBackgroundImage() error {
	return a.settingsService.ClearBackgroundImage()
}

// GetBackgroundImageDataURL returns the current wallpaper as a data URL.
func (a *App) GetBackgroundImageDataURL() (string, error) {
	return a.settingsService.GetBackgroundImageDataURL()
}

// GetMonitoringData retrieves monitoring data for a session
func (a *App) GetMonitoringData(sessionID string) (*ssh.MonitoringData, error) {
	return a.monitorService.GetMonitoringData(sessionID)
}

// ListFiles lists files in a directory
func (a *App) ListFiles(sessionID string, path string) ([]ssh.FileInfo, error) {
	return a.sftpService.ListFiles(sessionID, path)
}

// ChangeDirectory changes the current working directory for a session
func (a *App) ChangeDirectory(sessionID string, path string) error {
	return a.sftpService.ChangeDirectory(sessionID, path)
}

// GetCurrentPath returns the current working directory
func (a *App) GetCurrentPath(sessionID string) (string, error) {
	return a.sftpService.GetCurrentPath(sessionID)
}

// UpdateCurrentPath updates the tracked working directory for a session
func (a *App) UpdateCurrentPath(sessionID string, path string) error {
	return a.sftpService.UpdateCurrentPath(sessionID, path)
}

// GetFileManagerSettings returns file manager settings for a specific connection
func (a *App) GetFileManagerSettings(connectionId string) (config.FileManagerSettings, error) {
	return a.settingsService.GetFileManagerSettings(connectionId), nil
}

// UpdateFileManagerSettings updates file manager settings for a specific connection
func (a *App) UpdateFileManagerSettings(connectionId string, settings map[string]interface{}) error {
	updates := map[string]interface{}{
		"connection_id":         connectionId,
		"file_manager_settings": settings,
	}
	return a.configManager.UpdateSettings(updates)
}

// UploadFiles uploads multiple files
func (a *App) UploadFiles(sessionID string, localPaths []string, remotePath string) ([]string, error) {
	return a.sftpService.UploadFiles(sessionID, localPaths, remotePath, func(progress ssh.TransferProgress) {
		// Emit event to frontend
		runtime.EventsEmit(a.ctx, sftpProgressEventPrefix+progress.TransferID, progress)
	})
}

// ExpandUploadPaths walks local files and folders into a remote-relative tree.
func (a *App) ExpandUploadPaths(localPaths []string) ([]service.LocalUploadItem, error) {
	return service.ExpandLocalUploadPaths(localPaths)
}

// UploadExpandedItems uploads a previously expanded local tree, including renamed RelPaths.
func (a *App) UploadExpandedItems(sessionID string, remotePath string, items []service.LocalUploadItem) ([]string, error) {
	return a.sftpService.UploadItems(sessionID, remotePath, items, func(progress ssh.TransferProgress) {
		runtime.EventsEmit(a.ctx, sftpProgressEventPrefix+progress.TransferID, progress)
	})
}

// DownloadFile downloads a single file
func (a *App) DownloadFile(sessionID string, remotePath string, localPath string) (string, error) {
	// Use service with Wails-specific progress callback
	return a.sftpService.DownloadFile(sessionID, remotePath, localPath, func(progress ssh.TransferProgress) {
		// Emit event to frontend
		runtime.EventsEmit(a.ctx, sftpProgressEventPrefix+progress.TransferID, progress)
	})
}

// DownloadFiles downloads multiple files
func (a *App) DownloadFiles(sessionID string, remotePaths []string, localPath string) ([]string, error) {
	return a.sftpService.DownloadFiles(sessionID, remotePaths, localPath, func(progress ssh.TransferProgress) {
		// Emit event to frontend
		runtime.EventsEmit(a.ctx, sftpProgressEventPrefix+progress.TransferID, progress)
	})
}

// DeleteFile deletes a single file or directory
func (a *App) DeleteFile(sessionID string, path string) error {
	return a.sftpService.DeleteFile(sessionID, path)
}

// DeleteFiles deletes multiple files or directories
func (a *App) DeleteFiles(sessionID string, paths []string) error {
	return a.sftpService.DeleteFiles(sessionID, paths)
}

// RenameFile renames a file or directory
func (a *App) RenameFile(sessionID string, oldPath string, newPath string) error {
	return a.sftpService.RenameFile(sessionID, oldPath, newPath)
}

// CreateDirectory creates a new directory
func (a *App) CreateDirectory(sessionID string, path string) error {
	return a.sftpService.CreateDirectory(sessionID, path)
}

// CreateFile creates an empty remote file
func (a *App) CreateFile(sessionID string, path string) error {
	return a.sftpService.CreateFile(sessionID, path)
}

// CopyFile copies a remote file to another remote path
func (a *App) CopyFile(sessionID string, srcPath string, dstPath string) error {
	return a.sftpService.CopyFile(sessionID, srcPath, dstPath)
}

// ChmodFile updates remote file permissions with an octal mode such as 644
func (a *App) ChmodFile(sessionID string, path string, mode string) error {
	return a.sftpService.ChmodFile(sessionID, path, mode)
}

// GetFileInfo gets information about a file
func (a *App) GetFileInfo(sessionID string, path string) (*ssh.FileInfo, error) {
	return a.sftpService.GetFileInfo(sessionID, path)
}

// SearchDirectories searches for directories matching the query recursively
func (a *App) SearchDirectories(sessionID string, searchPath string, query string, maxDepth int, maxResults int) ([]ssh.SearchResult, error) {
	return a.sftpService.SearchDirectories(sessionID, searchPath, query, maxDepth, maxResults)
}

// CancelTransfer cancels a file transfer
func (a *App) CancelTransfer(transferID string) error {
	return a.sftpService.CancelTransfer(transferID)
}

// GetTransferStatus gets the status of a transfer
func (a *App) GetTransferStatus(transferID string) (*ssh.TransferProgress, error) {
	return a.sftpService.GetTransferStatus(transferID)
}

// GetClipboardFilePaths returns local file and folder paths on the OS clipboard.
func (a *App) GetClipboardFilePaths() ([]string, error) {
	return service.ReadClipboardFilePaths()
}

// SelectUploadFiles opens a file picker for selecting files to upload
func (a *App) SelectUploadFiles() ([]string, error) {
	filePaths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要上传的文件",
	})

	if err != nil {
		return nil, err
	}

	return filePaths, nil
}

// SelectUploadDirectory opens a directory picker for uploading a local folder.
func (a *App) SelectUploadDirectory() (string, error) {
	dirPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择要上传的文件夹",
	})
	if err != nil {
		return "", err
	}
	return dirPath, nil
}

// SelectDownloadDirectory opens a directory picker for selecting download destination
func (a *App) SelectDownloadDirectory() (string, error) {
	dirPath, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择下载目录",
	})

	if err != nil {
		return "", err
	}

	return dirPath, nil
}

// ============================================================================
// DevTools Methods - 开发工具集相关方法
// ============================================================================

// FormatJSON 格式化JSON字符串
func (a *App) FormatJSON(input string) (string, error) {
	return a.devToolsService.FormatJSON(input)
}

// ValidateJSON 验证JSON字符串
func (a *App) ValidateJSON(input string) (service.JSONValidationResult, error) {
	return a.devToolsService.ValidateJSON(input)
}

// MinifyJSON 压缩JSON
func (a *App) MinifyJSON(input string) (string, error) {
	return a.devToolsService.MinifyJSON(input)
}

// EscapeJSON 转义JSON字符串
func (a *App) EscapeJSON(input string) (string, error) {
	return a.devToolsService.EscapeJSON(input)
}

// EncodeBase64 将字符串编码为 Base64
func (a *App) EncodeBase64(input string) (string, error) {
	return a.devToolsService.EncodeBase64(input)
}

// DecodeBase64 将 Base64 字符串解码
func (a *App) DecodeBase64(input string) (string, error) {
	return a.devToolsService.DecodeBase64(input)
}

// CalculateHash 计算字符串的哈希值
func (a *App) CalculateHash(input, algorithm string) (string, error) {
	return a.devToolsService.CalculateHash(input, algorithm)
}

// EncryptText 对文本进行加密，返回 Base64 密文
func (a *App) EncryptText(input, algorithm, keyHex, ivHex string) (string, error) {
	return a.devToolsService.EncryptText(input, algorithm, keyHex, ivHex)
}

// DecryptText 对 Base64 密文进行解密
func (a *App) DecryptText(input, algorithm, keyHex, ivHex string) (string, error) {
	return a.devToolsService.DecryptText(input, algorithm, keyHex, ivHex)
}

// TimestampToDateTime 将 Unix 时间戳转换为日期时间字符串
func (a *App) TimestampToDateTime(timestamp int64, format string) (string, error) {
	return a.devToolsService.TimestampToDateTime(timestamp, format)
}

// TimestampToDateTimeMs 将 Unix 毫秒时间戳转换为日期时间字符串
func (a *App) TimestampToDateTimeMs(timestampMs int64, format string) (string, error) {
	return a.devToolsService.TimestampToDateTimeMs(timestampMs, format)
}

// DateTimeToTimestamp 将日期时间字符串转换为 Unix 时间戳
func (a *App) DateTimeToTimestamp(datetime, format string) (int64, error) {
	return a.devToolsService.DateTimeToTimestamp(datetime, format)
}

// DateTimeToTimestampMs 将日期时间字符串转换为 Unix 毫秒时间戳
func (a *App) DateTimeToTimestampMs(datetime, format string) (int64, error) {
	return a.devToolsService.DateTimeToTimestampMs(datetime, format)
}

// GetCurrentTimestamp 获取当前 Unix 时间戳
func (a *App) GetCurrentTimestamp() int64 {
	return a.devToolsService.GetCurrentTimestamp()
}

// GetCurrentTimestampMs 获取当前 Unix 毫秒时间戳
func (a *App) GetCurrentTimestampMs() int64 {
	return a.devToolsService.GetCurrentTimestampMs()
}

// GenerateUUIDv4 生成 UUID v4
func (a *App) GenerateUUIDv4() (string, error) {
	return a.devToolsService.GenerateUUIDv4()
}

// URLEncode 对字符串进行 URL 编码
func (a *App) URLEncode(input, mode string) (service.URLEncodeResult, error) {
	return a.devToolsService.URLEncode(input, mode)
}

// URLDecode 对 URL 编码的字符串进行解码
func (a *App) URLDecode(input, mode string) (service.URLDecodeResult, error) {
	return a.devToolsService.URLDecode(input, mode)
}

// ParseURL 解析 URL 返回各个组成部分
func (a *App) ParseURL(input string) (map[string]interface{}, error) {
	return a.devToolsService.ParseURL(input)
}

// ShowAboutDialog 显示关于对话框
func (a *App) ShowAboutDialog() {
	runtime.EventsEmit(a.ctx, "app:show-about")
}

// ============================================================================
// Connection Export/Import Methods
// ============================================================================

const (
	passphrasePrefix = "penc:"
	passphraseKDF    = "argon2id"
)

// PasswordEncryption describes how passwords are encrypted in export data
type PasswordEncryption struct {
	Mode string `json:"mode"`
	Salt string `json:"salt"`
	KDF  string `json:"kdf"`
}

// ExportData represents exported connection data
type ExportData struct {
	Version            string                    `json:"version"`
	ExportedAt         string                    `json:"exported_at"`
	Connections        []config.ConnectionConfig `json:"connections"`
	Passwords          map[string]string         `json:"passwords,omitempty"`
	PasswordEncryption *PasswordEncryption       `json:"password_encryption,omitempty"`
}

// ExportConnections exports all connections with passwords to JSON
func (a *App) ExportConnections(encryptPasswords bool) (string, error) {
	conns, err := a.connectionService.GetConnections()
	if err != nil {
		return "", err
	}

	exportData := ExportData{
		Version:     "1.0",
		ExportedAt:  time.Now().UTC().Format(time.RFC3339),
		Connections: conns,
		Passwords:   make(map[string]string),
	}

	for _, conn := range conns {
		if encryptPasswords {
			password, err := a.connectionService.GetEncryptedPassword(conn.ID)
			if err == nil && password != "" {
				exportData.Passwords[conn.ID] = "enc:" + password
			}
			continue
		}

		password, err := a.connectionService.GetPassword(conn.ID)
		if err == nil && password != "" {
			exportData.Passwords[conn.ID] = password
		}
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export data: %w", err)
	}

	return string(data), nil
}

// ExportConnectionsByIDs exports selected connections to JSON
func (a *App) ExportConnectionsByIDs(connectionIDs []string, encryptPasswords bool) (string, error) {
	filtered, err := a.selectedConnections(connectionIDs)
	if err != nil {
		return "", err
	}

	exportData := ExportData{
		Version:     "1.0",
		ExportedAt:  time.Now().UTC().Format(time.RFC3339),
		Connections: filtered,
		Passwords:   make(map[string]string),
	}

	for _, conn := range filtered {
		if encryptPasswords {
			password, err := a.connectionService.GetEncryptedPassword(conn.ID)
			if err == nil && password != "" {
				exportData.Passwords[conn.ID] = "enc:" + password
			}
			continue
		}

		password, err := a.connectionService.GetPassword(conn.ID)
		if err == nil && password != "" {
			exportData.Passwords[conn.ID] = password
		}
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export data: %w", err)
	}

	return string(data), nil
}

// ExportConnectionsByIDsWithPassphrase exports selected connections using passphrase encryption
func (a *App) ExportConnectionsByIDsWithPassphrase(connectionIDs []string, passphrase string) (string, error) {
	if len(connectionIDs) == 0 {
		return "", fmt.Errorf("no connections selected")
	}
	if strings.TrimSpace(passphrase) == "" {
		return "", fmt.Errorf("passphrase required")
	}
	filtered, err := a.selectedConnections(connectionIDs)
	if err != nil {
		return "", err
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	key := derivePassphraseKey(passphrase, salt)
	encodedSalt := base64.StdEncoding.EncodeToString(salt)

	exportData := ExportData{
		Version:     "1.0",
		ExportedAt:  time.Now().UTC().Format(time.RFC3339),
		Connections: filtered,
		Passwords:   make(map[string]string),
		PasswordEncryption: &PasswordEncryption{
			Mode: "passphrase",
			Salt: encodedSalt,
			KDF:  passphraseKDF,
		},
	}

	for _, conn := range filtered {
		password, err := a.connectionService.GetPassword(conn.ID)
		if err != nil || password == "" {
			continue
		}
		ciphertext, err := encryptWithKey(password, key)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt password: %w", err)
		}
		exportData.Passwords[conn.ID] = passphrasePrefix + ciphertext
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export data: %w", err)
	}

	return string(data), nil
}

func (a *App) selectedConnections(connectionIDs []string) ([]config.ConnectionConfig, error) {
	if len(connectionIDs) == 0 {
		return nil, fmt.Errorf("no connections selected")
	}
	connections, err := a.connectionService.GetConnections()
	if err != nil {
		return nil, err
	}
	selected := make(map[string]struct{}, len(connectionIDs))
	for _, id := range connectionIDs {
		if id != "" {
			selected[id] = struct{}{}
		}
	}
	filtered := make([]config.ConnectionConfig, 0, len(selected))
	for _, connection := range connections {
		if _, ok := selected[connection.ID]; ok {
			filtered = append(filtered, connection)
		}
	}
	return filtered, nil
}

// ImportConnections imports connections from JSON data
func (a *App) ImportConnections(jsonData string) (int, error) {
	var exportData ExportData

	if err := json.Unmarshal([]byte(jsonData), &exportData); err != nil {
		return 0, fmt.Errorf("failed to parse import data: %w", err)
	}

	if exportData.PasswordEncryption != nil && exportData.PasswordEncryption.Mode == "passphrase" {
		return 0, fmt.Errorf("passphrase required")
	}

	importedCount := 0

	for _, conn := range exportData.Connections {
		originalID := conn.ID
		existingConns, _ := a.connectionService.GetConnections()
		idExists := false
		for _, existing := range existingConns {
			if existing.ID == conn.ID {
				idExists = true
				break
			}
		}

		if idExists {
			conn.ID = generateNewID()
		}

		if err := a.connectionService.AddConnection(conn); err != nil {
			fmt.Printf("Failed to add connection %s: %v\n", conn.ID, err)
			continue
		}

		if password, hasPassword := exportData.Passwords[originalID]; hasPassword {
			if strings.HasPrefix(password, "enc:") {
				encrypted := strings.TrimPrefix(password, "enc:")
				if err := a.connectionService.StoreEncryptedPassword(conn.ID, encrypted); err != nil {
					fmt.Printf("Failed to store encrypted password for %s: %v\n", conn.ID, err)
				}
			} else if isEncryptedPassword(password) {
				if err := a.connectionService.StoreEncryptedPassword(conn.ID, password); err != nil {
					fmt.Printf("Failed to store encrypted password for %s: %v\n", conn.ID, err)
				}
			} else if err := a.connectionService.SavePassword(conn.ID, password); err != nil {
				fmt.Printf("Failed to save password for %s: %v\n", conn.ID, err)
			}
		}

		importedCount++
	}

	return importedCount, nil
}

// ImportConnectionsWithPassphrase imports passphrase-encrypted connections
func (a *App) ImportConnectionsWithPassphrase(jsonData, passphrase string) (int, error) {
	var exportData ExportData
	if err := json.Unmarshal([]byte(jsonData), &exportData); err != nil {
		return 0, fmt.Errorf("failed to parse import data: %w", err)
	}

	if exportData.PasswordEncryption == nil || exportData.PasswordEncryption.Mode != "passphrase" {
		return a.ImportConnections(jsonData)
	}

	if strings.TrimSpace(passphrase) == "" {
		return 0, fmt.Errorf("passphrase required")
	}

	salt, err := base64.StdEncoding.DecodeString(exportData.PasswordEncryption.Salt)
	if err != nil {
		return 0, fmt.Errorf("invalid passphrase salt")
	}

	key := derivePassphraseKey(passphrase, salt)

	importedCount := 0
	for _, conn := range exportData.Connections {
		originalID := conn.ID
		existingConns, _ := a.connectionService.GetConnections()
		idExists := false
		for _, existing := range existingConns {
			if existing.ID == conn.ID {
				idExists = true
				break
			}
		}

		if idExists {
			conn.ID = generateNewID()
		}

		if err := a.connectionService.AddConnection(conn); err != nil {
			fmt.Printf("Failed to add connection %s: %v\n", conn.ID, err)
			continue
		}

		if password, hasPassword := exportData.Passwords[originalID]; hasPassword {
			if strings.HasPrefix(password, passphrasePrefix) {
				ciphertext := strings.TrimPrefix(password, passphrasePrefix)
				plaintext, err := decryptWithKey(ciphertext, key)
				if err != nil {
					return 0, fmt.Errorf("invalid passphrase")
				}
				if err := a.connectionService.SavePassword(conn.ID, plaintext); err != nil {
					fmt.Printf("Failed to save password for %s: %v\n", conn.ID, err)
				}
			} else if err := a.connectionService.SavePassword(conn.ID, password); err != nil {
				fmt.Printf("Failed to save password for %s: %v\n", conn.ID, err)
			}
		}

		importedCount++
	}

	return importedCount, nil
}

// ImportConnectionsFromFile imports connections from a JSON file
func (a *App) ImportConnectionsFromFile(filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read import file: %w", err)
	}

	return a.ImportConnections(string(data))
}

// ImportConnectionsFromFileWithPassphrase imports passphrase-encrypted connections from file
func (a *App) ImportConnectionsFromFileWithPassphrase(filePath, passphrase string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read import file: %w", err)
	}

	return a.ImportConnectionsWithPassphrase(string(data), passphrase)
}

// SelectImportFile opens a file picker for selecting connection import file
func (a *App) SelectImportFile() (string, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择导入文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "JSON 文件 (*.json)",
				Pattern:     "*.json",
			},
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func (a *App) ConnectDatabase(sessionID, host string, port int, user, password, dbType, database string) error {
	return a.databaseService.ConnectDatabase(sessionID, host, port, user, password, dbType, database)
}

func (a *App) ConnectDatabaseWithProfile(sessionID, host string, port int, user, password, dbType, database, driverProfileID string) error {
	return a.databaseService.ConnectDatabaseWithProfile(sessionID, host, port, user, password, dbType, database, driverProfileID)
}

func (a *App) ConnectDatabaseWithOptions(sessionID, host string, port int, user, password, dbType, database, driverProfileID string, properties map[string]string) error {
	return a.databaseService.ConnectDatabaseWithProfileAndProperties(sessionID, host, port, user, password, dbType, database, driverProfileID, properties)
}

func (a *App) ExecuteDatabaseQuery(sessionID, query string) (string, error) {
	result, err := a.databaseService.ExecuteQuery(sessionID, query)
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal query result: %w", err)
	}

	return string(jsonData), nil
}

func (a *App) ListDatabaseTables(sessionID string) ([]string, error) {
	return a.databaseService.ListTables(sessionID)
}

func (a *App) ListDatabases(sessionID string) ([]string, error) {
	return a.databaseService.ListDatabases(sessionID)
}

func (a *App) ListDatabaseTablesInDatabase(sessionID, database string) ([]string, error) {
	return a.databaseService.ListTablesInDatabase(sessionID, database)
}

func (a *App) ListDatabaseSchemas(sessionID, database string) ([]string, error) {
	return a.databaseService.ListSchemas(sessionID, database)
}

func (a *App) ListDatabaseTablesInSchema(sessionID, database, schema string) ([]string, error) {
	return a.databaseService.ListTablesInSchema(sessionID, database, schema)
}

func (a *App) ListDatabaseObjects(sessionID, database, schema string, types []string) ([]string, error) {
	return a.databaseService.ListObjects(sessionID, database, schema, types)
}

func (a *App) ListDatabaseRoutines(sessionID, database, schema string, functions bool) ([]string, error) {
	return a.databaseService.ListRoutines(sessionID, database, schema, functions)
}

func (a *App) TestNativeDatabaseConnection(host string, port int, user, password, databaseType, database string) error {
	service, cfg, ctx, cancel, err := a.nativeDatabaseRequest(host, port, user, password, databaseType, database)
	if err != nil {
		return err
	}
	defer cancel()
	return service.TestConnection(ctx, cfg)
}

func (a *App) ConnectNativeDatabase(sessionID, host string, port int, user, password, databaseType, database string) error {
	service, cfg, ctx, cancel, err := a.nativeDatabaseRequest(host, port, user, password, databaseType, database)
	if err != nil {
		return err
	}
	defer cancel()
	return service.Connect(ctx, sessionID, cfg)
}

func (a *App) ListNativeDatabaseResources(sessionID string) ([]service.NativeResource, error) {
	if a.nativeDatabaseService == nil {
		return nil, fmt.Errorf("原生数据库服务尚未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.nativeDatabaseService.ListPrimaryResources(ctx, sessionID)
}

func (a *App) ListNativeDatabaseChildResources(sessionID, parent string) ([]service.NativeResource, error) {
	if a.nativeDatabaseService == nil {
		return nil, fmt.Errorf("原生数据库服务尚未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.nativeDatabaseService.ListSecondaryResources(ctx, sessionID, parent)
}

// ListNativeDatabaseChildResourcesPage lists child resources with pattern + cursor pagination.
func (a *App) ListNativeDatabaseChildResourcesPage(sessionID, parent, pattern, cursor string, limit int) (service.NativeResourcePage, error) {
	if a.nativeDatabaseService == nil {
		return service.NativeResourcePage{}, fmt.Errorf("原生数据库服务尚未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.nativeDatabaseService.ListSecondaryResourcesPage(ctx, sessionID, parent, pattern, cursor, limit)
}

// DescribeNativeDatabaseResource returns a bounded, read-only resource preview.
func (a *App) DescribeNativeDatabaseResource(sessionID, parent, name string) (service.NativeResourceDetails, error) {
	if a.nativeDatabaseService == nil {
		return service.NativeResourceDetails{}, fmt.Errorf("原生数据库服务尚未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.nativeDatabaseService.DescribeResource(ctx, sessionID, parent, name)
}

// DescribeNativeDatabaseSession returns session/cluster overview when supported.
func (a *App) DescribeNativeDatabaseSession(sessionID string) (service.NativeResourceDetails, error) {
	if a.nativeDatabaseService == nil {
		return service.NativeResourceDetails{}, fmt.Errorf("原生数据库服务尚未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return a.nativeDatabaseService.DescribeSession(ctx, sessionID)
}

func (a *App) CloseNativeDatabase(sessionID string) error {
	if a.nativeDatabaseService == nil {
		return fmt.Errorf("原生数据库服务尚未初始化")
	}
	return a.nativeDatabaseService.Close(sessionID)
}

func (a *App) ExecuteNativeDatabaseQuery(sessionID, parent, name, query string) (service.NativeQueryResult, error) {
	if a.nativeDatabaseService == nil {
		return service.NativeQueryResult{}, fmt.Errorf("原生数据库服务尚未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return a.nativeDatabaseService.ExecuteQuery(ctx, sessionID, parent, name, query)
}

func (a *App) MutateNativeDatabaseResource(sessionID, parent, name, operation, payload string) (service.NativeMutationResult, error) {
	if a.nativeDatabaseService == nil {
		return service.NativeMutationResult{}, fmt.Errorf("原生数据库服务尚未初始化")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return a.nativeDatabaseService.MutateResource(ctx, sessionID, parent, name, operation, payload)
}

func (a *App) nativeDatabaseRequest(host string, port int, user, password, databaseType, database string) (*service.NativeDatabaseService, service.NativeDatabaseConfig, context.Context, context.CancelFunc, error) {
	if a.nativeDatabaseService == nil {
		return nil, service.NativeDatabaseConfig{}, nil, nil, fmt.Errorf("原生数据库服务尚未初始化")
	}
	databaseType = strings.TrimSpace(databaseType)
	if databaseType == "" {
		return nil, service.NativeDatabaseConfig{}, nil, nil, fmt.Errorf("原生数据库类型不能为空")
	}
	user = strings.TrimSpace(user)
	if service.NativeDatabaseType(databaseType) == service.NativeDatabaseTypeRedis {
		// Redis 的默认认证仅使用 requirepass；忽略旧连接中遗留的用户名。
		user = ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return a.nativeDatabaseService, service.NativeDatabaseConfig{
		Type:     service.NativeDatabaseType(databaseType),
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
		Timeout:  10 * time.Second,
	}, ctx, cancel, nil
}

func (a *App) ListJDBCDrivers() ([]service.DriverView, error) {
	return a.jdbcCatalog.ListDriversWithInstallStatus()
}

func (a *App) InstallJDBCDriver(driverID, version string) error {
	driver, profile, err := a.jdbcCatalog.GetProfile(driverID, version)
	if err != nil {
		return err
	}
	_, err = a.jdbcInstaller.InstallProfile(context.Background(), *driver, *profile)
	return err
}

func (a *App) ImportJDBCDriverPackage(path string) error {
	_, err := a.jdbcInstaller.ImportOfflinePackage(path)
	return err
}

func (a *App) ValidateJDBCDriver(driverID, version string) error {
	installPath := filepath.Join(a.jdbcPaths.DriversDir, driverID, version)
	if _, err := os.Stat(installPath); err != nil {
		return fmt.Errorf("JDBC 驱动未安装或不可访问: %w", err)
	}
	return nil
}

func (a *App) RemoveJDBCDriver(driverID, version string) error {
	profile, err := a.jdbcDriverProfile(driverID, version)
	if err != nil {
		return err
	}
	users, err := a.jdbcDriverUsers(driverID, *profile)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return fmt.Errorf("[DRIVER_IN_USE] JDBC 驱动 %s %s 正被使用：%s。请先修改这些连接的驱动版本或关闭活动会话后再卸载", driverID, profile.Version, strings.Join(users, "；"))
	}
	installPath := filepath.Join(a.jdbcPaths.DriversDir, driverID, version)
	if err := os.RemoveAll(installPath); err != nil {
		return fmt.Errorf("删除 JDBC 驱动失败: %w", err)
	}
	return nil
}

func (a *App) jdbcDriverProfile(driverID, version string) (*config.JDBCDriverProfile, error) {
	if a.jdbcCatalog == nil {
		return nil, fmt.Errorf("JDBC 驱动目录尚未初始化")
	}
	_, profile, err := a.jdbcCatalog.GetProfile(driverID, version)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (a *App) jdbcDriverUsers(driverID string, target config.JDBCDriverProfile) ([]string, error) {
	users, err := a.savedJDBCDriverUsers(driverID, target)
	if err != nil {
		return nil, err
	}
	users = append(users, a.activeJDBCDriverUsers(driverID, target)...)
	return users, nil
}

func (a *App) savedJDBCDriverUsers(driverID string, target config.JDBCDriverProfile) ([]string, error) {
	if a.connectionService == nil {
		return nil, nil
	}
	connections, err := a.connectionService.GetConnections()
	if err != nil {
		return nil, fmt.Errorf("检查 JDBC 驱动连接引用失败: %w", err)
	}
	users := make([]string, 0)
	for _, connection := range connections {
		if isJDBCDriverProfileInUse(driverID, target, connection.Metadata["db_type"], connection.Metadata["driver_profile_id"], a.jdbcCatalog) {
			users = append(users, "已保存连接 "+jdbcConnectionName(connection))
		}
	}
	return users, nil
}

func (a *App) activeJDBCDriverUsers(driverID string, target config.JDBCDriverProfile) []string {
	if a.jdbcGateway == nil {
		return nil
	}
	users := make([]string, 0)
	for sessionID, session := range a.jdbcGateway.ActiveSessionConfigs() {
		if isJDBCDriverProfileInUse(driverID, target, session.DBType, session.DriverProfileID, a.jdbcCatalog) {
			users = append(users, "活动会话 "+sessionID)
		}
	}
	return users
}

func jdbcConnectionName(connection config.ConnectionConfig) string {
	if name := strings.TrimSpace(connection.Name); name != "" {
		return name
	}
	return connection.ID
}

func isJDBCDriverProfileInUse(driverID string, target config.JDBCDriverProfile, databaseType, profileID string, catalog *service.DriverCatalogService) bool {
	if databaseType != driverID {
		return false
	}
	profileID = strings.TrimSpace(profileID)
	if profileID != "" {
		return profileID == target.ID || profileID == target.Version
	}
	if catalog == nil {
		return false
	}
	_, recommended, err := catalog.GetRecommendedProfile(driverID)
	return err == nil && (recommended.ID == target.ID || recommended.Version == target.Version)
}

func (a *App) GetJDBCRuntimeStatus() (service.RuntimeStatus, error) {
	selected, err := a.jdbcRuntime.SelectRuntime()
	if err != nil {
		return service.RuntimeStatus{}, err
	}
	return service.RuntimeStatus{
		Kind:     selected.Kind,
		JavaPath: selected.JavaPath,
		Version:  selected.Version,
	}, nil
}

func (a *App) InstallJDBCManagedRuntime() (service.JDBCRuntimeActivationResult, error) {
	_, err := a.jdbcRuntime.InstallManagedRuntime(context.Background())
	if err != nil {
		return service.JDBCRuntimeActivationResult{}, err
	}
	return a.activateJDBCRuntime("managed", "")
}

func (a *App) ImportJDBCRuntimeArchive(path string) (service.JDBCRuntimeActivationResult, error) {
	_, err := a.jdbcRuntime.ImportRuntimeArchive(path)
	if err != nil {
		return service.JDBCRuntimeActivationResult{}, err
	}
	return a.activateJDBCRuntime("managed", "")
}

func (a *App) SelectJDBCRuntimeArchive() (string, error) {
	return a.jdbcDialogs().SelectRuntimeArchive(a.ctx)
}

func (a *App) SelectJDBCDriverPackage() (string, error) {
	return a.jdbcDialogs().SelectDriverPackage(a.ctx)
}

func (a *App) SelectJDBCJavaExecutable() (string, error) {
	return a.jdbcDialogs().SelectJavaExecutable(a.ctx)
}

func (a *App) jdbcDialogs() jdbcFileDialogs {
	if a.jdbcFileDialogs == nil {
		return wailsJDBCFileDialogs{}
	}
	return a.jdbcFileDialogs
}

func (a *App) GetJDBCAgentStatus() (service.JDBCAgentStatus, error) {
	if a.jdbcAgentSupervisor == nil {
		return service.JDBCAgentStatus{}, &service.JDBCError{Code: service.JDBCErrorAgentUnavailable, Message: jdbcAgentSupervisorUnavailableMessage}
	}
	status := a.jdbcAgentSupervisor.RefreshStatus(context.Background())
	if a.jdbcRuntime != nil {
		selected, err := a.jdbcRuntime.SelectRuntime()
		if err != nil {
			return service.JDBCAgentStatus{}, err
		}
		status.RuntimeKind = selected.Kind
	}
	return status, nil
}

func (a *App) GetJDBCAgentLogTail(maxBytes int64) (service.JDBCLogTail, error) {
	return service.NewJDBCLogTailService(a.jdbcPaths).Read(maxBytes)
}

func (a *App) SetJDBCRuntimeMode(mode, path string) (service.JDBCRuntimeActivationResult, error) {
	return a.activateJDBCRuntime(mode, path)
}

func (a *App) activateJDBCRuntime(mode, path string) (service.JDBCRuntimeActivationResult, error) {
	if a.jdbcRuntime == nil {
		return service.JDBCRuntimeActivationResult{}, fmt.Errorf("JDBC 运行时服务未初始化")
	}
	if a.jdbcAgentSupervisor == nil {
		return service.JDBCRuntimeActivationResult{}, &service.JDBCError{Code: service.JDBCErrorAgentUnavailable, Message: jdbcAgentSupervisorUnavailableMessage}
	}
	settings := a.jdbcRuntimeSettings
	if settings == nil {
		settings = a.configManager
	}
	if settings == nil {
		return service.JDBCRuntimeActivationResult{}, fmt.Errorf("JDBC 运行时配置存储未初始化")
	}

	before := a.jdbcRuntime.Snapshot()
	if err := a.jdbcRuntime.ApplyMode(mode, path); err != nil {
		return service.JDBCRuntimeActivationResult{}, err
	}
	if err := settings.UpdateJDBCRuntimeSettings(mode, path); err != nil {
		a.jdbcRuntime.Restore(before)
		return service.JDBCRuntimeActivationResult{}, fmt.Errorf("保存 JDBC 运行时设置失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, restartErr := a.jdbcAgentSupervisor.Restart(ctx)
	result, statusErr := a.jdbcRuntimeActivationResult()
	if statusErr != nil {
		return result, statusErr
	}
	return result, restartErr
}

func (a *App) jdbcRuntimeActivationResult() (service.JDBCRuntimeActivationResult, error) {
	selected, err := a.jdbcRuntime.SelectRuntime()
	if err != nil {
		return service.JDBCRuntimeActivationResult{}, err
	}
	return service.JDBCRuntimeActivationResult{
		Runtime: service.RuntimeStatus{Kind: selected.Kind, JavaPath: selected.JavaPath, Version: selected.Version},
		Agent:   a.jdbcAgentSupervisor.Status(),
	}, nil
}

func (a *App) RestartJDBCAgent() error {
	if a.jdbcAgentSupervisor == nil {
		return &service.JDBCError{Code: service.JDBCErrorAgentUnavailable, Message: jdbcAgentSupervisorUnavailableMessage}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := a.jdbcAgentSupervisor.Restart(ctx)
	return err
}

func (a *App) GetTableColumns(sessionID, table string) ([]string, error) {
	schema, err := a.databaseService.GetTableSchema(sessionID, table)
	if err != nil {
		return nil, err
	}

	columns := make([]string, 0, len(schema.Columns))
	for _, col := range schema.Columns {
		columns = append(columns, col.Name)
	}

	return columns, nil
}

// GetTableSchemaInSchema returns structured column metadata for the table structure panel.
func (a *App) GetTableSchemaInSchema(sessionID, database, schema, table string) (*config.TableSchema, error) {
	return a.databaseService.GetTableSchemaInSchema(sessionID, database, schema, table)
}

func (a *App) CloseDatabase(sessionID string) error {
	if a.copilotService != nil {
		a.copilotService.Cancel(sessionID)
	}
	a.clearCopilotSession(sessionID)
	return a.databaseService.CloseDatabase(sessionID)
}

func (a *App) TestDatabaseConnection(host string, port int, user, password, dbType, database string) error {
	return a.databaseService.TestConnection(host, port, user, password, dbType, database)
}

func (a *App) TestDatabaseConnectionWithOptions(host string, port int, user, password, dbType, database string, properties map[string]string) error {
	return a.databaseService.TestConnectionWithProperties(host, port, user, password, dbType, database, properties)
}

// isEncryptedPassword checks if a password is in AES-GCM encrypted format
func isEncryptedPassword(password string) bool {
	decoded, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return false
	}

	return len(decoded) >= 29
}

func generateNewID() string {
	return fmt.Sprintf("conn-%d", time.Now().UnixNano())
}

func derivePassphraseKey(passphrase string, salt []byte) []byte {
	return argon2.IDKey([]byte(passphrase), salt, 3, 64*1024, 4, 32)
}

func encryptWithKey(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decryptWithKey(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GetTableDDL returns the CREATE TABLE statement for a table
func (a *App) GetTableDDL(sessionID, database, table string) (*service.TableDDL, error) {
	return a.databaseService.GetTableDDL(sessionID, database, table)
}

// GetTableDDLInSchema returns table DDL scoped to a PostgreSQL-compatible schema.
func (a *App) GetTableDDLInSchema(sessionID, database, schema, table string) (*service.TableDDL, error) {
	return a.databaseService.GetTableDDLInSchema(sessionID, database, schema, table)
}
