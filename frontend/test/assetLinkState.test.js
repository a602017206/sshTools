import assert from 'node:assert/strict';
import test from 'node:test';

import { assetLinkStateLabel, getAssetLinkState } from '../src/lib/assetLinkState.js';

const sshAsset = { id: 'ssh-1', type: 'ssh', status: 'online' };
const dbAsset = { id: 'db-1', type: 'database', status: 'online', dbConnected: false };

test('未连接时 SSH 资产显示 idle，不受 status=online 影响', () => {
  assert.equal(getAssetLinkState(sshAsset, []), 'idle');
});

test('存在已连接会话时显示 online', () => {
  const sessions = [{ connection: { id: 'ssh-1' }, connected: true }];
  assert.equal(getAssetLinkState(sshAsset, sessions), 'online');
});

test('连接进行中显示 connecting', () => {
  const sessions = [{ connection: { id: 'ssh-1' }, connected: false }];
  assert.equal(getAssetLinkState(sshAsset, sessions), 'connecting');
});

test('数据库资产根据 dbConnected 或会话判断在线状态', () => {
  assert.equal(getAssetLinkState(dbAsset, []), 'idle');
  assert.equal(getAssetLinkState({ ...dbAsset, dbConnected: true }, []), 'online');
  assert.equal(
    getAssetLinkState(dbAsset, [{ connection: { id: 'db-1' }, connected: true, panelType: 'database-list' }]),
    'online'
  );
});

test('连接失败状态优先于 idle', () => {
  assert.equal(getAssetLinkState({ ...sshAsset, status: 'error' }, []), 'error');
});

test('状态标签提供中文 tooltip', () => {
  assert.equal(assetLinkStateLabel('online'), '已连接');
  assert.equal(assetLinkStateLabel('idle'), '未连接');
});
