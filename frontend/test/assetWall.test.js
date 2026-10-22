import assert from 'node:assert/strict';
import test from 'node:test';
import {
  ASSET_STATE_CONNECTED,
  ASSET_STATE_EXCLUDED,
  ASSET_STATE_OFFLINE,
  ASSET_STATE_ONLINE,
  ASSET_STATE_UNKNOWN,
  buildAssetCards,
  groupCards,
  stateLabel,
  summarizeCards
} from '../src/lib/assetWall.js';

const assets = [
  { id: 'ssh-1', name: '生产服务器', host: '10.0.0.1', port: 22, username: 'root', group: '生产', type: 'ssh', metadata: {} },
  { id: 'ssh-2', name: '测试服务器', host: '10.0.0.2', port: 22, username: 'ubuntu', group: '测试', type: 'ssh', metadata: {} },
  { id: 'db-1', name: '订单库', host: '10.0.0.3', port: 3306, username: 'app', group: '生产', type: 'database', metadata: { db_type: 'mysql' }, dbConnected: true, dbSessionId: 'sess-db' },
  { id: 'ssh-3', name: '未探资产', host: '10.0.0.4', port: 22, username: 'root', group: '测试', type: 'ssh', metadata: {} }
];

test('已连接会话优先于探活结果，并带出真实指标', () => {
  const sessions = [
    { sessionId: 'sess-1', type: 'ssh', connected: true, connection: { id: 'ssh-1' } },
    { sessionId: 'sess-db', type: 'database', connected: true, connection: { id: 'db-1' } }
  ];
  const probes = new Map([
    ['ssh-1', { reachable: false, latency_ms: 0, error: '连接失败' }],
    ['db-1', { reachable: true, latency_ms: 3 }]
  ]);
  const metrics = new Map([
    ['sess-1', { cpu: { overall: 42.26 }, memory: { used_percent: 63.74 } }]
  ]);
  const cards = buildAssetCards(assets, sessions, probes, metrics);
  assert.equal(cards[0].state, ASSET_STATE_CONNECTED);
  assert.equal(cards[0].cpuPercent, 42.3);
  assert.equal(cards[0].memPercent, 63.7);
  assert.equal(cards[2].state, ASSET_STATE_CONNECTED);
  assert.equal(cards[2].sessionId, 'sess-db');
  assert.equal(cards[2].dbType, 'mysql');
});

test('未连接资产按探活结果分在线/离线/未知', () => {
  const probes = new Map([
    ['ssh-2', { reachable: true, latency_ms: 12 }],
    ['db-1', { reachable: false, latency_ms: 0, error: '连接失败: timeout' }]
  ]);
  const cards = buildAssetCards(assets, [], probes, new Map());
  assert.equal(cards[1].state, ASSET_STATE_ONLINE);
  assert.equal(cards[1].latencyMs, 12);
  const dbAssets = [{ ...assets[2], dbConnected: false, dbSessionId: null }];
  const dbCards = buildAssetCards(dbAssets, [], probes, new Map());
  assert.equal(dbCards[0].state, ASSET_STATE_OFFLINE);
  assert.match(dbCards[0].probeError, /timeout/);
  assert.equal(cards[3].state, ASSET_STATE_UNKNOWN);
  assert.equal(cards[3].latencyMs, null);
});

test('汇总口径：可用 = 已连接 + 在线', () => {
  const cards = [
    { state: ASSET_STATE_CONNECTED },
    { state: ASSET_STATE_ONLINE },
    { state: ASSET_STATE_ONLINE },
    { state: ASSET_STATE_OFFLINE },
    { state: ASSET_STATE_UNKNOWN }
  ];
  assert.deepEqual(summarizeCards(cards), {
    total: 5,
    connected: 1,
    online: 2,
    offline: 1,
    unknown: 1,
    excluded: 0,
    available: 3
  });
});

test('分组保持首次出现顺序，组内保持资产顺序', () => {
  const cards = buildAssetCards(assets, [], new Map(), new Map());
  const groups = groupCards(cards);
  assert.deepEqual(groups.map((g) => g.name), ['生产', '测试']);
  assert.deepEqual(groups[0].cards.map((c) => c.id), ['ssh-1', 'db-1']);
  assert.deepEqual(groups[1].cards.map((c) => c.id), ['ssh-2', 'ssh-3']);
});

test('关闭探活的资产单列状态，但已连接时仍显示已连接', () => {
  const excluded = [
    { id: 'x1', name: '敏感机', host: '10.0.9.1', port: 22, type: 'ssh', metadata: { status_wall_excluded: 'true' } }
  ];
  const probes = new Map([['x1', { reachable: true, latency_ms: 1 }]]);
  const cards = buildAssetCards(excluded, [], probes, new Map());
  assert.equal(cards[0].state, ASSET_STATE_EXCLUDED);
  assert.equal(cards[0].probeExcluded, true);
  assert.equal(stateLabel(cards[0].state), '已关闭探活');

  const sessions = [{ sessionId: 's-x1', type: 'ssh', connected: true, connection: { id: 'x1' } }];
  const connected = buildAssetCards(excluded, sessions, probes, new Map());
  assert.equal(connected[0].state, ASSET_STATE_CONNECTED);
});

test('状态文案齐全', () => {
  assert.equal(stateLabel(ASSET_STATE_CONNECTED), '已连接');
  assert.equal(stateLabel(ASSET_STATE_ONLINE), '在线');
  assert.equal(stateLabel(ASSET_STATE_OFFLINE), '离线');
  assert.equal(stateLabel('whatever'), '未探测');
});
