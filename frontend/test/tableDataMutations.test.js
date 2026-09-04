import assert from 'node:assert/strict';
import test from 'node:test';
import {
  buildBatchUpdateStatements,
  buildDeleteSQL,
  buildDeleteStatements,
  buildInsertSQL,
  buildInsertStatements,
  buildUpdateSQL
} from '../src/lib/tableDataMutations.js';

test('生成带主键条件的更新和删除语句', () => {
  const input = { databaseType: 'mysql', table: '`app`.`users`', columns: ['id', 'name'], row: [7, "O'Reilly"], primaryKeys: ['id'], changes: { name: 'Ada' } };
  assert.equal(buildUpdateSQL(input), "UPDATE `app`.`users` SET `name` = 'Ada' WHERE `id` = 7;");
  assert.equal(buildDeleteSQL(input), "DELETE FROM `app`.`users` WHERE `id` = 7;");
});

test('主键元数据与查询列名大小写不一致时仍可删除记录', () => {
  const input = { databaseType: 'mysql', table: '`app`.`users`', columns: ['id', 'name'], row: [7, 'Ada'], primaryKeys: ['ID'] };
  assert.equal(buildDeleteSQL(input), 'DELETE FROM `app`.`users` WHERE `id` = 7;');
});

test('生成 INSERT 语句并正确处理 NULL 与引号', () => {
  assert.equal(buildInsertSQL({ databaseType: 'mysql', table: '`app`.`users`', columns: ['id', 'name', 'note'], row: [7, "O'Reilly", null] }), "INSERT INTO `app`.`users` (`id`, `name`, `note`) VALUES (7, 'O''Reilly', NULL);");
});

test('PostgreSQL 使用双引号生成更新和删除语句', () => {
  const input = { databaseType: 'postgresql', table: '"public"."users"', columns: ['id', 'name'], row: [7, 'Ada'], primaryKeys: ['id'], changes: { name: 'Grace' } };
  assert.equal(buildUpdateSQL(input), 'UPDATE "public"."users" SET "name" = \'Grace\' WHERE "id" = 7;');
  assert.equal(buildDeleteSQL(input), 'DELETE FROM "public"."users" WHERE "id" = 7;');
});

test('无主键时 MySQL 以整行原始值构造删除条件', () => {
  const input = { databaseType: 'mysql', table: '`app`.`events`', columns: ['date', 'code'], row: ['2025-06-01', '000001'], primaryKeys: [] };
  assert.equal(buildDeleteSQL(input), "DELETE FROM `app`.`events` WHERE `date` <=> '2025-06-01' AND `code` <=> '000001';");
});

test('无主键时 PostgreSQL 以整行原始值构造更新条件', () => {
  const input = { databaseType: 'postgresql', table: '"public"."events"', columns: ['date', 'code'], row: ['2025-06-01', null], primaryKeys: [], changes: { code: '000001' } };
  assert.equal(buildUpdateSQL(input), "UPDATE \"public\".\"events\" SET \"code\" = '000001' WHERE \"date\" IS NOT DISTINCT FROM '2025-06-01' AND \"code\" IS NOT DISTINCT FROM NULL;");
});

test('buildDeleteStatements 为多行生成删除语句', () => {
  const base = { databaseType: 'mysql', table: '`app`.`users`', columns: ['id', 'name'], primaryKeys: ['id'] };
  assert.deepEqual(
    buildDeleteStatements(base, [[7, 'Ada'], [8, 'Bob']]),
    [
      'DELETE FROM `app`.`users` WHERE `id` = 7;',
      'DELETE FROM `app`.`users` WHERE `id` = 8;'
    ]
  );
});

test('buildBatchUpdateStatements 为选中行批量更新同一字段', () => {
  const base = { databaseType: 'mysql', table: '`app`.`users`', columns: ['id', 'status'], primaryKeys: ['id'] };
  assert.deepEqual(
    buildBatchUpdateStatements(base, [[7, 'old'], [8, 'old']], 'status', 'active'),
    [
      "UPDATE `app`.`users` SET `status` = 'active' WHERE `id` = 7;",
      "UPDATE `app`.`users` SET `status` = 'active' WHERE `id` = 8;"
    ]
  );
});

test('Oracle 主键含 NULL 时使用 IS NULL 构造删除条件', () => {
  const input = {
    databaseType: 'oracle',
    table: '"PEMS"."DW_CP_CONTROL_CHECK"',
    columns: ['PID', 'SPECIALTY_ID'],
    row: ['abc', null],
    primaryKeys: ['PID', 'SPECIALTY_ID']
  };
  assert.equal(
    buildDeleteSQL(input),
    'DELETE FROM "PEMS"."DW_CP_CONTROL_CHECK" WHERE "PID" = \'abc\' AND "SPECIALTY_ID" IS NULL;'
  );
});

test('buildInsertStatements 为多行生成 INSERT', () => {
  const base = { databaseType: 'mysql', table: '`app`.`users`', columns: ['id', 'name'], primaryKeys: ['id'] };
  assert.deepEqual(
    buildInsertStatements(base, [[1, 'Ada'], [2, 'Bob']]),
    [
      "INSERT INTO `app`.`users` (`id`, `name`) VALUES (1, 'Ada');",
      "INSERT INTO `app`.`users` (`id`, `name`) VALUES (2, 'Bob');"
    ]
  );
});
