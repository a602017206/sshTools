export function createSelectionMemory() {
  return { text: '', pointerSelecting: false };
}

/**
 * tail -f 会在松手之后继续改缓冲区。选区坐标还在时，getSelection() 可能已经变成空白。
 * 拖动过程中记下非空文本；只有用户正在框选空白时才把快照清掉。
 */
export function reduceSelectionMemory(state, event) {
  const current = state || createSelectionMemory();
  const live = event?.liveText || '';

  if (event?.type === 'pointerdown') {
    return { text: current.text, pointerSelecting: true };
  }

  if (event?.type === 'pointerup') {
    return {
      pointerSelecting: false,
      text: live || current.text
    };
  }

  if (event?.type === 'change') {
    if (live) {
      return { text: live, pointerSelecting: current.pointerSelecting };
    }
    if (current.pointerSelecting && event.hasSelection) {
      return { text: '', pointerSelecting: true };
    }
    return current;
  }

  return current;
}

export function resolveTerminalCopyText(liveText, hasSelection, snapshotText) {
  const live = liveText || '';
  const snapshot = snapshotText || '';
  if (live.trim()) {
    return live;
  }
  if (hasSelection && snapshot.trim()) {
    return snapshot;
  }
  return live;
}

export function prepareTerminalPaste(text, bracketedPasteMode) {
  const normalized = String(text ?? '').replace(/\r\n/g, '\r').replace(/\n/g, '\r');
  if (!normalized) {
    return '';
  }
  if (!bracketedPasteMode) {
    return normalized;
  }
  return `\u001b[200~${normalized}\u001b[201~`;
}

export function copyUsingHiddenTextarea(text, doc) {
  if (!text || !doc?.body || typeof doc.createElement !== 'function' || typeof doc.execCommand !== 'function') {
    return false;
  }
  const area = doc.createElement('textarea');
  area.value = text;
  area.setAttribute?.('readonly', '');
  if (area.style) {
    area.style.position = 'fixed';
    area.style.opacity = '0';
  }
  doc.body.appendChild(area);
  area.focus?.();
  area.select?.();
  let ok = false;
  try {
    ok = doc.execCommand('copy') !== false;
  } catch {
    ok = false;
  }
  doc.body.removeChild(area);
  return ok;
}
