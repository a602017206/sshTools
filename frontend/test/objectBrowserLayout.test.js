import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('对象列表在主内容区内滚动并保留可见滚动轨道', async () => {
  const source = await readFile(new URL('../src/components/SelectedDatabaseObjects.svelte', import.meta.url), 'utf8');

  assert.match(source, /\.object-browser__main \{[^}]*min-height: 0/);
  assert.match(source, /\.object-browser__table \{[^}]*overflow-y: auto/);
  assert.match(source, /\.object-browser__table::-webkit-scrollbar/);
});

test('选中表后在详情侧栏加载只读表结构预览', async () => {
  const source = await readFile(new URL('../src/components/SelectedDatabaseObjects.svelte', import.meta.url), 'utf8');

  assert.match(source, /GetTableSchemaInSchema/);
  assert.match(source, /表结构/);
  assert.match(source, /selectedTableSchema/);
  assert.match(source, /object-browser__table-structure/);
});

test('详情侧栏支持信息与 DDL 切换、拖拽调整和显示开关', async () => {
  const source = await readFile(new URL('../src/components/SelectedDatabaseObjects.svelte', import.meta.url), 'utf8');

  assert.match(source, /GetTableDDLInSchema/);
  assert.match(source, /detailsMode/);
  assert.match(source, /startDetailsResize/);
  assert.match(source, /detailsVisible/);
  assert.match(source, /\.object-browser__main \{[^}]*flex: 1 1 0/);
});
