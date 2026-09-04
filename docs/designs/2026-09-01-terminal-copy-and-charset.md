# 设计：终端复制与连接编码

## 背景

macOS 上按 Command 时 xterm 把修饰键当作用户输入并 `scrollToBottom`，滚出缓冲区上方的选区，Cmd+C 失效；右键被设成选词，没有复制菜单。粘贴中文乱码是因为远端常见 GBK 环境，而客户端始终按 UTF-8 写 PTY，连接表单也没有编码选项。

## 决策

- 修饰键单独按下（Meta/Control/Alt/Shift）交给终端快捷键层吞掉，不让 xterm 滚动。
- 关闭 `rightClickSelectsWord`，右键弹出复制/粘贴。定位菜单时不要清掉 xterm 选区。
- SSH 连接增加编码（默认 UTF-8，可选 GBK / GB2312 / GB18030 / Big5），存 `metadata.encoding`。
- 输入：`SendSSHData` 按会话编码把 UTF-8 转成远端字节。ZMODEM 仍走 `SendSSHDataBinary`，不转码。
- 输出：ZMODEM sentry 仍吃原始字节；写入 xterm 前用 `TextDecoder` 按编码解码成 Unicode。
- 已打开的会话改编码立即生效：更新该连接下所有 SSH 会话的 `SetSessionCharset` 与 xterm 解码，无需重连。终端内也可直接切换编码。
