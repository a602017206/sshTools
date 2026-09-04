import assert from 'node:assert/strict';
import test from 'node:test';

import { extractShellCommand } from '../src/lib/terminalCommandLine.js';

test('extractShellCommand 剥离 root # prompt', () => {
  assert.equal(
    extractShellCommand('[root@localhost pems]# cd lnpems-idc-hiddendanger/'),
    'cd lnpems-idc-hiddendanger/'
  );
});

test('extractShellCommand 剥离 user $ prompt', () => {
  assert.equal(extractShellCommand('user@host:~$ ls -la /opt/pems'), 'ls -la /opt/pems');
});

test('extractShellCommand 无 prompt 时返回整行 trim', () => {
  assert.equal(extractShellCommand('  echo hi  '), 'echo hi');
});

test('extractShellCommand 空行', () => {
  assert.equal(extractShellCommand(''), '');
  assert.equal(extractShellCommand('   '), '');
});
