# 缺陷：从 Finder 启动后终端复制得到空剪贴板

## 背景

SSH 控制台选中内容后，用 Ctrl/Cmd+C 或右键「复制」，剪贴板都是空的。

2026-09-24 的修复（`2026-09-24-terminal-copy-braces.md`）把终端复制改为只走 Wails 的 `ClipboardSetText`。在 macOS 上，它实际执行的是 `pbcopy`。从 Finder 或 Dock 启动的应用进程没有 `LANG`/`LC_*` 环境变量，`pbcopy` 会把输入当作非 UTF-8 文本处理：只要内容里有中文等非 ASCII 字符，就写入空剪贴板。`pbpaste` 同理，粘贴中文会读出乱码。`wails dev` 从终端启动时继承了终端的 `LANG`，所以开发阶段复现不了。

复现命令：

```bash
printf '中文 {"a":1}' | env -i PATH=/usr/bin:/bin /usr/bin/pbcopy; pbpaste   # 输出为空
printf '中文 {"a":1}' | env -i PATH=/usr/bin:/bin LC_CTYPE=UTF-8 /usr/bin/pbcopy; pbpaste   # 正常
```

## 范围

- macOS 启动时，如果当前生效的 locale（按 `LC_ALL` → `LC_CTYPE` → `LANG` 的优先级）不是 UTF-8，就设置 `LC_CTYPE=UTF-8`；如果 `LC_ALL` 被设成了非 UTF-8 的值，则把 `LC_ALL` 改为 `en_US.UTF-8`。`pbcopy`/`pbpaste` 子进程会继承这个设置。
- 前端复制逻辑不变。

## 修改文件

- `clipboard_locale.go`
- `clipboard_locale_test.go`
- `main.go`
- `.github/workflows/release.yml`（发布说明「修复问题」）
- 本文档

## 验证

```bash
go test -run TestUTF8LocaleOverride -v .
go vet .
go build .
```

已在命令行用 `env -i` 模拟 Finder 启动环境，确认补上 `LC_CTYPE=UTF-8` 后，`pbcopy`/`pbpaste` 能正确处理中文。尚未在打包后的 `.app` 中手工验证。

## 剩余风险

- 本地终端会话会继承 `LC_CTYPE=UTF-8`。这与 macOS Terminal.app 的默认行为一致。
- 由于 macOS 原生 Edit 菜单的存在，按 Cmd+C 仍会先触发一次浏览器 copy 事件（会被拦截，可能清空剪贴板），随后才由 `pbcopy` 写入，两者是先后顺序关系。如果打包后 Cmd+C 仍偶发为空而右键复制正常，需要再排查这一时序问题。
