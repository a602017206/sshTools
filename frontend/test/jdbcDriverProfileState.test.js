import assert from 'node:assert/strict';
import test from 'node:test';

import { jdbcProfileActionState } from '../src/lib/jdbcDriverProfileState.js';

test('未安装的所选 profile 显示安装操作，即使同一驱动的其他 profile 已安装', () => {
  const state = jdbcProfileActionState({
    driverInstalled: true,
    profileInstalled: false
  });

  assert.deepEqual(state, {
    canInstall: true,
    canValidate: false,
    canRemove: false
  });
});

test('已安装的所选 profile 显示校验、重新安装和卸载操作', () => {
  const state = jdbcProfileActionState({
    driverInstalled: false,
    profileInstalled: true
  });

  assert.deepEqual(state, {
    canInstall: false,
    canValidate: true,
    canRemove: true
  });
});
