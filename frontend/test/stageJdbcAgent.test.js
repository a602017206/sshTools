import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

test('前端构建使用跨平台 Node 脚本暂存 JDBC agent', async () => {
  const packageJSON = JSON.parse(await readFile(new URL('../package.json', import.meta.url), 'utf8'));

  assert.match(packageJSON.scripts.build, /node \.\.\/scripts\/stage-jdbc-agent\.mjs/);
  assert.doesNotMatch(packageJSON.scripts.build, /stage-jdbc-agent\.sh/);

  const stagingScript = await readFile(new URL('../../scripts/stage-jdbc-agent.mjs', import.meta.url), 'utf8');
  assert.match(stagingScript, /gradlew\.bat/);
  assert.match(stagingScript, /copyFile/);
});
