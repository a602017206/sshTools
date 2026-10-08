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

const LONG_TEXT_TYPE_PATTERN = /text|json|xml|clob|blob|bytea/i;
const MAX_CELL_TITLE_LENGTH = 100;

export function isLongTextColumn(column) {
  return LONG_TEXT_TYPE_PATTERN.test(String(column?.type || ''));
}

export function getCellTitle(cell, column) {
  if (cell === null || cell === undefined) return 'NULL';
  const text = String(cell);
  if (isLongTextColumn(column) || text.length > MAX_CELL_TITLE_LENGTH || /[\r\n]/.test(text)) return undefined;
  return text;
}

export function buildGridTemplateColumns(columns, columnWidths = {}, columnMetadata = {}) {
  const widths = columns.map(column => {
    const width = columnWidths[column] ?? getInitialColumnWidth(column, columnMetadata[column]);
    return `${clampColumnWidth(width)}px`;
  });

  return [`${ROW_NUMBER_WIDTH}px`, ...widths].join(' ');
}
