import assert from 'node:assert/strict';
import test from 'node:test';
import {
  buildBatchPayload,
  decorateSessionTabs,
  isBatchableSession,
  listBatchTargets,
  sessionAccentColor,
  sessionTabSubtitle,
  sessionTabTitle
} from '../src/lib/sessionTabs.js';

const ssh = (id, assetId, extra = {}) => ({
  sessionId: id,
  connected: true,
  connection: { id: assetId, name: '192.168.121.110', host: '192.168.121.110', port: 22, username: 'root' },
  ...extra
});

test('同资产多页签自动编号，单页签不带序号，改名后不再加序号', () => {
  const sessions = [ssh('s1', 'a1'), ssh('s2', 'a1'), ssh('s3', 'a2')];
  const ordinals = decorateSessionTabs(sessions);
  assert.equal(sessionTabTitle(sessions[0], ordinals.get('s1'), 2), '192.168.121.110 #1');
  assert.equal(sessionTabTitle(sessions[1], ordinals.get('s2'), 2), '192.168.121.110 #2');
  assert.equal(sessionTabTitle(sessions[2], ordinals.get('s3'), 1), '192.168.121.110');
  const renamed = ssh('s4', 'a1', { tabName: '跳板窗口' });
  assert.equal(sessionTabTitle(renamed, 2, 2), '跳板窗口');
});

test('副标题显示用户@主机:端口，断开有标注', () => {
  assert.equal(sessionTabSubtitle(ssh('s1', 'a1')), 'root@192.168.121.110:22');
  assert.equal(sessionTabSubtitle(ssh('s1', 'a1', { connected: false })), 'root@192.168.121.110:22 · 已断开');
  assert.equal(sessionTabSubtitle({ sessionId: 'l1', type: 'local', connected: true, connection: { name: 'Local' } }), '本地终端');
  assert.equal(sessionTabSubtitle({ sessionId: 'd1', type: 'database', connected: true, connection: { id: 'x' } }), '数据库');
});

test('配色按 sessionId 稳定且落在调色板内', () => {
  assert.equal(sessionAccentColor('session_abc'), sessionAccentColor('session_abc'));
  assert.match(sessionAccentColor('session_abc'), /^#[0-9a-f]{6}$/i);
});

test('批量目标只含已连接的远程 SSH 会话', () => {
  const sessions = [
    ssh('s1', 'a1'),
    ssh('s2', 'a1', { connected: false }),
    { sessionId: 'd1', type: 'database', connected: true, connection: { id: 'db', host: 'h' } },
    { sessionId: 'l1', type: 'local', connected: true, connection: { host: 'localhost' } }
  ];
  assert.equal(isBatchableSession(sessions[0]), true);
  const targets = listBatchTargets(sessions);
  assert.deepEqual(targets.map((t) => t.sessionId), ['s1']);
  assert.equal(targets[0].subtitle, 'root@192.168.121.110:22');
});

test('批量命令规整为回车分隔，尾部回车可选', () => {
  assert.equal(buildBatchPayload('uptime\nfree -m\n', true), 'uptime\rfree -m\r');
  assert.equal(buildBatchPayload('uptime', false), 'uptime');
  assert.equal(buildBatchPayload('  \n', true), '');
});
