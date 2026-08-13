import assert from 'node:assert/strict';
import test from 'node:test';

import { getRowContextMenuPosition, shouldCloseRowContextMenu } from '../src/lib/databaseRowContextMenu.js';

test('数据行菜单以被右键单元格为锚点显示在其下方', () => {
  assert.deepEqual(getRowContextMenuPosition({ left: 624, bottom: 344 }), { x: 624, y: 344 });
});

test('点击菜单外部时关闭数据行菜单，点击菜单内部时保持打开', () => {
  assert.equal(shouldCloseRowContextMenu({ menuContainsTarget: false }), true);
  assert.equal(shouldCloseRowContextMenu({ menuContainsTarget: true }), false);
});
