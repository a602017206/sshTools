import assert from 'node:assert/strict';
import test from 'node:test';

import {
  formatColumnLength,
  formatColumnType,
  formatColumnDescription
} from '../src/lib/tableStructureMetadata.js';

test('字段元数据格式化展示类型、长度和描述', () => {
  const column = {
    type: 'DECIMAL',
    column_size: 12,
    decimal_digits: 2,
    description: '订单金额'
  };

  assert.equal(formatColumnType(column), 'DECIMAL(12,2)');
  assert.equal(formatColumnLength(column), '12');
  assert.equal(formatColumnDescription(column), '订单金额');
});

test('字段元数据对无长度和描述的类型显示占位符', () => {
  assert.equal(formatColumnType({ type: 'DATE' }), 'DATE');
  assert.equal(formatColumnLength({ type: 'DATE', column_size: 0 }), '-');
  assert.equal(formatColumnDescription({ description: '   ' }), '-');
});
