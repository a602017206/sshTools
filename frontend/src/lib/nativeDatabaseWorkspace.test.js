import test from 'node:test';
import assert from 'node:assert/strict';
import { nativeDatabaseWorkspace } from './nativeDatabaseWorkspace.js';

test('Redis、Elasticsearch 和 Kafka 工作区声明支持资源详情', () => {
  for (const databaseType of ['redis', 'elasticsearch', 'kafka']) {
    assert.equal(nativeDatabaseWorkspace(databaseType).canDescribe, true, databaseType);
  }
});
