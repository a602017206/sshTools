import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('表结构面板提供 Navicat 式字段设计与新建模式', async () => {
  const [panel, store, objects, terminal] = await Promise.all([
    readFile(new URL('../src/components/TableStructurePanel.svelte', import.meta.url), 'utf8'),
    readFile(new URL('../src/stores.js', import.meta.url), 'utf8'),
    readFile(new URL('../src/components/SelectedDatabaseObjects.svelte', import.meta.url), 'utf8'),
    readFile(new URL('../src/components/TerminalPanel.svelte', import.meta.url), 'utf8')
  ]);

  assert.match(panel, /设计表/);
  assert.match(panel, /添加字段/);
  assert.match(panel, /保存新表/);
  assert.match(panel, /保存结构/);
  assert.match(panel, /buildAlterTableStatements/);
  assert.match(panel, /'kingbase'/);
  assert.match(panel, /disabled={fieldsReadOnly}/);
  assert.match(panel, /fieldDrafts/);
  assert.match(panel, /buildCreateTableSQL\(\{ databaseType, databaseName, schemaName, tableName: draftTableName, fields: fieldDrafts \}\)/);
  assert.match(store, /mode: 'design'/);
  assert.match(objects, /mode: 'create'/);
  assert.match(terminal, /<TableStructurePanel/);
});
