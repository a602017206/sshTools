import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

import { formatRemoteText, resolveTextLanguage } from '../src/lib/remoteTextFormat.js';

test('按扩展名和内容选择语言', () => {
  assert.equal(resolveTextLanguage('app.json', '').id, 'json');
  assert.equal(resolveTextLanguage('deploy.yaml', 'a: 1\n').id, 'yaml');
  assert.equal(resolveTextLanguage('nginx.conf', 'server {\n  listen 80;\n}\n').id, 'nginx');
  assert.equal(resolveTextLanguage('app.conf', '{"a":1}\n').id, 'json');
  assert.equal(resolveTextLanguage('notes.txt', '<root><item>1</item></root>').id, 'xml');
  assert.equal(resolveTextLanguage('query.txt', 'select id from users').id, 'sql');
  assert.equal(resolveTextLanguage('.env', 'A=1\n').id, 'properties');
});

test('JSON、YAML、XML 和 SQL 可以格式化，失败时不改内容', () => {
  const json = formatRemoteText('app.conf', '{"a":1,"b":[2]}');
  assert.equal(json.ok, true);
  assert.equal(json.changed, true);
  assert.equal(json.text, '{\n  "a": 1,\n  "b": [\n    2\n  ]\n}\n');
  assert.equal(formatRemoteText('app.json', json.text).changed, false);

  const invalid = formatRemoteText('app.json', '{');
  assert.equal(invalid.ok, false);
  assert.match(invalid.error, /JSON/);

  const yaml = formatRemoteText('app.yml', 'a: 1 # keep\nb:\n  - x\n');
  assert.equal(yaml.ok, true);
  assert.match(yaml.text, /# keep/);
  assert.equal(formatRemoteText('app.yml', yaml.text).changed, false);

  const xml = formatRemoteText('page.html', '<div><span>hi</span></div>');
  assert.equal(xml.language, 'html');
  assert.equal(xml.text, '<div>\n  <span>\n    hi\n  </span>\n</div>\n');

  const sql = formatRemoteText('q.sql', "SELECT id FROM users WHERE name = 'select'");
  assert.equal(sql.text, "SELECT id\nFROM users\nWHERE name = 'select'\n");
  assert.equal(formatRemoteText('q.sql', sql.text).changed, false);
});

test('普通文本只整理空白，原文本留给调用方撤销', () => {
  const source = 'echo hi  \n\n';
  const result = formatRemoteText('run.sh', source);
  assert.equal(result.ok, true);
  assert.equal(result.text, 'echo hi\n');
  assert.equal(source, 'echo hi  \n\n');
});

test('编辑器把格式化做成可撤销修改，并用 Ctrl+S 写回', async () => {
  const source = await readFile(new URL('../src/components/RemoteTextEditorDialog.svelte', import.meta.url), 'utf8');
  assert.match(source, /formatRemoteText/);
  assert.match(source, /undo\(view\)/);
  assert.match(source, /SaveRemoteTextFile/);
  assert.match(source, /event\.preventDefault\(\)/);
  assert.match(source, /changes: \{ from: 0, to: view\.state\.doc\.length, insert: result\.text \}/);
});
