import assert from 'node:assert/strict';
import test from 'node:test';

import {
  listDatabaseSessionsToClose,
  resolveDatabaseCloseBinding,
  resolveDatabaseSessionId
} from '../src/lib/databaseSessionClose.js';

test('resolveDatabaseSessionId 优先使用 dbSessionId', () => {
  assert.equal(resolveDatabaseSessionId({ id: 'a1', dbSessionId: 'db-a1' }), 'db-a1');
  assert.equal(resolveDatabaseSessionId({ id: 'a1' }), 'db-a1');
});

test('listDatabaseSessionsToClose 优先返回父会话，由父会话收起子面板', () => {
  const asset = { id: 'redis-1', dbSessionId: 'db-redis-1', type: 'database' };
  const sessions = [
    { sessionId: 'db-redis-1', type: 'database', panelType: 'native-database', dbSessionId: 'db-redis-1' },
    { sessionId: 'child-1', type: 'database', panelType: 'database-query', dbSessionId: 'db-redis-1' },
    { sessionId: 'other', type: 'database', panelType: 'native-database', dbSessionId: 'db-other' }
  ];
  assert.deepEqual(listDatabaseSessionsToClose(sessions, asset), ['db-redis-1']);
});

test('无父会话时关闭残留子面板', () => {
  const asset = { id: 'es-1', dbSessionId: 'db-es-1' };
  const sessions = [
    { sessionId: 'child-a', type: 'database', panelType: 'database-query', dbSessionId: 'db-es-1' }
  ];
  assert.deepEqual(listDatabaseSessionsToClose(sessions, asset), ['child-a']);
});

test('resolveDatabaseCloseBinding 按原生/JDBC 选择关闭 API', () => {
  const bindings = {
    CloseDatabase: async () => {},
    CloseNativeDatabase: async () => {}
  };
  assert.equal(
    resolveDatabaseCloseBinding({ metadata: { db_type: 'redis' } }, bindings),
    bindings.CloseNativeDatabase
  );
  assert.equal(
    resolveDatabaseCloseBinding({ metadata: { db_type: 'mysql' } }, bindings),
    bindings.CloseDatabase
  );
});
