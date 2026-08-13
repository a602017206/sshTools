export function shouldResetBlankConnectionForm({
  isOpen,
  requestVersion,
  appliedRequestVersion,
  editingAsset,
  cloningAsset,
}) {
  return Boolean(
    isOpen &&
    requestVersion !== appliedRequestVersion &&
    !editingAsset &&
    !cloningAsset
  );
}
