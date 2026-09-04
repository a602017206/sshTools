import assert from 'node:assert/strict';
import test from 'node:test';
import {
  allRowsSelected,
  rowsFromIndexes,
  toggleAllRowSelection,
  toggleRowSelection
} from '../src/lib/tableRowSelection.js';

test('toggleRowSelection 增删单行选中', () => {
  assert.deepEqual([...toggleRowSelection(new Set(), 1)], [1]);
  assert.deepEqual([...toggleRowSelection(new Set([1]), 1)], []);
});

test('toggleAllRowSelection 全选与清空', () => {
  assert.deepEqual([...toggleAllRowSelection(new Set(), 3, true)], [0, 1, 2]);
  assert.deepEqual([...toggleAllRowSelection(new Set([0, 1, 2]), 3, false)], []);
});

test('allRowsSelected 判断当前页是否全选', () => {
  assert.equal(allRowsSelected(new Set([0, 1]), 2), true);
  assert.equal(allRowsSelected(new Set([0]), 2), false);
});

test('rowsFromIndexes 按索引顺序取行', () => {
  const rows = rowsFromIndexes([[1], [2], [3]], new Set([2, 0]));
  assert.deepEqual(rows, [[1], [3]]);
});
