import assert from 'node:assert/strict';
import test from 'node:test';

import {
  DOCKER_RESOURCE_TABS,
  filterResourceRows,
  formatDockerSize,
  imageDisplayTags,
  removeResourceConfirmCopy,
  resourceDetailFields,
  resourceRowKey,
  resourceSearchText,
} from '../src/lib/dockerResources.js';

test('资源标签页顺序与中文文案', () => {
  assert.deepEqual(DOCKER_RESOURCE_TABS, [
    { id: 'containers', label: '容器' },
    { id: 'images', label: '镜像' },
    { id: 'networks', label: '网络' },
    { id: 'volumes', label: '卷' },
  ]);
});

test('formatDockerSize 按 B/KB/MB/GB 格式化', () => {
  assert.equal(formatDockerSize(0), '0 B');
  assert.equal(formatDockerSize(512), '512 B');
  assert.equal(formatDockerSize(1024), '1.0 KB');
  assert.equal(formatDockerSize(1536), '1.5 KB');
  assert.equal(formatDockerSize(5 * 1024 * 1024), '5.0 MB');
  assert.equal(formatDockerSize(1_500_000_000), '1.4 GB');
  assert.equal(formatDockerSize(-1), '-');
  assert.equal(formatDockerSize(null), '-');
  assert.equal(formatDockerSize(undefined), '-');
  assert.equal(formatDockerSize('abc'), '-');
});

test('imageDisplayTags 支持 Tags 与 RepoTags，无标签显示 <none>', () => {
  assert.equal(imageDisplayTags({ Tags: ['nginx:latest', 'nginx:1.25'] }), 'nginx:latest, nginx:1.25');
  assert.equal(imageDisplayTags({ RepoTags: ['redis:7'] }), 'redis:7');
  assert.equal(imageDisplayTags({ Tags: [] }), '<none>');
  assert.equal(imageDisplayTags({ Tags: ['<none>:<none>'] }), '<none>');
  assert.equal(imageDisplayTags({}), '<none>');
  assert.equal(imageDisplayTags(null), '<none>');
});

test('resourceRowKey 按资源类型生成稳定 key', () => {
  assert.equal(resourceRowKey('images', { ID: 'abc', Tags: ['nginx:latest', 'x'] }), 'abc|nginx:latest');
  assert.equal(resourceRowKey('images', { ID: 'abc', Tags: [] }), 'abc|');
  assert.equal(resourceRowKey('networks', { ID: 'n1', Name: 'bridge' }), 'n1|bridge');
  assert.equal(resourceRowKey('volumes', { Name: 'data' }), 'data');
  assert.equal(resourceRowKey('containers', { ID: 'c1' }), 'c1');
  assert.equal(resourceRowKey('images', null), '');
});

test('removeResourceConfirmCopy 返回中文确认文案', () => {
  const img = removeResourceConfirmCopy('images', 'nginx:latest', false);
  assert.match(img.title, /删除镜像/);
  assert.match(img.message, /nginx:latest/);
  assert.equal(img.confirmLabel, '删除');

  const forced = removeResourceConfirmCopy('images', 'nginx:latest', true);
  assert.match(forced.title, /强制删除镜像/);
  assert.match(forced.message, /强制/);
  assert.equal(forced.confirmLabel, '强制删除');

  assert.match(removeResourceConfirmCopy('networks', 'net1', false).title, /删除网络/);
  const vol = removeResourceConfirmCopy('volumes', 'data', false);
  assert.match(vol.title, /删除卷/);
  assert.match(vol.message, /数据/);
  assert.match(removeResourceConfirmCopy('containers', 'web', true).title, /强制删除容器/);
});

test('resourceDetailFields 为各类型生成侧栏字段并跳过空值', () => {
  const img = resourceDetailFields('images', {
    ID: 'sha256:abc',
    Tags: ['nginx:latest'],
    Size: 2048,
    Architecture: 'amd64',
    Os: 'linux',
    Author: '',
  });
  assert.deepEqual(img.slice(0, 3), [
    { label: 'ID', value: 'sha256:abc' },
    { label: '标签', value: 'nginx:latest' },
    { label: '大小', value: '2.0 KB' },
  ]);
  assert.ok(img.some((f) => f.label === '架构' && f.value === 'amd64'));
  assert.ok(img.some((f) => f.label === '系统' && f.value === 'linux'));
  assert.ok(!img.some((f) => f.label === '作者'));

  const net = resourceDetailFields('networks', {
    ID: 'n1', Name: 'bridge', Driver: 'bridge', Scope: 'local', Internal: true, Attachable: false, IPAM: '172.17.0.0/16',
  });
  assert.ok(net.some((f) => f.label === '名称' && f.value === 'bridge'));
  assert.ok(net.some((f) => f.label === '内部网络' && f.value === '是'));
  assert.ok(net.some((f) => f.label === '可挂载' && f.value === '否'));
  assert.ok(net.some((f) => f.label === 'IPAM' && f.value === '172.17.0.0/16'));

  const vol = resourceDetailFields('volumes', {
    Name: 'data', Driver: 'local', Mountpoint: '/var/lib/docker/volumes/data/_data', Scope: 'local', Labels: { a: '1', b: '2' },
  });
  assert.ok(vol.some((f) => f.label === '挂载点' && f.value === '/var/lib/docker/volumes/data/_data'));
  assert.ok(vol.some((f) => f.label === '标签' && f.value === 'a=1, b=2'));

  assert.deepEqual(resourceDetailFields('images', null), []);
  assert.deepEqual(resourceDetailFields('unknown', { ID: 'x' }), []);
});

test('filterResourceRows 各类型可见列子串匹配', () => {
  assert.equal(filterResourceRows('images', [{ ID: 'sha256:abcdef0123456789', Tags: ['nginx:latest'], Size: 1024, Created: 1700000000 }], '').length, 1);
  assert.deepEqual(
    filterResourceRows('images', [{ ID: 'sha256:abcdef0123456789', Tags: ['nginx:latest'], Size: 1024, Created: 0 }], 'NGINX').map((r) => r.Tags[0]),
    ['nginx:latest']
  );
  assert.equal(filterResourceRows('images', [{ ID: 'sha256:abcdef0123456789', Tags: [], Size: 2048, Created: 0 }], '2.0 KB').length, 1);
  assert.equal(filterResourceRows('networks', [{ ID: 'netid1234567', Name: 'bridge', Driver: 'bridge', Scope: 'local' }], 'LOCAL').length, 1);
  assert.equal(filterResourceRows('volumes', [{ Name: 'data', Driver: 'local', Mountpoint: '/var/lib/docker/volumes/data', Created: 0 }], 'mountpoint-miss').length, 0);
  assert.equal(filterResourceRows('volumes', [{ Name: 'data', Driver: 'local', Mountpoint: '/var/lib/docker/volumes/data', Created: 0 }], 'volumes/data').length, 1);
  assert.ok(resourceSearchText('networks', { ID: 'x', Name: 'bridge', Driver: 'bridge', Scope: 'local' }).includes('bridge'));
});
