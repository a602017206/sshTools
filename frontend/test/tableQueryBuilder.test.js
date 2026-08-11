import assert from 'node:assert/strict';
import test from 'node:test';

import { buildQualifiedTableName, buildTableBrowseSQL } from '../src/lib/tableQueryBuilder.js';

const fromSQL = '`app`.`users`';

test('表数据筛选支持 AND/OR、包含和比较操作', () => {
  const sql = buildTableBrowseSQL({
    fromSQL,
    databaseType: 'mysql',
    filters: [
      { field: 'name', operation: 'contains', value: "O'Reilly" },
      { connector: 'OR', field: 'age', operation: 'greater_or_equal', value: '18' }
    ],
    sorters: [{ field: 'created_at', direction: 'DESC' }]
  });

  assert.equal(sql, "SELECT * FROM `app`.`users` WHERE `name` LIKE '%O''Reilly%' OR `age` >= '18' ORDER BY `created_at` DESC LIMIT 100;");
});

test('表数据筛选支持 NULL、列表和多字段排序', () => {
  const sql = buildTableBrowseSQL({
    fromSQL,
    databaseType: 'mysql',
    filters: [
      { field: 'deleted_at', operation: 'is_null' },
      { connector: 'AND', field: 'role', operation: 'not_in', value: 'admin, guest' },
      { connector: 'AND', field: 'email', operation: 'not_contains', value: 'example.com' }
    ],
    sorters: [
      { field: 'role', direction: 'ASC' },
      { field: 'id', direction: 'DESC' }
    ]
  });

  assert.equal(sql, "SELECT * FROM `app`.`users` WHERE `deleted_at` IS NULL AND `role` NOT IN ('admin', 'guest') AND `email` NOT LIKE '%example.com%' ORDER BY `role` ASC, `id` DESC LIMIT 100;");
});

test('无有效筛选时生成默认表数据查询', () => {
  assert.equal(buildTableBrowseSQL({ fromSQL, databaseType: 'mysql', filters: [], sorters: [], limit: 50 }), 'SELECT * FROM `app`.`users` LIMIT 50;');
});

test('Oracle 使用 FETCH FIRST 而不是 LIMIT', () => {
  assert.equal(
    buildTableBrowseSQL({ fromSQL: '"HR"."USERS"', databaseType: 'oracle', filters: [], sorters: [], limit: 100 }),
    'SELECT * FROM "HR"."USERS" FETCH FIRST 100 ROWS ONLY;'
  );
});

test('Oracle 分页使用 OFFSET/FETCH NEXT', () => {
  assert.equal(
    buildTableBrowseSQL({ fromSQL: '"HR"."USERS"', databaseType: 'oracle', filters: [], sorters: [], limit: 50, offset: 100 }),
    'SELECT * FROM "HR"."USERS" OFFSET 100 ROWS FETCH NEXT 50 ROWS ONLY;'
  );
});

test('SQL Server 分页使用 OFFSET/FETCH，并补默认 ORDER BY', () => {
  assert.equal(
    buildTableBrowseSQL({ fromSQL: '[dbo].[users]', databaseType: 'sqlserver', filters: [], sorters: [], limit: 25, offset: 50 }),
    'SELECT * FROM [dbo].[users] ORDER BY (SELECT NULL) OFFSET 50 ROWS FETCH NEXT 25 ROWS ONLY;'
  );
});

test('Oracle 限定表名使用 schema 而不是服务名', () => {
  assert.equal(
    buildQualifiedTableName({ databaseType: 'oracle', databaseName: 'pdb', schemaName: 'HR', tableName: 'USERS' }),
    '"HR"."USERS"'
  );
});

test('MySQL 限定表名仍使用 database.table', () => {
  assert.equal(
    buildQualifiedTableName({ databaseType: 'mysql', databaseName: 'app', schemaName: 'ignored', tableName: 'users' }),
    '`app`.`users`'
  );
});