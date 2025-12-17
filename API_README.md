# sshTools API Server

## 概述

sshTools 现在提供两种前端接入方式：

1. **Wails 桌面版**（Svelte + xterm.js）- 适用于 macOS/Windows/Linux 桌面
2. **REST API + WebSocket**（新增）- 适用于 Flutter 全平台前端

## 快速启动

### 构建 API 服务器

```bash
# 构建
./scripts/build-api-server.sh

# 或者直接使用 go build
go build -o build/bin/sshtools-api cmd/apiserver/main.go
```

### 运行 API 服务器

```bash
# 默认端口 8080
./build/bin/sshtools-api

# 自定义端口
PORT=3000 ./build/bin/sshtools-api
```

服务器启动后：
- **HTTP API**: `http://localhost:8080/api/v1`
- **WebSocket**: `ws://localhost:8080/api/v1/ws`
- **健康检查**: `http://localhost:8080/api/v1/health`

---

## REST API 端点

### 连接管理

#### 获取所有连接
```http
GET /api/v1/connections
```

**响应示例：**
```json
{
  "data": [
    {
      "id": "conn_1",
      "name": "Production Server",
      "host": "192.168.1.100",
      "port": 22,
      "user": "root",
      "auth_type": "password",
      "tags": ["production"]
    }
  ]
}
```

#### 添加连接
```http
POST /api/v1/connections
Content-Type: application/json

{
  "name": "My Server",
  "host": "192.168.1.100",
  "port": 22,
  "user": "admin",
  "auth_type": "password"
}
```

#### 更新连接
```http
PUT /api/v1/connections/:id
Content-Type: application/json

{
  "name": "Updated Server Name",
  "host": "192.168.1.101",
  "port": 22,
  "user": "admin",
  "auth_type": "password"
}
```

#### 删除连接
```http
DELETE /api/v1/connections/:id
```

#### 测试连接
```http
POST /api/v1/connections/test
Content-Type: application/json

{
  "host": "192.168.1.100",
  "port": 22,
  "user": "root",
  "auth_type": "password",
  "auth_value": "your_password"
}
```

### SSH 会话

#### 建立 SSH 连接
```http
POST /api/v1/sessions/connect
Content-Type: application/json

{
  "session_id": "session_123",
  "host": "192.168.1.100",
  "port": 22,
  "user": "root",
  "auth_type": "password",
  "auth_value": "your_password",
  "cols": 80,
  "rows": 24
}
```

**响应示例：**
```json
{
  "data": {
    "session_id": "session_123",
    "message": "SSH session started: root@192.168.1.100:22"
  }
}
```

#### 发送数据到会话
```http
POST /api/v1/sessions/:id/send
Content-Type: application/json

{
  "data": "ls -la\n"
}
```

#### 调整终端大小
```http
POST /api/v1/sessions/:id/resize
Content-Type: application/json

{
  "cols": 120,
  "rows": 30
}
```

#### 关闭会话
```http
DELETE /api/v1/sessions/:id
```

#### 列出所有会话
```http
GET /api/v1/sessions
```

---

## WebSocket 协议

### 连接

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws');
```

### 客户端 → 服务器消息

#### 订阅 SSH 会话输出
```json
{
  "action": "subscribe",
  "target": "session_123"
}
```

#### 取消订阅
```json
{
  "action": "unsubscribe",
  "target": "session_123"
}
```

### 服务器 → 客户端消息

#### SSH 输出
```json
{
  "type": "ssh:output",
  "session_id": "session_123",
  "data": "Welcome to Ubuntu 20.04 LTS\n",
  "timestamp": 1702345678
}
```

#### 文件传输进度
```json
{
  "type": "transfer:progress",
  "transfer_id": "transfer_456",
  "data": {
    "filename": "file.txt",
    "bytes_sent": 512000,
    "total_bytes": 1024000,
    "percentage": 50.0,
    "speed": 102400,
    "status": "running"
  },
  "timestamp": 1702345678
}
```

---

## 完整的 SSH 会话示例

### 1. 建立连接
```bash
curl -X POST http://localhost:8080/api/v1/sessions/connect \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "my_session",
    "host": "192.168.1.100",
    "port": 22,
    "user": "root",
    "auth_type": "password",
    "auth_value": "password123",
    "cols": 80,
    "rows": 24
  }'
```

### 2. 通过 WebSocket 接收输出

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws');

ws.onopen = () => {
  // 订阅会话输出
  ws.send(JSON.stringify({
    action: 'subscribe',
    target: 'my_session'
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  if (message.type === 'ssh:output' && message.session_id === 'my_session') {
    console.log('SSH Output:', message.data);
    // 在终端显示 message.data
  }
};
```

### 3. 发送命令
```bash
curl -X POST http://localhost:8080/api/v1/sessions/my_session/send \
  -H "Content-Type: application/json" \
  -d '{"data": "ls -la\n"}'
```

### 4. 关闭会话
```bash
curl -X DELETE http://localhost:8080/api/v1/sessions/my_session
```

---

## 架构说明

### 目录结构
```
internal/
├── service/           # 业务逻辑层（Wails 和 REST API 共用）
│   ├── connection_service.go
│   ├── session_service.go
│   ├── sftp_service.go
│   ├── monitor_service.go
│   └── settings_service.go
│
├── api/              # REST API 层
│   ├── server.go     # HTTP 服务器
│   ├── middleware.go # CORS/日志/恢复中间件
│   ├── handlers/     # REST API 处理器
│   │   ├── connection.go
│   │   └── session.go
│   ├── websocket/    # WebSocket 支持
│   │   ├── hub.go    # 连接池管理
│   │   ├── client.go # 客户端连接
│   │   └── message.go# 消息协议
│   └── dto/          # 数据传输对象
│       └── response.go
│
└── ...
```

### 通信流程

```
Flutter 前端
    ↓ HTTP/WebSocket
API 服务器 (Gin + gorilla/websocket)
    ↓ 服务层调用
业务逻辑服务 (service/)
    ↓
SSH/SFTP 核心模块 (internal/ssh/)
```

---

## 开发计划

### 已完成 ✅
- [x] 服务层抽象（连接、会话、SFTP、监控、设置）
- [x] REST API 基础框架（Gin）
- [x] WebSocket 支持（Hub + Client）
- [x] 连接管理 API
- [x] SSH 会话 API
- [x] 构建脚本

### 待实现 🚧
- [ ] SFTP 文件操作 API
- [ ] 系统监控 API
- [ ] 设置管理 API
- [ ] 凭证管理 API
- [ ] Flutter 前端实现

---

## 注意事项

1. **安全性**：
   - 当前 CORS 允许所有来源（`*`），生产环境需要限制
   - 建议添加 JWT 认证机制
   - 使用 HTTPS/WSS 加密通信

2. **性能**：
   - WebSocket 连接池支持多会话并发
   - 心跳机制：60 秒无活动自动断开
   - 消息缓冲区：256 条消息

3. **配置**：
   - 配置文件：`~/.sshtools/config.json`
   - 凭证存储：`~/.sshtools/credentials.enc`（AES-GCM 加密）

---

## 测试

```bash
# 健康检查
curl http://localhost:8080/api/v1/health

# 获取连接列表
curl http://localhost:8080/api/v1/connections

# 测试 SSH 连接
curl -X POST http://localhost:8080/api/v1/connections/test \
  -H "Content-Type: application/json" \
  -d '{
    "host": "192.168.1.100",
    "port": 22,
    "user": "root",
    "auth_type": "password",
    "auth_value": "password123"
  }'
```

---

## 故障排查

### API 服务器无法启动
- 检查端口 8080 是否被占用：`lsof -i :8080`
- 使用自定义端口：`PORT=3000 ./build/bin/sshtools-api`

### WebSocket 连接失败
- 确认服务器已启动
- 检查 WebSocket URL：`ws://localhost:8080/api/v1/ws`
- 查看浏览器控制台错误信息

### SSH 连接失败
- 验证主机、端口、用户名和认证信息
- 使用 `/connections/test` 端点测试连接
- 检查服务器日志输出

---

## 许可证

与主项目相同。
