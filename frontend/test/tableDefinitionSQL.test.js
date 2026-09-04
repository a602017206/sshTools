import assert from 'node:assert/strict';
import test from 'node:test';
import { buildCreateTableSQL } from '../src/lib/tableDefinitionSQL.js';

const fields = [
  { name: 'id', type: 'BIGINT', length: '20', nullable: false, primary: true, defaultValue: '' },
  { name: 'name', type: 'VARCHAR', length: '255', nullable: true, primary: false, defaultValue: '' }
];

test('MySQL 建表使用数据库前缀和反引号', () => {
  assert.equal(buildCreateTableSQL({ databaseType: 'mysql', databaseName: 'app', schemaName: '', tableName: 'users', fields }), 'CREATE TABLE `app`.`users` (\n  `id` BIGINT(20) NOT NULL,\n  `name` VARCHAR(255),\n  PRIMARY KEY (`id`)\n);');
});

test('PostgreSQL 建表使用 schema 前缀和双引号', () => {
  assert.equal(buildCreateTableSQL({ databaseType: 'postgresql', databaseName: 'app', schemaName: 'public', tableName: 'users', fields }), 'CREATE TABLE "public"."users" (\n  "id" BIGINT NOT NULL,\n  "name" VARCHAR(255),\n  PRIMARY KEY ("id")\n);');
});

test('人大金仓建表复用 PostgreSQL 的 schema 与双引号语法', () => {
  assert.equal(buildCreateTableSQL({ databaseType: 'kingbase', databaseName: 'app', schemaName: 'public', tableName: 'users', fields }), 'CREATE TABLE "public"."users" (\n  "id" BIGINT NOT NULL,\n  "name" VARCHAR(255),\n  PRIMARY KEY ("id")\n);');
});

test('Oracle 建表使用 schema 前缀，并把通用类型映射为 Oracle 类型', () => {
  assert.equal(
    buildCreateTableSQL({ databaseType: 'oracle', databaseName: 'pdb', schemaName: 'PEMS', tableName: 'DW_CP_CONTROL_TYPE', fields }),
    'CREATE TABLE "PEMS"."DW_CP_CONTROL_TYPE" (\n  "id" NUMBER(20) NOT NULL,\n  "name" VARCHAR2(255),\n  PRIMARY KEY ("id")\n);'
  );
});
