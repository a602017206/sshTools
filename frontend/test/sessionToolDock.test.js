import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('会话工具在切换文件/性能/日志时保持面板挂载', async () => {
  const source = await readFile(new URL('../src/components/SessionToolDock.svelte', import.meta.url), 'utf8');
  assert.match(source, /sshToolPanelHidden\(toolTab, 'files'\)/);
  assert.match(source, /sshToolPanelHidden\(toolTab, 'performance'\)/);
  assert.match(source, /sshToolPanelHidden\(toolTab, 'logs'\)/);
  assert.match(source, /<FileManager \/>/);
  assert.match(source, /<ServerMonitor/);
  assert.match(source, /<SessionLogPanel/);
  assert.match(source, /connectionId/);
  assert.doesNotMatch(source, /{:else if toolTab === 'performance'}/);
});
