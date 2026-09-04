# Windows 打包仍显示 Wails 默认图标

## 背景

`wails build` 在 Windows 上把 `build/windows/icon.ico` 编进 exe 资源。这份 ico 是项目初始化时的默认斜体 W，后来只改了 `build/appicon.png`。Wails 在 ico **已存在**时不会再从 png 生成，所以资源管理器、任务栏、窗口标题栏和安装程序都还是 W。

## 决策

1. 用当前 `appicon.png` 重新生成多尺寸 `icon.ico` 并提交。
2. 发布工作流在 Windows 构建前删除 `icon.ico`，让 Wails 按 `appicon.png` 再生成，避免以后再改 png 却忘了 ico。
3. 用单测拦住「又变回默认约 21KB 的 W 图标」。

不改窗口代码：Wails 2.11 的 Windows 窗口图标就是从 exe 资源读取的。

## 风险

Windows 资源管理器会缓存 exe 图标。更新后若仍看到旧 W，可重启资源管理器或换目录复制一份 exe 再看。
