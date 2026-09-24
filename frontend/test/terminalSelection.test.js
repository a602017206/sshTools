import assert from 'node:assert/strict';
import test from 'node:test';

import {
  copyUsingHiddenTextarea,
  createSelectionMemory,
  prepareTerminalPaste,
  reduceSelectionMemory,
  resolveTerminalCopyText
} from '../src/lib/terminalSelection.js';

test('拖动时记下选区，tail 把缓冲区改空后仍保留快照', () => {
  let state = createSelectionMemory();
  state = reduceSelectionMemory(state, { type: 'pointerdown' });
  state = reduceSelectionMemory(state, { type: 'change', liveText: 'ERROR disk full', hasSelection: true });
  state = reduceSelectionMemory(state, { type: 'pointerup', liveText: 'ERROR disk full' });
  state = reduceSelectionMemory(state, { type: 'change', liveText: '', hasSelection: true });
  assert.equal(state.text, 'ERROR disk full');
  assert.equal(resolveTerminalCopyText('', true, state.text), 'ERROR disk full');
});

test('实时选区只有空白时改用快照', () => {
  assert.equal(resolveTerminalCopyText('\n  \n', true, 'log line'), 'log line');
});

test('没有选区时不拿旧快照冒充本次复制', () => {
  assert.equal(resolveTerminalCopyText('', false, 'old log'), '');
});

test('用户正在框选空白时清掉上一次快照', () => {
  let state = reduceSelectionMemory(createSelectionMemory(), {
    type: 'change',
    liveText: 'keep',
    hasSelection: true
  });
  state = reduceSelectionMemory(state, { type: 'pointerdown' });
  state = reduceSelectionMemory(state, { type: 'change', liveText: '', hasSelection: true });
  assert.equal(state.text, '');
});

test('粘贴把换行收成回车，并在远端开启时加括号粘贴', () => {
  assert.equal(prepareTerminalPaste('a\nb\r\nc', false), 'a\rb\rc');
  assert.equal(prepareTerminalPaste('a\nb', true), '\u001b[200~a\rb\u001b[201~');
  assert.equal(prepareTerminalPaste('', true), '');
});

test('同步复制把文本放进隐藏输入框再执行 copy', () => {
  const area = { value: '', style: {}, setAttribute() {}, focus() {}, select() {} };
  const doc = {
    body: {
      node: null,
      appendChild(node) { this.node = node; },
      removeChild() { this.node = null; }
    },
    createElement() { return area; },
    execCommand(command) {
      return command === 'copy' && area.value === 'log line';
    }
  };
  assert.equal(copyUsingHiddenTextarea('log line', doc), true);
  assert.equal(area.value, 'log line');
  assert.equal(doc.body.node, null);
});
