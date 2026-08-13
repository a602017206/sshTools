import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

import { shouldResetBlankConnectionForm } from '../src/lib/connectionDialogRequest.js';

test('新的新建请求必须重置表单，即使弹窗此前展示的是编辑内容', () => {
  assert.equal(shouldResetBlankConnectionForm({
    isOpen: true,
    requestVersion: 2,
    appliedRequestVersion: 1,
    editingAsset: null,
    cloningAsset: null,
  }), true);
});

test('编辑、克隆、未变化请求或关闭状态不会当作新建重置', () => {
  const base = { isOpen: true, requestVersion: 2, appliedRequestVersion: 1 };

  assert.equal(shouldResetBlankConnectionForm({ ...base, editingAsset: { id: 'A' }, cloningAsset: null }), false);
  assert.equal(shouldResetBlankConnectionForm({ ...base, editingAsset: null, cloningAsset: { name: 'A copy' } }), false);
  assert.equal(shouldResetBlankConnectionForm({ ...base, appliedRequestVersion: 2, editingAsset: null, cloningAsset: null }), false);
  assert.equal(shouldResetBlankConnectionForm({ ...base, isOpen: false, editingAsset: null, cloningAsset: null }), false);
});

test('每个连接弹窗请求均重建弹窗实例，避免保留上一种操作的局部状态', async () => {
  const app = await readFile(new URL('../src/App.svelte', import.meta.url), 'utf8');

  assert.match(app, /\{#key connectionDialogRequestVersion\}/);
  assert.match(app, /\{\/key\}/);
});
