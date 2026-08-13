export function getRowContextMenuPosition(anchorRect) {
  return { x: anchorRect.left, y: anchorRect.bottom };
}

export function shouldCloseRowContextMenu({ menuContainsTarget }) {
  return !menuContainsTarget;
}
