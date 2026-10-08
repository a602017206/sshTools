import assert from 'node:assert/strict';
import test from 'node:test';
import {
  buildGridTemplateColumns,
  clampColumnWidth,
  getCellTitle,
  getInitialColumnWidth
} from '../src/lib/tableGridColumns.js';

test('text 等长文本列不显示悬停预览，普通列和 NULL 保持提示', () => {
  assert.equal(getCellTitle('<p>html</p>', { type: 'text' }), undefined);
  assert.equal(getCellTitle('{}', { type: 'LONGTEXT' }), undefined);
  assert.equal(getCellTitle('{}', { type: 'jsonb' }), undefined);
  assert.equal(getCellTitle(null, { type: 'text' }), 'NULL');
  assert.equal(getCellTitle('root', { type: 'varchar' }), 'root');
  assert.equal(getCellTitle(1, undefined), '1');
});

test('缺少列元数据时，超长或多行内容也不显示悬停预览', () => {
  assert.equal(getCellTitle('<p>'.padEnd(500, 'x'), undefined), undefined);
  assert.equal(getCellTitle('line1\nline2', undefined), undefined);
  assert.equal(getCellTitle('短文本', undefined), '短文本');
});

test('根据字段名和类型生成紧凑的初始列宽，不使用字段声明长度或描述撑宽', () => {
  const width = getInitialColumnWidth('config_title', {
    type: 'VARCHAR',
    column_size: 255,
    description: '用户检索条件的标题'
  });

  assert.equal(width, 144);
});

test('网格轨道使用固定列宽，单列不会占满剩余可用空间', () => {
  const template = buildGridTemplateColumns(
    ['id', 'config_title'],
    { id: 96 },
    { config_title: { type: 'VARCHAR', column_size: 255 } }
  );

  assert.equal(template, '48px 96px 144px');
});

test('手动调整列宽时限制最小和最大值', () => {
  assert.equal(clampColumnWidth(20), 88);
  assert.equal(clampColumnWidth(260), 260);
  assert.equal(clampColumnWidth(1000), 640);
});
