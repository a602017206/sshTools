import assert from 'node:assert/strict';
import test from 'node:test';

import { applyNativeArtifactToElasticsearch, applyNativeArtifactToRedis } from '../src/lib/nativeCopilotApply.js';

test('ES 填入 DSL 到 Discover', () => {
  const next = applyNativeArtifactToElasticsearch({
    type: 'native_query',
    content: '{"query":{"match_all":{}},"size":5}',
    summary: 'match all'
  }, { queryText: '' });
  assert.equal(next.activeTab, 'discover');
  assert.match(next.queryText, /match_all/);
});

test('ES 填入 Dev Tools', () => {
  const next = applyNativeArtifactToElasticsearch({
    type: 'native_query',
    content: JSON.stringify({ method: 'GET', path: '/logs/_mapping' })
  }, {});
  assert.equal(next.activeTab, 'devtools');
  assert.equal(next.devPath, '/logs/_mapping');
});

test('Redis 填入 CLI', () => {
  const next = applyNativeArtifactToRedis({
    type: 'native_query',
    content: 'GET session:1'
  }, {});
  assert.equal(next.activeTab, 'cli');
  assert.equal(next.applyTarget, 'cli');
  assert.equal(next.cliCommand, 'GET session:1');
});

test('Redis SCAN 填入 MATCH 工具栏而不是 CLI', () => {
  const next = applyNativeArtifactToRedis({
    type: 'native_query',
    content: JSON.stringify({ command: 'SCAN', cursor: 0, match: 'mini*', count: 100 })
  }, {});
  assert.equal(next.applyTarget, 'match');
  assert.equal(next.activeTab, 'key');
  assert.equal(next.keyPattern, 'mini*');
  assert.equal(next.shouldReloadKeys, true);
});

test('Redis command 数组 SCAN 填入 MATCH 并保留大小写', () => {
  const next = applyNativeArtifactToRedis({
    type: 'native_query',
    content: JSON.stringify({ command: ['SCAN', '0', 'MATCH', 'mini*', 'COUNT', '100'] })
  }, {});
  assert.equal(next.keyPattern, 'mini*');
  assert.equal(next.applyTarget, 'match');
});

test('Redis 纯通配模式填入 MATCH', () => {
  const next = applyNativeArtifactToRedis({
    type: 'native_query',
    content: 'user:*'
  }, {});
  assert.equal(next.applyTarget, 'match');
  assert.equal(next.keyPattern, 'user:*');
});

test('Redis 可强制 target=cli', () => {
  const next = applyNativeArtifactToRedis({
    type: 'native_query',
    target: 'cli',
    content: JSON.stringify({ command: ['SCAN', '0', 'MATCH', 'mini*', 'COUNT', '100'] })
  }, {});
  assert.equal(next.applyTarget, 'cli');
  assert.equal(next.cliCommand, 'SCAN 0 MATCH mini* COUNT 100');
});

test('Redis 填入逗号拼接残留可修复为 CLI 命令', () => {
  const next = applyNativeArtifactToRedis({
    type: 'native_query',
    target: 'cli',
    content: 'SCAN, 0, MATCH, mini*, COUNT, 100'
  }, {});
  assert.equal(next.cliCommand, 'SCAN 0 MATCH mini* COUNT 100');
});
