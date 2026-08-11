# 设置支持自定义背景图片

## 背景

运维工作区改为霜白浅色壳层后，用户希望在「全局设置」中自行添加背景图，作为底层氛围，而不影响面板可读性。

## 范围

- 在应用设置中增加背景图开关、路径、填充方式（铺满 / 完整显示）、强度（透明度）。
- 通过系统文件对话框选择 jpg/png/webp/gif，复制到 `~/.ahasshtools/backgrounds/`（单文件最大 8MB）。
- 选择后实时预览；点「保存」才写入配置；「取消」恢复先前外观。
- 启动时按配置加载 Data URL 并应用到 `.ops-shell`。

## 修改文件

- `internal/config/config.go`：背景图相关设置字段
- `internal/service/settings_background.go`：安装 / 清除 / Data URL
- `internal/service/settings_background_test.go`：格式与复制唯一名测试
- `internal/service/settings_service.go`：更新设置时清理旧背景文件
- `app.go`：暴露 `SelectBackgroundImage` / `ClearBackgroundImage` / `GetBackgroundImageDataURL`
- `frontend/wailsjs/go/main/App.js`（及已有 d.ts/models）：前端绑定
- `frontend/src/settings/appearance.js`：`applyBackgroundImage`
- `frontend/src/styles/app.css`：`data-bg-image` 样式
- `frontend/src/components/GlobalSettingsDialog.svelte`：外观分区 UI
- `frontend/src/App.svelte`：加载 / 持久化背景字段

## 验证

- `go test ./internal/service -run 'MimeForBackground|NormalizeBackground|EncodeBackground|CopyBackground' -count=1`：通过
- `go build -o /dev/null .`：通过
- `cd frontend && npx vite build`：通过
- 手工建议：`wails dev` → 设置 → 选择图片 → 预览/保存/重启仍生效；清除后保存恢复霜白底；取消不保留新图

## 剩余风险

- 取消前多次选择会留下未引用的临时文件（唯一文件名），长期可能占少量磁盘；可后续做目录清理。
- 超大图虽限制 8MB，仍可能拖慢启动时 Data URL 编码；极端场景可再改为本地 file URL 或缩略图。
- 深色主题下强对比壁纸可能影响侧栏对比度，需用户自行调低强度。
