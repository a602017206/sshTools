import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';
import { buildCopyTableStatements, buildDropTableSQL } from '../src/lib/tableObjectMutations.js';

test('MySQL 表删除与复制语句使用数据库限定标识符', () => {
  assert.equal(buildDropTableSQL({ databaseType: 'mysql', databaseName: 'inventory', tableName: 'stock' }), 'DROP TABLE `inventory`.`stock`;');
  assert.deepEqual(
    buildCopyTableStatements({ databaseType: 'mysql', databaseName: 'inventory', sourceTable: 'stock', targetTable: 'stock_copy', includeData: true }),
    ['CREATE TABLE `inventory`.`stock_copy` LIKE `inventory`.`stock`;', 'INSERT INTO `inventory`.`stock_copy` SELECT * FROM `inventory`.`stock`;']
  );
});

test('PostgreSQL 表复制保留结构并可选择复制数据', () => {
  assert.equal(buildDropTableSQL({ databaseType: 'postgresql', schemaName: 'public', tableName: 'stock' }), 'DROP TABLE "public"."stock";');
  assert.deepEqual(
    buildCopyTableStatements({ databaseType: 'postgresql', schemaName: 'public', sourceTable: 'stock', targetTable: 'stock_copy', includeData: false }),
    ['CREATE TABLE "public"."stock_copy" (LIKE "public"."stock" INCLUDING ALL);']
  );
});

test('人大金仓表复制与删除复用 PostgreSQL 兼容语法', () => {
  assert.equal(buildDropTableSQL({ databaseType: 'kingbase', schemaName: 'public', tableName: 'stock' }), 'DROP TABLE "public"."stock";');
  assert.deepEqual(
    buildCopyTableStatements({ databaseType: 'kingbase', schemaName: 'public', sourceTable: 'stock', targetTable: 'stock_copy', includeData: true }),
    ['CREATE TABLE "public"."stock_copy" (LIKE "public"."stock" INCLUDING ALL);', 'INSERT INTO "public"."stock_copy" SELECT * FROM "public"."stock";']
  );
});

test('对象页为表行提供右键打开、设计、复制和删除操作', async () => {
  const source = await readFile(new URL('../src/components/SelectedDatabaseObjects.svelte', import.meta.url), 'utf8');

  assert.match(source, /on:contextmenu/);
  assert.match(source, /打开表/);
  assert.match(source, /设计表/);
  assert.match(source, /复制表.*结构和数据/);
  assert.match(source, /删除表/);
});
