import test from 'node:test';
import assert from 'node:assert/strict';
import {
  clampDockerDetailWidth,
  DEFAULT_DOCKER_DETAIL_WIDTH,
  dockerTransportLabel,
  dockerTransportOf,
  formatContainerCreated,
  isContainerRunning,
  MAX_DOCKER_DETAIL_WIDTH,
  MIN_DOCKER_DETAIL_WIDTH,
  usedCliFallback,
  visibleContainers
} from '../src/lib/dockerWorkspace.js';

test('dockerTransportOf 读取 metadata 并有合理回退', () => {
  assert.equal(dockerTransportOf({ metadata: { transport: 'SSH' } }), 'ssh');
  assert.equal(dockerTransportOf({ metadata: { ssh_connection_id: 'x' } }), 'ssh');
  assert.equal(dockerTransportOf({ host: '10.0.0.1' }), 'tcp');
  assert.equal(dockerTransportOf({}), 'local');
  assert.equal(dockerTransportLabel({ metadata: { transport: 'tcp' } }), 'TCP');
});

test('isContainerRunning / visibleContainers', () => {
  const list = [
    { ID: 'a', Status: 'Up 2 hours' },
    { ID: 'b', Status: 'Exited (0) 3 days ago' }
  ];
  assert.equal(isContainerRunning(list[0]), true);
  assert.equal(isContainerRunning(list[1]), false);
  assert.deepEqual(visibleContainers(list, true).map((c) => c.ID), ['a']);
  assert.equal(visibleContainers(list, false).length, 2);
});

test('isContainerRunning 识别 Inspect 形态的详情（State/Status 为 running）', () => {
  assert.equal(isContainerRunning({ ID: 'a', Name: 'web', State: 'running', Status: 'running' }), true);
  assert.equal(isContainerRunning({ State: 'Running' }), true);
  assert.equal(isContainerRunning({ State: 'exited', Status: 'exited' }), false);
  assert.equal(isContainerRunning({ State: 'paused', Status: 'paused' }), false);
  assert.equal(isContainerRunning({ State: 'restarting', Status: 'restarting' }), false);
  assert.equal(isContainerRunning(null), false);
});

test('formatContainerCreated 与 usedCliFallback', () => {
  assert.equal(formatContainerCreated(0), '-');
  assert.match(formatContainerCreated(1700000000), /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/);
  assert.equal(usedCliFallback([{ via: 'cli' }]), true);
  assert.equal(usedCliFallback([{}]), false);
});

test('clampDockerDetailWidth 默认与硬限', () => {
  assert.equal(DEFAULT_DOCKER_DETAIL_WIDTH, 420);
  assert.equal(MIN_DOCKER_DETAIL_WIDTH, 320);
  assert.equal(MAX_DOCKER_DETAIL_WIDTH, 720);
  assert.equal(clampDockerDetailWidth(Number.NaN, 1000), 420);
  assert.equal(clampDockerDetailWidth(200, 1000), 320);
  assert.equal(clampDockerDetailWidth(900, 2000), 720);
  assert.equal(clampDockerDetailWidth(500, 1000), 500);
});

test('clampDockerDetailWidth 受主区 55% 约束', () => {
  // 1000 * 0.55 = 550 → 900 硬夹 720 后再受 55% 约束
  assert.equal(clampDockerDetailWidth(900, 1000), 550);
  // 800 * 0.55 = 440 → 上限 440
  assert.equal(clampDockerDetailWidth(600, 800), 440);
  // 500 * 0.55 = 275 → 与 MIN 取大得 320，宽度 400 被夹到 320
  assert.equal(clampDockerDetailWidth(400, 500), 320);
  // mainWidth 无效时仅硬限
  assert.equal(clampDockerDetailWidth(600, 0), 600);
  assert.equal(clampDockerDetailWidth(600, Number.NaN), 600);
});
