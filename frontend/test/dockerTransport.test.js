import test from 'node:test';
import assert from 'node:assert/strict';
import { buildDockerMetadata, defaultDockerName } from '../src/lib/dockerTransport.js';
import { resolveMode, modeForAsset, sessionMatchesMode } from '../src/lib/workspaceTabs.js';

test('buildDockerMetadata local', () => {
  const m = buildDockerMetadata({ transport: 'local' });
  assert.equal(m.transport, 'local');
  assert.equal(m.domain, 'docker');
});

test('docker is its own mode', () => {
  assert.equal(resolveMode('docker'), 'docker');
  assert.equal(modeForAsset({ type: 'docker' }), 'docker');
  assert.equal(sessionMatchesMode({ type: 'docker' }, 'docker'), true);
  assert.equal(sessionMatchesMode({ type: 'docker' }, 'ssh'), false);
});

test('defaultDockerName suffixes ssh asset name', () => {
  assert.equal(defaultDockerName({ name: 'prod' }), 'prod-Docker');
});
