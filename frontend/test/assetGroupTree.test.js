import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import test from 'node:test';

import {
  ancestorGroupPaths,
  applyPreferredGroupToFormData,
  buildAssetGroupForest,
  flattenVisibleGroupTree,
  resolvePreferredConnectionGroup,
  splitGroupPath,
} from '../src/lib/assetGroupTree.js';

test('分组路径按斜杠拆成多级目录，无斜杠时保持原名', () => {
  assert.deepEqual(splitGroupPath('生产环境'), ['生产环境']);
  assert.deepEqual(splitGroupPath('生产/华东/核心'), ['生产', '华东', '核心']);
  assert.deepEqual(splitGroupPath(' 生产 / 华东 '), ['生产', '华东']);
  assert.deepEqual(splitGroupPath(''), ['']);
});

test('祖先路径包含自身，便于搜索时展开匹配节点', () => {
  assert.deepEqual(ancestorGroupPaths('生产/华东/核心'), ['生产', '生产/华东', '生产/华东/核心']);
  assert.deepEqual(ancestorGroupPaths('开发环境'), ['开发环境']);
});

test('同前缀的分组会收成嵌套文件夹，计数包含后代', () => {
  const forest = buildAssetGroupForest({
    '生产/华东': [{ id: 'a', name: '华东机' }],
    '生产/华北': [{ id: 'b', name: '华北机' }],
    开发环境: [{ id: 'c', name: '开发机' }],
    生产: [{ id: 'd', name: '生产入口' }],
  });

  assert.equal(forest.length, 2);
  assert.equal(forest[0].name, '开发环境');
  assert.equal(forest[0].assetCount, 1);
  assert.equal(forest[1].name, '生产');
  assert.equal(forest[1].path, '生产');
  assert.equal(forest[1].assetCount, 3);
  assert.equal(forest[1].assets.map((asset) => asset.id).join(','), 'd');
  assert.deepEqual(forest[1].children.map((child) => child.path), ['生产/华北', '生产/华东']);
});

test('仅展开的文件夹会露出子文件夹和本组连接', () => {
  const forest = buildAssetGroupForest({
    '生产/华东': [{ id: 'a' }],
    生产: [{ id: 'd' }],
  });
  const collapsed = flattenVisibleGroupTree(forest, new Set());
  assert.deepEqual(collapsed.map((row) => row.kind + ':' + (row.node?.path || row.asset.id)), ['group:生产']);

  const expanded = flattenVisibleGroupTree(forest, new Set(['生产', '生产/华东']));
  assert.deepEqual(
    expanded.map((row) => row.kind + ':' + (row.node?.path || row.asset.id)),
    ['group:生产', 'group:生产/华东', 'asset:a', 'asset:d'],
  );
});

test('新建连接只接受分组字符串，忽略点击事件对象', () => {
  assert.equal(resolvePreferredConnectionGroup('生产/华东'), '生产/华东');
  assert.equal(resolvePreferredConnectionGroup('  生产环境  '), '生产环境');
  assert.equal(resolvePreferredConnectionGroup(''), '');
  assert.equal(resolvePreferredConnectionGroup({ type: 'click' }), '');
  assert.equal(resolvePreferredConnectionGroup(undefined), '');
});

test('空白表单可带入右键文件夹对应的分组', () => {
  const form = applyPreferredGroupToFormData({ group: '' }, '生产/华东');
  assert.equal(form.group, '生产/华东');
  assert.equal(applyPreferredGroupToFormData({ group: '' }, { type: 'click' }).group, '');
});

test('资产树文件夹提供右键新建连接，并把路径带进新建弹窗', async () => {
  const list = await readFile(new URL('../src/components/AssetList.svelte', import.meta.url), 'utf8');
  const app = await readFile(new URL('../src/App.svelte', import.meta.url), 'utf8');
  const dialog = await readFile(new URL('../src/components/AddAssetDialog.svelte', import.meta.url), 'utf8');

  assert.match(list, /on:contextmenu=\{\(event\) => openGroupContextMenu/);
  assert.match(list, /新建连接/);
  assert.match(list, /onAddClick\?\.\(groupContextPath\)/);
  assert.match(app, /preferredGroup=\{preferredGroup\}/);
  assert.match(dialog, /export let preferredGroup/);
  assert.match(dialog, /applyPreferredGroupToFormData/);
  assert.match(dialog, /appliedPreferredGroupVersion/);
});
