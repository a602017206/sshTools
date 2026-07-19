export function formatColumnType(column) {
  const type = String(column?.type || '').trim();
  if (!type) return '-';

  const size = Number(column?.column_size || 0);
  const scale = Number(column?.decimal_digits || 0);
  const normalizedType = type.toLowerCase();
  if (size > 0 && (normalizedType.includes('char') || normalizedType.includes('varchar') || normalizedType.includes('decimal') || normalizedType.includes('numeric'))) {
    return scale > 0 && (normalizedType.includes('decimal') || normalizedType.includes('numeric'))
      ? `${type}(${size},${scale})`
      : `${type}(${size})`;
  }
  return type;
}

export function formatColumnLength(column) {
  const size = Number(column?.column_size || 0);
  return size > 0 ? String(size) : '-';
}

export function formatColumnDescription(column) {
  const description = String(column?.description || '').trim();
  return description || '-';
}
