/**
 * 从 xterm 当前光标行读取可见文本（含 prompt）。
 * Tab 补全后的完整命令会出现在这一行，而不会出现在本地按键缓冲里。
 */
export function readXtermCursorLine(terminal) {
  if (!terminal?.buffer?.active) {
    return '';
  }
  const buffer = terminal.buffer.active;
  const line = buffer.getLine(buffer.baseY + buffer.cursorY);
  if (!line) {
    return '';
  }

  let text = '';
  for (let col = 0; col < line.length; col += 1) {
    const cell = line.getCell(col);
    if (!cell) {
      continue;
    }
    const chars = cell.getChars();
    if (chars) {
      text += chars;
    } else if (cell.getWidth() > 0) {
      text += ' ';
    }
  }
  return text.replace(/\s+$/g, '');
}

/**
 * 从整行终端文本中剥离常见 shell prompt，得到用户命令。
 * 例：`[root@host pems]# cd /opt` → `cd /opt`
 */
export function extractShellCommand(lineText) {
  const raw = String(lineText || '').replace(/\s+$/g, '');
  if (!raw) {
    return '';
  }

  const withSpace = raw.match(/.*[#$%>]\s+(.*)$/);
  if (withSpace) {
    return withSpace[1].replace(/\s+$/g, '');
  }

  const tight = raw.match(/.*[#$%>](.*)$/);
  if (tight) {
    return tight[1].replace(/\s+$/g, '');
  }

  return raw.trim();
}

/** 读取并解析当前光标行上的实际命令（优先于按键缓冲）。 */
export function readExecutedCommandFromTerminal(terminal) {
  return extractShellCommand(readXtermCursorLine(terminal));
}
