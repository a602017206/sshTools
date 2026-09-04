import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

import {
  clearWindowSelection,
  getViewportMenuPosition,
  resolveContextMenuPoint,
} from '../src/lib/contextMenu.js';

test('视口菜单用 client 坐标并在窗口边缘夹紧', () => {
  assert.deepEqual(
    getViewportMenuPosition({
      clientX: 24,
      clientY: 36,
      menuWidth: 160,
      menuHeight: 40,
      viewWidth: 800,
      viewHeight: 600,
    }),
    { x: 24, y: 36 },
  );
  assert.deepEqual(
    getViewportMenuPosition({
      clientX: 780,
      clientY: 590,
      menuWidth: 160,
      menuHeight: 80,
      viewWidth: 800,
      viewHeight: 600,
    }),
    { x: 632, y: 512 },
  );
});

test('右键事件取出视口坐标并取消默认菜单', () => {
  const calls = [];
  const event = {
    clientX: 120,
    clientY: 80,
    preventDefault() { calls.push('prevent'); },
    stopPropagation() { calls.push('stop'); },
  };
  assert.deepEqual(resolveContextMenuPoint(event, { menuWidth: 160, menuHeight: 40, viewWidth: 800, viewHeight: 600 }), {
    x: 120,
    y: 80,
  });
  assert.deepEqual(calls, ['prevent', 'stop']);
});

test('clearWindowSelection 会清掉当前选区', () => {
  const removed = [];
  const previous = globalThis.window;
  globalThis.window = {
    getSelection() {
      return { removeAllRanges() { removed.push(true); } };
    },
  };
  try {
    clearWindowSelection();
    assert.equal(removed.length, 1);
  } finally {
    if (previous === undefined) delete globalThis.window;
    else globalThis.window = previous;
  }
});

test('资产树、标签栏和文件管理右键菜单挂到 body，避免 backdrop-filter 偏移', async () => {
  const [list, tabs, files, menu] = await Promise.all([
    readFile(new URL('../src/components/AssetList.svelte', import.meta.url), 'utf8'),
    readFile(new URL('../src/components/TerminalPanel.svelte', import.meta.url), 'utf8'),
    readFile(new URL('../src/components/FileManager.svelte', import.meta.url), 'utf8'),
    readFile(new URL('../src/components/FileManagerContextMenu.svelte', import.meta.url), 'utf8'),
  ]);
  assert.match(list, /use:portalToBody/);
  assert.match(list, /resolveContextMenuPoint/);
  assert.match(tabs, /use:portalToBody/);
  assert.match(files, /getViewportMenuPosition/);
  assert.match(menu, /use:portalToBody/);
  assert.match(menu, /fixed/);
});
