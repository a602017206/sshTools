# sshTools（AHaSSHTools）

[![最新版本](https://img.shields.io/github/v/release/a602017206/sshTools?label=最新版本)](https://github.com/a602017206/sshTools/releases/latest)
[![总下载量](https://img.shields.io/github/downloads/a602017206/sshTools/total?label=总下载量)](https://github.com/a602017206/sshTools/releases)
[![最新版下载量](https://img.shields.io/github/downloads/a602017206/sshTools/latest/total?label=最新版下载量)](https://github.com/a602017206/sshTools/releases/latest)
![访问次数](https://hits.sh/github.com/a602017206/sshTools.svg?label=%E8%AE%BF%E9%97%AE%E6%AC%A1%E6%95%B0&color=0e75b6&style=flat)

跨平台运维桌面客户端：SSH 终端、SFTP、关系库 / 缓存 / 搜索 / 消息队列，以及 AI 助手与本地开发工具，基于 **Go + Wails + Svelte** 构建。

## 下载

| 平台 | 安装包 | 最新版下载量 |
| --- | --- | --- |
| Windows | [`AHaSSHTools.exe`](https://github.com/a602017206/sshTools/releases/latest) | ![Windows](https://img.shields.io/github/downloads/a602017206/sshTools/latest/AHaSSHTools.exe?label=下载) |
| macOS（Apple Silicon） | [`AHaSSHTools-macos-arm64.zip`](https://github.com/a602017206/sshTools/releases/latest) | ![macOS arm64](https://img.shields.io/github/downloads/a602017206/sshTools/latest/AHaSSHTools-macos-arm64.zip?label=下载) |
| macOS（Intel） | [`AHaSSHTools-macos-amd64.zip`](https://github.com/a602017206/sshTools/releases/latest) | ![macOS amd64](https://img.shields.io/github/downloads/a602017206/sshTools/latest/AHaSSHTools-macos-amd64.zip?label=下载) |

macOS 下载后若无法打开，在终端执行：

```bash
xattr -cr AHaSSHTools.app && open AHaSSHTools.app
```

徽章由 [shields.io](https://shields.io) 读取 GitHub Releases（仅统计安装包）；访问次数由 [hits.sh](https://hits.sh) 统计。更多版本见 [Releases](https://github.com/a602017206/sshTools/releases)。

## 功能概览

按成熟度区分：**完整** ≈ 日常可用；**可用** ≈ 主流程可用，能力仍在加深；**雏形** ≈ 可连通或只读浏览，深度操作有限。

### SSH 与终端 — 完整

- 多标签会话、重命名、批量关闭（全部 / 左侧 / 右侧 / 其它）
- 密码 / SSH 密钥（RSA、Ed25519、ECDSA）与 Passphrase；密码可选 AES-GCM 本地加密保存
- 终端编码可选（UTF-8 / GBK 等），会话内可切换
- 界面主题与终端主题分离（浅色 / 深色 / 跟随界面）
- 会话日志：自动记录、搜索、导出、保留天数与敏感信息过滤
- 常用命令提示：按连接统计频率，Tab / 点击填入后再执行
- 本地 Shell；资产树支持分组文件夹与克隆连接

### 文件管理（SFTP）— 完整

- 目录浏览、面包屑、收藏路径、目录跟踪
- 上传 / 下载 / 删除 / 重命名 / 新建目录；支持文件夹上传与粘贴本地文件
- 传输进度、取消、同名冲突处理（覆盖 / 重命名）
- 远程文本在线编辑（语法高亮、格式化、Ctrl+S / ⌘S 保存）

### 系统监控 — 完整

- CPU（含核心）、内存、磁盘、网络与主机信息
- 「实时」开关（按会话，默认关）与手动刷新，切走面板时停止采集

### 关系型数据库（JDBC）— 可用

通过内置 JDBC Agent（需本机或托管 JRE + 驱动）：

| 类型 | 说明 |
| --- | --- |
| MySQL、PostgreSQL、SQLite | 对象浏览、SQL、表数据与结构 |
| Oracle、SQL Server | 同上；Oracle 支持服务名 / SID |
| 达梦、人大金仓、openGauss | 国产库，能力持续对齐 |

主要能力：库 / Schema 对象树、SQL 查询、表数据分页与筛选排序、单元格编辑与批量改删、表设计器（按方言字段类型）、运行 SQL 文件、空闲断连自动重连。

### 缓存 · 搜索 · 消息 — 可用 / 雏形

| 类型 | 成熟度 | 现状 |
| --- | --- | --- |
| Redis / KeyDB | 可用 | 逻辑库、键扫描、类型预览编辑、CLI（白名单） |
| Elasticsearch / OpenSearch | 可用 | 索引浏览、DSL 查询、建删索引等 |
| Kafka、RocketMQ、RabbitMQ | 可用 | 连通与 Topic / Queue 元数据浏览（不生产/消费消息） |
| MongoDB、Memcached、Cassandra、Couchbase、InfluxDB、Neo4j | 雏形 | 多为只读资源浏览，深度操作有限 |

### AI 助手（Copilot）— 可用

- 多服务商 / 多模型（DeepSeek、OpenAI、Ollama、自定义），密钥按服务商加密存储
- 可按服务商拉取模型列表；对话旁切换模型
- 结合当前 SSH / 库表 / Redis·ES 等上下文，支持只读查询与确认后执行变更

### 开发工具 — 完整

JSON 格式化、Base64、URL 编解码、哈希 / AES / SM4、时间戳、UUID。

### 尚未完成 / 入口预留

- Docker 容器连接：界面有入口，后端能力未就绪
- SSH 端口转发 / 隧道、ProxyJump 跳板机
- 系统密钥链、标签拖拽排序与会话持久化、终端分屏等

更细的阶段清单见 [DEVELOPMENT_PLAN.md](./DEVELOPMENT_PLAN.md)（部分条目可能滞后于代码，以本 README「功能概览」为准）。

## 技术栈

| 层级 | 技术 |
| --- | --- |
| 桌面壳 | Wails v2 |
| 后端 | Go 1.24、`x/crypto/ssh`、SFTP、JDBC Agent（Java）、部分原生 Go 客户端（Redis / Kafka 等） |
| 前端 | Svelte + Vite、xterm.js |
| 数据 | `~/.sshtools/`（或兼容路径）本地配置与加密凭据；会话日志目录见应用设置 |

## 快速开始

### 环境

- Go 1.24+
- Node.js 18+
- [Wails CLI](https://wails.io) v2

macOS 需 Xcode Command Line Tools。关系库功能还需可用的 Java 运行时（应用内可安装 / 导入）。

### 开发与构建

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd frontend && npm install && cd ..
wails dev          # 热重载开发
wails build        # 生产构建，输出 build/bin/
./scripts/build-mac.sh   # macOS 分发（ad-hoc 签名）
```

跨平台示例：

```bash
wails build -platform darwin/arm64
wails build -platform darwin/amd64
wails build -platform windows/amd64
wails build -platform linux/amd64
```

上手步骤见 [QUICK_START.md](./QUICK_START.md)。

## 文档

| 文档 | 说明 |
| --- | --- |
| [QUICK_START.md](./QUICK_START.md) | 安装与基本使用 |
| [CLAUDE.md](./CLAUDE.md) / [AGENTS.md](./AGENTS.md) | 架构与开发约定 |
| [DEVTOOLS_GUIDE.md](./DEVTOOLS_GUIDE.md) | 开发工具说明 |
| [MACOS_SIGNING.md](./MACOS_SIGNING.md) | macOS 签名与分发 |
| [docs/README.md](./docs/README.md) | 设计 / 开发 / 变更文档地图 |

## 安全与隐私

- 连接配置与密码默认仅存本机；密码可选 AES-GCM 加密（密钥与机器特征相关）
- SSH 私钥 Passphrase 不落盘
- 无强制云同步；AI 仅在使用 Copilot 时按你配置的服务商发出请求
- 会话日志可关闭，并支持敏感信息过滤

## 欢迎赞助

如果项目对你有帮助，欢迎扫码赞助，支持继续维护。

<img src="zfb.png" alt="支付宝" width="200"/>
<img src="wx.png" alt="微信支付" width="200"/>

## 贡献与许可

欢迎提 [Issue](https://github.com/a602017206/sshTools/issues) 与 Pull Request。

Apache License 2.0
