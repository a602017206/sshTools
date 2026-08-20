import assert from 'node:assert/strict';
import test from 'node:test';

import { getTerminalShortcutAction } from '../src/lib/terminalShortcuts.js';

test('有选区时 Ctrl 或 Cmd+C 复制，未选中时保留 Ctrl+C 中断语义', () => {
  assert.equal(getTerminalShortcutAction({ key: 'c', ctrlKey: true }, true), 'copy');
  assert.equal(getTerminalShortcutAction({ key: 'C', metaKey: true }, true), 'copy');
  assert.equal(getTerminalShortcutAction({ key: 'c', ctrlKey: true }, false), null);
});

test('macOS Cmd+C 无选区时吞掉按键，避免泄漏到 PTY', () => {
  assert.equal(getTerminalShortcutAction({ key: 'c', metaKey: true }, false), 'noop');
  assert.equal(getTerminalShortcutAction({ code: 'KeyC', metaKey: true }, false), 'noop');
});

test('识别终端常用粘贴快捷键', () => {
  assert.equal(getTerminalShortcutAction({ key: 'v', ctrlKey: true }, false), 'paste');
  assert.equal(getTerminalShortcutAction({ key: 'V', metaKey: true }, false), 'paste');
  assert.equal(getTerminalShortcutAction({ key: 'v', ctrlKey: true, shiftKey: true }, false), 'paste');
  assert.equal(getTerminalShortcutAction({ key: 'Insert', shiftKey: true }, false), 'paste');
});

test('Ctrl+Shift+C 和 Ctrl+Insert 仅在有选区时复制', () => {
  assert.equal(getTerminalShortcutAction({ key: 'c', ctrlKey: true, shiftKey: true }, true), 'copy');
  assert.equal(getTerminalShortcutAction({ key: 'Insert', ctrlKey: true }, true), 'copy');
  assert.equal(getTerminalShortcutAction({ key: 'c', ctrlKey: true, shiftKey: true }, false), null);
});

test('支持通过 event.code 识别复制粘贴键', () => {
  assert.equal(getTerminalShortcutAction({ code: 'KeyC', metaKey: true }, true), 'copy');
  assert.equal(getTerminalShortcutAction({ code: 'KeyV', ctrlKey: true }, false), 'paste');
  assert.equal(getTerminalShortcutAction({ code: 'Insert', ctrlKey: true }, true), 'copy');
});

test('带有冲突修饰键的组合不被终端快捷键接管', () => {
  assert.equal(getTerminalShortcutAction({ key: 'v', ctrlKey: true, metaKey: true }, false), null);
  assert.equal(getTerminalShortcutAction({ key: 'c', altKey: true, ctrlKey: true }, true), null);
});
