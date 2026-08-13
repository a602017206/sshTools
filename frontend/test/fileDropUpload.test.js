import assert from 'node:assert/strict';
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
