export function shouldApplyClone({ isOpen, cloningAsset, appliedCloningAsset }) {
  return Boolean(isOpen && cloningAsset && cloningAsset !== appliedCloningAsset);
}
