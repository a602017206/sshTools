package mobile

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

// TestStartStopEmbeddedServer 验证内嵌服务器能在沙盒目录下启动、
// 应答健康检查并被干净关闭。移动端 Flutter 侧依赖的正是这条路径。
func TestStartStopEmbeddedServer(t *testing.T) {
	dataDir := t.TempDir()

	if IsRunning() {
		t.Fatal("server should not be running before Start")
	}
	if err := Start(dataDir); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer Stop()

	if !IsRunning() {
		t.Fatal("server should be running after Start")
	}
	p := Port()
	if p <= 0 {
		t.Fatalf("invalid port %d", p)
	}

	// 重复 Start 应幂等
	if err := Start(dataDir); err != nil {
		t.Fatalf("second Start should be a no-op, got: %v", err)
	}
	if Port() != p {
		t.Fatalf("port changed after second Start: %d -> %d", p, Port())
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/api/v1/health", p)
	deadline := time.Now().Add(5 * time.Second)
	var resp *http.Response
	var err error
	for time.Now().Before(deadline) {
		resp, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("health check failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health check status = %d", resp.StatusCode)
	}

	if got := DataDir(); got == "" {
		t.Fatal("DataDir should not be empty after Start")
	}

	Stop()
	if IsRunning() {
		t.Fatal("server should not be running after Stop")
	}
	if Port() != 0 {
		t.Fatalf("port should reset to 0 after Stop, got %d", Port())
	}
}

func TestStartRejectsEmptyDataDir(t *testing.T) {
	if err := Start(""); err == nil {
		t.Fatal("Start with empty dataDir should fail")
	}
}
