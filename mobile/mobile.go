// Package mobile 是供 gomobile 绑定的移动端入口。
//
// 它把桌面端 cmd/apiserver 的服务装配（配置、凭证、SSH 会话管理）
// 以库的形式内嵌进 Android/iOS 应用进程内运行，而不是像桌面端那样
// 拉起一个独立的 API 服务器进程——iOS 不允许子进程，Android 也不适合
// 以二进制方式常驻。Flutter 侧通过 MethodChannel 调用 Start/Stop，
// 然后照常走 http://127.0.0.1:<port>/api/v1 与 ws://.../api/v1/ws。
//
// 本包只暴露 gomobile 支持的签名（string/int/bool/error），
// 供 `gomobile bind` 生成 AAR / XCFramework。
package mobile

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"AHaSSHTools/internal/api"
	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/ssh"
	"AHaSSHTools/internal/store"
)

var (
	mu      sync.Mutex
	server  *api.Server
	port    int
	started bool
)

// Start 在进程内启动内嵌 API 服务器，仅监听 127.0.0.1 的随机可用端口。
//
// dataDir 是原生侧（Flutter）传入的应用沙盒目录。现有 config/store 都是
// 按 $HOME/.ahasshtools 解析路径的，因此这里在装配服务前把 HOME 指到
// 沙盒目录，使配置与凭证自然落进应用私有空间；这是当前最小侵入的做法，
// 后续若给 config 增加显式 DataDir 参数，应优先改用显式参数。
//
// 重复调用 Start 是幂等的：已运行时直接返回 nil。
func Start(dataDir string) error {
	mu.Lock()
	defer mu.Unlock()

	if started {
		return nil
	}

	if dataDir == "" {
		return fmt.Errorf("dataDir is empty")
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	if err := os.Setenv("HOME", dataDir); err != nil {
		return fmt.Errorf("set HOME: %w", err)
	}

	freePort, err := pickFreeLoopbackPort()
	if err != nil {
		return err
	}

	configManager, err := config.NewConfigManager()
	if err != nil {
		return fmt.Errorf("init config manager: %w", err)
	}
	credentialStore := store.NewCredentialStore()
	sessionManager := ssh.NewSessionManager()
	transferManager := ssh.NewTransferManager()

	connectionService := service.NewConnectionService(configManager, credentialStore)
	// 修复早期版本存量数据中缺少 ID 的连接（否则删除/更新都会失败）
	if err := connectionService.EnsureConnectionIDs(); err != nil {
		return fmt.Errorf("ensure connection ids: %w", err)
	}

	services := &api.Services{
		Connection: connectionService,
		Session:    service.NewSessionService(sessionManager),
		SFTP:       service.NewSFTPService(sessionManager, transferManager),
		Monitor:    service.NewMonitorService(sessionManager),
		Settings:   service.NewSettingsService(configManager),
	}

	srv := api.NewServerWithHost(services, "127.0.0.1", freePort)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	// 等待服务真正可应答（或启动失败），避免 Flutter 侧拿到端口后立即请求被拒。
	healthURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/health", freePort)
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-errCh:
			return fmt.Errorf("api server exited early: %w", err)
		default:
		}

		resp, err := http.Get(healthURL) //nolint:bodyclose // 探活请求，立即关闭
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				server = srv
				port = freePort
				started = true
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	_ = srv.Shutdown(context.Background())
	return fmt.Errorf("api server did not become ready on port %d", freePort)
}

// Stop 优雅关闭内嵌 API 服务器。未运行时调用是无操作。
func Stop() {
	mu.Lock()
	defer mu.Unlock()

	if !started || server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	server = nil
	port = 0
	started = false
}

// Port 返回内嵌服务器监听的端口；未运行时返回 0。
func Port() int {
	mu.Lock()
	defer mu.Unlock()
	return port
}

// IsRunning 报告内嵌服务器是否正在运行。
func IsRunning() bool {
	mu.Lock()
	defer mu.Unlock()
	return started
}

// DataDir 返回当前生效的数据目录（即 Start 传入的沙盒目录），
// 仅用于诊断与日志。
func DataDir() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".ahasshtools")
}

func pickFreeLoopbackPort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("pick free port: %w", err)
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port, nil
}
