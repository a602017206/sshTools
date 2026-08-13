import test from 'node:test';
import assert from 'node:assert/strict';
import { fileExtension, resolveFileIconKind, formatFileModified } from './fileType.js';

test('按扩展名解析文件图标类型', () => {
  assert.equal(resolveFileIconKind({ is_dir: true }), 'folder');
  assert.equal(resolveFileIconKind({ is_parent: true }), 'parent');
  assert.equal(resolveFileIconKind({ name: 'deploy.sh' }), 'code');
  assert.equal(resolveFileIconKind({ name: 'app.json' }), 'json');
  assert.equal(resolveFileIconKind({ name: 'pkg.zip' }), 'archive');
  assert.equal(resolveFileIconKind({ name: 'lib.jar' }), 'java');
  assert.equal(resolveFileIconKind({ name: 'cover.png' }), 'image');
  assert.equal(resolveFileIconKind({ name: 'app.log' }), 'log');
  assert.equal(resolveFileIconKind({ name: 'readme.md' }), 'markdown');
  assert.equal(resolveFileIconKind({ name: 'unknown.bin' }), 'file');
});

test('解析扩展名', () => {
  assert.equal(fileExtension('a/b/deploy.sh'), 'sh');
  assert.equal(fileExtension('.env'), 'env');
});

test('格式化修改时间，兼容 mod_time', () => {
  assert.equal(formatFileModified({}), '');
  assert.equal(formatFileModified({ modified: undefined }), '');
  assert.match(formatFileModified({ mod_time: '2026-08-12T02:00:00.000Z' }), /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/);
});
