import assert from 'node:assert/strict';
import test from 'node:test';

import {
  TERMINAL_CHARSET_OPTIONS,
  applyCharsetToSessionId,
  applyCharsetToSessionMap,
  decodeTerminalOutput,
  normalizeTerminalCharset,
  terminalContextMenuItems,
} from '../src/lib/terminalCharset.js';

test('连接编码默认 UTF-8，并识别常见中文编码别名', () => {
  assert.equal(normalizeTerminalCharset(''), 'utf-8');
  assert.equal(normalizeTerminalCharset('UTF8'), 'utf-8');
  assert.equal(normalizeTerminalCharset('GBK'), 'gbk');
  assert.equal(normalizeTerminalCharset('gb2312'), 'gb2312');
  assert.equal(normalizeTerminalCharset('GB18030'), 'gb18030');
  assert.equal(normalizeTerminalCharset('Big5'), 'big5');
  assert.equal(normalizeTerminalCharset('latin1'), 'utf-8');
});

test('编码选项覆盖 SSH 客户端常见中文编码', () => {
  assert.deepEqual(
    TERMINAL_CHARSET_OPTIONS.map((item) => item.id),
    ['utf-8', 'gbk', 'gb2312', 'gb18030', 'big5']
  );
});

test('GBK 字节解码为 Unicode，UTF-8 保持原始字节', () => {
  const gbkNiHao = Uint8Array.from([0xc4, 0xe3, 0xba, 0xc3]);
  assert.equal(decodeTerminalOutput(gbkNiHao, 'gbk'), '你好');
  const utf8 = Uint8Array.from([0xe4, 0xbd, 0xa0]);
  assert.equal(decodeTerminalOutput(utf8, 'utf-8'), utf8);
});

test('终端右键菜单在有选区时允许复制，始终提供粘贴', () => {
  assert.deepEqual(terminalContextMenuItems(false), [
    { id: 'copy', label: '复制', disabled: true },
    { id: 'paste', label: '粘贴', disabled: false }
  ]);
  assert.equal(terminalContextMenuItems(true)[0].disabled, false);
});

test('改编码会更新同一连接下已打开的 SSH 会话，不碰数据库会话', () => {
  const sessions = new Map([
    ['ssh-1', { sessionId: 'ssh-1', type: 'ssh', connection: { id: 'c1', metadata: { encoding: 'utf-8' } } }],
    ['ssh-2', { sessionId: 'ssh-2', type: 'ssh', connection: { id: 'c1', metadata: {} } }],
    ['other', { sessionId: 'other', type: 'ssh', connection: { id: 'c2', metadata: { encoding: 'utf-8' } } }],
    ['db-1', { sessionId: 'db-1', type: 'database', connection: { id: 'c1' } }]
  ]);
  const { sessions: next, sessionIds, charset } = applyCharsetToSessionMap(sessions, 'c1', 'GBK');
  assert.equal(charset, 'gbk');
  assert.deepEqual(sessionIds, ['ssh-1', 'ssh-2']);
  assert.equal(next.get('ssh-1').connection.metadata.encoding, 'gbk');
  assert.equal(next.get('ssh-2').connection.encoding, 'gbk');
  assert.equal(next.get('other').connection.metadata.encoding, 'utf-8');
  assert.equal(next.get('db-1').connection.metadata, undefined);
});

test('单会话热切换编码只改当前标签', () => {
  const sessions = new Map([
    ['ssh-1', { sessionId: 'ssh-1', type: 'ssh', connection: { id: 'c1', metadata: { encoding: 'utf-8' } } }],
    ['ssh-2', { sessionId: 'ssh-2', type: 'ssh', connection: { id: 'c1', metadata: { encoding: 'utf-8' } } }]
  ]);
  const { sessions: next } = applyCharsetToSessionId(sessions, 'ssh-2', 'gb18030');
  assert.equal(next.get('ssh-1').connection.metadata.encoding, 'utf-8');
  assert.equal(next.get('ssh-2').connection.metadata.encoding, 'gb18030');
});
