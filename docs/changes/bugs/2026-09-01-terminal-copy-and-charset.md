# 缺陷：SSH 终端无法复制与中文粘贴乱码

## 背景

macOS 上按 Command，xterm 会把修饰键当输入并滚到最底部，上方选区丢失，Cmd+C 失效；右键被设成选词，没有复制菜单。粘贴中文乱码是因为始终按 UTF-8 写入 PTY，连接没有编码设置。

## 范围

- 单独按下修饰键不再交给 xterm 滚动
- 终端右键提供复制/粘贴
- SSH 连接增加编码选项，输入按编码转码，输出在写入 xterm 前解码
- 已打开会话改编码立即生效，无需重连
- ZMODEM 二进制通道不转码

## 修改文件

- `docs/designs/2026-09-01-terminal-copy-and-charset.md`
- `docs/development/2026-09-01-terminal-copy-and-charset.md`
- `frontend/src/lib/terminalShortcuts.js`
- `frontend/test/terminalShortcuts.test.js`
- `frontend/src/lib/terminalCharset.js`
- `frontend/test/terminalCharset.test.js`
- `frontend/src/lib/connectionFormData.js`
- `frontend/test/connectionFormData.test.js`
- `frontend/src/components/Terminal.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/App.svelte`
- `internal/ssh/charset.go`
- `internal/ssh/charset_test.go`
- `app.go`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- `.github/workflows/release.yml`

## 验证

```bash
cd frontend && node --test test/terminalShortcuts.test.js test/terminalCharset.test.js test/connectionFormData.test.js
go test ./internal/ssh -count=1 -run 'TestEncodeFromUTF8|TestNormalizeCharset'
```

## 剩余风险

已打开的会话改编码立即生效。Node/WebView 对 GB2312 标签的支持可能回落到替换字符；GBK 在 Chromium 中可用。未在真实 GBK 远端上手工粘贴验证。切换编码不会重绘已经显示在缓冲区里的旧输出。
