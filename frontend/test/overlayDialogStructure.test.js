import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('开发工具复用全局设置使用的 Dialog 门户组件', async () => {
  const component = await readFile(new URL('../src/components/DevToolsPanel.svelte', import.meta.url), 'utf8');

  assert.match(component, /import Dialog from '\.\/ui\/Dialog\.svelte';/);
  assert.match(component, /<Dialog bind:isOpen=\{isOpen\}/);
  assert.doesNotMatch(component, /class="fixed inset-0 z-\[200\]/);
});

test('上传任务使用独立的 Dialog 组件，不直接插入主工作区', async () => {
  const app = await readFile(new URL('../src/App.svelte', import.meta.url), 'utf8');

  assert.match(app, /import UploadTaskDialog from '\.\/components\/UploadTaskDialog\.svelte';/);
  assert.match(app, /<UploadTaskDialog\s*\/>/);
  assert.doesNotMatch(app, /\{#if \$uploadStore\.isPanelOpen\}[\s\S]*?class="fixed inset-0 z-\[190\]/);
});
