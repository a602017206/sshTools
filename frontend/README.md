# SSH Tools Frontend New

基于 Figma 设计的新前端，使用 **Svelte 4 + Vite 5 + Tailwind CSS 3** 构建，集成 **Wails v2** 后端。

## ✅ 功能特性

### 已实现功能

**资产列表管理**
- ✅ 分组显示（按环境分组：生产、开发、测试）
- ✅ 搜索过滤功能（实时搜索名称、主机、用户名、分组）
- ✅ 添加/编辑/删除资产
- ✅ 连接状态指示灯（在线绿色、离线灰色）
- ✅ 支持多种资产类型（SSH、Database、Docker）
- ✅ 动态端口默认值（SSH:22, MySQL:3306, PostgreSQL:5432 等）
- ✅ 悬停显示编辑/删除按钮

**终端面板**
- ✅ 多标签页支持（无数量限制）
- ✅ 标签切换（单击切换到对应终端）
- ✅ 标签关闭（关闭前确认）
- ✅ 标签重命名（双击编辑名称）
- ✅ **xterm.js 集成**：
  - 完整的终端模拟
  - 支持 ANSI 颜色代码
  - 光标闪烁
  - 自动适应容器大小
  - WebLinks 链接点击
  - 支持浅色/深色主题
- ✅ 标签页工具栏（复制、最小化、最大化）
- ✅ 终端尺寸与后端同步
- ✅ **SSH 连接功能**（通过 Wails API）：
  - `ConnectSSH` - 建立连接
  - `SendSSHData` - 发送数据
  - `ResizeSSH` - 调整终端大小
  - `CloseSSH` - 关闭会话
- ✅ Wails 绑定动态加载（支持独立开发）

**文件管理器**
- ✅ 文件列表显示（图标区分文件夹/文件）
- ✅ 目录导航（面包屑）
- ✅ 文件权限显示（drwxr-xr-x）
- ✅ 刷新按钮
- ✅ 上传/下载/删除按钮
- ✅ 文件大小和修改时间显示
- ✅ **API 调用集成**（待 wails dev 生成绑定后启用）：
  - `ListFiles` - 列出文件
  - `ChangeDirectory` - 切换目录
  - `UploadFile` - 上传文件
  - `DownloadFile` - 下载文件
  - `SelectUploadFiles` - 选择上传文件
  - `SelectDownloadDirectory` - 选择下载目录

**服务器监控**
- ✅ CPU/内存/磁盘/网络监控卡片
- ✅ 实时数据更新（每 2 秒轮询）
- ✅ SVG 图表（CPU 和 内存曲线图）
- ✅ 状态颜色指示（绿色 <50%, 黄色 <80%, 红色 >80%）
- ✅ 磁盘使用进度条
- ✅ 网络流量实时显示（入站/出站）
- ✅ 系统信息（OS、内核、运行时间、进程数）
- ✅ **API 集成**（待 wails dev 生成绑定后启用）：
  - `GetMonitoringData` - 获取监控数据

**开发工具集**
- ✅ 工具列表展示（抽屉式面板）
- ✅ JSON 格式化工具（实时验证）
- ✅ 其他工具占位（Base64、Hash、时间戳、UUID）
- ✅ 模态框/抽屉式 UI

**主应用**
- ✅ 三栏布局（资产列表 + 终端 + 监控/文件）
- ✅ 顶部标题栏（Logo + 开发工具按钮）
- ✅ **主题系统**（深色/浅色切换，支持自动检测系统主题）
- ✅ 侧边栏可折叠
- ✅ 侧边栏宽度可拖动调整
- ✅ 对话框组件（添加资产）

## 🚀 快速开始

### 前置要求

- Node.js 18+
- Go 1.25+
- Wails CLI v2.11.0+

### 安装依赖

```bash
cd frontend
npm install
```

### 开发模式

#### 1. 完整 Wails 开发（推荐）

```bash
# 在项目根目录
cd /Users/dingwei/go/sshTools
wails dev
```

这将：
- 启动 Go 后端
- 运行 Vite 开发服务器（http://localhost:5174）
- 自动生成 Wails 绑定到 `frontend_new/wailsjs/`
- 启用热重载

#### 2. 独立前端开发（不连接后端）

```bash
cd frontend
npm run dev
```

前端将在 http://localhost:5174 运行，但后端 API 不可用（会显示友好提示）。

### 构建

```bash
cd frontend
npm run build
```

构建输出到 `../build/frontend_new/`。

### Wails 构建

```bash
wails build
```

构建桌面应用，包含 Go 后端和前端。

## 📁 项目结构

```
frontend_new/
├── src/
│   ├── components/          # 组件
│   │   ├── ui/           # 基础 UI 组件
│   │   │   ├── Button.svelte
│   │   │   ├── Input.svelte
│   │   │   ├── Dialog.svelte
│   │   │   └── Select.svelte
│   │   ├── AssetList.svelte        # 资产列表
│   │   ├── TerminalPanel.svelte     # 终端面板（标签）
│   │   ├── Terminal.svelte          # 终端组件（xterm.js）
│   │   ├── FileManager.svelte       # 文件管理器
│   │   ├── ServerMonitor.svelte     # 服务器监控
│   │   ├── DevToolsPanel.svelte    # 开发工具集
│   │   └── AddAssetDialog.svelte    # 添加资产对话框
│   ├── stores/             # Svelte Store 状态管理
│   │   └── stores.js
│   ├── styles/             # 样式文件
│   │   └── app.css
│   ├── App.svelte          # 主组件
│   └── main.js             # 入口文件
├── package.json
├── vite.config.js
├── svelte.config.js
├── tailwind.config.js
├── postcss.config.js
├── index.html
└── README.md
```

## 🛠 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Svelte | 4.2.0 | 前端框架 |
| Vite | 5.0.0 | 构建工具 |
| Tailwind CSS | 3.4.0 | 样式框架 |
| xterm.js | 5.5.0 | 终端模拟器 |
| Wails | 2.x | 桌面应用框架 |
| Go | 1.25+ | 后端语言 |

## 🎨 主题系统

支持深色/浅色主题切换，使用 Tailwind CSS `dark:` 前缀：

```javascript
// src/stores.js
import { themeStore } from './stores.js';

function toggleTheme() {
  themeStore.update(t => t === 'light' ? 'dark' : 'light');
}
```

### 终端主题

xterm.js 支持独立主题配置，会自动跟随应用主题切换。

## 📊 状态管理

使用 Svelte Store 进行全局状态管理：

```javascript
// src/stores.js
export const assetsStore = writable([]);
export const themeStore = writable('light');
export const connectionsStore = writable(new Map());
export const activeSessionIdStore = writable(null);
```

## 🧪 Wails API 集成

### 动态加载绑定

在 `App.svelte` 中动态加载 Wails 绑定：

```javascript
onMount(async () => {
  try {
    const wails = await import('../wailsjs/go/main/App.js');
    window.wailsBindings = wails;
    console.log('Wails bindings loaded successfully');
  } catch (error) {
    console.warn('Wails bindings not available (run with wails dev):', error);
  }
});
```

### 已集成的 API

- **SSH 连接**: `ConnectSSH(sessionId, host, port, user, authType, authValue, passphrase, cols, rows)`
- **发送数据**: `SendSSHData(sessionId, data)`
- **调整大小**: `ResizeSSH(sessionId, cols, rows)`
- **关闭连接**: `CloseSSH(sessionId)`
- **监控数据**: `GetMonitoringData(sessionId)`
- **文件操作**: `ListFiles`, `ChangeDirectory`, `UploadFile`, `DownloadFile`, `DeleteFile`, `RenameFile`, `CreateDirectory`
- **资产管理**: `GetConnections`, `AddConnection`, `UpdateConnection`, `RemoveConnection`
- **开发工具**: `FormatJSON`, `ValidateJSON`, `MinifyJSON`, `EscapeJSON`

## 🔧 开发工具集

### JSON 格式化工具

```javascript
// 输入
const input = '{"name": "test"}';

// 格式化
const formatted = JSON.stringify(JSON.parse(input), null, 2);

// 输出
// {
//   "name": "test"
// }
```

### 工具列表

- ✅ JSON 格式化（已实现）
- ⏸️ Base64 编解码（UI 占位，待实现后端）
- ⏸️ Hash 计算（UI 占位，待实现后端）
- ⏸️ 时间戳转换（UI 占位，待实现后端）
- ⏸️ UUID 生成（UI 占位，待实现后端）

## 📊 监控数据格式

```typescript
interface MonitoringData {
  cpu: number;        // CPU 使用率 0-100
  memory: number;     // 内存使用率 0-100
  disk: number;        // 磁盘使用率 0-100
  network: {
    in: number;    // 入站流量 MB/s
    out: number;   // 出站流量 MB/s
  };
}
```

## 🎯 下一步开发

### 优先级：高

1. **完善开发工具**
   - 实现 Base64 编解码
   - 实现 Hash 计算
   - 实现时间戳转换
   - 实现 UUID 生成

2. **完善文件管理**
   - 实现文件选择对话框
   - 实现传输进度显示
   - 添加批量操作

3. **增强终端功能**
   - 添加终端配置（字体大小、光标样式等）
   - 实现终端日志保存
   - 添加快捷键支持

### 优先级：中

4. **改进图表**
   - 添加网络流量历史
   - 添加磁盘分区详情
   - 添加进程列表

5. **UI 优化**
   - 修复 A11y 警告
   - 添加加载骨架屏
   - 添加空状态提示

## 📝 与旧前端的对比

| 功能 | 旧前端 (frontend/) | 新前端 (frontend_new/) |
|------|-------------------|----------------------|
| 框架 | Svelte 3 | Svelte 4 |
| 样式 | 自定义 CSS | Tailwind CSS 3 |
| UI 设计 | 简约风格 | Figma 现代设计 |
| 组件结构 | 功能性组件 | 组件化设计 |
| 状态管理 | 分离的 store 文件 | 统一的 stores.js |
| 响应式 | 基础 | 完整响应式 |
| 终端 | 基础 | 完整（xterm.js + Fit + WebLinks） |
| 监控 | 基础 | 完整（SVG 图表） |
| 主题 | 基础 | 完整（自动检测系统主题） |

## 🐛 已知问题

### A11y 警告

构建时会显示一些无障碍警告：

```
A11y: visible, non-interactive elements with an on:click event must be accompanied by a keyboard event handler.
```

这些可以在后续优化中添加键盘事件处理或改用 `<button>` 元素。

### Wails 绑定

Wails 绑定在 `wails dev` 时才会生成。独立运行 `npm run dev` 时会显示友好提示。

## 📖 相关文档

- [../README.md](../README.md) - 主项目 README
- [FRONTEND_NEW_IMPLEMENTATION.md](../FRONTEND_NEW_IMPLEMENTATION.md) - 详细实施总结
- [../AGENTS.md](../AGENTS.md) - 开发者指南

## 🚀 如何开始

1. **克隆项目**
```bash
git clone <repository-url>
cd sshTools
```

2. **安装依赖**
```bash
# 前端依赖
cd frontend
npm install

# 后端依赖（如果需要）
cd ..
go mod download
```

3. **启动开发**
```bash
# 完整应用
wails dev

# 或仅前端
cd frontend
npm run dev
```

4. **访问应用**
- 应用会自动在浏览器中打开
- 或手动访问 http://localhost:5174

## 📄 许可证

Apache License 2.0
