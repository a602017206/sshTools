import assert from 'node:assert/strict';
import test from 'node:test';

import {
  buildMySQLObjectQuery,
  buildPostgreSQLObjectQuery,
  buildPostgreSQLSchemaQuery,
  databaseObjectCategories,
  databaseSidebarCategories
} from '../src/lib/databaseObjectTree.js';

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

test('MySQL 对象树使用 information_schema 并安全限定数据库', () => {
  const query = buildMySQLObjectQuery("app's", 'views');

  assert.match(query, /information_schema\.views/);
  assert.match(query, /TABLE_SCHEMA = 'app''s'/);
  assert.deepEqual(databaseObjectCategories('mysql').map(item => item.id), [
    'tables', 'views', 'procedures', 'functions', 'events'
  ]);
});

test('左侧对象树分类携带标准 JDBC 元数据类型或例程类型', () => {
  assert.deepEqual(databaseSidebarCategories('kingbase'), [
    { id: 'tables', label: '表', icon: '▦', types: ['TABLE'] },
    { id: 'views', label: '视图', icon: '◉', types: ['VIEW'] },
    { id: 'procedures', label: '存储过程', icon: '▤', functions: false },
    { id: 'functions', label: '函数', icon: 'ƒ', functions: true }
  ]);
});
