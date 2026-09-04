import assert from 'node:assert/strict';
import test from 'node:test';

import {
  categoryForDbType,
  groupDatabaseTypesByCategory,
  JDBC_DRIVER_CATEGORIES
} from '../src/lib/databaseTypeCatalog.js';

test('按类型解析连接 category', () => {
  assert.equal(categoryForDbType('mysql'), 'relational');
  assert.equal(categoryForDbType('dm'), 'domestic');
  assert.equal(categoryForDbType('sqlite'), 'lightweight');
  assert.equal(categoryForDbType('redis'), 'document_cache_search');
  assert.equal(categoryForDbType('neo4j'), 'graph_timeseries');
  assert.equal(categoryForDbType('kafka'), 'mq');
  assert.equal(categoryForDbType('rocketmq'), 'mq');
  assert.equal(categoryForDbType('rabbitmq'), 'mq');
});

test('groupDatabaseTypesByCategory 按细分组聚合', () => {
  const groups = groupDatabaseTypesByCategory([
    { value: 'mysql', label: 'MySQL' },
    { value: 'dm', label: 'DM' },
    { value: 'sqlite', label: 'SQLite' },
    { value: 'redis', label: 'Redis' },
    { value: 'kafka', label: 'Kafka' },
    { value: 'rocketmq', label: 'RocketMQ' },
    { value: 'rabbitmq', label: 'RabbitMQ' }
  ]);
  assert.deepEqual(groups.map((group) => group.id), [
    'relational',
    'domestic',
    'lightweight',
    'document_cache_search',
    'mq'
  ]);
  assert.deepEqual(groups.find((group) => group.id === 'mq').types.map((item) => item.value), [
    'kafka',
    'rocketmq',
    'rabbitmq'
  ]);
});

test('JDBC 驱动分类不含消息队列', () => {
  assert.ok(!JDBC_DRIVER_CATEGORIES.some((item) => item.id === 'mq'));
  assert.ok(JDBC_DRIVER_CATEGORIES.some((item) => item.id === 'relational'));
});
