export const FILE_MANAGER_MENU_WIDTH = 248;
export const FILE_MANAGER_SUBMENU_WIDTH = 200;
export const FILE_MANAGER_MENU_HEIGHT_FILE = 448;
export const FILE_MANAGER_MENU_HEIGHT_BLANK = 336;
export const FILE_MANAGER_MORE_INLINE_HEIGHT = 136;
export const FILE_MANAGER_SUBMENU_OVERLAP = 6;

export function isMacPlatform(userAgent = '') {
  return /Mac|iPhone|iPad/i.test(userAgent);
}

export function joinRemotePath(dir, name) {
  const base = String(dir || '/').replace(/\/+$/, '') || '';
  const leaf = String(name || '').replace(/^\/+/, '');
  if (!leaf) return base || '/';
  if (!base || base === '/') return `/${leaf}`;
  return `${base}/${leaf}`;
}

export function splitFileName(name) {
  const value = String(name || '');
  const idx = value.lastIndexOf('.');
  if (idx <= 0) return { stem: value, ext: '' };
  return { stem: value.slice(0, idx), ext: value.slice(idx) };
}

export function uniqueCopyName(existingNames, desired) {
  const names = new Set(existingNames || []);
  if (!names.has(desired)) return desired;
  const { stem, ext } = splitFileName(desired);
  let index = 1;
  while (true) {
    const next = index === 1 ? `${stem} copy${ext}` : `${stem} copy ${index}${ext}`;
    if (!names.has(next)) return next;
    index += 1;
  }
}

export function unixModeToOctal(mode) {
  const raw = String(mode || '');
  const perms = raw.replace(/^[d\-lcbsp]/, '').slice(0, 9);
  if (perms.length < 9) return '644';
  const isSet = [
    (ch) => ch === 'r',
    (ch) => ch === 'w',
    (ch) => 'xstST'.includes(ch),
  ];
  let value = 0;
  for (let i = 0; i < 9; i += 1) {
    if (isSet[i % 3](perms[i])) value |= 1 << (8 - i);
  }
  return value.toString(8).padStart(3, '0');
}

export function isValidOctalMode(value) {
  return /^[0-7]{3,4}$/.test(String(value || '').trim());
}

export function isPathFavorite(history, path) {
  return Array.isArray(history) && history.includes(path);
}

export function toggleFavoriteHistory(history, path, limit = 5) {
  const next = [...(history || [])];
  const index = next.indexOf(path);
  if (index >= 0) {
    next.splice(index, 1);
    return next;
  }
  next.unshift(path);
  return next.slice(0, Math.max(1, limit));
}

export function getContextMenuPosition({
  clientX,
  clientY,
  root,
  menuWidth = FILE_MANAGER_MENU_WIDTH,
  menuHeight = FILE_MANAGER_MENU_HEIGHT_FILE,
}) {
  if (!root) return { x: clientX, y: clientY };
  return {
    x: Math.max(8, Math.min(clientX - root.left, root.width - menuWidth - 8)),
    y: Math.max(8, Math.min(clientY - root.top, root.height - menuHeight - 8)),
  };
}

export function getSubmenuPlacement(
  menuX,
  rootWidth,
  menuWidth = FILE_MANAGER_MENU_WIDTH,
  submenuWidth = FILE_MANAGER_SUBMENU_WIDTH,
) {
  if (!rootWidth) return 'right';
  const rightFits = menuX + menuWidth + submenuWidth - FILE_MANAGER_SUBMENU_OVERLAP <= rootWidth - 8;
  if (rightFits) return 'right';
  const leftFits = menuX + FILE_MANAGER_SUBMENU_OVERLAP >= submenuWidth + 8;
  if (leftFits) return 'left';
  return 'down';
}

export function getSubmenuSide(menuX, rootWidth, menuWidth = FILE_MANAGER_MENU_WIDTH, submenuWidth = FILE_MANAGER_SUBMENU_WIDTH) {
  return getSubmenuPlacement(menuX, rootWidth, menuWidth, submenuWidth);
}

export function shiftMenuTopForInlineMore(y, rootHeight, menuHeight = FILE_MANAGER_MENU_HEIGHT_FILE) {
  if (!rootHeight) return y;
  return Math.max(8, Math.min(y, rootHeight - menuHeight - FILE_MANAGER_MORE_INLINE_HEIGHT - 8));
}

export function getFileManagerMenuFlags({
  file = null,
  currentPath = '/',
  history = [],
  historyEnabled = true,
  clipboard = null,
} = {}) {
  const hasFile = Boolean(file && !file.is_parent);
  const isDir = hasFile && Boolean(file.is_dir);
  const isFile = hasFile && !file.is_dir;
  const canPaste = true;
  return {
    hasFile,
    isDir,
    isFile,
    canOpen: hasFile,
    canRename: hasFile,
    canDelete: hasFile,
    canCut: hasFile,
    canCopy: isFile,
    canPaste,
    canDownload: isFile,
    canChmod: hasFile,
    canFavorite: Boolean(historyEnabled && currentPath),
    isFavorite: isPathFavorite(history, currentPath),
  };
}

export function matchFileManagerShortcut(event, { isMac = false } = {}) {
  if (!event) return null;
  const meta = isMac ? event.metaKey : event.ctrlKey;
  const { altKey: alt, shiftKey: shift, key: rawKey } = event;
  const key = String(rawKey || '').toLowerCase();

  const directAction = directFileManagerShortcut(rawKey);
  if (directAction) return directAction;
  if (alt && shift && !meta) {
    return altShiftFileManagerShortcut(key);
  }
  if (meta && !alt && !shift) {
    return primaryFileManagerShortcut(key);
  }
  return null;
}

function directFileManagerShortcut(key) {
  return {
    Enter: 'openLocal',
    F2: 'rename',
    Backspace: 'delete',
    Delete: 'delete',
  }[key] || null;
}

function altShiftFileManagerShortcut(key) {
  return { c: 'copyPath', n: 'newFile', m: 'chmod' }[key] || null;
}

function primaryFileManagerShortcut(key) {
  return { r: 'refresh', c: 'copy', x: 'cut', v: 'paste' }[key] || null;
}
