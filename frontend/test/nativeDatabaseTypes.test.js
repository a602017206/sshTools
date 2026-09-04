import assert from 'node:assert/strict';
import test from 'node:test';

import {
  assetSupportsJdbcSidebar,
  connectionAllowsEmptyPassword,
  databaseTypeConfig,
  databaseTypeRequiresPassword,
  databaseTypeRequiresUsername,
  databaseTypeSupportsAuthNone,
  isNativeDatabaseType
} from '../src/lib/nativeDatabaseTypes.js';

test('原生数据库类型不需要 JDBC profile', () => {
  assert.equal(isNativeDatabaseType('redis'), true);
  assert.equal(isNativeDatabaseType('mongodb'), true);
  assert.equal(isNativeDatabaseType('elasticsearch'), true);
  assert.equal(isNativeDatabaseType('opensearch'), true);
  assert.equal(isNativeDatabaseType('OpenSearch'), true);
  assert.equal(isNativeDatabaseType('memcached'), true);
  assert.equal(isNativeDatabaseType('cassandra'), true);
  assert.equal(isNativeDatabaseType('couchbase'), true);
  assert.equal(isNativeDatabaseType('influxdb'), true);
  assert.equal(isNativeDatabaseType('neo4j'), true);
  assert.equal(isNativeDatabaseType('kafka'), true);
  assert.equal(isNativeDatabaseType('rocketmq'), true);
  assert.equal(isNativeDatabaseType('rabbitmq'), true);
  assert.equal(isNativeDatabaseType('mysql'), false);
});

test('缓存/搜索/消息等原生资产不展开 JDBC 库树', () => {
  assert.equal(assetSupportsJdbcSidebar({ type: 'database', metadata: { db_type: 'mysql' } }), true);
  assert.equal(assetSupportsJdbcSidebar({ type: 'database', metadata: { db_type: 'postgresql' } }), true);
  assert.equal(assetSupportsJdbcSidebar({ type: 'database', metadata: { db_type: 'redis' } }), false);
  assert.equal(assetSupportsJdbcSidebar({ type: 'database', metadata: { db_type: 'elasticsearch' } }), false);
  assert.equal(assetSupportsJdbcSidebar({ type: 'database', metadata: { db_type: 'opensearch' } }), false);
  assert.equal(assetSupportsJdbcSidebar({ type: 'database', metadata: { db_type: 'kafka' } }), false);
  assert.equal(assetSupportsJdbcSidebar({ type: 'database', metadata: { db_type: 'rocketmq' } }), false);
  assert.equal(assetSupportsJdbcSidebar({ type: 'database', metadata: { db_type: 'rabbitmq' } }), false);
  assert.equal(assetSupportsJdbcSidebar({ type: 'database', metadata: { db_type: 'mongodb' } }), false);
  assert.equal(assetSupportsJdbcSidebar({ type: 'ssh' }), false);
  assert.equal(assetSupportsJdbcSidebar(null), false);
});

test('原生数据库类型提供默认端口和资源标签', () => {
  assert.deepEqual(databaseTypeConfig('redis'), {
    port: '6379',
    resourceLabel: '键',
    requiresUsername: false
  });
  assert.deepEqual(databaseTypeConfig('mongodb'), {
    port: '27017',
    resourceLabel: '集合'
  });
  assert.deepEqual(databaseTypeConfig('elasticsearch'), {
    port: '9200',
    resourceLabel: '索引'
  });
  assert.deepEqual(databaseTypeConfig('memcached'), {
    port: '11211',
    resourceLabel: '统计项'
  });
  assert.deepEqual(databaseTypeConfig('cassandra'), {
    port: '9042',
    resourceLabel: '表'
  });
  assert.deepEqual(databaseTypeConfig('couchbase'), {
    port: '8091',
    resourceLabel: '集合'
  });
  assert.deepEqual(databaseTypeConfig('influxdb'), { port: '8086', resourceLabel: '资源' });
  assert.deepEqual(databaseTypeConfig('neo4j'), { port: '7687', resourceLabel: '资源' });
  assert.deepEqual(databaseTypeConfig('kafka'), {
    port: '9092',
    resourceLabel: '主题',
    requiresUsername: false,
    requiresPassword: false,
    supportsAuthNone: true
  });
  assert.deepEqual(databaseTypeConfig('rocketmq'), {
    port: '9876',
    resourceLabel: '主题',
    requiresUsername: false,
    requiresPassword: false,
    supportsAuthNone: true
  });
  assert.deepEqual(databaseTypeConfig('rabbitmq'), {
    port: '5672',
    resourceLabel: '队列',
    requiresUsername: false,
    requiresPassword: false,
    supportsAuthNone: true
  });
});

test('Redis 使用仅密码认证，其他原生数据库默认需要用户名', () => {
  assert.equal(databaseTypeRequiresUsername('redis'), false);
  assert.equal(databaseTypeRequiresUsername('mongodb'), true);
  assert.equal(databaseTypeRequiresUsername('kafka'), false);
});

test('Kafka 支持无认证且密码非必填', () => {
  assert.equal(databaseTypeRequiresPassword('kafka'), false);
  assert.equal(databaseTypeRequiresPassword('rocketmq'), false);
  assert.equal(databaseTypeRequiresPassword('rabbitmq'), false);
  assert.equal(databaseTypeRequiresPassword('mysql'), true);
  assert.equal(databaseTypeRequiresPassword('redis'), true);
  assert.equal(databaseTypeSupportsAuthNone('kafka'), true);
  assert.equal(databaseTypeSupportsAuthNone('rocketmq'), true);
  assert.equal(databaseTypeSupportsAuthNone('rabbitmq'), true);
  assert.equal(databaseTypeSupportsAuthNone('redis'), false);
  assert.equal(connectionAllowsEmptyPassword({ metadata: { db_type: 'kafka' }, auth_type: 'none' }), true);
  assert.equal(connectionAllowsEmptyPassword({ metadata: { db_type: 'kafka' }, auth_type: 'password' }), true);
  assert.equal(connectionAllowsEmptyPassword({ metadata: { db_type: 'mysql' }, auth_type: 'password' }), false);
});
