import test from 'node:test';
import assert from 'node:assert/strict';
import { databaseSchemaMenuItems, sqlFileProgressFromEvent } from '../src/lib/databaseSchemaMenu.js';

test('schema node menu includes query, sql file, refresh and disconnect', () => {
  const labels = databaseSchemaMenuItems().map((item) => item.label);
  assert.deepEqual(labels, ['新建查询', '运行 SQL 文件…', '刷新', '断开']);
});

test('sql file progress normalizes camelCase and PascalCase event payloads', () => {
  const fromCamel = sqlFileProgressFromEvent({
    sessionId: 'db-1',
    fileName: 'init.sql',
    fileSize: 200,
    bytesRead: 50,
    statements: 4,
    done: false
  });
  assert.equal(fromCamel.sessionId, 'db-1');
  assert.equal(fromCamel.percent, 25);

  const fromGo = sqlFileProgressFromEvent({
    SessionID: 'db-2',
    FileName: 'seed.sql',
    FileSize: 100,
    BytesRead: 100,
    Done: true,
    Error: '失败',
    FailedSQL: 'SELECT 1'
  });
  assert.equal(fromGo.sessionId, 'db-2');
  assert.equal(fromGo.percent, 100);
  assert.equal(fromGo.error, '失败');
  assert.equal(fromGo.failedSql, 'SELECT 1');
});
