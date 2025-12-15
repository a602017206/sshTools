# sshTools

一个跨平台的SSH桌面客户端工具，使用Go和Wails构建。

## 特性

### 当前实现 ✨
- ✅ **完整的终端UI** - 基于xterm.js的现代化终端界面
- ✅ **SSH连接管理** - 保存和管理多个SSH连接配置
- ✅ **实时SSH会话** - 密码认证、PTY支持、完整ANSI颜色
- ✅ **响应式界面** - 侧边栏连接管理 + 主终端显示区域
- ✅ **终端功能** - 自适应大小、滚动历史、链接点击
- ✅ **配置持久化** - 连接配置保存在本地
- ✅ **事件驱动通信** - 前后端实时数据传输

### 计划中 🚀
- 🔸 SSH密钥认证
- 🔸 多标签页（同时连接多个服务器）
- 🔸 SFTP文件传输
- 🔸 端口转发/SSH隧道
- 🔸 跳板机支持
- 🔸 系统密钥链集成
- 🔸 会话日志记录
- 🔸 终端分屏
- 🔸 主题切换系统

详细的开发计划请查看 [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md)

## 技术栈

- **后端**: Go 1.25
  - Wails v2 - 桌面应用框架
  - golang.org/x/crypto/ssh - SSH协议实现

- **前端**: Svelte + Vite
  - 现代化的响应式UI
  - 通过Wails绑定与Go后端通信

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

应用配置存储在 `~/.sshtools/config.json`，包含:
- SSH连接配置
- 应用设置（主题、字体等）

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

## License

MIT
