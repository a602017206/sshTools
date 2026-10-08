import assert from 'node:assert/strict';
import test from 'node:test';
import { columnTypeOptions, typeAcceptsLength } from '../src/lib/columnTypeOptions.js';
import { buildCreateTableSQL } from '../src/lib/tableDefinitionSQL.js';

test('MySQL 字段类型包含 LONGTEXT、DATETIME 等常用类型', () => {
  const types = columnTypeOptions('mysql');
  for (const type of ['LONGTEXT', 'MEDIUMTEXT', 'DATETIME', 'TINYINT', 'JSON', 'LONGBLOB']) {
    assert.ok(types.includes(type), `missing ${type}`);
  }
  assert.ok(!types.includes('VARCHAR2'));
});

test('不同数据库返回各自方言的类型', () => {
  assert.ok(columnTypeOptions('kingbase').includes('JSONB'));
  assert.ok(columnTypeOptions('oracle').includes('VARCHAR2'));
  assert.ok(!columnTypeOptions('oracle').includes('LONGTEXT'));
});

test('字段现有类型不在列表中时仍保留，避免下拉框空白', () => {
  assert.equal(columnTypeOptions('mysql', 'enum')[0], 'ENUM');
  assert.equal(columnTypeOptions('mysql', 'LONGTEXT').filter(type => type === 'LONGTEXT').length, 1);
});

test('仅需要长度的类型拼接长度', () => {
  assert.equal(typeAcceptsLength('VARCHAR'), true);
  assert.equal(typeAcceptsLength('TINYINT'), true);
  assert.equal(typeAcceptsLength('LONGTEXT'), false);
  assert.equal(typeAcceptsLength('DATETIME'), false);

  const sql = buildCreateTableSQL({
    databaseType: 'mysql',
    databaseName: 'demo',
    tableName: 't',
    fields: [
      { name: 'flag', type: 'TINYINT', length: '1', nullable: true },
      { name: 'body', type: 'LONGTEXT', length: '255', nullable: true },
      { name: 'created_at', type: 'DATETIME', length: '', nullable: false }
    ]
  });
  assert.match(sql, /`flag` TINYINT\(1\)/);
  assert.match(sql, /`body` LONGTEXT,/);
  assert.match(sql, /`created_at` DATETIME NOT NULL/);
});
