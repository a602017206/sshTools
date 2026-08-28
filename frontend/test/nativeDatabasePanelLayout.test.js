import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('原生数据库面板按类型路由到独立工作区', async () => {
  const source = await readFile(new URL('../src/components/NativeDatabasePanel.svelte', import.meta.url), 'utf8');

  assert.match(source, /resolveNativeWorkspaceKind/);
  assert.match(source, /RedisWorkspace/);
  assert.match(source, /ElasticsearchWorkspace/);
  assert.match(source, /KafkaWorkspace/);
  assert.match(source, /GenericNativeWorkspace/);
});

test('Redis 工作区保留键编辑器与逻辑库下拉', async () => {
  const source = await readFile(new URL('../src/components/workspaces/RedisWorkspace.svelte', import.meta.url), 'utf8');

  assert.match(source, /RedisKeyEditor/);
  assert.match(source, /__db-select/);
  assert.match(source, /native-database-panel__context/);
  assert.match(source, /native-database-panel__inspector/);
});

test('Elasticsearch 工作区保留集群概览、索引搜索与 mapping', async () => {
  const source = await readFile(new URL('../src/components/workspaces/ElasticsearchWorkspace.svelte', import.meta.url), 'utf8');

  assert.match(source, /canSearchResources|搜索|cluster|mapping|Mapping/i);
  assert.match(source, /native-database-panel__context/);
  assert.match(source, /native-database-panel__inspector/);
});

test('Kafka 工作区使用 Topic 列表与 kafka 工作区标题', async () => {
  const source = await readFile(new URL('../src/components/workspaces/KafkaWorkspace.svelte', import.meta.url), 'utf8');

  assert.match(source, /Topic|nativeDatabaseWorkspace\('kafka'\)/);
  assert.match(source, /native-database-panel__context/);
  assert.match(source, /native-database-panel__inspector/);
});
