<script>
  import Terminal from './components/Terminal.svelte';
  import TabBar from './components/TabBar.svelte';
  import ConnectionManagerSimple from './components/ConnectionManagerSimple.svelte';
  import MonitorPanel from './components/MonitorPanel.svelte';
  import { onMount, onDestroy, tick } from 'svelte';
  import { ConnectSSH, SendSSHData, ResizeSSH, CloseSSH } from '../wailsjs/go/main/App.js';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import { showConfirm } from './utils/dialog.js';
  import { themeStore } from './stores/theme.js';

  let sessions = new Map(); // sessionId -> session metadata
  let activeSessionId = null;
  let tabOrder = []; // Array of sessionIds (maintains tab order)
  let terminalRefs = {}; // sessionId -> Terminal component ref
  let sessionUnsubscribers = new Map(); // sessionId -> event unsubscribe function

  // Sidebar dragging state
  let sidebarWidth = 300;
  let isDragging = false;
  let startX = 0;
  let startWidth = 0;

  // Subscribe to theme store
  themeStore.subscribe(state => {
    if (!isDragging) {
      sidebarWidth = state.sidebarWidth;
    }
  });

  // Reactive declarations
  $: activeSession = sessions.get(activeSessionId);
  $: tabsList = Array.from(sessions.values()).map(session => ({
    id: session.sessionId,
    name: session.tabName || session.connection.name,
    displayName: session.tabName || `${session.connection.user}@${session.connection.host}`,
    connectionName: session.connection.name,
    userAtHost: `${session.connection.user}@${session.connection.host}:${session.connection.port}`,
    isActive: session.sessionId === activeSessionId
  }));

  async function handleConnect(connection, authValue, passphrase = '') {
    // Generate unique session ID
    const sessionId = `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

    console.log('Connecting to:', connection);

    // Create session metadata
    const newSession = {
      sessionId,
      connection,
      authValue,
      passphrase,
      tabName: '', // Empty = use connection.name
      connected: false,
      createdAt: Date.now(),
      lastActivity: Date.now()
    };

    // Add to sessions and tab order
    sessions.set(sessionId, newSession);
    tabOrder.push(sessionId);
    tabOrder = tabOrder; // Trigger reactivity

    // Set as active
    activeSessionId = sessionId;

    // Wait for Terminal to mount, then get size
    await tick();
    const size = terminalRefs[sessionId]?.getSize() || { cols: 80, rows: 24 };

    // Subscribe to output events
    const eventName = `ssh:output:${sessionId}`;
    const unsubscribe = EventsOn(eventName, (data) => {
      const terminal = terminalRefs[sessionId];
      if (terminal) {
        terminal.write(data);
      }
    });
    sessionUnsubscribers.set(sessionId, unsubscribe);

    // Show connecting message
    const terminal = terminalRefs[sessionId];
    if (terminal) {
      const authType = connection.auth_type === 'key' ? 'SSH key' : 'password';
      terminal.writeln(`正在连接 ${connection.user}@${connection.host}:${connection.port} (${authType})...`);
      terminal.writeln('');
    }

    try {
      // Connect to SSH
      await ConnectSSH(
        sessionId,
        connection.host,
        connection.port,
        connection.user,
        connection.auth_type || 'password',
        authValue,
        passphrase,
        size.cols,
        size.rows
      );

      // Mark as connected
      newSession.connected = true;
      sessions.set(sessionId, newSession);

      console.log('SSH connection established:', sessionId);

    } catch (error) {
      console.error('Failed to connect:', error);
      if (terminal) {
        terminal.writeln(`\r\n连接失败: ${error}`);
      }

      // Clean up failed session
      await closeSession(sessionId);
    }
  }

  async function closeSession(sessionId) {
    // Unsubscribe from events
    const unsubscribe = sessionUnsubscribers.get(sessionId);
    if (unsubscribe) {
      unsubscribe();
      sessionUnsubscribers.delete(sessionId);
    }

    // Dispose terminal (component's onDestroy handles cleanup)
    delete terminalRefs[sessionId];

    // Close backend session
    try {
      await CloseSSH(sessionId);
    } catch (error) {
      console.error('Failed to close session:', error);
    }

    // Remove from state
    sessions.delete(sessionId);
    tabOrder = tabOrder.filter(id => id !== sessionId);

    // Switch to another tab or show welcome
    if (activeSessionId === sessionId) {
      if (tabOrder.length > 0) {
        // Activate first remaining tab
        activeSessionId = tabOrder[0];

        // Focus terminal after switch
        setTimeout(() => {
          const terminal = terminalRefs[activeSessionId];
          if (terminal) {
            terminal.focus();
          }
        }, 50);
      } else {
        activeSessionId = null;
      }
    }
  }

  function handleTabChange(sessionId) {
    if (!sessions.has(sessionId)) return;

    activeSessionId = sessionId;

    // Focus terminal after render
    setTimeout(() => {
      const terminal = terminalRefs[sessionId];
      if (terminal) {
        terminal.focus();

        // Sync terminal size with backend
        const size = terminal.getSize();
        ResizeSSH(sessionId, size.cols, size.rows).catch(console.error);
      }
    }, 50);
  }

  async function handleTabClose(sessionId) {
    const session = sessions.get(sessionId);
    if (!session) return;

    // Confirm if session is connected
    if (session.connected) {
      const confirmed = await showConfirm('确定关闭此 SSH 会话吗？');
      if (!confirmed) return;
    }

    await closeSession(sessionId);
  }

  function handleTabRename(sessionId, newName) {
    const session = sessions.get(sessionId);
    if (!session) return;

    session.tabName = newName.trim();
    sessions.set(sessionId, session);
    sessions = sessions; // Trigger reactivity
  }

  function handleNewTab() {
    // User can select a connection from sidebar to create new tab
    console.log('New tab clicked - select a connection from sidebar');
  }

  async function handleTerminalData(sessionId, data) {
    if (!sessions.has(sessionId)) {
      return;
    }

    try {
      await SendSSHData(sessionId, data);
    } catch (error) {
      console.error('Failed to send data:', error);
    }
  }

  async function handleTerminalResize(sessionId, cols, rows) {
    if (!sessions.has(sessionId)) {
      return;
    }

    try {
      await ResizeSSH(sessionId, cols, rows);
    } catch (error) {
      console.error('Failed to resize terminal:', error);
    }
  }

  onMount(async () => {
    // Initialize theme from settings
    await themeStore.init();
  });

  onDestroy(() => {
    // Unsubscribe from all events
    sessionUnsubscribers.forEach((unsubscribe) => {
      unsubscribe();
    });

    // Close all sessions
    const sessionIds = Array.from(sessions.keys());
    sessionIds.forEach((sessionId) => {
      CloseSSH(sessionId).catch(console.error);
    });
  });

  // Sidebar dragging handlers
  function handleDragStart(event) {
    isDragging = true;
    startX = event.clientX;
    startWidth = sidebarWidth;

    document.addEventListener('mousemove', handleDragMove);
    document.addEventListener('mouseup', handleDragEnd);

    document.body.style.userSelect = 'none';
    document.body.style.cursor = 'col-resize';
  }

  function handleDragMove(event) {
    if (!isDragging) return;

    const delta = event.clientX - startX;
    const newWidth = Math.max(200, Math.min(600, startWidth + delta));
    sidebarWidth = newWidth;
  }

  async function handleDragEnd() {
    if (!isDragging) return;

    isDragging = false;

    document.removeEventListener('mousemove', handleDragMove);
    document.removeEventListener('mouseup', handleDragEnd);

    document.body.style.userSelect = '';
    document.body.style.cursor = '';

    try {
      await themeStore.setSidebarWidth(sidebarWidth);
    } catch (error) {
      console.error('Failed to save sidebar width:', error);
    }
  }
</script>

<main>
  <div class="app-container">
    <aside class="sidebar" style="width: {sidebarWidth}px; min-width: {sidebarWidth}px;">
      <ConnectionManagerSimple onConnect={handleConnect} />
    </aside>

    <div
      class="sidebar-resizer"
      class:dragging={isDragging}
      on:mousedown={handleDragStart}
    ></div>

    <div class="main-content">
      <!-- Tab bar (only show if sessions exist) -->
      {#if sessions.size > 0}
        <div class="tab-bar-container">
          <TabBar
            tabs={tabsList}
            activeTabId={activeSessionId}
            onTabChange={handleTabChange}
            onTabClose={handleTabClose}
            onTabRename={handleTabRename}
            onNewTab={handleNewTab}
          />
        </div>
      {/if}

      <!-- Terminal area with multiple instances -->
      <div class="terminal-area">
        {#if activeSessionId}
          {#each tabOrder as sessionId (sessionId)}
            {#if sessionId === activeSessionId}
              <div class="terminal-wrapper">
                <Terminal
                  bind:this={terminalRefs[sessionId]}
                  sessionId={sessionId}
                  onData={handleTerminalData}
                  onResize={handleTerminalResize}
                />
              </div>
            {/if}
          {/each}
        {:else}
          <div class="welcome">
            <h1>SSH Tools</h1>
            <p>选择一个连接开始使用</p>
          </div>
        {/if}
      </div>
    </div>

    <!-- Monitor Panel -->
    {#if activeSessionId}
      <MonitorPanel activeSessionId={activeSessionId} />
    {/if}
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
  }

  main {
    width: 100vw;
    height: 100vh;
    overflow: hidden;
  }

  .app-container {
    display: flex;
    height: 100%;
    background-color: var(--bg-primary);
  }

  .sidebar {
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-primary);
    overflow-y: auto;
    -webkit-app-region: no-drag !important;
    transition: none;
  }

  .sidebar-resizer {
    width: 4px;
    background: transparent;
    cursor: col-resize;
    flex-shrink: 0;
    position: relative;
    -webkit-app-region: no-drag;
  }

  .sidebar-resizer:hover,
  .sidebar-resizer.dragging {
    background: var(--accent-primary);
  }

  .sidebar-resizer::before {
    content: '';
    position: absolute;
    top: 0;
    left: -2px;
    right: -2px;
    bottom: 0;
  }

  .main-content {
    flex: 1;
    min-width: 0; /* Allow shrinking */
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .tab-bar-container {
    height: 40px;
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border-primary);
    overflow: hidden;
  }

  .terminal-area {
    flex: 1;
    overflow: hidden;
    position: relative;
  }

  .terminal-wrapper {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
  }

  .welcome {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: flex-start;
    padding: 40px;
    color: var(--text-secondary);
  }

  .welcome h1 {
    font-size: 32px;
    margin-bottom: 10px;
    color: var(--text-primary);
  }

  .welcome p {
    font-size: 16px;
  }
</style>
