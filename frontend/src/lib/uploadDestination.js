export function resolveUploadDestination({ currentPath = '/', selectedPaths = [], files = [] } = {}) {
  const current = String(currentPath || '/') || '/';
  if (!Array.isArray(selectedPaths) || selectedPaths.length !== 1) {
    return current;
  }
  const selected = selectedPaths[0];
  const file = (files || []).find((item) => item?.path === selected);
  if (file?.is_dir && !file.is_parent) {
    return file.path;
  }
  return current;
}
