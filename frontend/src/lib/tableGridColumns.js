const MIN_COLUMN_WIDTH = 88;
const MAX_COLUMN_WIDTH = 640;
const ROW_NUMBER_WIDTH = 48;

export function clampColumnWidth(width) {
  const numericWidth = Number(width);
  if (!Number.isFinite(numericWidth)) return MIN_COLUMN_WIDTH;
  return Math.min(MAX_COLUMN_WIDTH, Math.max(MIN_COLUMN_WIDTH, Math.round(numericWidth)));
}

export function getInitialColumnWidth(columnName, column = {}) {
  const labelLength = Math.max(
    String(columnName || '').length,
    String(column?.type || '').length
  );

  return clampColumnWidth(48 + labelLength * 8);
}

export function buildGridTemplateColumns(columns, columnWidths = {}, columnMetadata = {}) {
  const widths = columns.map(column => {
    const width = columnWidths[column] ?? getInitialColumnWidth(column, columnMetadata[column]);
    return `${clampColumnWidth(width)}px`;
  });

  return [`${ROW_NUMBER_WIDTH}px`, ...widths].join(' ');
}
