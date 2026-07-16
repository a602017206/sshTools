import assert from 'node:assert/strict';
import test from 'node:test';

import { buildTableMetadataQuery, tableMetadataFromNames } from '../src/lib/tableMetadataQuery.js';

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

test('右侧表信息列表保留 JDBC 表列表 API 返回的表名', () => {
  assert.deepEqual(tableMetadataFromNames(['Sheet1', 'alarm_file']), [
    {
      tableName: 'Sheet1',
      rowCount: null,
      dataLength: null,
      engine: '-',
      createTime: '-',
      updateTime: '-',
      collation: '-',
      comment: '-'
    },
    {
      tableName: 'alarm_file',
      rowCount: null,
      dataLength: null,
      engine: '-',
      createTime: '-',
      updateTime: '-',
      collation: '-',
      comment: '-'
    }
  ]);
});
