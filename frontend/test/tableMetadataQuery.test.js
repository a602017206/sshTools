import assert from 'node:assert/strict';
import test from 'node:test';

import { buildTableMetadataQuery } from '../src/lib/tableMetadataQuery.js';

test('人大金仓使用 PostgreSQL 兼容的表元数据查询', () => {
  const query = buildTableMetadataQuery('kingbase', 'application');

  assert.match(query, /FROM pg_catalog\.pg_class/);
  assert.doesNotMatch(query, /TABLE_ROWS/);
});

test('MySQL 保持 information_schema 表元数据查询', () => {
  const query = buildTableMetadataQuery('mysql', 'application');

  assert.match(query, /FROM information_schema\.tables/);
  assert.match(query, /TABLE_SCHEMA = 'application'/);
});
