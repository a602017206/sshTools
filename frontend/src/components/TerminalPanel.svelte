<script>
  import { assetsStore, connectionsStore, activeSessionIdStore } from '../stores.js';
  import Terminal from './Terminal.svelte';
  import SelectedDatabaseObjects from './SelectedDatabaseObjects.svelte';
  import DatabaseTablePanel from './DatabaseTablePanel.svelte';
  import TableStructurePanel from './TableStructurePanel.svelte';
  import NativeDatabasePanel from './NativeDatabasePanel.svelte';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  import InputDialog from './ui/InputDialog.svelte';
  import { onMount, onDestroy, tick } from 'svelte';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import { isNativeDatabaseType } from '../lib/nativeDatabaseTypes.js';
  import { resolveMode, sessionMatchesMode } from '../lib/workspaceTabs.js';
  import { copilotStore } from '../stores/copilot.js';

  export let activeMode = 'ssh';

  let terminalRefs = {};
  let sessionsList = [];
  let visibleSessions = [];
  let sessionUnsubscribers = new Map();
  let sessionCloseUnsubscribers = new Map();
  let osc7PendingBuffers = new Map();
  let handleDatabaseTableSelectEvent = null;
  let handleDatabaseTableStructureEvent = null;
  let databaseQuerySequence = 0;
  let databaseDesignerSequence = 0;

  // Close confirmation dialog state
  let showCloseConfirm = false;
  let sessionToClose = null;
  let showAuthInput = false;
  let authInputTitle = '';
  let authInputMessage = '';
  let authInputPlaceholder = '';
  let authInputDefault = '';
  let authInputType = 'text';
  let authInputAllowEmpty = false;
  let authInputTrim = true;
  let resolveAuthInput = null;
  let showSavePasswordConfirm = false;
  let resolveSavePasswordConfirm = null;

  $: mode = resolveMode(activeMode);
  $: sessionsList = $connectionsStore ? Array.from($connectionsStore.values()) : [];
  $: visibleSessions = sessionsList.filter((session) => sessionMatchesMode(session, mode));
  $: activeVisibleSession = visibleSessions.find((session) => session.sessionId === $activeSessionIdStore) || null;
  $: viewportIsConsole = !activeVisibleSession || activeVisibleSession.type !== 'database';
  $: {
    const activeId = $activeSessionIdStore;
    const activeSession = activeId
      ? sessionsList.find((session) => session.sessionId === activeId)
      : null;
    if (visibleSessions.length === 0) {
      // 当前模式无会话时不强制改写 activeId，避免影响另一模式的会话保活
    } else if (!activeSession || !sessionMatchesMode(activeSession, mode)) {
      const preferred =
        visibleSessions.find((session) => session.connected) ||
        visibleSessions[visibleSessions.length - 1];
      if (preferred?.sessionId && preferred.sessionId !== activeId) {
        activeSessionIdStore.set(preferred.sessionId);
      }
    }
  }
  export function insertCopilotText(sessionId, text) {
    if (!sessionId || text == null) return;
    handleTerminalData(sessionId, String(text));
  }

  function buildDbListSession(asset, sessionId) {
    const isNativeDatabase = isNativeDatabaseType(asset?.metadata?.db_type);
    return {
      sessionId,
      connection: asset,
      connected: true,
      createdAt: Date.now(),
      lastActivity: Date.now(),
      type: 'database',
      panelType: isNativeDatabase ? 'native-database' : 'database-list',
      dbSessionId: sessionId,
      tabName: `${asset.name} · 数据库`
    };
  }

  function buildDbTablePanelId(dbSessionId, databaseName, schemaName, tableName) {
    const dbPart = databaseName || '__default__';
    const schemaPart = schemaName || '__default__';
    return `dbtable_${dbSessionId}_${dbPart}_${schemaPart}_${tableName}`;
  }

  function updateAssetDbStateBySession(sessionId, connected) {
    if (!sessionId) return;
    assetsStore.update(items => items.map(item => {
      if (item.dbSessionId === sessionId) {
        return {
          ...item,
          dbConnected: connected,
          dbSessionId: connected ? sessionId : null
        };
      }
      return item;
    }));
  }

  export function openDatabaseSession({ asset, sessionId }) {
    if (!asset || !sessionId) return;

    const existing = $connectionsStore.get(sessionId);
    if (existing) {
      activeSessionIdStore.set(sessionId);
      return;
    }

    const newSession = buildDbListSession(asset, sessionId);
    connectionsStore.update(conns => {
      conns.set(sessionId, newSession);
      return conns;
    });
    activeSessionIdStore.set(sessionId);
  }

  function openDatabaseTablePanel({ sessionId, databaseName, schemaName = '', tableName }) {
    if (!sessionId || !tableName) return;

    const parentSession = $connectionsStore.get(sessionId);
    if (!parentSession) return;

    const panelId = buildDbTablePanelId(sessionId, databaseName, schemaName, tableName);
    const existing = $connectionsStore.get(panelId);
    if (existing) {
      activeSessionIdStore.set(panelId);
      return;
    }

    const dbLabel = databaseName ? `${databaseName}.` : '';
    const tableSession = {
      sessionId: panelId,
      connection: parentSession.connection,
      connected: true,
      createdAt: Date.now(),
      lastActivity: Date.now(),
      type: 'database',
      panelType: 'database-table',
      dbSessionId: sessionId,
      databaseName,
      schemaName,
      tableName,
      tabName: `${dbLabel}${tableName}`
    };

    connectionsStore.update(conns => {
      conns.set(panelId, tableSession);
      return conns;
    });
    activeSessionIdStore.set(panelId);
  }

  function openDatabaseQueryPanel({ sessionId, databaseName = '', schemaName = '', initialQuery = '' }) {
    if (!sessionId) return;
    const parentSession = $connectionsStore.get(sessionId);
    if (!parentSession) return;

    const panelId = `dbquery_${sessionId}_${Date.now()}_${++databaseQuerySequence}`;
    const querySession = {
      sessionId: panelId,
      connection: parentSession.connection,
      connected: true,
      createdAt: Date.now(),
      lastActivity: Date.now(),
      type: 'database',
      panelType: 'database-query',
      dbSessionId: sessionId,
      databaseName,
      schemaName,
      initialQuery,
      tabName: '查询'
    };

    connectionsStore.update(conns => {
      conns.set(panelId, querySession);
      return conns;
    });
    activeSessionIdStore.set(panelId);
  }

  function openDatabaseTableDesignerPanel({ sessionId, databaseName = '', schemaName = '', tableName = 'new_table', mode = 'design' }) {
    if (!sessionId) return;

    const parentSession = $connectionsStore.get(sessionId);
    if (!parentSession) return;

    const dbPart = databaseName || '__default__';
    const schemaPart = schemaName || '__default__';
    const panelId = mode === 'create'
      ? `dbdesigner_${sessionId}_${Date.now()}_${++databaseDesignerSequence}`
      : `dbdesigner_${sessionId}_${dbPart}_${schemaPart}_${tableName}`;
    const existing = $connectionsStore.get(panelId);
    if (existing) {
      activeSessionIdStore.set(panelId);
      return;
    }

    const dbLabel = databaseName ? `${databaseName}.` : '';
    const designerSession = {
      sessionId: panelId,
      connection: parentSession.connection,
      connected: true,
      createdAt: Date.now(),
      lastActivity: Date.now(),
      type: 'database',
      panelType: 'database-table-designer',
      dbSessionId: sessionId,
      databaseName,
      schemaName,
      tableName,
      designerMode: mode,
      tabName: mode === 'create' ? '新建表' : `设计 · ${dbLabel}${tableName}`
    };

    connectionsStore.update(conns => {
      conns.set(panelId, designerSession);
      return conns;
    });
    activeSessionIdStore.set(panelId);
  }

  // 当会话列表更新时，更新全局 terminalRefs
  $: if (window.terminalRefs !== undefined) {
    window.terminalRefs = terminalRefs;
  }
  
  // 导出 handleConnect 供 App.svelte 调用
  export async function handleConnect(asset) {
    console.log('Connecting to:', asset);

    // 检查 Wails 绑定是否可用
    if (!window.wailsBindings) {
      console.error('Wails bindings not loaded');
      alert('Wails 绑定未加载，请使用 wails dev 运行');
      return;
    }

    const { ConnectSSH, GetPassword, HasPassword, SavePassword } = window.wailsBindings;

    if (typeof ConnectSSH !== 'function') {
      console.error('ConnectSSH not available');
      alert('SSH 连接功能不可用');
      return;
    }

    // 获取认证信息
    let authValue = '';
    let passphrase = '';

    if (asset.auth_type === 'key') {
      // 密钥认证：提示输入 passphrase（如果密钥已加密）
      passphrase = await requestInput({
        title: `连接到 ${asset.name}`,
        message: '如果 SSH 密钥已加密，请输入 Passphrase（否则留空）：',
        placeholder: 'Passphrase（可留空）',
        inputType: 'password',
        allowEmpty: true,
        trimValue: false
      });
      if (passphrase === null) {
        return;
      }
      authValue = asset.key_path || '';
    } else {
      // 密码认证：尝试获取保存的密码
      try {
        const hasSaved = typeof HasPassword === 'function' && await HasPassword(asset.id);
        if (hasSaved) {
          authValue = await GetPassword(asset.id);
          console.log('Using saved password');
        } else {
          // 没有保存的密码，提示用户输入
          authValue = await requestInput({
            title: `连接到 ${asset.name}`,
            message: '请输入密码：',
            placeholder: '密码',
            inputType: 'password',
            allowEmpty: false,
            trimValue: false
          });
          if (authValue === null) {
            return;
          }
          // 询问是否保存密码
          if (authValue && await requestConfirm('保存密码', '是否保存密码以便下次自动连接？', 'warning') && typeof SavePassword === 'function') {
            try {
              await SavePassword(asset.id, authValue);
              console.log('Password saved successfully');
            } catch (error) {
              console.error('Failed to save password:', error);
            }
          }
        }
      } catch (error) {
        console.error('Failed to get saved password:', error);
        authValue = await requestInput({
          title: `连接到 ${asset.name}`,
          message: '请输入密码：',
          placeholder: '密码',
          inputType: 'password',
          allowEmpty: false,
          trimValue: false
        });
        if (authValue === null) {
          return;
        }
      }
    }

    // 如果没有认证信息，取消连接
    if (!authValue) {
      console.log('Connection cancelled - no auth value provided');
      return;
    }

    // 生成唯一的 session ID
    const sessionId = `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    console.log('Connecting to server:', asset.host, asset.port, asset.username, 'auth_type:', asset.auth_type);

    // 创建会话元数据
    const newSession = {
      sessionId,
      connection: asset,
      connected: false,
      createdAt: Date.now(),
      lastActivity: Date.now()
    };

      // 添加到连接存储
     connectionsStore.update(conns => {
       conns.set(sessionId, newSession);
       return conns;
     });

    // 设置为活动会话
    activeSessionIdStore.set(sessionId);

    // 获取终端尺寸 - 多次等待确保组件渲染完成
    await tick();
    await new Promise(resolve => setTimeout(resolve, 50));
    let size = terminalRefs[sessionId]?.getSize();
    if (!size) {
      size = { cols: 80, rows: 24 };
      console.warn('Terminal not ready, using default size:', size);
    }

    // 订阅输出事件（必须在 ConnectSSH 之前）
    console.log('Subscribing to events for session:', sessionId);
    subscribeToOutput(sessionId);
    await new Promise(resolve => setTimeout(resolve, 100));

    // 再次验证终端可用
    if (!terminalRefs[sessionId]) {
      console.error('Terminal component not ready after delay');
      return;
    }

    // 显示连接消息
    const terminal = terminalRefs[sessionId];
    if (!terminal) {
      console.error('Failed to get terminal reference for session:', sessionId);
      return;
    }

    console.log('Terminal ready, attempting to connect...');
    const authType = asset.auth_type === 'key' ? 'SSH key' : 'password';
    terminal.writeln(`正在连接 ${asset.username}@${asset.host}:${asset.port} (${authType})...`);
    terminal.writeln('');

    try {
      // 调用 Wails ConnectSSH API
      await ConnectSSH(
        sessionId,
        asset.host,
        asset.port,
        asset.username,
        asset.auth_type || 'password',
        authValue,
        passphrase,
        size.cols,
        size.rows
      );

      // 连接成功，更新会话状态
      newSession.connected = true;
      connectionsStore.update(conns => {
        conns.set(sessionId, newSession);
        return conns;
      });

      // 连接成功后聚焦终端
      setTimeout(() => {
        terminal.focus();
      }, 100);
    } catch (error) {
      console.error('Failed to connect:', error);

      // 显示错误消息
      if (terminal) {
        terminal.writeln(`\r\n连接失败: ${error.message || error}`);
      }

      // 清理失败的会话
      await closeSession(sessionId);
    }
  }

  function requestInput({ title, message, placeholder, defaultValue, inputType, allowEmpty, trimValue }) {
    return new Promise(resolve => {
      authInputTitle = title || '';
      authInputMessage = message || '';
      authInputPlaceholder = placeholder || '';
      authInputDefault = defaultValue || '';
      authInputType = inputType || 'text';
      authInputAllowEmpty = Boolean(allowEmpty);
      authInputTrim = trimValue !== false;
      resolveAuthInput = resolve;
      showAuthInput = true;
    });
  }

  function handleAuthInputConfirm(value) {
    showAuthInput = false;
    if (resolveAuthInput) {
      resolveAuthInput(value);
      resolveAuthInput = null;
    }
  }

  function handleAuthInputCancel() {
    showAuthInput = false;
    if (resolveAuthInput) {
      resolveAuthInput(null);
      resolveAuthInput = null;
    }
  }

  function requestConfirm(title, message, type = 'default') {
    return new Promise(resolve => {
      confirmTitle = title;
      confirmMessage = message;
      confirmType = type;
      resolveSavePasswordConfirm = resolve;
      showSavePasswordConfirm = true;
    });
  }

  let confirmTitle = '';
  let confirmMessage = '';
  let confirmType = 'default';

  function handleSavePasswordConfirm() {
    showSavePasswordConfirm = false;
    if (resolveSavePasswordConfirm) {
      resolveSavePasswordConfirm(true);
      resolveSavePasswordConfirm = null;
    }
  }

  function handleSavePasswordCancel() {
    showSavePasswordConfirm = false;
    if (resolveSavePasswordConfirm) {
      resolveSavePasswordConfirm(false);
      resolveSavePasswordConfirm = null;
    }
  }
  
  async function closeSession(sessionId) {
    await removeSession(sessionId, { closeBackend: true });
  }

  async function removeSession(sessionId, { closeBackend = true } = {}) {
    const session = $connectionsStore.get(sessionId);
    const isDatabaseListPanel = session?.type === 'database' && session?.panelType === 'database-list';
    const isNativeDatabasePanel = session?.type === 'database' && session?.panelType === 'native-database';
    const isDatabaseTablePanel = session?.type === 'database' && session?.panelType === 'database-table';
    const isDatabaseQueryPanel = session?.type === 'database' && session?.panelType === 'database-query';
    const isDatabaseDesignerPanel = session?.type === 'database' && session?.panelType === 'database-table-designer';

    // 取消订阅
    const unsubscribe = sessionUnsubscribers.get(sessionId);
    if (unsubscribe) {
      unsubscribe();
      sessionUnsubscribers.delete(sessionId);
    }
    const closeUnsubscribe = sessionCloseUnsubscribers.get(sessionId);
    if (closeUnsubscribe) {
      closeUnsubscribe();
      sessionCloseUnsubscribers.delete(sessionId);
    }

    // 释放终端引用
    delete terminalRefs[sessionId];

    if (closeBackend && (isDatabaseListPanel || isNativeDatabasePanel)) {
      const { CloseDatabase, CloseNativeDatabase } = window.wailsBindings || {};
      const close = isNativeDatabasePanel ? CloseNativeDatabase : CloseDatabase;
      if (typeof close === 'function') {
        try {
          await close(sessionId);
        } catch (error) {
          console.error('Failed to close database session:', error);
        }
      }
      updateAssetDbStateBySession(sessionId, false);
    } else if (closeBackend && !isDatabaseTablePanel && !isDatabaseQueryPanel && !isDatabaseDesignerPanel) {
      const { CloseSSH } = window.wailsBindings || {};
      if (typeof CloseSSH === 'function') {
        try {
          await CloseSSH(sessionId);
        } catch (error) {
          console.error('Failed to close session:', error);
        }
      }
    }

    connectionsStore.update(conns => {
      conns.delete(sessionId);

      if (isDatabaseListPanel || isNativeDatabasePanel) {
        const relatedPanels = Array.from(conns.entries())
          .filter(([_, value]) => value?.type === 'database' && value?.dbSessionId === sessionId)
          .map(([key]) => key);
        relatedPanels.forEach(key => {
          copilotStore.clearSession(key);
          conns.delete(key);
        });
      }

      return conns;
    });

    copilotStore.clearSession(sessionId);
    if (session?.dbSessionId) {
      copilotStore.clearSession(session.dbSessionId);
    }

    // 切换到另一个会话
    const remainingSessions = Array.from($connectionsStore.keys());
    if ($activeSessionIdStore === sessionId) {
      if (remainingSessions.length > 0) {
        activeSessionIdStore.set(remainingSessions[0]);
      } else {
        activeSessionIdStore.set(null);
      }
    }
  }

  async function handleBackendSessionClosed(sessionId) {
    const session = $connectionsStore.get(sessionId);
    if (!session) return;

    session.connected = false;
    connectionsStore.update(conns => {
      conns.set(sessionId, session);
      return conns;
    });

    await removeSession(sessionId, { closeBackend: false });
  }
  
  function handleTabChange(sessionId) {
    if (!$connectionsStore.has(sessionId)) return;
    activeSessionIdStore.set(sessionId);

    // 聚焦终端
    setTimeout(() => {
      const terminal = terminalRefs[sessionId];
      if (terminal) {
        terminal.focus();
        
        // 同步终端尺寸与后端
        const size = terminal.getSize();
        const { ResizeSSH } = window.wailsBindings || {};
        if (typeof ResizeSSH === 'function') {
          ResizeSSH(sessionId, size.cols, size.rows).catch(console.error);
        }
      }
    }, 50);
  }
  
  function handleTabClose(sessionId, event) {
    event.stopPropagation();
    const session = $connectionsStore.get(sessionId);
    if (!session) return;

    // 如果已连接，显示确认对话框
    if (session.type !== 'database' && session.connected) {
      sessionToClose = sessionId;
      showCloseConfirm = true;
    } else {
      closeSession(sessionId);
    }
  }

  function handleConfirmClose() {
    if (sessionToClose) {
      closeSession(sessionToClose);
      sessionToClose = null;
    }
    showCloseConfirm = false;
  }

  function handleCancelClose() {
    sessionToClose = null;
    showCloseConfirm = false;
  }
  
  function handleTabRename(sessionId, newName) {
    const session = $connectionsStore.get(sessionId);
    if (!session) return;

    session.tabName = newName.trim();
    connectionsStore.update(conns => {
      conns.set(sessionId, session);
      return conns;
    });
  }
  
  // 标签页双击编辑
  let editingTabId = null;
  let editingTabName = '';
  
  function startEditTab(sessionId) {
    const session = $connectionsStore.get(sessionId);
    if (session) {
      editingTabId = sessionId;
      editingTabName = session.tabName || session.connection.name;
    }
  }
  
  function finishEditTab() {
    if (editingTabId && editingTabName.trim()) {
      handleTabRename(editingTabId, editingTabName.trim());
    }
    editingTabId = null;
    editingTabName = '';
  }
  
  function cancelEditTab() {
    editingTabId = null;
    editingTabName = '';
  }
  
  function handleKeyDown(event) {
    if (event.key === 'Enter') {
      finishEditTab();
    } else if (event.key === 'Escape') {
      cancelEditTab();
    }
  }

  async function handleNewLocalTerminal() {
    if (!window.wailsBindings) {
      console.error('Wails bindings not loaded');
      alert('Wails 绑定未加载，请使用 wails dev 运行');
      return;
    }

    // Open local terminal directly (PowerShell on Windows, default shell on Unix)
    await openLocalTerminal('');
  }

  async function openLocalTerminal(shellType) {
    const { ConnectLocalShell } = window.wailsBindings;

    // Generate unique session ID
    const sessionId = `local_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    // Create local session metadata
    const newSession = {
      sessionId,
      connection: {
        name: 'Local Shell',
        host: 'localhost',
        port: 0,
        username: ''
      },
      connected: false,
      createdAt: Date.now(),
      lastActivity: Date.now(),
      tabName: 'Local',
      type: 'local'
    };

    connectionsStore.update(conns => {
      conns.set(sessionId, newSession);
      return conns;
    });

    // Set as active session
    activeSessionIdStore.set(sessionId);

    // Get terminal size
    await tick();
    await new Promise(resolve => setTimeout(resolve, 50));
    let size = terminalRefs[sessionId]?.getSize();
    if (!size) {
      size = { cols: 80, rows: 24 };
      console.warn('Terminal not ready, using default size:', size);
    }

    // Subscribe to local output events
    console.log('Subscribing to local events for session:', sessionId);
    subscribeToLocalOutput(sessionId);
    await new Promise(resolve => setTimeout(resolve, 100));

    // Verify terminal is ready
    if (!terminalRefs[sessionId]) {
      console.error('Terminal component not ready after delay');
      return;
    }

    // Display connection message
    const terminal = terminalRefs[sessionId];
    if (!terminal) {
      console.error('Failed to get terminal reference for session:', sessionId);
      return;
    }

    console.log('Terminal ready, starting local shell...');
    terminal.writeln(`正在启动本地终端${shellType ? ` (${shellType})` : ''}...`);
    terminal.writeln('');

    try {
      // Call Wails ConnectLocalShell API
      await ConnectLocalShell(sessionId, shellType, size.cols, size.rows);

      // Connection successful, update session state
      newSession.connected = true;
      connectionsStore.update(conns => {
        conns.set(sessionId, newSession);
        return conns;
      });

      // Focus terminal
      setTimeout(() => {
        terminal.focus();
      }, 100);
    } catch (error) {
      console.error('Failed to start local shell:', error);

      // Display error message
      if (terminal) {
        terminal.writeln(`\r\n启动本地终端失败: ${error.message || error}`);
      }

      // Clean up failed session
      await closeSession(sessionId);
    }
  }

  // Subscribe to local output events
  function subscribeToLocalOutput(sessionId) {
    const eventName = `local:output:${sessionId}`;
    const unsubscribe = EventsOn(eventName, (encodedData) => {
      const terminal = terminalRefs[sessionId];
      if (terminal) {
        // 解码 base64 数据以获取原始二进制字节
        // ZMODEM 协议需要原始二进制数据
        const decodedData = atob(encodedData);
        const octets = new Uint8Array(decodedData.length);
        for (let i = 0; i < decodedData.length; i++) {
          octets[i] = decodedData.charCodeAt(i);
        }
        maybeUpdateCwdFromOutput(sessionId, octets, true);
        terminal.write(octets);
      }
    });
    sessionUnsubscribers.set(sessionId, unsubscribe);
    subscribeToSessionClosed(sessionId, 'local');
  }

  function maybeUpdateCwdFromOutput(sessionId, octets, isLocal) {
    if (!sessionId || !octets || octets.length === 0 || isLocal) {
      return;
    }

    const { paths, pending } = extractOsc7PathsBuffered(sessionId, octets);
    osc7PendingBuffers.set(sessionId, pending);
    if (paths.length === 0) {
      return;
    }

    const latestPath = paths[paths.length - 1];
    if (!latestPath) {
      return;
    }

    const { UpdateCurrentPath } = window.wailsBindings || {};
    if (typeof UpdateCurrentPath === 'function') {
      UpdateCurrentPath(sessionId, latestPath).catch(error => {
        console.error('Failed to update current path:', error);
      });
    }
  }

  function extractOsc7PathsBuffered(sessionId, octets) {
    const pending = osc7PendingBuffers.get(sessionId);
    const combined = pending && pending.length > 0
      ? concatOctets(pending, octets)
      : octets;

    const paths = extractOsc7Paths(combined);
    const nextPending = extractOsc7Pending(combined);

    return { paths, pending: nextPending };
  }

  function concatOctets(left, right) {
    const combined = new Uint8Array(left.length + right.length);
    combined.set(left, 0);
    combined.set(right, left.length);
    return combined;
  }

  function extractOsc7Pending(octets) {
    const prefix = [0x1b, 0x5d, 0x37, 0x3b];
    let lastPrefixIndex = -1;

    for (let i = octets.length - prefix.length; i >= 0; i--) {
      if (
        octets[i] === prefix[0] &&
        octets[i + 1] === prefix[1] &&
        octets[i + 2] === prefix[2] &&
        octets[i + 3] === prefix[3]
      ) {
        lastPrefixIndex = i;
        break;
      }
    }

    if (lastPrefixIndex === -1) {
      return octets.slice(Math.max(0, octets.length - (prefix.length - 1)));
    }

    const terminatorIndex = findOsc7Terminator(octets, lastPrefixIndex + prefix.length);
    if (terminatorIndex === -1) {
      return octets.slice(lastPrefixIndex);
    }

    return octets.slice(Math.max(0, octets.length - (prefix.length - 1)));
  }

  function findOsc7Terminator(octets, startIndex) {
    for (let i = startIndex; i < octets.length; i++) {
      if (octets[i] === 0x07) {
        return i;
      }
      if (octets[i] === 0x1b && i + 1 < octets.length && octets[i + 1] === 0x5c) {
        return i;
      }
    }

    return -1;
  }

  function extractOsc7Paths(octets) {
    const paths = [];
    const prefix = [0x1b, 0x5d, 0x37, 0x3b];
    let i = 0;
    while (i <= octets.length - prefix.length) {
      if (
        octets[i] === prefix[0] &&
        octets[i + 1] === prefix[1] &&
        octets[i + 2] === prefix[2] &&
        octets[i + 3] === prefix[3]
      ) {
        i += prefix.length;
        const start = i;
        let end = -1;
        while (i < octets.length) {
          if (octets[i] === 0x07) {
            end = i;
            i += 1;
            break;
          }
          if (octets[i] === 0x1b && i + 1 < octets.length && octets[i + 1] === 0x5c) {
            end = i;
            i += 2;
            break;
          }
          i += 1;
        }

        if (end > start) {
          const content = octetsToString(octets.slice(start, end));
          const path = parseOsc7Path(content);
          if (path) {
            paths.push(path);
          }
        }
        continue;
      }
      i += 1;
    }

    return paths;
  }

  function octetsToString(octets) {
    let text = '';
    for (let i = 0; i < octets.length; i++) {
      text += String.fromCharCode(octets[i]);
    }
    return text;
  }

  function parseOsc7Path(content) {
    if (!content) {
      return '';
    }

    let value = content;
    if (value.startsWith('file://')) {
      value = value.slice('file://'.length);
      const slashIndex = value.indexOf('/');
      if (slashIndex === -1) {
        return '/';
      }
      value = value.slice(slashIndex);
    } else if (value.startsWith('file:')) {
      value = value.slice('file:'.length);
    }

    try {
      value = decodeURIComponent(value);
    } catch (error) {
      return value;
    }

    return value;
  }

  // 处理终端数据
  function handleTerminalData(sessionId, data) {
    if (!$connectionsStore.has(sessionId)) {
      return;
    }

    const session = $connectionsStore.get(sessionId);
    const { SendSSHData, SendLocalShellData } = window.wailsBindings || {};

    if (session && session.type === 'local') {
      if (typeof SendLocalShellData === 'function') {
        SendLocalShellData(sessionId, data).catch(error => {
          console.error('Failed to send local shell data:', error);
        });
      }
    } else {
      if (typeof SendSSHData === 'function') {
        SendSSHData(sessionId, data).catch(error => {
          console.error('Failed to send SSH data:', error);
        });
      }
    }
  }

  function encodeBinaryString(octets) {
    let binary = '';
    for (let i = 0; i < octets.length; i++) {
      binary += String.fromCharCode(octets[i]);
    }
    return binary;
  }

  function encodeBase64(octets) {
    return btoa(encodeBinaryString(octets));
  }

  function handleZModemTransfer(sessionId, octets) {
    if (!$connectionsStore.has(sessionId)) {
      return;
    }

    const session = $connectionsStore.get(sessionId);
    const {
      SendSSHData,
      SendSSHDataBinary,
      SendLocalShellData,
      SendLocalShellDataBinary
    } = window.wailsBindings || {};

    if (session && session.type === 'local') {
      if (typeof SendLocalShellDataBinary === 'function') {
        const encoded = encodeBase64(octets);
        SendLocalShellDataBinary(sessionId, encoded).catch(error => {
          console.error('Failed to send local shell binary data:', error);
        });
      } else if (typeof SendLocalShellData === 'function') {
        const binary = encodeBinaryString(octets);
        SendLocalShellData(sessionId, binary).catch(error => {
          console.error('Failed to send local shell data:', error);
        });
      }
    } else {
      if (typeof SendSSHDataBinary === 'function') {
        const encoded = encodeBase64(octets);
        SendSSHDataBinary(sessionId, encoded).catch(error => {
          console.error('Failed to send SSH binary data:', error);
        });
      } else if (typeof SendSSHData === 'function') {
        const binary = encodeBinaryString(octets);
        SendSSHData(sessionId, binary).catch(error => {
          console.error('Failed to send SSH data:', error);
        });
      }
    }
  }

  // 处理终端大小调整
  function handleTerminalResize(sessionId, cols, rows) {
    if (!$connectionsStore.has(sessionId)) {
      return;
    }

    const session = $connectionsStore.get(sessionId);
    const { ResizeSSH, ResizeLocalShell } = window.wailsBindings || {};

    if (session && session.type === 'local') {
      if (typeof ResizeLocalShell === 'function') {
        ResizeLocalShell(sessionId, cols, rows).catch(error => {
          console.error('Failed to resize local terminal:', error);
        });
      }
    } else {
      if (typeof ResizeSSH === 'function') {
        ResizeSSH(sessionId, cols, rows).catch(error => {
          console.error('Failed to resize SSH terminal:', error);
        });
      }
    }
  }

  // 订阅输出事件（导出供 App.svelte 使用）
  export function subscribeToOutput(sessionId) {
    // Wails 事件系统
    const eventName = `ssh:output:${sessionId}`;
    const unsubscribe = EventsOn(eventName, (encodedData) => {
      const terminal = terminalRefs[sessionId];
      if (terminal) {
        // 解码 base64 数据以获取原始二进制字节
        // ZMODEM 协议需要原始二进制数据
        const decodedData = atob(encodedData);
        const octets = new Uint8Array(decodedData.length);
        for (let i = 0; i < decodedData.length; i++) {
          octets[i] = decodedData.charCodeAt(i);
        }
        maybeUpdateCwdFromOutput(sessionId, octets, false);
        terminal.write(octets);
      }
    });
    sessionUnsubscribers.set(sessionId, unsubscribe);
    subscribeToSessionClosed(sessionId, 'ssh');
  }

  function subscribeToSessionClosed(sessionId, type) {
    const eventName = `${type}:closed:${sessionId}`;
    const unsubscribe = EventsOn(eventName, () => {
      handleBackendSessionClosed(sessionId).catch(error => {
        console.error('Failed to remove closed session:', error);
      });
    });
    sessionCloseUnsubscribers.set(sessionId, unsubscribe);
  }

  onDestroy(() => {
    // 取消所有订阅
    sessionUnsubscribers.forEach(unsubscribe => {
      unsubscribe();
    });
    sessionCloseUnsubscribers.forEach(unsubscribe => {
      unsubscribe();
    });

    if (handleDatabaseTableSelectEvent) {
      window.removeEventListener('database:table-select', handleDatabaseTableSelectEvent);
    }

    if (handleDatabaseTableStructureEvent) {
      window.removeEventListener('database:table-structure', handleDatabaseTableStructureEvent);
    }
  });


  onMount(async () => {
    // 加载 Wails 绑定到全局
    if (window.wailsBindings) {
      console.log('Wails bindings already loaded');
    }

    // 将 terminalRefs 存储到全局，供 App.svelte 访问
    window.terminalRefs = terminalRefs;

    // 将 sessionUnsubscribers 存储到全局
    if (!window.sessionUnsubscribers) {
      window.sessionUnsubscribers = sessionUnsubscribers;
    }

    console.log('TerminalPanel mounted, subscribing to events for sessions:', sessionsList);

    handleDatabaseTableSelectEvent = (event) => {
      const detail = event?.detail;
      if (!detail) return;
      openDatabaseTablePanel(detail);
    };
    window.addEventListener('database:table-select', handleDatabaseTableSelectEvent);
    handleDatabaseTableStructureEvent = (event) => {
      const detail = event?.detail;
      if (!detail) return;
      openDatabaseTableDesignerPanel({ ...detail, mode: 'design' });
    };
    window.addEventListener('database:table-structure', handleDatabaseTableStructureEvent);

    // 聚焦当前活动的终端
    await tick();
    if ($activeSessionIdStore && terminalRefs[$activeSessionIdStore]) {
      setTimeout(() => {
        terminalRefs[$activeSessionIdStore].focus();
      }, 100);
    }
  });
</script>

<div class="h-full flex flex-col ops-workspace-chrome">
  <!-- 标签栏 -->
  <div class="session-tabbar flex items-center border-b overflow-x-auto" style="border-color: var(--glass-border);">
    {#if visibleSessions.length === 0}
      <div class="px-4 py-2 text-xs ops-muted">
        {mode === 'database' ? '没有数据库会话' : '没有活动连接'}
      </div>
    {:else}
      {#each visibleSessions as session (session.sessionId)}
        <div
          class="session-tab group flex items-center gap-2 px-3 py-2 border-r cursor-pointer transition-all min-w-[168px] {
            $activeSessionIdStore === session.sessionId
              ? 'session-tab--active text-gray-900 dark:text-white'
              : 'text-gray-600 dark:text-gray-300'
          }"
          style="border-color: var(--glass-border);"
          role="button"
          tabindex="0"
          on:click={() => handleTabChange(session.sessionId)}
          on:dblclick={() => startEditTab(session.sessionId)}
          on:keydown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.preventDefault();
              handleTabChange(session.sessionId);
            }
          }}
        >
          <div class={`w-2 h-2 rounded-full flex-shrink-0 ${
            session.connected ? 'bg-green-500' : 'bg-gray-400 dark:bg-gray-600'
          }`} />

          {#if editingTabId === session.sessionId}
            <input
              type="text"
              bind:value={editingTabName}
              on:blur={finishEditTab}
              on:keydown={handleKeyDown}
              class="flex-1 bg-transparent text-sm outline-none px-1"
              focus
            />
          {:else}
            <div class="flex-1 min-w-0 flex flex-col leading-tight">
              <span class="text-xs font-medium truncate">{session.tabName || session.connection.name}</span>
              <span class="text-[10px] ops-muted truncate">
                {#if session.type === 'database'}
                  数据库
                {:else if session.connected}
                  已连接
                {:else}
                  断开
                {/if}
              </span>
            </div>
          {/if}

          <button
            on:click={(e) => handleTabClose(session.sessionId, e)}
            class="opacity-0 group-hover:opacity-100 p-0.5 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-opacity"
          >
            <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
      {/each}
    {/if}

    {#if mode === 'ssh'}
    <button
      on:click={handleNewLocalTerminal}
      class="ops-icon-button flex items-center gap-2 px-3 py-2 transition-colors min-w-[44px]"
      title="打开本地终端"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
      </svg>
    </button>
    {/if}
  </div>

  <!-- 终端内容：保留全部会话挂载，仅按当前活动标签显示 -->
  {#if sessionsList.length > 0}
    <div class="terminal-stage flex-1 min-h-0">
    <div class="terminal-content-area relative flex-1 flex flex-col ops-terminal-viewport" class:ops-terminal-viewport--console={viewportIsConsole} class:ops-terminal-viewport--panel={!viewportIsConsole}>
      {#each sessionsList as session (session.sessionId)}
        <div
          class="terminal-wrapper {
            $activeSessionIdStore === session.sessionId && sessionMatchesMode(session, mode) ? 'active' : 'inactive'
          }"
        >
          <!-- 终端 / 数据库面板窗口：信息已在标签栏，不再重复工具条 -->
          <div class="flex-1 overflow-hidden h-full">
            {#if session.type === 'database' && session.panelType === 'database-list'}
      <SelectedDatabaseObjects
        sessionId={session.sessionId}
        dbConfig={session.connection}
        on:open-table-structure={(event) => openDatabaseTableDesignerPanel({ ...event.detail, mode: 'design' })}
        on:open-table-designer={(event) => openDatabaseTableDesignerPanel(event.detail)}
        on:open-table-data={(event) => openDatabaseTablePanel(event.detail)}
        on:database:new-query={(event) => openDatabaseQueryPanel(event.detail)}
      />
            {:else if session.type === 'database' && session.panelType === 'database-table'}
              <DatabaseTablePanel
                sessionId={session.dbSessionId}
                dbConfig={session.connection}
                databaseName={session.databaseName}
                schemaName={session.schemaName}
                tableName={session.tableName}
              />
            {:else if session.type === 'database' && session.panelType === 'database-query'}
              <DatabaseTablePanel
                sessionId={session.dbSessionId}
                dbConfig={session.connection}
                databaseName={session.databaseName}
                schemaName={session.schemaName}
                initialQuery={session.initialQuery}
              />
            {:else if session.type === 'database' && session.panelType === 'database-table-designer'}
              <TableStructurePanel
                sessionId={session.dbSessionId}
                dbConfig={session.connection}
                databaseName={session.databaseName}
                schemaName={session.schemaName}
                tableName={session.tableName}
                mode={session.designerMode}
              />
            {:else if session.type === 'database' && session.panelType === 'native-database'}
              <NativeDatabasePanel sessionId={session.sessionId} dbConfig={session.connection} />
            {:else}
              <Terminal
                bind:this={terminalRefs[session.sessionId]}
                sessionId={session.sessionId}
                onData={handleTerminalData}
                onResize={handleTerminalResize}
                onZModemTransfer={handleZModemTransfer}
              />
            {/if}
          </div>
        </div>
      {/each}
    </div>
    </div>
  {:else}
    <div class="terminal-stage flex-1 min-h-0 flex items-center justify-center ops-muted">
      <div class="text-center ops-glass rounded-2xl px-8 py-7">
        <div class="text-base font-medium mb-2" style="color: var(--text-primary);">未选择连接</div>
        <div class="text-xs">从左侧资产列表选择一个服务器开始连接</div>
      </div>
    </div>
  {/if}
</div>

<ConfirmDialog
  bind:isOpen={showCloseConfirm}
  title="关闭 SSH 会话"
  message="确定要关闭此 SSH 会话吗？"
  type="warning"
  confirmText="确定关闭"
  cancelText="取消"
  onConfirm={handleConfirmClose}
  onCancel={handleCancelClose}
/>

<ConfirmDialog
  bind:isOpen={showSavePasswordConfirm}
  title={confirmTitle}
  message={confirmMessage}
  type={confirmType}
  confirmText="确定"
  cancelText="取消"
  onConfirm={handleSavePasswordConfirm}
  onCancel={handleSavePasswordCancel}
/>

<InputDialog
  bind:isOpen={showAuthInput}
  title={authInputTitle}
  message={authInputMessage}
  placeholder={authInputPlaceholder}
  defaultValue={authInputDefault}
  inputType={authInputType}
  allowEmpty={authInputAllowEmpty}
  trimValue={authInputTrim}
  confirmText="确定"
  cancelText="取消"
  onConfirm={handleAuthInputConfirm}
  onCancel={handleAuthInputCancel}
/>

<style>
  .session-tabbar {
    background: color-mix(in srgb, var(--glass-bg) 75%, transparent);
    backdrop-filter: blur(14px) saturate(var(--glass-saturate));
    -webkit-backdrop-filter: blur(14px) saturate(var(--glass-saturate));
  }

  .session-tab {
    background: transparent;
    border-bottom: 2px solid transparent;
    transition: background-color var(--trans-fast), border-color var(--trans-fast), color var(--trans-fast);
  }

  .session-tab:hover {
    background: color-mix(in srgb, var(--glass-highlight) 55%, transparent);
  }

  .session-tab--active {
    background: color-mix(in srgb, var(--glass-bg-strong) 88%, var(--accent-subtle));
    border-bottom-color: var(--ops-signal);
  }

  .terminal-stage {
    display: flex;
    flex-direction: column;
    padding: 10px 12px 12px;
    min-height: 0;
  }

  .terminal-content-area {
    position: relative;
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-height: 0;
  }

  .terminal-wrapper {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    padding: 0;
    display: flex;
    flex-direction: column;
    opacity: 0;
    pointer-events: none;
    transition: opacity 0.15s ease;
  }

  .terminal-wrapper.inactive {
    opacity: 0;
    pointer-events: none;
  }

  .terminal-wrapper.active {
    opacity: 1;
    pointer-events: auto;
    z-index: 1;
  }

  @media (prefers-reduced-motion: reduce) {
    .session-tab,
    .terminal-wrapper {
      transition: none;
    }
  }
</style>
