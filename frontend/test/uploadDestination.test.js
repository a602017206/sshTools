import assert from 'node:assert/strict';
import test from 'node:test';

import { resolveUploadDestination } from '../src/lib/uploadDestination.js';

test('未选中或选中文件时上传到当前目录', () => {
  const files = [
    { path: '/home/a.txt', is_dir: false, is_parent: false },
    { path: '/home/logs', is_dir: true, is_parent: false },
    { path: '/home', is_dir: true, is_parent: true },
  ];
  assert.equal(resolveUploadDestination({ currentPath: '/home', selectedPaths: [], files }), '/home');
  assert.equal(resolveUploadDestination({
    currentPath: '/home',
    selectedPaths: ['/home/a.txt'],
    files,
  }), '/home');
});

test('恰好选中一个文件夹时上传到该文件夹', () => {
  const files = [
    { path: '/home/logs', name: 'logs', is_dir: true, is_parent: false },
    { path: '/home/src', name: 'src', is_dir: true, is_parent: false },
  ];
  assert.equal(resolveUploadDestination({
    currentPath: '/home',
    selectedPaths: ['/home/logs'],
    files,
  }), '/home/logs');
  assert.equal(resolveUploadDestination({
    currentPath: '/home',
    selectedPaths: ['/home/logs', '/home/src'],
    files,
  }), '/home');
});
