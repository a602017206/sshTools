# 变更：macOS 窗口最大化按钮被禁用

## 背景

macOS 标题栏左侧交通灯的绿色按钮呈灰色不可用。Wails v2.11 在 Darwin 上把 `zoomable` 默认初始化为 0；只有在提供 `options.App.Mac` 且 `DisableZoom` 为 false 时才会打开缩放。应用此前未配置 `Mac` 选项，因此窗口虽可拖拽改大小，缩放按钮仍被显式禁用。

## 范围

- 在 `main.go` 中补充 `Mac: &mac.Options{DisableZoom: false}`，让绿色缩放按钮可用。
- 不改变现有启动最大化行为（`WindowStartState: options.Maximised`）。

## 修改文件

- `main.go`
- `docs/changes/bugs/2026-08-27-macos-zoom-button.md`

## 验证

- 阅读 Wails v2.11 Darwin 窗口创建逻辑，确认 `Mac == nil` 时 `zoomable` 保持为 0，并调用 `NSWindowZoomButton setEnabled:NO`。
- 需重新启动桌面应用后，绿色按钮应变为可点击；再次点击应在最大化与 1024×768 之间切换。

## 剩余风险

- 此修复依赖重新启动 Wails 进程才会生效；热重载前端不会重建原生窗口。
- 若后续把 `DisableZoom` 设为 true，按钮会再次禁用。
