const shortcutKey = (event) => event.key?.toLowerCase();

function matchesKey(event, letter, code) {
  return shortcutKey(event) === letter || event.code === code;
}

const MODIFIER_ONLY_CODES = new Set([
  'MetaLeft', 'MetaRight', 'ControlLeft', 'ControlRight',
  'AltLeft', 'AltRight', 'ShiftLeft', 'ShiftRight'
]);

export function isModifierOnlyKey(event) {
  const key = event?.key;
  if (key === 'Meta' || key === 'Control' || key === 'Alt' || key === 'Shift' || key === 'OS' || key === 'Ctrl') {
    return true;
  }
  return MODIFIER_ONLY_CODES.has(event?.code);
}

/** 终端滚屏后，方向键会被 xterm 用来翻历史输出；先回到底部再发给 shell 读命令历史。 */
export function shouldScrollToBottomBeforeArrowKey(event, viewportY = 0) {
  if (viewportY <= 0) return false;
  if (event.shiftKey || event.ctrlKey || event.metaKey || event.altKey) return false;
  return event.key === 'ArrowUp' || event.key === 'ArrowDown';
}

/**
 * 返回应由终端客户端处理的复制、粘贴快捷键；null 表示交由 xterm/远端终端处理。
 * 调用方应仅在 keydown 时使用返回值。
 */
export function getTerminalShortcutAction(event, hasSelection) {
  if (isModifierOnlyKey(event)) {
    return 'noop';
  }

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

  // Ctrl+V / Cmd+V 粘贴。无选区的 Ctrl+C 仍交给远端做中断。
  if ((event.metaKey && isV) || (event.ctrlKey && isV) || (event.shiftKey && isInsert)) {
    return 'paste';
  }

  return null;
}
