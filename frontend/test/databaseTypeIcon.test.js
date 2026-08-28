import assert from 'node:assert/strict';
import test from 'node:test';

import {
  databaseTypeLabel,
  resolveDatabaseType,
  resolveDatabaseTypeIconKind
} from '../src/lib/databaseTypeIcon.js';

test('从资产 metadata 或 dbType 解析数据库类型', () => {
  assert.equal(resolveDatabaseType({ metadata: { db_type: 'redis' } }), 'redis');
  assert.equal(resolveDatabaseType({ dbType: 'Elasticsearch' }), 'elasticsearch');
  assert.equal(resolveDatabaseType({ db_type: 'MYSQL' }), 'mysql');
  assert.equal(resolveDatabaseType({}), '');
});

test('已知数据库类型返回专属图标 kind', () => {
  assert.equal(resolveDatabaseTypeIconKind('redis'), 'redis');
  assert.equal(resolveDatabaseTypeIconKind('elasticsearch'), 'elasticsearch');
  assert.equal(resolveDatabaseTypeIconKind('mysql'), 'mysql');
  assert.equal(resolveDatabaseTypeIconKind('unknown-driver'), 'database');
  assert.equal(resolveDatabaseTypeIconKind(''), 'database');
});

test('数据库类型标签用于图标 tooltip', () => {
  assert.equal(databaseTypeLabel('redis'), 'Redis / KeyDB');
  assert.equal(databaseTypeLabel('elasticsearch'), 'Elasticsearch / OpenSearch');
  assert.equal(databaseTypeLabel('custom'), 'CUSTOM');
  assert.equal(databaseTypeLabel(''), 'Database');
});
