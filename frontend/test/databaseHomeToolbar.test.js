import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('数据库主页提供独立于表的数据操作入口', async () => {
  const [home, terminal] = await Promise.all([
    readFile(new URL('../src/components/SelectedDatabaseObjects.svelte', import.meta.url), 'utf8'),
    readFile(new URL('../src/components/TerminalPanel.svelte', import.meta.url), 'utf8')
  ]);

  assert.match(home, /title="新增查询"/);
  assert.match(home, /title="设计表"/);
  assert.match(home, /title="新建表"/);
  assert.match(home, /title="刷新"/);
  assert.match(home, /database:new-query/);
  assert.match(home, /function createTable/);
  assert.match(terminal, /openDatabaseQueryPanel/);
  assert.match(terminal, /openDatabaseTableDesignerPanel/);
  assert.match(terminal, /panelType: 'database-table-designer'/);
  assert.match(terminal, /panelType: 'database-query'/);
});
