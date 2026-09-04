import assert from 'node:assert/strict';
import test from 'node:test';

import {
  buildRemoteIndex,
  classifyConflict,
  conflictsForItems,
  indexRemoteEntries,
  namesInParent,
  normalizeUploadItems,
  remapPrefix,
  resolveUploadConflicts,
} from '../src/lib/uploadConflict.js';

const files = (relPaths) => relPaths.map((relPath) => ({
  localPath: `/tmp/${relPath}`,
  relPath,
  isDir: false,
}));

const dir = (relPath) => ({
  localPath: `/tmp/${relPath}`,
  relPath,
  isDir: true,
});

test('规范化后端展开结果，兼容 camelCase 与 Go 导出字段', () => {
  assert.deepEqual(
    normalizeUploadItems([
      { LocalPath: '/a/b.txt', RelPath: 'b.txt', IsDir: false },
      { localPath: '/a/docs', relPath: 'docs', isDir: true },
      { localPath: '', relPath: 'skip' },
    ]),
    [
      { localPath: '/a/b.txt', relPath: 'b.txt', isDir: false },
      { localPath: '/a/docs', relPath: 'docs', isDir: true },
    ],
  );
});

test('同名文件算冲突，同名文件夹只合并不算冲突', () => {
  const remote = new Map([
    ['readme.txt', { isDir: false }],
    ['docs', { isDir: true }],
  ]);
  assert.equal(classifyConflict({ relPath: 'readme.txt', isDir: false }, remote.get('readme.txt')), 'file');
  assert.equal(classifyConflict({ relPath: 'docs', isDir: true }, remote.get('docs')), null);
  assert.equal(classifyConflict({ relPath: 'docs', isDir: false }, remote.get('docs')), 'type');
  assert.equal(classifyConflict({ relPath: 'new.txt', isDir: false }, remote.get('new.txt')), null);
});

test('从目录列表建立远程相对路径索引', () => {
  const index = indexRemoteEntries('', [
    { name: '..', is_dir: true, is_parent: true },
    { name: 'app.yml', is_dir: false },
    { name: 'logs', is_dir: true },
  ]);
  indexRemoteEntries('logs', [{ name: 'out.txt', is_dir: false }], index);
  assert.deepEqual([...index.keys()].sort(), ['app.yml', 'logs', 'logs/out.txt']);
  assert.equal(index.get('logs').isDir, true);
});

test('已存在的远程文件夹会继续列出子目录', async () => {
  const listings = {
    '': [{ name: 'project', is_dir: true }],
    project: [{ name: 'src', is_dir: true }, { name: 'README.md', is_dir: false }],
    'project/src': [{ name: 'main.go', is_dir: false }],
  };
  const index = await buildRemoteIndex(
    [dir('project'), dir('project/src'), { localPath: '/x', relPath: 'project/src/main.go', isDir: false }],
    async (relDir) => listings[relDir] || [],
  );
  assert.equal(index.get('project/src/main.go')?.isDir, false);
});

test('覆盖同名文件时保留原相对路径', async () => {
  const remote = new Map([['a.txt', { isDir: false }]]);
  const result = await resolveUploadConflicts(files(['a.txt']), remote, {
    promptFolder: async () => {
      throw new Error('should not prompt folder');
    },
    promptFile: async () => 'overwrite',
  });
  assert.deepEqual(result.items.map((item) => item.relPath), ['a.txt']);
  assert.deepEqual(result.deleteRemote, []);
});

test('重命名同名文件时使用 copy 后缀', async () => {
  const remote = new Map([['a.txt', { isDir: false }], ['a copy.txt', { isDir: false }]]);
  const result = await resolveUploadConflicts(files(['a.txt']), remote, {
    promptFolder: async () => {
      throw new Error('should not prompt folder');
    },
    promptFile: async () => 'rename',
  });
  assert.deepEqual(result.items.map((item) => item.relPath), ['a copy 2.txt']);
});

test('取消同名文件则跳过该文件', async () => {
  const remote = new Map([['a.txt', { isDir: false }]]);
  const result = await resolveUploadConflicts(
    files(['a.txt', 'b.txt']),
    remote,
    {
      promptFolder: async () => {
        throw new Error('should not prompt folder');
      },
      promptFile: async () => 'cancel',
    },
  );
  assert.deepEqual(result.items.map((item) => item.relPath), ['b.txt']);
});

test('已存在文件夹可全部覆盖内部同名文件，不再逐个询问', async () => {
  const remote = new Map([
    ['project', { isDir: true }],
    ['project/a.txt', { isDir: false }],
    ['project/b.txt', { isDir: false }],
  ]);
  const items = [
    dir('project'),
    { localPath: '/x/a.txt', relPath: 'project/a.txt', isDir: false },
    { localPath: '/x/b.txt', relPath: 'project/b.txt', isDir: false },
    { localPath: '/x/c.txt', relPath: 'project/c.txt', isDir: false },
  ];
  let filePrompts = 0;
  const result = await resolveUploadConflicts(items, remote, {
    promptFolder: async ({ name, conflictCount }) => {
      assert.equal(name, 'project');
      assert.equal(conflictCount, 2);
      return 'overwriteAll';
    },
    promptFile: async () => {
      filePrompts += 1;
      return 'overwrite';
    },
  });
  assert.equal(filePrompts, 0);
  assert.deepEqual(result.items.map((item) => item.relPath), [
    'project',
    'project/a.txt',
    'project/b.txt',
    'project/c.txt',
  ]);
});

test('已存在文件夹可重命名整棵树', async () => {
  const remote = new Map([
    ['project', { isDir: true }],
    ['project/a.txt', { isDir: false }],
  ]);
  const items = [
    dir('project'),
    { localPath: '/x/a.txt', relPath: 'project/a.txt', isDir: false },
  ];
  const result = await resolveUploadConflicts(items, remote, {
    promptFolder: async () => 'rename',
    promptFile: async () => {
      throw new Error('renamed folder should not prompt files');
    },
  });
  assert.deepEqual(result.items.map((item) => item.relPath), ['project copy', 'project copy/a.txt']);
});

test('逐个选择时按文件询问，全部覆盖可套用到剩余文件', async () => {
  const remote = new Map([
    ['project', { isDir: true }],
    ['project/a.txt', { isDir: false }],
    ['project/b.txt', { isDir: false }],
  ]);
  const items = [
    dir('project'),
    { localPath: '/x/a.txt', relPath: 'project/a.txt', isDir: false },
    { localPath: '/x/b.txt', relPath: 'project/b.txt', isDir: false },
  ];
  const fileDecisions = [];
  const result = await resolveUploadConflicts(items, remote, {
    promptFolder: async () => 'oneByOne',
    promptFile: async (spec) => {
      fileDecisions.push(spec.relPath);
      return spec.relPath.endsWith('a.txt') ? 'overwriteAll' : 'rename';
    },
  });
  assert.deepEqual(fileDecisions, ['project/a.txt']);
  assert.deepEqual(result.items.map((item) => item.relPath), [
    'project',
    'project/a.txt',
    'project/b.txt',
  ]);
});

test('文件与文件夹类型冲突覆盖前需要删除远程对象', async () => {
  const remote = new Map([['data', { isDir: true }]]);
  const result = await resolveUploadConflicts(files(['data']), remote, {
    promptFolder: async () => {
      throw new Error('file vs folder should use file prompt');
    },
    promptFile: async ({ kind }) => {
      assert.equal(kind, 'type');
      return 'overwrite';
    },
  });
  assert.deepEqual(result.deleteRemote, ['data']);
});

test('remapPrefix 只替换完整路径前缀', () => {
  assert.equal(remapPrefix('project/src/main.go', 'project', 'project copy'), 'project copy/src/main.go');
  assert.equal(remapPrefix('project', 'project', 'project copy'), 'project copy');
  assert.equal(remapPrefix('projector/a', 'project', 'project copy'), 'projector/a');
});

test('父目录下的已占用名称用于生成 copy 后缀', () => {
  const remote = new Map([
    ['docs', { isDir: true }],
    ['docs/a.txt', { isDir: false }],
    ['docs/a copy.txt', { isDir: false }],
  ]);
  assert.deepEqual([...namesInParent(remote, 'docs')].sort(), ['a copy.txt', 'a.txt']);
});

test('无冲突时不弹窗直接上传', async () => {
  const result = await resolveUploadConflicts(files(['fresh.txt']), new Map(), {
    promptFolder: async () => {
      throw new Error('no prompt');
    },
    promptFile: async () => {
      throw new Error('no prompt');
    },
  });
  assert.deepEqual(result.items.map((item) => item.relPath), ['fresh.txt']);
});
