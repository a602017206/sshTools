import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('对象列表在主内容区内滚动并保留可见滚动轨道', async () => {
  const source = await readFile(new URL('../src/components/SelectedDatabaseObjects.svelte', import.meta.url), 'utf8');

  assert.match(source, /\.object-browser__main \{[^}]*min-height: 0/);
  assert.match(source, /\.object-browser__table \{[^}]*overflow-y: auto/);
  assert.match(source, /\.object-browser__table::-webkit-scrollbar/);
});
