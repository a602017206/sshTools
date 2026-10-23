import test from 'node:test';
import assert from 'node:assert/strict';
import {
  appendLogText,
  asStringList,
  createThrottledAppender,
  dockerLogsEventName,
  isNearBottom,
  logChunkText,
  normalizeLogTail
} from '../src/lib/dockerLogs.js';

test('dockerLogsEventName 与后端事件名一致', () => {
  assert.equal(dockerLogsEventName('s1', 'abc'), 'docker:logs:s1:abc');
});

test('normalizeLogTail 仅接受预设值，其余回退 200', () => {
  assert.equal(normalizeLogTail('500'), 500);
  assert.equal(normalizeLogTail(1000), 1000);
  assert.equal(normalizeLogTail(7), 200);
  assert.equal(normalizeLogTail(undefined), 200);
});

test('logChunkText 兼容对象与字符串 payload', () => {
  assert.equal(logChunkText({ chunk: 'hi\n' }), 'hi\n');
  assert.equal(logChunkText('raw'), 'raw');
  assert.equal(logChunkText(null), '');
  assert.equal(logChunkText({ chunk: 1 }), '');
});

test('appendLogText 超限时保留尾部并从整行截断', () => {
  assert.equal(appendLogText('a\n', 'b\n', 100), 'a\nb\n');
  assert.equal(appendLogText('line1\nline2\n', 'line3\n', 12), 'line2\nline3\n');
  assert.equal(appendLogText('line1\nline2\n', 'line3\n', 14), 'line2\nline3\n');
  assert.equal(appendLogText('', 'abcdefghij', 4), 'ghij');
});

test('isNearBottom 判定贴底', () => {
  assert.equal(isNearBottom({ scrollTop: 976, scrollHeight: 1000, clientHeight: 24 }), true);
  assert.equal(isNearBottom({ scrollTop: 900, scrollHeight: 1000, clientHeight: 24 }), false);
  assert.equal(isNearBottom(null), true);
});

test('createThrottledAppender 合并同一窗口内的多个块', () => {
  const calls = [];
  let scheduled = null;
  const timers = {
    setTimeout: (fn) => {
      scheduled = fn;
      return 1;
    },
    clearTimeout: () => {
      scheduled = null;
    }
  };
  const appender = createThrottledAppender((text) => calls.push(text), 50, timers);
  appender.push('a');
  appender.push('b');
  appender.push('');
  assert.deepEqual(calls, []);
  scheduled();
  assert.deepEqual(calls, ['ab']);
  appender.push('c');
  appender.flush();
  assert.deepEqual(calls, ['ab', 'c']);
  appender.push('d');
  appender.cancel();
  assert.equal(scheduled, null);
  assert.deepEqual(calls, ['ab', 'c']);
});

test('asStringList 非数组返回空列表', () => {
  assert.deepEqual(asStringList(['a', 1]), ['a', '1']);
  assert.deepEqual(asStringList(null), []);
});
