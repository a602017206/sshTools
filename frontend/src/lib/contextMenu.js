export function getViewportMenuPosition({
  clientX,
  clientY,
  menuWidth = 160,
  menuHeight = 48,
  viewWidth = Number.POSITIVE_INFINITY,
  viewHeight = Number.POSITIVE_INFINITY,
  margin = 8,
} = {}) {
  const maxX = Number.isFinite(viewWidth) ? viewWidth - menuWidth - margin : clientX;
  const maxY = Number.isFinite(viewHeight) ? viewHeight - menuHeight - margin : clientY;
  return {
    x: Math.max(margin, Math.min(Number(clientX) || 0, maxX)),
    y: Math.max(margin, Math.min(Number(clientY) || 0, maxY)),
  };
}

export function resolveContextMenuPoint(event, options = {}) {
  event?.preventDefault?.();
  event?.stopPropagation?.();
  clearWindowSelection();
  const viewWidth = options.viewWidth ?? (typeof window !== 'undefined' ? window.innerWidth : Number.POSITIVE_INFINITY);
  const viewHeight = options.viewHeight ?? (typeof window !== 'undefined' ? window.innerHeight : Number.POSITIVE_INFINITY);
  return getViewportMenuPosition({
    clientX: event?.clientX,
    clientY: event?.clientY,
    menuWidth: options.menuWidth,
    menuHeight: options.menuHeight,
    viewWidth,
    viewHeight,
    margin: options.margin,
  });
}

export function clearWindowSelection() {
  const selection = typeof window !== 'undefined' ? window.getSelection?.() : null;
  selection?.removeAllRanges?.();
}

export function portalToBody(node) {
  if (typeof document === 'undefined') return {};
  const host = document.createElement('div');
  host.className = 'ops-context-menu-host';
  document.body.appendChild(host);
  host.appendChild(node);
  return {
    destroy() {
      if (host.parentNode) host.parentNode.removeChild(host);
    },
  };
}

export function viewSize() {
  if (typeof window === 'undefined') {
    return { viewWidth: Number.POSITIVE_INFINITY, viewHeight: Number.POSITIVE_INFINITY };
  }
  return { viewWidth: window.innerWidth, viewHeight: window.innerHeight };
}
