# AHaSSHTools 宣传站与 GitHub Pages 发布设计

日期：2026-10-28  
状态：待实现（设计已口头确认；已增补 three.js 氛围层与多主题）

## 背景

产品功能已较完整（SSH 终端、SFTP、监控、关系库 / 缓存 / 搜索 / 消息、AI、开发工具、桌面 + 移动端），需要独立宣传站介绍产品，并经由 GitHub Pages（`gh-pages`）对外发布。现有流水线仅有发版（`release.yml`）与移动端构建（`mobile.yml`），不含站点部署。宣传站在保留纯静态、低维护的前提下，增加 three.js 氛围特效与多色彩主题，提升浏览记忆点。

## 目标与非目标

### 目标

- 以**产品介绍**为主、下载入口为辅
- 少量多页：首页、功能详情、下载说明
- 中英双语（默认中文，可切换并记住偏好）
- 复用仓库已有界面截图与应用图标
- `master` 上 `website/**`（或部署 workflow）变更时自动发布
- 低维护：纯静态，无构建工具链；three.js 经 CDN 引入（ES module）
- 首页（及可选内页）具备粒子 / 鼠标跟随氛围层；顶栏可切换多套色彩主题

### 非目标

- 不做文档站 / 手册站
- 不接入发版流水线、不随 `v*` tag 强制同步发布
- 不做自定义域名（首期）
- 不在本站托管安装包二进制（仅链到 GitHub Releases）
- 不做重度 3D 产品拆解 / 可交互场景编辑器；特效仅作背景氛围，不挡阅读与点击

## 方案取舍

| 方案 | 说明 | 结论 |
| --- | --- | --- |
| 纯静态多页 HTML | 零构建、CI 只部署目录 | **采用** |
| Vite + Svelte/Vue | 与桌面前端接近，但多一套依赖 | 过重，否决 |
| Astro SSG | 多页友好，引入新框架 | 收益不足，否决 |

## 信息架构

### 仓库目录

```
website/
  index.html
  features.html
  download.html
  assets/
    css/site.css
    js/i18n.js
    js/nav.js
    js/theme.js       # 多主题：CSS 变量 + localStorage
    js/scene.js       # three.js 粒子场 + 鼠标跟随
    img/              # 从现有素材拷贝/压缩后的宣传用图
```

与 `frontend/` 完全隔离，避免与 Wails 前端构建混淆。three.js 不 vendor 进仓库，运行时从可信 CDN（如 jsDelivr / esm.sh 固定版本）加载。

### 页面职责

| 页面 | 职责 |
| --- | --- |
| `index.html` | 定位叙事、能力摘要、平台覆盖、次要下载/开源 CTA |
| `features.html` | 按功能域展开，标注成熟度（完整 / 可用 / 雏形），与 README 对齐 |
| `download.html` | 各平台安装包链接、安装注意、指向 `releases/latest` |

### 全局导航

顶栏：Logo + 首页 / 功能 / 下载 + **主题切换** + 中英切换 + GitHub。页脚：许可、仓库链接；赞助可链 README 或页内轻量区块。

## 视觉、氛围特效与文案

### 视觉方向

- 隐喻：运维节点网络（连接、会话、资产），而非通用「宇宙粒子」
- 默认主题对齐应用「石墨青」：深青灰底 + 克制青绿点缀
- 字体：等宽展示标题 + 无衬线正文（如 IBM Plex Mono + Source Sans 3，CDN）
- 签名：首页 hero 区 = **three.js 粒子背景** + 前景品牌文案 + 真实截图条带；截图仍是产品证据，粒子只做氛围
- 避免：紫色渐变模板、酸绿霓虹堆叠、大面积 glow、浮动徽章墙

### 多色彩主题

- 实现：`data-theme` + CSS 自定义属性；`theme.js` 读写 `localStorage`（键如 `sshtools-site-theme`）
- 首期固定 **4** 套（名称可微调，语义固定）：

| id | 气质 | 用途 |
| --- | --- | --- |
| `graphite` | 深青灰 + 青绿（默认） | 对齐桌面 Graphite Cyan |
| `amber` | 深底 + 琥珀终端黄 | 经典终端感 |
| `aurora` | 深底 + 蓝青双色粒子 | 更「宣传」一点的冷色 |
| `paper` | 浅底 + 墨青强调 | 明亮阅读；粒子降饱和、降密度 |

- 切换主题时：同步更新 CSS 变量，并调用 `scene.js` 的 `setPalette()` 更新粒子/连线颜色（无需重载页）
- 顶栏用紧凑色点或下拉，不占主叙事

### three.js 氛围层（`scene.js`）

| 能力 | 规格 |
| --- | --- |
| 载体 | 全页或 hero 背后固定 `canvas`，`pointer-events: none`，不拦截链接/按钮 |
| 粒子 | Points 点云；隐喻「主机/会话节点」；数量桌面约 80–150，移动端约 40–60 |
| 连线 | 近邻节点淡线（BufferGeometry），暗示网络拓扑；密度随主题与性能档位调整 |
| 鼠标跟随 | 归一化指针位置作吸引力场：近处粒子轻微偏移/汇聚；可选极淡光点跟随光标（非花哨拖尾） |
| 范围 | **首页必开**；功能页 / 下载页可用更低密度同一模块，或仅保留静态渐变（实现时以首页完整、内页轻量为准） |
| 降级 | `prefers-reduced-motion: reduce` → 不启动 WebGL，改 CSS 静态渐变；WebGL 失败同降级 |
| 性能 | `requestAnimationFrame`；页不可见时 `visibilitychange` 暂停；resize 节流 |

依赖：three.js **固定版本** CDN（写入设计实现计划与 HTML 注释，避免浮动 latest）。

### 素材来源

- 界面：`docs/designs/assets/aether-glass/`（SSH / 数据库，浅色与深色）
- 图标：`build/appicon.png`（拷贝到 `website/assets/img/`）
- 实现时将图片复制进 `website/assets/img/` 并适当压缩，站点不直接依赖 `docs/` 路径

### 文案原则

- 中文为默认；英文为完整平行文案（非机翻占位）
- 功能成熟度诚实表述，与根目录 README「功能概览」一致
- 下载页不抢首页叙事

### i18n 机制

- `assets/js/i18n.js` 集中字典；DOM 用 `data-i18n`（或等价）标记
- 语言偏好写入 `localStorage`；顶栏切换即时生效
- 三页共用同一套脚本
- 主题与语言状态互相独立

## 部署与 CI

### 发布形态

- 目标 URL：`https://a602017206.github.io/sshTools/`
- 所有链接与静态资源使用**相对路径**，并正确处理仓库子路径（Pages project site）
- **选定**：GitHub Actions 作为 Pages 源，使用官方 `actions/configure-pages` + `upload-pages-artifact` + `deploy-pages`（`environment: github-pages`）
- 不采用 `peaceiris` 推分支方案，避免双源混淆；若官方 action 在仓库权限上不可用，再单独立项改方案

### 新工作流

文件：`.github/workflows/gh-pages.yml`

- **触发**：`push` → `master`，`paths`：`website/**`、`.github/workflows/gh-pages.yml`
- **不**修改 `release.yml` / `mobile.yml`
- **权限**：`contents: read`、`pages: write`、`id-token: write`
- **步骤概要**：checkout → 校验 `website/index.html` 存在 → configure-pages → upload `website/` 为 artifact → deploy-pages

### 仓库设置（人工一次性）

实现合并后需在 GitHub 仓库 Settings → Pages 中将 Source 设为 **GitHub Actions**。变更文档须把该项列为验收检查。

## 文档与入口

- 本设计：`docs/designs/2026-10-28-marketing-website-gh-pages-design.md`
- 规格副本：`docs/superpowers/specs/2026-10-28-marketing-website-gh-pages-design.md`
- 实现后：`docs/changes/process/` 记流程变更；根 README 增加宣传站链接一行

## 成功标准

1. 本地用简易静态服务打开站点可浏览三页、切换中英与四套主题
2. 首页可见粒子场与鼠标跟随；切换主题后面板色与粒子色同步变化
3. `prefers-reduced-motion` 或 WebGL 不可用时，页面仍可读、可下载，无控制台致命错误
4. 合并到 `master` 且触及 `website/**` 后，Actions 成功部署，公网 URL 可访问
5. 子路径下 CSS / JS / 图片 / CDN 资源无 404（CDN 需联网）
6. 下载链接指向官方 Releases latest，平台说明与 README 一致

## 风险与缓解

| 风险 | 缓解 |
| --- | --- |
| Pages 子路径导致资源 404 | 统一相对路径；部署后抽查 |
| 截图文件过大拖慢站点 | 拷贝时压缩；控制张数 |
| 中英文案漂移 | 字典集中维护；功能描述以 README 为源 |
| Pages Source 未改导致部署成功但站点不更新 | 变更文档列出人工检查项 |
| three.js / 粒子拖慢低端机或抢焦点 | 限粒子数、内页降密度、不可见暂停、reduced-motion 降级、canvas 不接收指针 |
| CDN 不可达导致白屏 | 场景脚本 `try/catch` + 无 WebGL 时 CSS 回退；核心文案与下载不依赖 three |
| 特效像通用 AI 站 | 节点连线隐喻 + 石墨青默认主题 + 截图仍占主视觉 |

## 实现范围边界

首期交付：`website/` 三页 + 共享 CSS/JS（含 `theme.js`、`scene.js`）+ CDN three.js + `gh-pages.yml` + 设计/变更文档 + README 链接。

不包含：自定义域名、博客、changelog 自动同步、截图自动化生成、自托管 three 构建管线、复杂 3D 模型或后处理（bloom 等）特效栈。
