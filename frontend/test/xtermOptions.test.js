import assert from 'node:assert/strict';
import test from 'node:test';

import { createXtermOptions } from '../src/lib/xtermOptions.js';

test('全屏程序的 LF 不能被折成回车，否则 vim 状态栏会散开', () => {
  const options = createXtermOptions({ fontSize: 14, fontFamily: 'monospace', theme: {} });
  assert.equal(options.convertEol, false);
});
