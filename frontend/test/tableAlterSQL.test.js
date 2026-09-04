import assert from 'node:assert/strict';
import test from 'node:test';
import { buildAlterTableStatements } from '../src/lib/tableAlterSQL.js';

test('MySQL 表结构差异生成字段修改、新增与主键变更语句', () => {
  const originalFields = [
    { _originalName: 'id', name: 'id', type: 'INT', length: '11', nullable: false, primary: true, defaultValue: '', comment: '' },
    { _originalName: 'name', name: 'name', type: 'VARCHAR', length: '50', nullable: true, primary: false, defaultValue: '', comment: '' }
  ];
  const fields = [
    { _originalName: 'id', name: 'id', type: 'BIGINT', length: '20', nullable: false, primary: true, defaultValue: '', comment: '' },
    { _originalName: 'name', name: 'display_name', type: 'VARCHAR', length: '100', nullable: false, primary: false, defaultValue: "'unknown'", comment: '显示名' },
    { _originalName: '', name: 'created_at', type: 'TIMESTAMP', length: '', nullable: false, primary: false, defaultValue: 'CURRENT_TIMESTAMP', comment: '' }
  ];

  assert.deepEqual(buildAlterTableStatements({ databaseType: 'mysql', databaseName: 'app', tableName: 'users', originalFields, fields }), [
    'ALTER TABLE `app`.`users` MODIFY COLUMN `id` BIGINT(20) NOT NULL;',
    "ALTER TABLE `app`.`users` CHANGE COLUMN `name` `display_name` VARCHAR(100) NOT NULL DEFAULT 'unknown' COMMENT '显示名';",
    'ALTER TABLE `app`.`users` ADD COLUMN `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;'
  ]);
});

test('PostgreSQL 表结构差异生成重命名、默认值与删除字段语句', () => {
  const originalFields = [
    { _originalName: 'id', name: 'id', type: 'BIGINT', length: '', nullable: false, primary: true, defaultValue: '', comment: '' },
    { _originalName: 'legacy', name: 'legacy', type: 'TEXT', length: '', nullable: true, primary: false, defaultValue: '', comment: '' }
  ];
  const fields = [
    { _originalName: 'id', name: 'user_id', type: 'BIGINT', length: '', nullable: false, primary: true, defaultValue: '', comment: '' },
    { _originalName: '', name: 'enabled', type: 'BOOLEAN', length: '', nullable: false, primary: false, defaultValue: 'true', comment: '启用状态' }
  ];

  assert.deepEqual(buildAlterTableStatements({ databaseType: 'postgresql', schemaName: 'public', tableName: 'users', originalFields, fields }), [
    'ALTER TABLE "public"."users" RENAME COLUMN "id" TO "user_id";',
    'ALTER TABLE "public"."users" ADD COLUMN "enabled" BOOLEAN NOT NULL DEFAULT true;',
    'COMMENT ON COLUMN "public"."users"."enabled" IS \'启用状态\';',
    'ALTER TABLE "public"."users" DROP COLUMN "legacy";'
  ]);
});

test('人大金仓表结构修改复用 PostgreSQL 兼容语法', () => {
  const originalFields = [{ _originalName: 'name', name: 'name', type: 'VARCHAR', length: '64', nullable: true, primary: false, defaultValue: '', comment: '' }];
  const fields = [{ _originalName: 'name', name: 'display_name', type: 'VARCHAR', length: '128', nullable: false, primary: false, defaultValue: "'unknown'", comment: '显示名' }];

  assert.deepEqual(buildAlterTableStatements({ databaseType: 'kingbase', schemaName: 'public', tableName: 'users', originalFields, fields }), [
    'ALTER TABLE "public"."users" RENAME COLUMN "name" TO "display_name";',
    'ALTER TABLE "public"."users" ALTER COLUMN "display_name" TYPE VARCHAR(128);',
    'ALTER TABLE "public"."users" ALTER COLUMN "display_name" SET NOT NULL;',
    "ALTER TABLE \"public\".\"users\" ALTER COLUMN \"display_name\" SET DEFAULT 'unknown';",
    'COMMENT ON COLUMN "public"."users"."display_name" IS \'显示名\';'
  ]);
});

test('Oracle 表结构修改使用 ADD/MODIFY/RENAME COLUMN，并以 schema 限定表名', () => {
  const originalFields = [
    { _originalName: 'NAME', name: 'NAME', type: 'VARCHAR2', length: '50', nullable: true, primary: false, defaultValue: '', comment: '' }
  ];
  const fields = [
    { _originalName: 'NAME', name: 'DISPLAY_NAME', type: 'VARCHAR2', length: '100', nullable: false, primary: false, defaultValue: "'unknown'", comment: '显示名' },
    { _originalName: '', name: 'CREATED_AT', type: 'TIMESTAMP', length: '', nullable: false, primary: false, defaultValue: 'SYSDATE', comment: '' }
  ];

  assert.deepEqual(buildAlterTableStatements({ databaseType: 'oracle', databaseName: 'pdb', schemaName: 'PEMS', tableName: 'DW_CP_CONTROL_TYPE', originalFields, fields }), [
    'ALTER TABLE "PEMS"."DW_CP_CONTROL_TYPE" RENAME COLUMN "NAME" TO "DISPLAY_NAME";',
    'ALTER TABLE "PEMS"."DW_CP_CONTROL_TYPE" MODIFY ("DISPLAY_NAME" VARCHAR2(100) NOT NULL DEFAULT \'unknown\');',
    'COMMENT ON COLUMN "PEMS"."DW_CP_CONTROL_TYPE"."DISPLAY_NAME" IS \'显示名\';',
    'ALTER TABLE "PEMS"."DW_CP_CONTROL_TYPE" ADD ("CREATED_AT" TIMESTAMP NOT NULL DEFAULT SYSDATE);'
  ]);
});
