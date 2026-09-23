import assert from 'node:assert/strict';
import test from 'node:test';

import { fileExtension, isEditableRemoteText, MAX_REMOTE_TEXT_BYTES } from '../src/lib/remoteTextFile.js';

test('文本和配置文件可以在线编辑，二进制和超限文件不行', () => {
  assert.equal(isEditableRemoteText({ name: 'app.yml', is_dir: false, size: 20 }), true);
  assert.equal(isEditableRemoteText({ name: 'nginx.conf', is_dir: false, size: 20 }), true);
  assert.equal(isEditableRemoteText({ name: 'Dockerfile', is_dir: false, size: 20 }), true);
  assert.equal(isEditableRemoteText({ name: '.env', is_dir: false, size: 0 }), true);
  assert.equal(isEditableRemoteText({ name: 'notes.txt', is_dir: false }), true);
  assert.equal(isEditableRemoteText({ name: 'photo.png', is_dir: false, size: 20 }), false);
  assert.equal(isEditableRemoteText({ name: 'backup.tar.gz', is_dir: false, size: 20 }), false);
  assert.equal(isEditableRemoteText({ name: 'app.json', is_dir: false, size: MAX_REMOTE_TEXT_BYTES + 1 }), false);
  assert.equal(isEditableRemoteText({ name: 'logs', is_dir: true, size: 0 }), false);
  assert.equal(fileExtension('.gitignore'), '');
  assert.equal(fileExtension('App.YML'), 'yml');
});
