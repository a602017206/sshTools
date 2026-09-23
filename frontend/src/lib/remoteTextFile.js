// 与 internal/ssh/text_file.go 的 MaxRemoteTextBytes 保持一致。
export const MAX_REMOTE_TEXT_BYTES = 1024 * 1024;

const BINARY_EXTENSIONS = new Set([
  'png', 'jpg', 'jpeg', 'gif', 'webp', 'ico', 'bmp', 'tif', 'tiff', 'heic', 'avif',
  'pdf', 'zip', 'gz', 'tgz', 'bz2', 'xz', '7z', 'rar', 'tar', 'zst',
  'jar', 'war', 'ear', 'class', 'so', 'dylib', 'dll', 'exe', 'bin', 'dmg', 'iso', 'img',
  'mp3', 'mp4', 'mov', 'avi', 'mkv', 'wav', 'ogg', 'flac', 'webm',
  'woff', 'woff2', 'ttf', 'otf', 'eot',
  'sqlite', 'db', 'pyc', 'o', 'a', 'wasm', 'parquet', 'avro', 'pcap', 'deb', 'rpm', 'apk',
]);

export function fileBaseName(name) {
  return String(name || '').split('/').pop() || '';
}

export function fileExtension(name) {
  const base = fileBaseName(name).toLowerCase();
  if (base.startsWith('.') && base.indexOf('.', 1) === -1) return '';
  const index = base.lastIndexOf('.');
  if (index <= 0) return '';
  return base.slice(index + 1);
}

export function isEditableRemoteText(file) {
  if (!file || file.is_dir || file.is_parent) return false;
  const size = Number(file.size);
  if (Number.isFinite(size) && size > MAX_REMOTE_TEXT_BYTES) return false;
  const extension = fileExtension(file.name);
  if (extension && BINARY_EXTENSIONS.has(extension)) return false;
  return true;
}
