import assert from 'node:assert/strict';
import test from 'node:test';
import {
  formatRowsAsTsv,
  formatSelectionAsTsv,
  isCellInRange,
  normalizeCellRange,
  rowIndexesFromCellRange,
  selectRowRange,
  selectionCellCount,
  toggleRowInSet
} from '../src/lib/tableGridSelection.js';

test('normalizeCellRange 规范化矩形选区', () => {
  assert.deepEqual(
    normalizeCellRange({ row: 3, col: 2 }, { row: 1, col: 0 }),
    { minRow: 1, maxRow: 3, minCol: 0, maxCol: 2 }
  );
});

test('formatSelectionAsTsv 导出选区为制表符文本', () => {
  const rows = [['a', 'b'], ['c', 'd']];
  const range = normalizeCellRange({ row: 0, col: 0 }, { row: 1, col: 1 });
  assert.equal(formatSelectionAsTsv(rows, range), 'a\tb\nc\td');
});

test('formatRowsAsTsv 导出多行', () => {
  assert.equal(formatRowsAsTsv([[1, 2], [3, null]]), '1\t2\n3\t');
});

test('selectRowRange 与 toggleRowInSet 管理行选中', () => {
  assert.deepEqual([...selectRowRange(1, 3, 5)], [1, 2, 3]);
  assert.deepEqual([...toggleRowInSet(new Set([1]), 1)], []);
  assert.deepEqual([...toggleRowInSet(new Set([1]), 2)], [1, 2]);
});

test('rowIndexesFromCellRange 识别整行选中的行', () => {
  const range = normalizeCellRange({ row: 0, col: 0 }, { row: 2, col: 1 });
  assert.deepEqual(rowIndexesFromCellRange(range, 2), [0, 1, 2]);
  assert.equal(isCellInRange(1, 1, range), true);
  assert.equal(selectionCellCount(range), 6);
});
