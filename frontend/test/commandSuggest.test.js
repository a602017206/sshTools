import assert from 'node:assert/strict';
import test from 'node:test';

import { pickSuggestFill, shouldOfferSuggest } from '../src/lib/commandSuggest.js';

test('pickSuggestFill 用建议替换整行', () => {
  assert.equal(pickSuggestFill('cd', 'cd /var/log'), 'cd /var/log');
  assert.equal(pickSuggestFill('partial command', 'git status'), 'git status');
});

test('shouldOfferSuggest 需 trim 后非空且开关开启', () => {
  assert.equal(shouldOfferSuggest('ls', true), true);
  assert.equal(shouldOfferSuggest('  docker ps  ', true), true);
  assert.equal(shouldOfferSuggest('', true), false);
  assert.equal(shouldOfferSuggest('   ', true), false);
  assert.equal(shouldOfferSuggest('ls', false), false);
});
