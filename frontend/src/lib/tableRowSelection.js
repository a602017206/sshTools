export function toggleRowSelection(selectedIndexes, rowIndex) {
  const next = new Set(selectedIndexes);
  if (next.has(rowIndex)) {
    next.delete(rowIndex);
  } else {
    next.add(rowIndex);
  }
  return next;
}

export function toggleAllRowSelection(selectedIndexes, rowCount, selectAll) {
  if (!selectAll) {
    return new Set();
  }
  return new Set(Array.from({ length: rowCount }, (_, index) => index));
}

export function allRowsSelected(selectedIndexes, rowCount) {
  return rowCount > 0 && selectedIndexes.size === rowCount;
}

export function rowsFromIndexes(filteredRows, selectedIndexes) {
  return [...selectedIndexes]
    .sort((left, right) => left - right)
    .map(index => filteredRows[index])
    .filter(Boolean);
}
