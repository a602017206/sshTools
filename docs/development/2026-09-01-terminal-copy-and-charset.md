# 开发记录：SSH 终端复制与连接编码

## 实现内容

- `getTerminalShortcutAction` 对单独的 Meta/Control/Alt/Shift 返回 `noop`，阻止 xterm 滚到底。
- 终端右键菜单复制/粘贴；打开菜单时记下选区文本，避免点击菜单丢失选区。
- 连接表单增加编码，写入 `metadata.encoding`。连接时 `SetSessionCharset`，`SendSSHData` 用 `EncodeFromUTF8` 转成远端字节。保存连接或终端内切换编码时，对已打开会话立即更新发送/解码，无需重连。
- 显示路径在 ZMODEM sentry 的 `to_terminal` 里用 `TextDecoder` 解码；UTF-8 仍写原始字节以免破坏协议探测。

## 验证

见变更文档中的测试命令。

## 剩余风险

切换编码不会重绘已经显示在缓冲区里的旧输出。ZMODEM 与文本混流时，非 UTF-8 解码只发生在 sentry 判定为终端输出之后。
