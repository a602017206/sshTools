# sshTools

一个跨平台的SSH桌面客户端工具，使用Go和Wails构建。

**功能齐全** · **安全可靠** · **跨平台** · **现代化UI**

## 特性

### 当前实现 ✨
- ✅ **完整的终端UI** - 基于xterm.js的现代化终端界面
- ✅ **SSH连接管理** - 保存和管理多个SSH连接配置
- ✅ **多种认证方式**
  - 密码认证（支持键盘交互式认证）
  - SSH密钥认证（支持RSA、Ed25519、ECDSA等）
  - 加密密钥Passphrase支持
- ✅ **密码管理**
  - AES-GCM加密存储密码
  - 自动保存和读取密码
  - 机器特征绑定的加密密钥
- ✅ **原生对话框** - Wails原生对话框集成，完美支持桌面端
- ✅ **SSH密钥选择器** - 图形化文件选择器，支持多种密钥格式
- ✅ **实时SSH会话** - PTY支持、完整ANSI颜色、实时双向通信
- ✅ **响应式界面** - 侧边栏连接管理 + 主终端显示区域
- ✅ **终端功能** - 自适应大小、滚动历史、链接点击
- ✅ **配置持久化** - 连接配置和加密凭证本地存储
- ✅ **事件驱动通信** - 前后端实时数据传输

### 计划中 🚀
- 🔸 多标签页（同时连接多个服务器）
- 🔸 SFTP文件传输
- 🔸 端口转发/SSH隧道
- 🔸 跳板机支持
- 🔸 系统密钥链集成（macOS Keychain / Windows Credential Manager）
- 🔸 会话日志记录
- 🔸 终端分屏
- 🔸 主题切换系统

详细的开发计划请查看 [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md)

## 技术栈

- **后端**: Go 1.25
  - Wails v2 - 桌面应用框架
  - golang.org/x/crypto/ssh - SSH协议实现
  - AES-GCM - 密码加密存储
  - Wails Runtime - 原生对话框和文件选择器

- **前端**: Svelte + Vite
  - xterm.js - 终端模拟器
  - 现代化的响应式UI
  - 通过Wails绑定与Go后端通信
  - 自定义模态对话框组件

## 快速开始

### 环境要求

- Go 1.25+
- Node.js 18+
- Wails CLI v2.11.0+

macOS还需要:
- Xcode Command Line Tools

### 安装

```bash
# 安装Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 进入项目目录
cd sshTools

# 安装前端依赖
cd frontend && npm install && cd ..
```

### 开发

```bash
# 开发模式（热重载）
wails dev

# 或使用完整路径
$HOME/go/bin/wails dev

# 构建生产版本
wails build
```

**📖 详细使用说明请查看 [QUICK_START.md](./QUICK_START.md)**

## 使用指南

### 创建 SSH 连接

1. **启动应用**: `wails dev` 或运行构建后的应用
2. **点击"+ 新建连接"**
3. **填写连接信息**:
   - 连接名称
   - 主机地址
   - 端口（默认22）
   - 用户名
4. **选择认证方式**:
   - **密码认证**: 输入密码用于测试
   - **SSH密钥认证**: 点击"选择文件"选择私钥（如 `~/.ssh/id_rsa`）
5. **测试连接**: 点击"测试连接"验证配置
6. **保存**: 点击"保存"按钮

### 连接到服务器

**使用密码认证**:
- 首次连接: 输入密码，可选择"保存密码（加密存储）"
- 再次连接: 自动使用保存的密码，无需重复输入

**使用密钥认证**:
- 如果密钥已加密: 系统会提示输入 Passphrase
- 如果密钥未加密: 直接连接

### 支持的 SSH 密钥格式

- RSA (`id_rsa`)
- Ed25519 (`id_ed25519`) - 推荐
- ECDSA (`id_ecdsa`)
- PEM 格式 (`*.pem`)
- 其他标准 SSH 私钥格式 (`*.key`)

## 项目结构

```
sshTools/
├── main.go              # 应用入口点
├── app.go               # 主应用逻辑和Wails绑定
├── internal/            # 内部包
│   ├── ssh/            # SSH核心功能
│   ├── config/         # 配置管理
│   ├── store/          # 凭证存储
│   ├── terminal/       # 终端模拟器
│   └── crypto/         # 加密相关
├── frontend/           # Svelte前端
│   ├── src/           # 源代码
│   └── wailsjs/       # Wails自动生成的绑定
├── build/             # 构建配置
└── assets/            # 资源文件
```

## 配置

应用配置和数据存储在 `~/.sshtools/` 目录:
- `config.json` - SSH连接配置和应用设置（主题、字体等）
- `credentials.enc` - AES-GCM加密的密码存储（仅当用户选择保存密码时）

**安全说明**：密码使用AES-GCM加密算法存储，加密密钥基于机器特征生成，确保数据安全。

## 安全和隐私

### 密码加密存储
- **加密算法**: AES-256-GCM（Galois/Counter Mode）
- **密钥派生**: 基于机器特征（主机名 + 用户目录）
- **存储位置**: `~/.sshtools/credentials.enc`（权限 0600）
- **数据保护**: 密码在传输和存储过程中始终加密

### 隐私保护
- ✅ 所有数据本地存储，无云端同步
- ✅ SSH 密钥 Passphrase 不会保存到磁盘
- ✅ 用户可选择是否保存密码
- ✅ 连接配置不包含敏感信息

### 安全建议
- 使用 SSH 密钥认证而非密码（更安全）
- 推荐使用 Ed25519 密钥（更快更安全）
- 定期轮换 SSH 密钥和密码
- 为 SSH 私钥设置强 Passphrase

## 开发文档

- [CLAUDE.md](./CLAUDE.md) - 项目架构和开发指南
- [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md) - 详细的开发计划和功能清单

## 构建

```bash
# macOS (Apple Silicon)
wails build -platform darwin/arm64

# macOS (Intel)
wails build -platform darwin/amd64

# Windows
wails build -platform windows/amd64

# Linux
wails build -platform linux/amd64
```

构建后的应用程序位于 `build/bin/` 目录。

## 常见问题

### 连接失败：ssh: unable to authenticate

这通常是认证方式不匹配导致的。解决方案：

1. **检查认证方式**: 确认服务器支持您选择的认证方式（密码或密钥）
2. **密钥认证**: 确保选择了正确的私钥文件
3. **密码认证**: 检查用户名和密码是否正确
4. **服务器配置**: 某些服务器可能禁用了密码认证，需要使用密钥

### 无法保存密码

- 确保应用有权限创建 `~/.sshtools/` 目录
- 检查磁盘空间是否充足
- macOS/Linux: 检查文件权限（应为 0600）

### SSH 密钥认证失败

1. **检查密钥格式**: 确保使用的是私钥文件（不是 `.pub` 公钥）
2. **Passphrase**: 如果密钥已加密，需要输入正确的 Passphrase
3. **密钥权限**: 私钥文件权限应为 600 或 400
4. **公钥配置**: 确保公钥已添加到服务器的 `~/.ssh/authorized_keys`

### 构建失败

```bash
# 清理并重新安装依赖
cd frontend
rm -rf node_modules
npm install
cd ..

# 重新生成绑定
wails dev
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## License

MIT
