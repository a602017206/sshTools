const shortcutKey = (event) => event.key?.toLowerCase();

/**
 * 返回应由终端客户端处理的复制、粘贴快捷键；null 表示交由 xterm/远端终端处理。
 */
export function getTerminalShortcutAction(event, hasSelection) {
  if (event.altKey || (event.ctrlKey && event.metaKey)) {
    return null;
  }

  const key = shortcutKey(event);
  const primaryModifier = event.ctrlKey || event.metaKey;

  if (hasSelection && primaryModifier && (key === 'c' || key === 'insert')) {
    return 'copy';
  }

  if ((primaryModifier && key === 'v') || (event.shiftKey && key === 'insert')) {
    return 'paste';
  }

  return null;
}
