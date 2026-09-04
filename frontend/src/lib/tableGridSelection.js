export function formatCellValue(value) {
  if (value === null || value === undefined) return '';
  return String(value);
}

export function normalizeCellRange(anchor, focus) {
  if (!anchor || !focus) return null;
  return {
    minRow: Math.min(anchor.row, focus.row),
    maxRow: Math.max(anchor.row, focus.row),
    minCol: Math.min(anchor.col, focus.col),
    maxCol: Math.max(anchor.col, focus.col)
  };
}

export function isCellInRange(row, col, range) {
  if (!range) return false;
  return row >= range.minRow
    && row <= range.maxRow
    && col >= range.minCol
    && col <= range.maxCol;
}

export function selectCellRange(anchor, focus) {
  return { anchor, focus };
}

export function extendCellRange(anchor, focus) {
  if (!anchor) return selectCellRange(focus, focus);
  return selectCellRange(anchor, focus);
}

export function selectRowRange(startRow, endRow, rowCount) {
  const minRow = Math.max(0, Math.min(startRow, endRow));
  const maxRow = Math.min(rowCount - 1, Math.max(startRow, endRow));
  const rows = new Set();
  for (let index = minRow; index <= maxRow; index += 1) {
    rows.add(index);
  }
  return rows;
}

export function toggleRowInSet(selectedRows, rowIndex) {
  const next = new Set(selectedRows);
  if (next.has(rowIndex)) next.delete(rowIndex);
  else next.add(rowIndex);
  return next;
}

export function formatSelectionAsTsv(rows, range) {
  if (!range || !rows?.length) return '';
  const lines = [];
  for (let rowIndex = range.minRow; rowIndex <= range.maxRow; rowIndex += 1) {
    const row = rows[rowIndex];
    if (!row) continue;
    const cells = [];
    for (let columnIndex = range.minCol; columnIndex <= range.maxCol; columnIndex += 1) {
      cells.push(formatCellValue(row[columnIndex]));
    }
    lines.push(cells.join('\t'));
  }
  return lines.join('\n');
}

export function formatRowsAsTsv(rows) {
  return rows.map(row => row.map(formatCellValue).join('\t')).join('\n');
}

export function rowsFromSelection(selectedRows) {
  return [...selectedRows].sort((left, right) => left - right);
}

export function selectionCellCount(range) {
  if (!range) return 0;
  return (range.maxRow - range.minRow + 1) * (range.maxCol - range.minCol + 1);
}

export function isRowFullySelected(rowIndex, range, columnCount) {
  if (!range || columnCount <= 0) return false;
  return rowIndex >= range.minRow
    && rowIndex <= range.maxRow
    && range.minCol <= 0
    && range.maxCol >= columnCount - 1;
}

export function rowIndexesFromCellRange(range, columnCount) {
  if (!range || columnCount <= 0) return [];
  const rows = [];
  for (let rowIndex = range.minRow; rowIndex <= range.maxRow; rowIndex += 1) {
    if (isRowFullySelected(rowIndex, range, columnCount)) rows.push(rowIndex);
  }
  return rows;
}
