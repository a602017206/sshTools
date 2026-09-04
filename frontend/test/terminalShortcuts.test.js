import assert from 'node:assert/strict';
import test from 'node:test';

import { getTerminalShortcutAction, shouldScrollToBottomBeforeArrowKey } from '../src/lib/terminalShortcuts.js';

test('有选区时 Ctrl 或 Cmd+C 复制，未选中时保留 Ctrl+C 中断语义', () => {
  assert.equal(getTerminalShortcutAction({ key: 'c', ctrlKey: true }, true), 'copy');
  assert.equal(getTerminalShortcutAction({ key: 'C', metaKey: true }, true), 'copy');
  assert.equal(getTerminalShortcutAction({ key: 'c', ctrlKey: true }, false), null);
});

test('macOS Cmd+C 无选区时吞掉按键，避免泄漏到 PTY', () => {
  assert.equal(getTerminalShortcutAction({ key: 'c', metaKey: true }, false), 'noop');
  assert.equal(getTerminalShortcutAction({ code: 'KeyC', metaKey: true }, false), 'noop');
});

test('识别终端常用粘贴快捷键，保留 Ctrl+V 给远端终端', () => {
  assert.equal(getTerminalShortcutAction({ key: 'v', ctrlKey: true }, false), null);
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
  assert.equal(getTerminalShortcutAction({ code: 'KeyV', ctrlKey: true }, false), null);
  assert.equal(getTerminalShortcutAction({ code: 'KeyV', ctrlKey: true, shiftKey: true }, false), 'paste');
  assert.equal(getTerminalShortcutAction({ code: 'Insert', ctrlKey: true }, true), 'copy');
});

test('带有冲突修饰键的组合不被终端快捷键接管', () => {
  assert.equal(getTerminalShortcutAction({ key: 'v', ctrlKey: true, metaKey: true }, false), null);
  assert.equal(getTerminalShortcutAction({ key: 'c', altKey: true, ctrlKey: true }, true), null);
});

test('单独按下 Command 等修饰键不交给 xterm，避免滚到最底部', () => {
  assert.equal(getTerminalShortcutAction({ key: 'Meta', metaKey: true }, true), 'noop');
  assert.equal(getTerminalShortcutAction({ key: 'Control', ctrlKey: true }, false), 'noop');
  assert.equal(getTerminalShortcutAction({ key: 'Alt', altKey: true }, false), 'noop');
  assert.equal(getTerminalShortcutAction({ key: 'Shift', shiftKey: true }, false), 'noop');
});

test('滚屏后上下方向键应先回到底部再发给 shell', () => {
  assert.equal(shouldScrollToBottomBeforeArrowKey({ key: 'ArrowUp' }, 3), true);
  assert.equal(shouldScrollToBottomBeforeArrowKey({ key: 'ArrowDown' }, 1), true);
  assert.equal(shouldScrollToBottomBeforeArrowKey({ key: 'ArrowUp' }, 0), false);
  assert.equal(shouldScrollToBottomBeforeArrowKey({ key: 'ArrowUp', shiftKey: true }, 5), false);
  assert.equal(shouldScrollToBottomBeforeArrowKey({ key: 'ArrowUp', ctrlKey: true }, 5), false);
});
