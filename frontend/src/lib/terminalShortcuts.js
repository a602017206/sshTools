const shortcutKey = (event) => event.key?.toLowerCase();

function matchesKey(event, letter, code) {
  return shortcutKey(event) === letter || event.code === code;
}

/**
 * 返回应由终端客户端处理的复制、粘贴快捷键；null 表示交由 xterm/远端终端处理。
 * 调用方应仅在 keydown 时使用返回值。
 */
export function getTerminalShortcutAction(event, hasSelection) {
  if (event.altKey || (event.ctrlKey && event.metaKey)) {
    return null;
  }

  const primaryModifier = event.ctrlKey || event.metaKey;
  const isC = matchesKey(event, 'c', 'KeyC');
  const isV = matchesKey(event, 'v', 'KeyV');
  const isInsert = matchesKey(event, 'insert', 'Insert');

  // macOS Cmd+C：始终拦截。有选区则复制；无选区吞掉按键，避免把 Meta 组合泄漏到 PTY。
  if (event.metaKey && !event.ctrlKey && isC) {
    return hasSelection ? 'copy' : 'noop';
  }

  if (hasSelection && primaryModifier && (isC || isInsert)) {
    return 'copy';
  }

  // Ctrl+V 是终端控制字符（例如 Vim 搜索中输入字面量），不能作为粘贴快捷键拦截。
  // macOS 使用 Cmd+V；其他平台仍支持 Ctrl+Shift+V 和 Shift+Insert 粘贴。
  if ((event.metaKey && isV) || (event.ctrlKey && event.shiftKey && isV) || (event.shiftKey && isInsert)) {
    return 'paste';
  }

  return null;
}
