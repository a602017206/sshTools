import assert from 'node:assert/strict';
import test from 'node:test';

import {
  defaultAssetTypeForDomain,
  filterAssetsByDomain,
  groupDatabaseTypesByDomain,
  resolveAssetDomain
} from '../src/lib/assetDomain.js';

test('按 db_type 推导资产域', () => {
  assert.equal(resolveAssetDomain({ type: 'ssh' }), 'ssh');
  assert.equal(resolveAssetDomain({ type: 'database', metadata: { db_type: 'redis' } }), 'cache');
  assert.equal(resolveAssetDomain({ type: 'database', metadata: { db_type: 'elasticsearch' } }), 'search');
  assert.equal(resolveAssetDomain({ type: 'database', metadata: { db_type: 'kafka' } }), 'mq');
  assert.equal(resolveAssetDomain({ type: 'database', metadata: { db_type: 'rocketmq' } }), 'mq');
  assert.equal(resolveAssetDomain({ type: 'database', metadata: { db_type: 'rabbitmq' } }), 'mq');
  assert.equal(resolveAssetDomain({ type: 'database', metadata: { db_type: 'mysql' } }), 'database');
  assert.equal(resolveAssetDomain({ type: 'database', metadata: { db_type: 'mongodb' } }), 'database');
});

test('显式 metadata.domain 优先于推导', () => {
  assert.equal(
    resolveAssetDomain({ type: 'database', metadata: { db_type: 'redis', domain: 'database' } }),
    'database'
  );
});

test('filterAssetsByDomain 按域过滤资产树', () => {
  const assets = [
    { id: '1', type: 'ssh', name: 'host' },
    { id: '2', type: 'database', metadata: { db_type: 'redis' }, name: 'redis' },
    { id: '3', type: 'database', metadata: { db_type: 'elasticsearch' }, name: 'es' },
    { id: '4', type: 'database', metadata: { db_type: 'kafka' }, name: 'kafka' },
    { id: '5', type: 'database', metadata: { db_type: 'mysql' }, name: 'mysql' }
  ];
  assert.equal(filterAssetsByDomain(assets, 'all').length, 5);
  assert.deepEqual(filterAssetsByDomain(assets, 'cache').map((item) => item.name), ['redis']);
  assert.deepEqual(filterAssetsByDomain(assets, 'search').map((item) => item.name), ['es']);
  assert.deepEqual(filterAssetsByDomain(assets, 'mq').map((item) => item.name), ['kafka']);
  assert.deepEqual(filterAssetsByDomain(assets, 'database').map((item) => item.name), ['mysql']);
  assert.deepEqual(filterAssetsByDomain(assets, 'ssh').map((item) => item.name), ['host']);
});

test('defaultAssetTypeForDomain 为新建连接提供默认类型', () => {
  assert.deepEqual(defaultAssetTypeForDomain('cache'), { assetType: 'database', dbType: 'redis' });
  assert.deepEqual(defaultAssetTypeForDomain('search'), { assetType: 'database', dbType: 'elasticsearch' });
  assert.deepEqual(defaultAssetTypeForDomain('ssh'), { assetType: 'ssh', dbType: '' });
});

test('groupDatabaseTypesByDomain 按域分组类型列表', () => {
  const groups = groupDatabaseTypesByDomain([
    { value: 'mysql', label: 'MySQL' },
    { value: 'redis', label: 'Redis' },
    { value: 'elasticsearch', label: 'ES' },
    { value: 'kafka', label: 'Kafka' },
    { value: 'mongodb', label: 'Mongo' }
  ]);
  assert.deepEqual(groups.map((group) => group.id), ['database', 'cache', 'search', 'mq']);
  assert.deepEqual(groups.find((group) => group.id === 'cache').types.map((item) => item.value), ['redis']);
  assert.deepEqual(groups.find((group) => group.id === 'database').types.map((item) => item.value), ['mysql', 'mongodb']);
});
