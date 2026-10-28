# Task 5 报告

## 完成内容

- 新增 `nav.js`：初始化主题和语言，支持四主题、双语按钮、当前导航标记及场景调色板更新。
- 新增首页、功能页和下载页；三页均使用相对站内路径、Google Fonts（IBM Plex Mono、Source Sans 3）和既有静态资源。
- 下载链接直达 GitHub 最新发布的 README 安装包；内页场景使用 `light` 密度，首页使用 `full` 密度。
- 新增导航单测，并补充变更记录 `docs/changes/features/2026-10-28-marketing-website-pages.md`。

## 验证

- `cd website && node --test test/*.test.js`：19 项测试全部通过。
- 本地静态服务器下，`/`、`/features.html`、`/download.html`、`/assets/css/site.css`、`/assets/js/nav.js` 均返回 HTTP 200。
- 已脚本校验三页 66 处 `data-i18n`/`data-i18n-html` 键均存在于 `MESSAGES` 的中英文表。

## 剩余关注点

- 无头验证无法实际确认 WebGL 粒子对鼠标移动的渲染响应；HTML、脚本挂载与 `scene.js` 接口已按要求接通，需在真实浏览器中作最终视觉检查。
# Task 5 报告

## 完成内容
- `app.go`：在 `InspectDockerContainer` 之后新增 9 个导出方法（镜像/网络/数据卷的 List、Inspect、Remove），均带 `dockerService == nil` 检查。`RemoveDockerImage` 含 `force`，`RemoveDockerVolume`/`RemoveDockerNetwork` 无 `force`。
- `frontend/wailsjs/go/main/App.js`、`App.d.ts`：按字母序手工插入 9 个绑定。
- `frontend/wailsjs/go/models.ts`：补充 `docker.ImageSummary/ImageDetail/NetworkSummary/NetworkDetail/VolumeSummary/VolumeDetail`（嵌入结构体展平，与 JSON 一致）。

## 验证
- `gofmt -l app.go` 无输出；`go build -o /dev/null .` 通过（需联网下载 wails 依赖，沙箱外执行）。

## 剩余风险
- 绑定为手工编写，下次 `wails dev/build` 重新生成时可能出现格式差异，但语义一致。
- 前端未使用这些 API（属后续任务）。
