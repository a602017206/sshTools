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

  if ((primaryModifier && isV) || (event.shiftKey && isInsert)) {
    return 'paste';
  }

  return null;
}
