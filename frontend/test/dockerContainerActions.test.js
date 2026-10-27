import test from 'node:test';
import assert from 'node:assert/strict';
import {
  actionSuccessLabel,
  canRemove,
  canExec,
  canRestart,
  canStart,
  canStop,
  containerActionMenu,
  isStillRunningError,
  removeConfirmCopy,
} from '../src/lib/dockerContainerActions.js';

const up = { ID: 'abc', Name: 'web', Status: 'Up 2 hours' };
const exited = { ID: 'def', Name: 'db', Status: 'Exited (0) 3 minutes ago' };
const restarting = { ID: 'ghi', Name: 'job', Status: 'Restarting (1) 5 seconds ago' };

test('canStart/canStop depend on state', () => {
  assert.equal(canStart(up), false);
  assert.equal(canStop(up), true);
  assert.equal(canStart(exited), true);
  assert.equal(canStop(exited), false);
  assert.equal(canStart(restarting), false);
  assert.equal(canStop(restarting), true);
  assert.equal(canStart(null), false);
  assert.equal(canStop(undefined), false);
});

test('lowercase status field is supported', () => {
  assert.equal(canStop({ status: 'up 1 minute' }), true);
});

test('restart/remove always available for existing container', () => {
  assert.equal(canRestart(exited), true);
  assert.equal(canRemove(up), true);
  assert.equal(canRestart(null), false);
});

test('containerActionMenu disables by state', () => {
  const byId = (c) => Object.fromEntries(containerActionMenu(c).map((i) => [i.id, i]));
  const running = byId(up);
  assert.equal(running.start.disabled, true);
  assert.equal(running.stop.disabled, false);
  assert.equal(running.remove.disabled, false);
  assert.equal(running.remove.danger, true);
  const stopped = byId(exited);
  assert.equal(stopped.start.disabled, false);
  assert.equal(stopped.stop.disabled, true);
  assert.deepEqual(
    containerActionMenu(up).map((i) => i.id),
    ['detail', 'exec', 'start', 'stop', 'restart', 'remove'],
  );
});

test('removeConfirmCopy differs for force', () => {
  const normal = removeConfirmCopy('web', false);
  const forced = removeConfirmCopy('web', true);
  assert.match(normal.message, /web/);
  assert.equal(normal.confirmText, '删除');
  assert.equal(forced.confirmText, '强制删除');
  assert.notEqual(normal.title, forced.title);
  assert.match(removeConfirmCopy('').message, /该容器/);
});

test('isStillRunningError detects running hint', () => {
  assert.equal(isStillRunningError(new Error('cannot remove a running container')), true);
  assert.equal(isStillRunningError('容器仍在运行'), true);
  assert.equal(isStillRunningError('permission denied'), false);
});

test('actionSuccessLabel', () => {
  assert.equal(actionSuccessLabel('start'), '已启动');
  assert.equal(actionSuccessLabel('x'), '操作完成');
});

test('canExec 仅运行中容器可用，菜单含 Exec', () => {
  assert.equal(canExec(up), true);
  assert.equal(canExec(exited), false);
  assert.equal(canExec(restarting), false);
  assert.equal(canExec(null), false);
  const exec = containerActionMenu(up)[1];
  assert.deepEqual(exec, { id: 'exec', label: 'Exec', disabled: false });
  assert.equal(containerActionMenu(exited)[1].disabled, true);
});
