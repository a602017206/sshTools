<script>
  import { connectionsStore, activeSessionIdStore } from '../stores.js';
  import Terminal from './Terminal.svelte';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  import { onMount, onDestroy, tick } from 'svelte';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';

  let terminalRefs = {};
  let sessionsList = [];
  let sessionUnsubscribers = new Map();

  // Close confirmation dialog state
  let showCloseConfirm = false;
  let sessionToClose = null;

  $: sessionsList = $connectionsStore ? Array.from($connectionsStore.values()) : [];

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
      passphrase = prompt(`连接到 ${asset.name}\n如果 SSH 密钥已加密，请输入 Passphrase（否则留空）：`) || '';
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
          authValue = prompt(`连接到 ${asset.name}\n请输入密码：`) || '';
          // 询问是否保存密码
          if (authValue && confirm('是否保存密码以便下次自动连接？') && typeof SavePassword === 'function') {
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
        authValue = prompt(`连接到 ${asset.name}\n请输入密码：`) || '';
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
  
  async function closeSession(sessionId) {
    // 取消订阅
    const unsubscribe = sessionUnsubscribers.get(sessionId);
    if (unsubscribe) {
      unsubscribe();
      sessionUnsubscribers.delete(sessionId);
    }

    // 释放终端引用
    delete terminalRefs[sessionId];

    // 关闭后端会话
    const { CloseSSH } = window.wailsBindings || {};
    if (typeof CloseSSH === 'function') {
      try {
        await CloseSSH(sessionId);
      } catch (error) {
        console.error('Failed to close session:', error);
      }
    }

    // 从状态中移除
    connectionsStore.update(conns => {
      conns.delete(sessionId);
      return conns;
    });

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
    if (session.connected) {
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

  // 处理终端数据
  function handleTerminalData(sessionId, data) {
    if (!$connectionsStore.has(sessionId)) {
      return;
    }

    const { SendSSHData } = window.wailsBindings || {};
    if (typeof SendSSHData === 'function') {
      SendSSHData(sessionId, data).catch(error => {
        console.error('Failed to send data:', error);
      });
    }
  }

  // 处理终端大小调整
  function handleTerminalResize(sessionId, cols, rows) {
    if (!$connectionsStore.has(sessionId)) {
      return;
    }

    const { ResizeSSH } = window.wailsBindings || {};
    if (typeof ResizeSSH === 'function') {
      ResizeSSH(sessionId, cols, rows).catch(error => {
        console.error('Failed to resize terminal:', error);
      });
    }
  }

  // 订阅输出事件（导出供 App.svelte 使用）
  export function subscribeToOutput(sessionId) {
    // Wails 事件系统
    const eventName = `ssh:output:${sessionId}`;
    const unsubscribe = EventsOn(eventName, (data) => {
      const terminal = terminalRefs[sessionId];
      if (terminal) {
        terminal.write(data);
      }
    });
    sessionUnsubscribers.set(sessionId, unsubscribe);
  }

  onDestroy(() => {
    // 取消所有订阅
    sessionUnsubscribers.forEach(unsubscribe => {
      unsubscribe();
    });
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

    // 聚焦当前活动的终端
    await tick();
    if ($activeSessionIdStore && terminalRefs[$activeSessionIdStore]) {
      setTimeout(() => {
        terminalRefs[$activeSessionIdStore].focus();
      }, 100);
    }
  });
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800">
  <!-- 标签栏 -->
  <div class="flex items-center bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 overflow-x-auto">
    {#if sessionsList.length === 0}
      <div class="px-4 py-2.5 text-sm text-gray-500 dark:text-gray-400">没有活动连接</div>
    {:else}
      {#each sessionsList as session (session.sessionId)}
        <div
          class="group flex items-center gap-2 px-4 py-2.5 border-r border-gray-200 dark:border-gray-700 cursor-pointer transition-all min-w-[180px] {
            $activeSessionIdStore === session.sessionId
              ? 'bg-white dark:bg-gray-800 text-gray-900 dark:text-white border-b-2 border-b-purple-600'
              : 'bg-gray-50 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-600'
          }"
          on:click={() => handleTabChange(session.sessionId)}
          on:dblclick={() => startEditTab(session.sessionId)}
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
            <span class="text-sm font-medium truncate flex-1">{session.tabName || session.connection.name}</span>
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
  </div>

  <!-- 终端内容 -->
  {#if sessionsList.length > 0}
    <div class="terminal-content-area relative flex-1 flex flex-col">
      {#each sessionsList as session (session.sessionId)}
        <div
          class="terminal-wrapper {
            $activeSessionIdStore === session.sessionId ? 'active' : 'inactive'
          }"
        >
          <!-- 工具栏 -->
          <div class="flex items-center justify-between px-4 py-2.5 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
            <div class="text-sm text-gray-600 dark:text-gray-400 font-mono">
              {session.connection.username}@{session.connection.host}:{session.connection.port}
            </div>
            <div class="flex items-center gap-1">
              <button class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors" title="复制">
                <svg class="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 01-2 2z" />
                </svg>
              </button>
              <button class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors" title="最小化">
                <svg class="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
                </svg>
              </button>
              <button class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors" title="最大化">
                <svg class="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 4l-5 5M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-6 0a2 2 0 00-2 2v4a2 2 0 002 2m0 0v4a2 2 0 002 2v-4a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 01-2 2z" />
                </svg>
              </button>
            </div>
          </div>

          <!-- 终端窗口 -->
          <div class="flex-1 overflow-hidden">
            <Terminal
              bind:this={terminalRefs[session.sessionId]}
              sessionId={session.sessionId}
              onData={handleTerminalData}
              onResize={handleTerminalResize}
            />
          </div>
        </div>
      {/each}
    </div>
  {:else}
    <div class="flex-1 flex items-center justify-center text-gray-500 dark:text-gray-400 bg-white dark:bg-gray-800">
      <div class="text-center">
        <div class="text-lg font-medium mb-2 text-gray-700 dark:text-gray-300">未选择连接</div>
        <div class="text-sm text-gray-500 dark:text-gray-400">从左侧资产列表选择一个服务器开始连接</div>
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

<style>
  .terminal-content-area {
    position: relative;
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
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
</style>
