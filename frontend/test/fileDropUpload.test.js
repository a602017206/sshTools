import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

import { canUploadDroppedFiles, normalizeDroppedFilePaths } from '../src/lib/fileDropUpload.js';

test('仅在远程已连接会话中接受拖入的本地文件路径', () => {
  assert.equal(canUploadDroppedFiles({ sessionId: 'ssh-1', connected: true, isLocal: false }), true);
  assert.equal(canUploadDroppedFiles({ sessionId: '', connected: true, isLocal: false }), false);
  assert.equal(canUploadDroppedFiles({ sessionId: 'ssh-1', connected: false, isLocal: false }), false);
  assert.equal(canUploadDroppedFiles({ sessionId: 'local-1', connected: true, isLocal: true }), false);
});

test('规范化拖放路径并去除重复和空项', () => {
  assert.deepEqual(
    normalizeDroppedFilePaths([' /Users/me/a.txt ', '', '/Users/me/a.txt', null, '/Users/me/b.txt']),
    ['/Users/me/a.txt', '/Users/me/b.txt']
  );
});

test('文件管理器粘贴本地剪贴板文件到当前或选中目录，并走串行上传', async () => {
  const manager = await readFile(new URL('../src/components/FileManager.svelte', import.meta.url), 'utf8');
  assert.match(manager, /GetClipboardFilePaths/);
  assert.match(manager, /handleSmartPaste/);
  assert.match(manager, /resolveUploadDestination/);
  assert.match(manager, /uploadLocalPaths\(localPaths, destPath\)/);
});

test('文件夹上传由后端串行排队，不再为每个文件单独订阅进度', async () => {
  const [manager, service] = await Promise.all([
    readFile(new URL('../src/components/FileManager.svelte', import.meta.url), 'utf8'),
    readFile(new URL('../../internal/service/sftp_service.go', import.meta.url), 'utf8'),
  ]);
  assert.match(service, /runFolderUpload/);
  assert.match(service, /runItemUpload/);
  assert.doesNotMatch(service, /startFileUpload\(\s*\n\s*sessionID,\s*\n\s*sftpClient/);
  assert.match(manager, /transferIDs\.forEach\(\(id\) => subscribeToTransfer\(id, 'upload'\)\)/);
});

test('文件管理器提供文件夹上传入口，拖放与按钮共用 UploadFiles', async () => {
  const [manager, menu] = await Promise.all([
    readFile(new URL('../src/components/FileManager.svelte', import.meta.url), 'utf8'),
    readFile(new URL('../src/components/FileManagerContextMenu.svelte', import.meta.url), 'utf8')
  ]);
  assert.match(manager, /SelectUploadDirectory/);
  assert.match(manager, /handleUploadFolder/);
  assert.match(manager, /case 'uploadFolder'/);
  assert.match(manager, /resolveUploadConflicts/);
  assert.match(manager, /ExpandUploadPaths/);
  assert.match(manager, /UploadExpandedItems/);
  assert.match(menu, /选择文件夹上传/);
});
