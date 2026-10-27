import test from 'node:test';
import assert from 'node:assert/strict';
import { buildDockerExecTabSession } from '../src/lib/dockerExecSession.js';
import { isBatchableSession, sessionTabSubtitle, sessionTabTitle, viewportIsConsole } from '../src/lib/sessionTabs.js';
import { sessionMatchesMode } from '../src/lib/workspaceTabs.js';
import { batchCloseConfirmCopy, batchCloseNeedsConfirm, dockerExecCloseFlags } from '../src/lib/sessionTabClose.js';

const build = (extra = {}) =>
  buildDockerExecTabSession({
    execSessionId: 'exec-1',
    parentDockerSessionId: 'dock-1',
    containerId: 'abcdef1234567890',
    containerName: '/web',
    connection: { id: 'c1', name: 'prod', host: 'h' },
    ...extra,
  });

test('buildDockerExecTabSession 构造 docker-exec 会话', () => {
  const s = build();
  assert.equal(s.sessionId, 'exec-1');
  assert.equal(s.type, 'docker-exec');
  assert.equal(s.connected, true);
  assert.equal(s.parentDockerSessionId, 'dock-1');
  assert.equal(s.containerId, 'abcdef1234567890');
  assert.equal(s.containerName, 'web');
  assert.equal(s.connection.id, 'c1');
});

test('docker-exec 标题与副标题', () => {
  assert.equal(sessionTabTitle(build()), 'web · exec');
  assert.equal(sessionTabTitle(build({ containerName: '' })), 'abcdef123456 · exec');
  assert.equal(sessionTabTitle({ ...build(), tabName: '自定义' }), '自定义');
  assert.equal(sessionTabSubtitle(build()), 'abcdef123456');
  assert.equal(sessionTabSubtitle(build({ containerId: '' })), 'Docker Exec');
});

test('docker-exec 不可批量下发，视口为 console，docker 模式可见', () => {
  assert.equal(isBatchableSession(build()), false);
  assert.equal(viewportIsConsole(build()), true);
  assert.equal(viewportIsConsole({ type: 'docker' }), false);
  assert.equal(viewportIsConsole(null), true);
  assert.equal(sessionMatchesMode(build(), 'docker'), true);
  assert.equal(sessionMatchesMode(build(), 'ssh'), false);
  assert.equal(sessionMatchesMode(build(), 'database'), false);
});

test('docker-exec 关闭路径', () => {
  const s = build();
  assert.deepEqual(dockerExecCloseFlags(s), { isDockerExec: true, closeBinding: 'CloseDockerExec', closeParent: false });
  assert.equal(dockerExecCloseFlags({ type: 'ssh' }).isDockerExec, false);
  assert.equal(batchCloseNeedsConfirm([s], ['exec-1']), true);
  assert.equal(batchCloseConfirmCopy(['exec-1'], [s]).title, '关闭容器 Exec');
});
