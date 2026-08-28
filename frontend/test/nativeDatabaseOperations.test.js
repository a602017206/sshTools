import assert from 'node:assert/strict';
import test from 'node:test';

import {
  buildRedisSavePayload,
  buildRedisSetPayload,
  createRedisEditorState,
  defaultElasticsearchQuery,
  parseElasticsearchQueryHits,
  redisDatabaseOptions,
  redisInspectorState
} from '../src/lib/nativeDatabaseOperations.js';

test('defaultElasticsearchQuery 提供 match_all 模板', () => {
  const parsed = JSON.parse(defaultElasticsearchQuery(10));
  assert.deepEqual(parsed.query, { match_all: {} });
  assert.equal(parsed.size, 10);
});

test('buildRedisSetPayload 作为 string save 兼容', () => {
  assert.deepEqual(JSON.parse(buildRedisSetPayload('hello', '30')), {
    type: 'string',
    value: 'hello',
    ttlSeconds: 30
  });
});

test('createRedisEditorState 解析各 Redis 类型', () => {
  assert.deepEqual(createRedisEditorState(JSON.stringify({
    type: 'list',
    items: ['a', 'b'],
    ttlSeconds: 30,
    length: 2
  })), {
    type: 'list',
    value: '',
    fields: [],
    items: ['a', 'b'],
    members: [],
    entries: [],
    ttlSeconds: 30,
    truncated: false,
    length: 2
  });
});

test('buildRedisSavePayload 生成 list/hash 保存内容', () => {
  assert.deepEqual(JSON.parse(buildRedisSavePayload({
    type: 'list',
    items: ['first', 'second']
  }, '60')), {
    type: 'list',
    items: ['first', 'second'],
    ttlSeconds: 60
  });
  assert.deepEqual(JSON.parse(buildRedisSavePayload({
    type: 'hash',
    fields: [{ field: 'name', value: 'Ada' }]
  }, '')), {
    type: 'hash',
    fields: { name: 'Ada' }
  });
});

test('redisInspectorState 识别全部可编辑类型', () => {
  for (const type of ['string', 'hash', 'list', 'set', 'zset']) {
    assert.equal(redisInspectorState(JSON.stringify({ type })).editable, true, type);
  }
});

test('redisDatabaseOptions 生成逻辑库下拉项', () => {
  assert.deepEqual(redisDatabaseOptions([{ name: '0' }, { name: '3' }]), [
    { value: '0', label: 'DB 0' },
    { value: '3', label: 'DB 3' }
  ]);
});

test('parseElasticsearchQueryHits 解析命中列表', () => {
  const hits = parseElasticsearchQueryHits(JSON.stringify({
    hits: [{ _id: 'p-1', _source: { name: 'Keyboard' } }]
  }));
  assert.equal(hits.length, 1);
  assert.equal(hits[0].id, 'p-1');
});
