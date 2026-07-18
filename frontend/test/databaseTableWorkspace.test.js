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
});
