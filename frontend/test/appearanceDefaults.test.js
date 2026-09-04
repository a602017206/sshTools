import assert from 'node:assert/strict';
import test from 'node:test';

import { getDefaultAppSettings } from '../src/settings/appearance.js';

test('getDefaultAppSettings includes session log and command suggest defaults', () => {
  const settings = getDefaultAppSettings();
  assert.equal(settings.session_log_enabled, true);
  assert.equal(settings.session_log_retention_days, 30);
  assert.equal(settings.session_log_redact_enabled, true);
  assert.equal(settings.command_suggest_enabled, true);
  assert.equal(settings.command_suggest_limit, 8);
});
