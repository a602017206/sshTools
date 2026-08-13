import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

import { shouldApplyClone } from '../src/lib/cloneDialogState.js';

test('打开的对话框收到新的克隆请求时应用预填数据一次', () => {
  const firstClone = { name: 'source copy 1' };
  const secondClone = { name: 'source copy 2' };

  assert.equal(shouldApplyClone({ isOpen: true, cloningAsset: firstClone, appliedCloningAsset: null }), true);
  assert.equal(shouldApplyClone({ isOpen: true, cloningAsset: firstClone, appliedCloningAsset: firstClone }), false);
  assert.equal(shouldApplyClone({ isOpen: true, cloningAsset: secondClone, appliedCloningAsset: firstClone }), true);
  assert.equal(shouldApplyClone({ isOpen: false, cloningAsset: firstClone, appliedCloningAsset: null }), false);
});

test('克隆连接落入表单时保留已获取的密码和保存密码标记', async () => {
  const component = await readFile(new URL('../src/components/AddAssetDialog.svelte', import.meta.url), 'utf8');
  const cloneFunction = component.match(/function applyCloningAsset\(\) \{[\s\S]*?\n  \}/);

  assert.ok(cloneFunction, '应能找到克隆表单预填函数');
  assert.doesNotMatch(cloneFunction[0], /password:\s*''/);
  assert.doesNotMatch(cloneFunction[0], /savePassword:\s*false/);
});
