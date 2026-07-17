import assert from 'node:assert/strict';
import test from 'node:test';

import { tableOpenEvents } from '../src/lib/databaseObjectActions.js';

test('表对象区分单击 DDL 与双击数据浏览事件', () => {
  assert.deepEqual(tableOpenEvents, {
    click: 'open-table-structure',
    doubleClick: 'open-table-data'
  });
});
