import assert from 'node:assert/strict';
import test from 'node:test';

import {
  createBlankConnectionFormData,
  createClonedConnectionFormData,
} from '../src/lib/connectionFormData.js';

test('新建连接表单为空白并使用 SSH 默认端口 22', () => {
  assert.deepEqual(createBlankConnectionFormData(), {
    id: '',
    name: '',
    host: '',
    port: '22',
    username: '',
    password: '',
    savePassword: false,
    keyPath: '',
    passphrase: '',
    group: '',
    encoding: 'utf-8',
    dbType: 'mysql',
    driverProfileID: '',
    database: '',
    oracleConnectionMode: 'service',
    sqlServerInstanceName: '',
  });
});

test('克隆连接保留配置与已保存密码，但不复用连接 ID 或私钥口令', () => {
  const source = {
    id: 'source-id',
    name: '生产 SSH',
    host: 'prod.example.com',
    port: 2202,
    username: 'deploy',
    password: 'source-password',
    savePassword: true,
    keyPath: '~/.ssh/production',
    passphrase: 'source-passphrase',
    group: '生产环境',
    dbType: 'oracle',
    driverProfileID: 'oracle-21',
    database: 'ORCL',
    oracleConnectionMode: 'sid',
    sqlServerInstanceName: 'instance-a',
    type: 'ssh',
    auth_type: 'key',
    tags: ['生产环境'],
    metadata: { db_type: 'oracle', region: 'cn-east-1', tunnel: { host: 'jump.example.com' } },
    settings: { terminal: { theme: 'dark' } },
  };

  const clone = createClonedConnectionFormData(
    source,
    new Date(2026, 7, 13, 9, 5, 7),
    { password: 'source-password', savePassword: true }
  );

  assert.deepEqual(clone, {
    id: '',
    name: '生产 SSH copy 20260813090507',
    host: 'prod.example.com',
    port: '2202',
    username: 'deploy',
    password: 'source-password',
    savePassword: true,
    keyPath: '~/.ssh/production',
    passphrase: '',
    group: '生产环境',
    encoding: 'utf-8',
    dbType: 'oracle',
    driverProfileID: 'oracle-21',
    database: 'ORCL',
    oracleConnectionMode: 'sid',
    sqlServerInstanceName: 'instance-a',
    type: 'ssh',
    authType: 'key',
    tags: ['生产环境'],
    metadata: { db_type: 'oracle', region: 'cn-east-1', tunnel: { host: 'jump.example.com' } },
    settings: { terminal: { theme: 'dark' } },
  });
  assert.notEqual(clone.tags, source.tags);
  assert.notEqual(clone.metadata, source.metadata);
  assert.notEqual(clone.settings, source.settings);
  clone.metadata.tunnel.host = 'other-jump.example.com';
  clone.settings.terminal.theme = 'light';
  assert.equal(source.metadata.tunnel.host, 'jump.example.com');
  assert.equal(source.settings.terminal.theme, 'dark');
});
