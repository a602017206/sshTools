import assert from 'node:assert/strict';
import test from 'node:test';

import { databaseTypeConfig, isNativeDatabaseType } from '../src/lib/nativeDatabaseTypes.js';

test('原生数据库类型不需要 JDBC profile', () => {
  assert.equal(isNativeDatabaseType('redis'), true);
  assert.equal(isNativeDatabaseType('mongodb'), true);
  assert.equal(isNativeDatabaseType('elasticsearch'), true);
  assert.equal(isNativeDatabaseType('memcached'), true);
  assert.equal(isNativeDatabaseType('cassandra'), true);
  assert.equal(isNativeDatabaseType('couchbase'), true);
  assert.equal(isNativeDatabaseType('mysql'), false);
});

test('原生数据库类型提供默认端口和资源标签', () => {
  assert.deepEqual(databaseTypeConfig('redis'), {
    port: '6379',
    resourceLabel: '键'
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
});
