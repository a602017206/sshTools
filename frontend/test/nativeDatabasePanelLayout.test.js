import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('原生数据库面板使用独立主题内容区和资源展示模型', async () => {
  const source = await readFile(new URL('../src/components/NativeDatabasePanel.svelte', import.meta.url), 'utf8');

  assert.match(source, /nativeDatabaseWorkspace\(databaseType\)/);
  assert.match(source, /background:\s*var\(--bg-primary\)/);
  assert.match(source, /border:\s*1px solid var\(--border-primary\)/);
  assert.match(source, /native-database-panel__context/);
  assert.match(source, /native-database-panel__resource-count/);
  assert.match(source, /native-database-panel__inspector/);
});
