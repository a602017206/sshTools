import test from 'node:test';
import assert from 'node:assert/strict';
import {
  dockerTransportLabel,
  dockerTransportOf,
  formatContainerCreated,
  isContainerRunning,
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

test('formatContainerCreated 与 usedCliFallback', () => {
  assert.equal(formatContainerCreated(0), '-');
  assert.match(formatContainerCreated(1700000000), /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}$/);
  assert.equal(usedCliFallback([{ via: 'cli' }]), true);
  assert.equal(usedCliFallback([{}]), false);
});
