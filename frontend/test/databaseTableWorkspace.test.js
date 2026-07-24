import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('表数据页面包含数据工作区的核心区域', async () => {
  const source = await readFile(new URL('../src/components/DatabaseTablePanel.svelte', import.meta.url), 'utf8');

  assert.match(source, /class="table-workspace"/);
  assert.match(source, /table-workspace__mode--active/);
  assert.match(source, /table-workspace__toolbar/);
  assert.match(source, /table-workspace__row-number/);
  assert.match(source, /table-workspace__details/);
  assert.match(source, /table-workspace__status/);
  assert.match(source, /table-workspace__query-builder/);
  assert.match(source, /table-workspace__filter-rule/);
  assert.match(source, /table-workspace__sort-rule/);
  assert.match(source, /buildTableBrowseSQL/);
  assert.match(source, /GetTableSchemaInSchema/);
  assert.match(source, /table-workspace__column-metadata/);
  assert.match(source, /formatColumnType/);
  assert.match(source, /formatColumnLength/);
  assert.match(source, /formatColumnDescription/);
  assert.match(source, /table-workspace__context/);
  assert.match(source, /table-workspace__query-strip/);
  assert.match(source, /table-workspace__column-accent/);
  assert.match(source, /table-workspace__inspector/);
  assert.doesNotMatch(source, /title="新增查询"/);
  assert.doesNotMatch(source, /title="设计表"/);
  assert.doesNotMatch(source, /title="新建表"/);
  assert.doesNotMatch(source, /\.table-workspace__tool:disabled \{ opacity: \.5; cursor: wait; \}/);
  assert.match(source, /column\.is_primary_key \|\| column\.primary_key \|\| column\.isPrimaryKey/);
  assert.doesNotMatch(source, /disabled=!hasEdits \|\| !hasPrimaryKey \|\| isMutating/);
  assert.doesNotMatch(source, /disabled=!hasPrimaryKey \|\| isMutating/);
  assert.match(source, /ConfirmDialog/);
  assert.match(source, /删除结果：未匹配到记录/);
  assert.doesNotMatch(source, /confirm\(message\)/);
});
