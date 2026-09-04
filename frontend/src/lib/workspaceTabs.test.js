import assert from 'node:assert/strict';
import test from 'node:test';

import {
  APP_MODES,
  SSH_TOOL_TABS,
  isDatabaseSession,
  modeForAsset,
  resolveMode,
  resolveSshToolTab,
  resolveWorkspace,
  sessionMatchesMode,
  sshToolPanelHidden
} from './workspaceTabs.js';

test('exposes ssh and database modes only', () => {
  assert.deepEqual(
    APP_MODES.map((mode) => mode.id),
    ['ssh', 'database']
  );
});

test('resolveMode keeps supported modes', () => {
  assert.equal(resolveMode('database'), 'database');
  assert.equal(resolveMode('ssh'), 'ssh');
});

test('resolveMode maps legacy workspace ids to dual modes', () => {
  assert.equal(resolveMode('dashboard'), 'ssh');
  assert.equal(resolveMode('terminal'), 'ssh');
  assert.equal(resolveMode('files'), 'ssh');
  assert.equal(resolveMode('performance'), 'ssh');
  assert.equal(resolveMode('docker'), 'ssh');
  assert.equal(resolveMode('logs'), 'ssh');
  assert.equal(resolveMode('not-a-mode'), 'ssh');
});

test('modeForAsset routes by asset type', () => {
  assert.equal(modeForAsset({ type: 'database' }), 'database');
  assert.equal(modeForAsset({ type: 'ssh' }), 'ssh');
  assert.equal(modeForAsset(null), 'ssh');
});

test('ssh tool tabs are files, performance and logs', () => {
  assert.deepEqual(
    SSH_TOOL_TABS.map((tab) => tab.id),
    ['files', 'performance', 'logs']
  );
  assert.equal(resolveSshToolTab('performance'), 'performance');
  assert.equal(resolveSshToolTab('logs'), 'logs');
  assert.equal(resolveSshToolTab('nope'), 'files');
});

test('ssh tool panels hide the inactive tab without implying unmount', () => {
  assert.equal(sshToolPanelHidden('files', 'files'), false);
  assert.equal(sshToolPanelHidden('files', 'performance'), true);
  assert.equal(sshToolPanelHidden('performance', 'files'), true);
  assert.equal(sshToolPanelHidden('performance', 'performance'), false);
  assert.equal(sshToolPanelHidden('logs', 'logs'), false);
  assert.equal(sshToolPanelHidden('logs', 'files'), true);
  assert.equal(sshToolPanelHidden('nope', 'files'), false);
});

test('resolveWorkspace stays compatible as a mode alias', () => {
  assert.equal(resolveWorkspace('database'), 'database');
  assert.equal(resolveWorkspace('files'), 'ssh');
});

test('sessionMatchesMode separates ssh and database tabs', () => {
  const ssh = { type: 'ssh', sessionId: 's1' };
  const local = { type: 'local', sessionId: 'l1' };
  const db = { type: 'database', sessionId: 'd1' };

  assert.equal(isDatabaseSession(db), true);
  assert.equal(sessionMatchesMode(ssh, 'ssh'), true);
  assert.equal(sessionMatchesMode(local, 'ssh'), true);
  assert.equal(sessionMatchesMode(db, 'ssh'), false);
  assert.equal(sessionMatchesMode(db, 'database'), true);
  assert.equal(sessionMatchesMode(ssh, 'database'), false);
});
