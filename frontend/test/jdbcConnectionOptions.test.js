import assert from 'node:assert/strict';
import test from 'node:test';
import { buildJDBCConnectionOptions } from '../src/lib/jdbcConnectionOptions.js';

test('Oracle 服务名和 SID 以受控连接模式传递', () => {
  assert.deepEqual(buildJDBCConnectionOptions('oracle', 'pdb1', 'service', ''), {
    database: 'pdb1',
    properties: { oracleConnectionMode: 'service' }
  });
  assert.deepEqual(buildJDBCConnectionOptions('oracle', 'ORCL', 'sid', ''), {
    database: 'ORCL',
    properties: { oracleConnectionMode: 'sid' }
  });
});

test('SQL Server 实例名作为 JDBC 属性传递', () => {
  assert.deepEqual(buildJDBCConnectionOptions('sqlserver', 'master', '', 'SQLEXPRESS'), {
    database: 'master',
    properties: { instanceName: 'SQLEXPRESS' }
  });
});
