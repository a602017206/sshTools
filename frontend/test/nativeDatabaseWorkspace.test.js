import assert from 'node:assert/strict';
import test from 'node:test';

import {
  nativeDatabaseWorkspace,
  resolveNativeWorkspaceKind
} from '../src/lib/nativeDatabaseWorkspace.js';

test('Redis 工作区使用顶部逻辑库下拉并支持读写', () => {
  const workspace = nativeDatabaseWorkspace('redis');
  assert.equal(workspace.dbSelector, 'dropdown');
  assert.equal(workspace.canExpand, false);
  assert.equal(workspace.canDescribe, true);
  assert.equal(workspace.canWrite, true);
  assert.equal(workspace.canDelete, true);
});

test('Elasticsearch 工作区支持查询、搜索与可调分栏', () => {
  const workspace = nativeDatabaseWorkspace('elasticsearch');
  assert.equal(workspace.canExpand, false);
  assert.equal(workspace.canDescribe, true);
  assert.equal(workspace.canQuery, true);
  assert.equal(workspace.canWrite, true);
  assert.equal(workspace.canDelete, true);
  assert.equal(workspace.canSearchResources, true);
  assert.equal(workspace.canResizeInspector, true);
  assert.equal(workspace.showSessionOverview, true);
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

test('resolveNativeWorkspaceKind 按类型映射到独立工作区', () => {
  assert.equal(resolveNativeWorkspaceKind('redis'), 'redis');
  assert.equal(resolveNativeWorkspaceKind('Redis'), 'redis');
  assert.equal(resolveNativeWorkspaceKind('elasticsearch'), 'elasticsearch');
  assert.equal(resolveNativeWorkspaceKind('opensearch'), 'elasticsearch');
  assert.equal(resolveNativeWorkspaceKind('kafka'), 'kafka');
  assert.equal(resolveNativeWorkspaceKind('mongodb'), 'generic');
  assert.equal(resolveNativeWorkspaceKind('cassandra'), 'generic');
  assert.equal(resolveNativeWorkspaceKind(''), 'generic');
  assert.equal(resolveNativeWorkspaceKind(null), 'generic');
});
