import assert from 'node:assert/strict';
import test from 'node:test';

import { nativeDatabaseWorkspace } from '../src/lib/nativeDatabaseWorkspace.js';

test('Redis 工作区将逻辑库标记为可展开，并说明键浏览范围', () => {
  assert.deepEqual(nativeDatabaseWorkspace('redis'), {
    title: 'Redis 键空间',
    resourceLabel: '逻辑库',
    childLabel: '键',
    description: '展开逻辑库可浏览其中的键；当前为只读浏览。',
    canExpand: true
  });
});

test('Elasticsearch 工作区将索引作为只读叶子资源', () => {
  assert.deepEqual(nativeDatabaseWorkspace('elasticsearch'), {
    title: 'Elasticsearch 索引',
    resourceLabel: '索引',
    childLabel: '',
    description: '当前可浏览索引；文档查询与编辑尚未提供。',
    canExpand: false
  });
});

test('未知原生类型使用安全的只读资源回退', () => {
  assert.deepEqual(nativeDatabaseWorkspace('custom'), {
    title: '原生数据库资源',
    resourceLabel: '资源',
    childLabel: '',
    description: '当前为只读资源浏览。',
    canExpand: false
  });
});
