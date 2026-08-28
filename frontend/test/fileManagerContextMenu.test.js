import assert from 'node:assert/strict';
import test from 'node:test';

import {
  getContextMenuPosition,
  getFileManagerMenuFlags,
  getSubmenuPlacement,
  isPathFavorite,
  isValidOctalMode,
  joinRemotePath,
  matchFileManagerShortcut,
  shiftMenuTopForInlineMore,
  toggleFavoriteHistory,
  uniqueCopyName,
  unixModeToOctal,
} from '../src/lib/fileManagerContextMenu.js';

test('拼接远程路径', () => {
  assert.equal(joinRemotePath('/', 'logs'), '/logs');
  assert.equal(joinRemotePath('/opt/', 'app.yml'), '/opt/app.yml');
  assert.equal(joinRemotePath('/opt/app', 'lib'), '/opt/app/lib');
});

test('复制重名时生成 copy 后缀', () => {
  assert.equal(uniqueCopyName(new Set(['a.txt']), 'b.txt'), 'b.txt');
  assert.equal(uniqueCopyName(new Set(['a.txt']), 'a.txt'), 'a copy.txt');
  assert.equal(uniqueCopyName(new Set(['a.txt', 'a copy.txt']), 'a.txt'), 'a copy 2.txt');
});

test('把 Unix 权限串转成八进制', () => {
  assert.equal(unixModeToOctal('-rw-r--r--'), '644');
  assert.equal(unixModeToOctal('drwxr-xr-x'), '755');
  assert.equal(unixModeToOctal('-rwxrwxrwx'), '777');
  assert.equal(isValidOctalMode('644'), true);
  assert.equal(isValidOctalMode('0755'), true);
  assert.equal(isValidOctalMode('88'), false);
});

test('收藏路径可加入和取消', () => {
  assert.equal(isPathFavorite(['/opt'], '/opt'), true);
  assert.deepEqual(toggleFavoriteHistory(['/a', '/b'], '/a', 5), ['/b']);
  assert.deepEqual(toggleFavoriteHistory(['/a'], '/b', 2), ['/b', '/a']);
});

test('右键菜单在面板内夹紧；子菜单优先右侧，左右都不够时向下展开', () => {
  assert.deepEqual(
    getContextMenuPosition({
      clientX: 900,
      clientY: 700,
      root: { left: 0, top: 0, width: 320, height: 400 },
      menuWidth: 248,
      menuHeight: 300,
    }),
    { x: 64, y: 92 },
  );
  assert.equal(getSubmenuPlacement(20, 600), 'right');
  assert.equal(getSubmenuPlacement(280, 520), 'left');
  assert.equal(getSubmenuPlacement(8, 300), 'down');
  assert.equal(shiftMenuTopForInlineMore(400, 500, 300), 56);
});

test('空白处右键禁用针对文件的动作，文件右键启用下载和复制', () => {
  const blank = getFileManagerMenuFlags({ currentPath: '/opt' });
  assert.equal(blank.hasFile, false);
  assert.equal(blank.canRename, false);
  assert.equal(blank.canDownload, false);
  assert.equal(blank.canFavorite, true);

  const file = getFileManagerMenuFlags({
    file: { name: 'app.yml', path: '/opt/app.yml', is_dir: false },
    currentPath: '/opt',
    clipboard: { files: [{ name: 'a' }] },
  });
  assert.equal(file.canCopy, true);
  assert.equal(file.canDownload, true);
  assert.equal(file.canPaste, true);
  assert.equal(file.canOpen, true);

  const dir = getFileManagerMenuFlags({
    file: { name: 'logs', path: '/opt/logs', is_dir: true },
  });
  assert.equal(dir.canCopy, false);
  assert.equal(dir.canCut, true);
  assert.equal(dir.canDownload, false);
});

test('识别文件管理快捷键', () => {
  assert.equal(matchFileManagerShortcut({ key: 'Enter' }), 'openLocal');
  assert.equal(matchFileManagerShortcut({ key: 'F2' }), 'rename');
  assert.equal(matchFileManagerShortcut({ key: 'r', metaKey: true }, { isMac: true }), 'refresh');
  assert.equal(matchFileManagerShortcut({ key: 'c', altKey: true, shiftKey: true }), 'copyPath');
  assert.equal(matchFileManagerShortcut({ key: 'c', metaKey: true }, { isMac: true }), 'copy');
});
