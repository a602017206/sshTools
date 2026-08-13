import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('用户名输入框关闭系统自动大写、自动更正和拼写检查', async () => {
  const component = await readFile(new URL('../src/components/AddAssetDialog.svelte', import.meta.url), 'utf8');
  const usernameInput = component.match(/<input\s+type="text"\s+id="connection-username"[\s\S]*?\/>/);

  assert.ok(usernameInput, '应能找到用户名输入框');
  assert.match(usernameInput[0], /autocapitalize="off"/);
  assert.match(usernameInput[0], /autocorrect="off"/);
  assert.match(usernameInput[0], /spellcheck=\{false\}/);
});
