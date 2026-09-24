# 设计：tail -f 时复制日志与 Ctrl+V 粘贴

## 背景

SSH 控制台里选中 `tail -f` 的日志后按 Ctrl+C，按键被客户端吃掉（tail 不会停），剪贴板却是空的。Ctrl+V 也没有把内容贴回去。

## 原因

- xterm 的选区是画在画布上的，不在 DOM selection 里。`hasSelection()` 为真时，`getSelection()` 仍可能是空白：后续输出改写了缓冲区，或复制时只剩被 trim 掉的空行。
- 复制路径若没有 `preventDefault`，随后的 `copy` 事件会把空的 DOM 选区写进剪贴板，把刚才写入的内容清掉。xterm 自己的 `copy` 监听在选区文本为空时也会 `setData('')`。
- macOS 上 Ctrl+C / Ctrl+V 不会走系统的拷贝/粘贴命令。Ctrl+V 之前被故意留给远端（vim 的字面量），所以控制台里按 Ctrl+V 不会粘贴。

## 决策

- 鼠标拖动期间记下最后一段非空选区。用户还按着键去框选空白时才清掉这段快照；松手之后的输出不再清它。
- 复制时优先用当前选区里的非空白文本，否则在选区仍在时用快照。
- 在终端容器的捕获阶段处理 `copy`，写完后阻止事件继续传给 xterm。快捷键复制同时 `preventDefault`，并用隐藏文本框同步执行一次 `copy`，再调用 Wails 剪贴板。
- Ctrl+V 与 Cmd+V 一样粘贴。换行收成 `\r`；远端开了 bracketed paste 时加上括号序列。
- 没有选区时 Ctrl+C 仍发送中断，不拿旧快照去复制。

## 取舍

vim 的 Ctrl+V（可视块、插入字面量）会变成粘贴。日志复制是当前更常见的路径；字面量输入不再占用 Ctrl+V。
