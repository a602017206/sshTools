package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/store"
	"github.com/gin-gonic/gin"
)

type testEnv struct {
	router  *gin.Engine
	service *service.ConnectionService
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	t.Setenv("HOME", t.TempDir())

	cm, err := config.NewConfigManager()
	if err != nil {
		t.Fatalf("NewConfigManager: %v", err)
	}
	svc := service.NewConnectionService(cm, store.NewCredentialStore())
	h := NewConnectionHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/connections", h.GetConnections)
	r.POST("/connections", h.AddConnection)
	r.PUT("/connections/:id", h.UpdateConnection)
	r.DELETE("/connections/:id", h.DeleteConnection)
	r.POST("/connections/test", h.TestConnection)
	return &testEnv{router: r, service: svc}
}

func doJSON(t *testing.T, env *testEnv, method, path string, body any) (int, []byte) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	return rec.Code, rec.Body.Bytes()
}

type connData struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	User        string `json:"user"`
	HasPassword bool   `json:"has_password"`
}

func decodeData(t *testing.T, raw []byte) connData {
	t.Helper()
	var resp struct {
		Data connData `json:"data"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("decode response %s: %v", raw, err)
	}
	return resp.Data
}

// TestAddConnectionGeneratesIDAndSavesPassword 覆盖移动端反馈的
// 「加了连接删不掉、改不了、密码不保存」：空 ID 新增必须自动生成 ID，
// 密码必须加密保存且只回 has_password 而不是密码本身。
func TestAddConnectionGeneratesIDAndSavesPassword(t *testing.T) {
	env := setupTestEnv(t)

	code, raw := doJSON(t, env, http.MethodPost, "/connections", map[string]any{
		"name": "srv", "host": "127.0.0.1", "port": 22, "user": "root",
		"auth_type": "password", "password": "s3cret",
	})
	if code != http.StatusCreated {
		t.Fatalf("add status = %d, body = %s", code, raw)
	}
	created := decodeData(t, raw)
	if created.ID == "" {
		t.Fatal("server must generate an ID when client sends empty ID")
	}
	if !created.HasPassword {
		t.Fatal("has_password should be true after adding with a password")
	}
	if bytes.Contains(raw, []byte("s3cret")) {
		t.Fatal("response must never contain the plaintext password")
	}

	// 更新用户名（此前空 ID 时 PUT 会 404，用户感知为「改不了」）
	code, raw = doJSON(t, env, http.MethodPut, "/connections/"+created.ID, map[string]any{
		"name": "srv", "host": "127.0.0.1", "port": 22, "user": "ubuntu",
		"auth_type": "password",
	})
	if code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", code, raw)
	}
	if updated := decodeData(t, raw); updated.User != "ubuntu" {
		t.Fatalf("username not updated: %q", updated.User)
	}
	// 未带密码更新不应清掉已存密码
	if !decodeData(t, raw).HasPassword {
		t.Fatal("update without password must keep the saved password")
	}

	// 测试连接不给 auth_value 时应回退到已保存密码：
	// 此处连一个不存在的端口，期望拿到「连接失败」而不是「缺凭证」，
	// 以此证明密码回退路径被走通（400 + connection failed 文案）。
	code, raw = doJSON(t, env, http.MethodPost, "/connections/test", map[string]any{
		"host": "127.0.0.1", "port": 1, "user": "root",
		"auth_type": "password", "connection_id": created.ID,
	})
	if code != http.StatusBadRequest || !bytes.Contains(raw, []byte("connection failed")) {
		t.Fatalf("test fallback status = %d, body = %s", code, raw)
	}

	// 删除后连接与密码应一并消失
	code, _ = doJSON(t, env, http.MethodDelete, "/connections/"+created.ID, nil)
	if code != http.StatusOK {
		t.Fatalf("delete status = %d", code)
	}
	if env.service.HasPassword(created.ID) {
		t.Fatal("saved password must be removed together with the connection")
	}
	code, raw = doJSON(t, env, http.MethodGet, "/connections", nil)
	if code != http.StatusOK || !bytes.Contains(raw, []byte(`"data":[]`)) && !bytes.Contains(raw, []byte(`"data": []`)) {
		// 空列表在不同序列化下空格可能不同，再做一次结构化断言
		var listResp struct {
			Data []connData `json:"data"`
		}
		if err := json.Unmarshal(raw, &listResp); err != nil || len(listResp.Data) != 0 {
			t.Fatalf("connections should be empty after delete, body = %s", raw)
		}
	}
}

// TestEnsureConnectionIDsBackfillsLegacyData 覆盖已装机用户存量数据：
// 旧版保存的空 ID 连接在启动修复后应获得 ID，从而能被删除。
func TestEnsureConnectionIDsBackfillsLegacyData(t *testing.T) {
	env := setupTestEnv(t)

	// 绕过 handler、直接经服务层写入一条空 ID 的旧数据（模拟旧版 App 的落盘结果）
	if err := env.service.AddConnection(config.ConnectionConfig{
		Name: "legacy", Host: "10.0.0.1", Port: 22, User: "root", AuthType: "password",
	}); err != nil {
		t.Fatalf("seed legacy connection: %v", err)
	}
	conns, err := env.service.GetConnections()
	if err != nil || len(conns) != 1 {
		t.Fatalf("GetConnections = %v, %v", conns, err)
	}
	if conns[0].ID != "" {
		t.Fatalf("seed data should have empty ID, got %q", conns[0].ID)
	}

	if err := env.service.EnsureConnectionIDs(); err != nil {
		t.Fatalf("EnsureConnectionIDs: %v", err)
	}
	conns, err = env.service.GetConnections()
	if err != nil || len(conns) != 1 {
		t.Fatalf("GetConnections after backfill = %v, %v", conns, err)
	}
	if conns[0].ID == "" {
		t.Fatal("legacy connection should have an ID after backfill")
	}

	// 回填出的 ID 必须真实落盘可用：能据此删掉这条连接
	code, _ := doJSON(t, env, http.MethodDelete, "/connections/"+conns[0].ID, nil)
	if code != http.StatusOK {
		t.Fatalf("delete backfilled connection status = %d", code)
	}
}
