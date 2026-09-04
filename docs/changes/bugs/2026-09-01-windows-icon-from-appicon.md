# 修复 Windows 打包仍用 Wails 默认图标

## 背景

Windows 安装包和 exe 的图标、窗口图标仍是 Wails 默认斜体 W，不是 `build/appicon.png` 里的终端 Logo。

## 范围

- 从 `appicon.png` 重新生成 `build/windows/icon.ico`
- 发布时在 Windows 构建前删除旧 ico，让 Wails 再生成
- 增加回归测试，避免再次提交默认 W 图标

不改 Flutter 客户端图标。

## 修改文件

- `build/windows/icon.ico`
- `scripts/sync-windows-icon.py`（新建）
- `windows_icon_test.go`（新建）
- `.github/workflows/release.yml`
- `docs/designs/2026-09-01-windows-icon-from-appicon.md`
- `docs/changes/bugs/2026-09-01-windows-icon-from-appicon.md`（本文）

## 验证

- `go test -count=1 -run TestWindowsIconIsNotDefaultWailsLogo .`
- 修复前该测试失败：`icon.ico` 为 21017 字节
- 抽查 ico 内 256 图为终端 `>_ aha` Logo，不是斜体 W

无法在本机实际跑 Windows 安装包点选验证。

## 剩余风险

- 若只改 `appicon.png` 却不重新生成或不走发布工作流，本地 `wails build` 仍会用仓库里的 ico；请运行 `python3 scripts/sync-windows-icon.py`
- Windows 资源管理器可能缓存旧图标
