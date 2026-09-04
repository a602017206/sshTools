import test from 'node:test';
import assert from 'node:assert/strict';
import {
  appendTerminalTail,
  buildChatHistory,
  buildCopilotWorkspaceContext,
  copilotChatPayload,
  copilotAssistantTitle,
  formatCopilotWorkspaceLabel,
  getTerminalTail,
  nativeCopilotObjectKind,
  resolveWorkspaceFocus
} from '../src/lib/copilotContext.js';

test('terminal tails are isolated by session and capped to the latest text', () => {
  let tails = {};
  tails = appendTerminalTail(tails, 'ssh-a', 'first-', 10);
  tails = appendTerminalTail(tails, 'ssh-a', 'second', 10);
  tails = appendTerminalTail(tails, 'ssh-b', 'other', 10);

  assert.equal(getTerminalTail(tails, 'ssh-a'), 'rst-second');
  assert.equal(getTerminalTail(tails, 'ssh-b'), 'other');
});

test('terminal tails ignore missing session IDs and empty output', () => {
  const tails = { existing: 'value' };
  assert.deepEqual(appendTerminalTail(tails, '', 'ignored'), tails);
  assert.deepEqual(appendTerminalTail(tails, 'existing', ''), tails);
});

test('chat history retains only the latest bounded turns', () => {
  const messages = Array.from({ length: 16 }, (_, index) => ({
    role: index % 2 ? 'assistant' : 'user', content: `message-${index}`
  }));
  const history = buildChatHistory(messages, { maxTurns: 4, maxChars: 1000 });
  assert.deepEqual(history.map((item) => item.Content), ['message-12', 'message-13', 'message-14', 'message-15']);
});

test('table tab context prefers the open table over sidebar selection', () => {
  const context = buildCopilotWorkspaceContext({
    session: {
      panelType: 'database-table',
      databaseName: 'orders_db',
      schemaName: 'sales',
      tableName: 'orders',
      connection: {
        host: 'db.example',
        user: 'alice',
        metadata: { db_type: 'postgresql', database: 'postgres' }
      }
    },
    navigation: { databaseName: 'other_db', schemaName: 'public' },
    focus: { objectName: 'customers', editorContent: 'SELECT * FROM orders' }
  });

  assert.equal(context.workspaceKind, 'jdbc');
  assert.equal(context.database, 'orders_db');
  assert.equal(context.schema, 'sales');
  assert.equal(context.objectKind, 'table');
  assert.equal(context.objectName, 'orders');
  assert.equal(context.editorContent, 'SELECT * FROM orders');
  assert.equal(context.dbType, 'postgresql');
});

test('jdbc home context uses navigation and selected table focus', () => {
  const context = buildCopilotWorkspaceContext({
    session: {
      type: 'database',
      connection: {
        host: 'db.example',
        user: 'alice',
        metadata: { db_type: 'mysql', database: 'app' }
      }
    },
    navigation: { databaseName: 'shop', schemaName: '' },
    focus: { objectKind: 'table', objectName: 'users' }
  });

  assert.equal(context.workspaceKind, 'jdbc');
  assert.equal(context.database, 'shop');
  assert.equal(context.objectName, 'users');
  assert.equal(context.objectKind, 'table');
});

test('native elasticsearch context carries the selected index and query', () => {
  const context = buildCopilotWorkspaceContext({
    session: {
      type: 'database',
      connection: {
        host: 'es.example',
        user: '',
        metadata: { db_type: 'elasticsearch' }
      }
    },
    focus: {
      objectKind: 'index',
      objectName: 'logs-2026',
      editorContent: '{"query":{"match_all":{}}}'
    }
  });

  assert.equal(context.workspaceKind, 'native');
  assert.equal(context.dbType, 'elasticsearch');
  assert.equal(context.objectKind, 'index');
  assert.equal(context.objectName, 'logs-2026');
  assert.equal(context.editorContent, '{"query":{"match_all":{}}}');
});

test('native object kind maps cache and search types', () => {
  assert.equal(nativeCopilotObjectKind('redis'), 'key');
  assert.equal(nativeCopilotObjectKind('elasticsearch'), 'index');
  assert.equal(nativeCopilotObjectKind('kafka'), 'topic');
});

test('workspace focus prefers the active tab then the backend session', () => {
  const focus = resolveWorkspaceFocus(
    {
      'dbtable_s1_shop_orders': { objectName: 'orders', editorContent: 'SELECT 1' },
      s1: { objectName: 'users' }
    },
    'dbtable_s1_shop_orders',
    's1'
  );
  assert.equal(focus.objectName, 'orders');
  assert.equal(resolveWorkspaceFocus({ s1: { objectName: 'users' } }, 'missing', 's1').objectName, 'users');
});

test('workspace label shows the open table or native resource', () => {
  assert.equal(
    formatCopilotWorkspaceLabel({ dbType: 'postgresql', database: 'shop', schema: 'sales', objectName: 'orders' }),
    'postgresql · shop.sales.orders'
  );
  assert.equal(
    formatCopilotWorkspaceLabel({ dbType: 'redis', objectParent: '0', objectName: 'session:1' }),
    'redis · 0 / session:1'
  );
});

test('SSH 会话即使误存 mysql db_type 也显示主机而不是数据库', () => {
  const context = buildCopilotWorkspaceContext({
    mode: 'ssh',
    session: {
      type: 'ssh',
      connection: {
        host: '10.0.0.8',
        user: 'root',
        metadata: { db_type: 'mysql', encoding: 'utf-8' }
      }
    }
  });
  assert.equal(context.workspaceKind, 'ssh');
  assert.equal(context.dbType, '');
  assert.equal(formatCopilotWorkspaceLabel(context), 'root@10.0.0.8');
});

test('chat payload includes schema and the open object', () => {
  const payload = copilotChatPayload(
    {
      host: 'db.example',
      user: 'alice',
      dbType: 'oracle',
      database: 'ORCL',
      schema: 'PEMS',
      objectKind: 'table',
      objectName: 'T_ORDER',
      objectParent: '',
      editorContent: 'SELECT 1 FROM dual'
    },
    {
      sessionID: 'db-1',
      mode: 'database',
      message: '给当前表加索引',
      history: [],
      terminalTail: ''
    }
  );

  assert.equal(payload.SessionID, 'db-1');
  assert.equal(payload.Schema, 'PEMS');
  assert.equal(payload.ObjectKind, 'table');
  assert.equal(payload.ObjectName, 'T_ORDER');
  assert.equal(payload.Database, 'ORCL');
  assert.equal(payload.EditorContent, 'SELECT 1 FROM dual');
  assert.equal(payload.DBType, 'oracle');
});

test('助手标题按 SSH / SQL / 缓存 / 搜索 分流', () => {
  assert.equal(copilotAssistantTitle({ workspaceKind: 'ssh' }, 'ssh'), 'Shell 助手');
  assert.equal(copilotAssistantTitle({ workspaceKind: 'jdbc', dbType: 'mysql' }, 'database'), 'SQL 助手');
  assert.equal(copilotAssistantTitle({ workspaceKind: 'native', dbType: 'redis' }, 'database'), '缓存助手');
  assert.equal(copilotAssistantTitle({ workspaceKind: 'native', dbType: 'elasticsearch' }, 'database'), '搜索助手');
  assert.equal(copilotAssistantTitle({ workspaceKind: 'native', dbType: 'kafka' }, 'database'), '消息助手');
  assert.equal(copilotAssistantTitle({ workspaceKind: 'native', dbType: 'mongodb' }, 'database'), '数据助手');
});
