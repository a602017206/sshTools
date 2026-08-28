import assert from 'node:assert/strict';
import test from 'node:test';

import { nativeDatabaseWorkspace } from '../src/lib/nativeDatabaseWorkspace.js';

test('Redis 工作区使用顶部逻辑库下拉并支持读写', () => {
  const workspace = nativeDatabaseWorkspace('redis');
  assert.equal(workspace.dbSelector, 'dropdown');
  assert.equal(workspace.canExpand, false);
  assert.equal(workspace.canDescribe, true);
  assert.equal(workspace.canWrite, true);
  assert.equal(workspace.canDelete, true);
});

test('Elasticsearch 工作区支持查询与文档变更', () => {
  const workspace = nativeDatabaseWorkspace('elasticsearch');
  assert.equal(workspace.canExpand, false);
  assert.equal(workspace.canDescribe, true);
  assert.equal(workspace.canQuery, true);
  assert.equal(workspace.canWrite, true);
  assert.equal(workspace.canDelete, true);
});

test('未知原生类型使用安全的只读资源回退', () => {
  assert.deepEqual(nativeDatabaseWorkspace('custom'), {
    title: '原生数据库资源',
    resourceLabel: '资源',
    childLabel: '',
    description: '当前为只读资源浏览。',
    canExpand: false,
    canDescribe: false
  });
});
