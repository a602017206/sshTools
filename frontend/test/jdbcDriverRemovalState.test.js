import assert from 'node:assert/strict';
import test from 'node:test';

import { jdbcDriverRemovalConfirmation } from '../src/lib/jdbcDriverRemovalState.js';

test('卸载确认信息使用当前选中的驱动 profile', () => {
  assert.deepEqual(
    jdbcDriverRemovalConfirmation({ name: '人大金仓', id: 'kingbase' }, { version: 'V8' }),
    {
      title: '卸载 JDBC 驱动',
      message: '确定卸载 人大金仓 V8 吗？',
      confirmText: '卸载'
    }
  );
});

test('缺少驱动或 profile 时不创建卸载确认信息', () => {
  assert.equal(jdbcDriverRemovalConfirmation(null, { version: '8.4.0' }), null);
  assert.equal(jdbcDriverRemovalConfirmation({ name: 'MySQL' }, null), null);
});
