import assert from 'node:assert/strict';
import test from 'node:test';

import { buildPostgreSQLObjectQuery, buildPostgreSQLSchemaQuery, databaseObjectCategories } from '../src/lib/databaseObjectTree.js';

test('PostgreSQL 兼容数据库通过系统目录读取 schema', () => {
  const query = buildPostgreSQLSchemaQuery();

  assert.match(query, /pg_catalog\.pg_namespace/);
  assert.match(query, /nspname NOT LIKE 'pg_toast%'/);
});

test('对象查询按 schema 安全过滤并按对象类型读取', () => {
  const query = buildPostgreSQLObjectQuery("app's", 'views');

  assert.match(query, /pg_catalog\.pg_views/);
  assert.match(query, /schemaname = 'app''s'/);
});

test('PostgreSQL 对象树包含 Navicat 式对象分类', () => {
  assert.deepEqual(databaseObjectCategories('kingbase').map(item => item.id), [
    'tables', 'views', 'materialized_views', 'procedures', 'functions', 'extensions'
  ]);
});
