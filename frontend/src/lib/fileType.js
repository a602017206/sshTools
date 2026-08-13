const ARCHIVE_EXT = new Set(['zip', 'tar', 'gz', 'tgz', 'bz2', 'xz', '7z', 'rar']);
const IMAGE_EXT = new Set(['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'ico', 'bmp']);
const CODE_EXT = new Set(['sh', 'bash', 'zsh', 'py', 'go', 'js', 'ts', 'jsx', 'tsx', 'c', 'h', 'cpp', 'rs', 'java']);
const CONFIG_EXT = new Set(['yml', 'yaml', 'toml', 'ini', 'conf', 'cfg', 'env', 'properties']);
const TEXT_EXT = new Set(['txt', 'md', 'markdown', 'rst', 'log']);

export function fileExtension(name = '') {
  const base = String(name).split('/').pop() || '';
  if (base.startsWith('.') && !base.slice(1).includes('.')) return base.slice(1).toLowerCase();
  const idx = base.lastIndexOf('.');
  return idx > 0 ? base.slice(idx + 1).toLowerCase() : '';
}

export function resolveFileIconKind(file) {
  if (file?.is_parent) return 'parent';
  if (file?.is_dir) return 'folder';
  const ext = fileExtension(file?.name);
  if (ARCHIVE_EXT.has(ext) || ext === 'jar' || ext === 'war') return ext === 'jar' || ext === 'war' ? 'java' : 'archive';
  if (IMAGE_EXT.has(ext)) return 'image';
  if (ext === 'json') return 'json';
  if (ext === 'log') return 'log';
  if (CONFIG_EXT.has(ext)) return 'config';
  if (ext === 'md' || ext === 'markdown') return 'markdown';
  if (CODE_EXT.has(ext)) return 'code';
  return 'file';
}

export function formatFileModified(file) {
  const raw = file?.mod_time || file?.modified || file?.ModTime;
  if (!raw) return '';
  const date = raw instanceof Date ? raw : new Date(raw);
  if (Number.isNaN(date.getTime())) return '';
  const pad = (n) => String(n).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
}
