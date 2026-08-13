export function shouldApplyEditConnectionResult({ isOpen, requestedId, activeId }) {
  return Boolean(isOpen && requestedId && requestedId === activeId);
}

export function shouldLoadEditConnection({ isOpen, targetId, loaded }) {
  return Boolean(isOpen && targetId && !loaded);
}
