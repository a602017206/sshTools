<script>
  import Terminal from './components/Terminal.svelte';
  import TabBar from './components/TabBar.svelte';
  import ConnectionManagerSimple from './components/ConnectionManagerSimple.svelte';
  import MonitorPanel from './components/MonitorPanel.svelte';
  import FileManager from './components/FileManager.svelte';
  import DevToolsPanel from './components/DevToolsPanel.svelte';
  import { onMount, onDestroy, tick } from 'svelte';
  import { ConnectSSH, SendSSHData, ResizeSSH, CloseSSH } from '../wailsjs/go/main/App.js';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import { showConfirm } from './utils/dialog.js';
  import { themeStore } from './stores/theme.js';
  import { fileManagerStore } from './stores/fileManager.js';
  import { devToolsStore } from './stores/devtools.js';

  let sessions = new Map(); // sessionId -> session metadata
  let activeSessionId = null;
  let tabOrder = []; // Array of sessionIds (maintains tab order)
  let terminalRefs = {}; // sessionId -> Terminal component ref
  let sessionUnsubscribers = new Map(); // sessionId -> event unsubscribe function

  // Sidebar dragging state
  let sidebarWidth = 300;
  let isDragging = false;
  let isSidebarCollapsed = false;
  let startX = 0;
  let startWidth = 0;

  // Subscribe to theme store
  themeStore.subscribe(state => {
    if (!isDragging) {
      sidebarWidth = state.sidebarWidth;
    }
  });

  function toggleSidebar() {
    isSidebarCollapsed = !isSidebarCollapsed;
  }

  function toggleTheme() {
    const newTheme = $themeStore.theme === 'light' ? 'dark' : 'light';
    themeStore.setTheme(newTheme);
  }

  function toggleDevTools() {
    devToolsStore.toggle();
  }

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
    // Initialize file manager from settings
    await fileManagerStore.init();
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
    <aside
      class="sidebar"
      class:collapsed={isSidebarCollapsed}
      class:dragging={isDragging}
      style="width: {isSidebarCollapsed ? 0 : sidebarWidth}px; min-width: {isSidebarCollapsed ? 0 : sidebarWidth}px;"
    >
      <div class="sidebar-content" style="width: {sidebarWidth}px">
        <ConnectionManagerSimple onConnect={handleConnect} on:collapse={toggleSidebar} />
      </div>
    </aside>

    {#if !isSidebarCollapsed}
      <div
        class="sidebar-resizer"
        class:dragging={isDragging}
        on:mousedown={handleDragStart}
      ></div>
    {/if}

    <div class="main-content">
      <!-- Tab bar (only show if sessions exist) -->
      <div class="tab-bar-container">
        {#if isSidebarCollapsed}
          <button class="expand-btn" on:click={toggleSidebar} title="展开侧边栏">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path fill-rule="evenodd" clip-rule="evenodd" d="M2 2.5H14V3.5H2V2.5ZM2 7.5H14V8.5H2V7.5ZM2 12.5H14V13.5H2V12.5ZM11 2.5V13.5H12V2.5H11Z"/>
            </svg>
          </button>
        {/if}
        {#if sessions.size > 0}
          <TabBar
            tabs={tabsList}
            activeTabId={activeSessionId}
            onTabChange={handleTabChange}
            onTabClose={handleTabClose}
            onTabRename={handleTabRename}
            onNewTab={handleNewTab}
          />
        {/if}
        <div class="header-spacer"></div>
        <button
          class="devtools-toggle-btn"
          class:active={$devToolsStore.isOpen}
          on:click={toggleDevTools}
          title="开发工具集"
        >
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path d="M8 4.754a3.246 3.246 0 1 0 0 6.492 3.246 3.246 0 0 0 0-6.492zM5.754 8a2.246 2.246 0 1 1 4.492 0 2.246 2.246 0 0 1-4.492 0z"/>
            <path d="M9.796 1.343c-.527-1.79-3.065-1.79-3.592 0l-.094.319a.873.873 0 0 1-1.255.52l-.292-.16c-1.64-.892-3.433.902-2.54 2.541l.159.292a.873.873 0 0 1-.52 1.255l-.319.094c-1.79.527-1.79 3.065 0 3.592l.319.094a.873.873 0 0 1 .52 1.255l-.16.292c-.892 1.64.901 3.434 2.541 2.54l.292-.159a.873.873 0 0 1 1.255.52l.094.319c.527 1.79 3.065 1.79 3.592 0l.094-.319a.873.873 0 0 1 1.255-.52l.292.16c1.64.893 3.434-.902 2.54-2.541l-.159-.292a.873.873 0 0 1 .52-1.255l.319-.094c1.79-.527 1.79-3.065 0-3.592l-.319-.094a.873.873 0 0 1-.52-1.255l.16-.292c.893-1.64-.902-3.433-2.541-2.54l-.292.159a.873.873 0 0 1-1.255-.52l-.094-.319z"/>
          </svg>
        </button>
        <button class="theme-toggle-btn" on:click={toggleTheme} title="切换主题">
          {#if $themeStore.theme === 'light'}
            <!-- Sun icon for light mode -->
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M8 12a4 4 0 1 1 0-8 4 4 0 0 1 0 8zm0-1.5a2.5 2.5 0 1 0 0-5 2.5 2.5 0 0 0 0 5zm5.657-8.157a.75.75 0 0 1 0 1.061l-1.061 1.06a.75.75 0 1 1-1.06-1.06l1.06-1.06a.75.75 0 0 1 1.06 0zm-9.193 9.193a.75.75 0 0 1 0 1.06l-1.06 1.061a.75.75 0 1 1-1.061-1.06l1.06-1.061a.75.75 0 0 1 1.061 0zM8 0a.75.75 0 0 1 .75.75v1.5a.75.75 0 0 1-1.5 0V.75A.75.75 0 0 1 8 0zM.75 8a.75.75 0 0 1 .75-.75h1.5a.75.75 0 0 1 0 1.5h-1.5A.75.75 0 0 1 .75 8zm12.25 0a.75.75 0 0 1 .75-.75h1.5a.75.75 0 0 1 0 1.5h-1.5a.75.75 0 0 1-.75-.75zM13.657 13.657a.75.75 0 0 1 0 1.061l-1.061 1.06a.75.75 0 1 1-1.06-1.06l1.06-1.06a.75.75 0 0 1 1.06 0zm-9.193-9.193a.75.75 0 0 1 0-1.06l-1.06-1.061a.75.75 0 1 1-1.061 1.06l1.06 1.061a.75.75 0 0 1 1.061 0z"/>
            </svg>
          {:else}
            <!-- Moon icon for dark mode -->
            <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
              <path d="M6 .278a.768.768 0 0 1 .08.858 7.208 7.208 0 0 0-.878 3.46c0 4.021 3.278 7.277 7.318 7.277.527 0 1.04-.055 1.533-.16a.787.787 0 0 1 .81.316.733.733 0 0 1-.031.893A8.349 8.349 0 0 1 8.344 16C3.734 16 0 12.286 0 7.71 0 4.266 2.114 1.312 5.124.06A.752.752 0 0 1 6 .278zM4.858 1.311A7.269 7.269 0 0 0 1.025 7.71c0 4.02 3.279 7.276 7.319 7.276a7.316 7.316 0 0 0 5.205-2.162c-.337.042-.68.063-1.029.063-4.61 0-8.343-3.714-8.343-8.29 0-1.167.242-2.278.681-3.286z"/>
            </svg>
          {/if}
        </button>
      </div>
      
      {#if sessions.size === 0 && isSidebarCollapsed}
        <button class="expand-btn floating" on:click={toggleSidebar} title="展开侧边栏">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <path fill-rule="evenodd" clip-rule="evenodd" d="M2 2.5H14V3.5H2V2.5ZM2 7.5H14V8.5H2V7.5ZM2 12.5H14V13.5H2V12.5ZM11 2.5V13.5H12V2.5H11Z"/>
          </svg>
        </button>
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

    <!-- File Manager Panel -->
    {#if activeSessionId}
      <FileManager activeSessionId={activeSessionId} />
    {/if}

    <!-- Monitor Panel -->
    {#if activeSessionId}
      <MonitorPanel activeSessionId={activeSessionId} />
    {/if}

    <!-- DevTools Panel -->
    <DevToolsPanel />
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    background-color: var(--bg-primary);
    color: var(--text-primary);
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

  /* Sidebar */
  .sidebar {
    background: var(--bg-secondary);
    border-right: 1px solid var(--border-primary);
    display: flex;
    flex-direction: column;
    position: relative;
    flex-shrink: 0;
    transition: width 0.2s cubic-bezier(0.16, 1, 0.3, 1), min-width 0.2s cubic-bezier(0.16, 1, 0.3, 1);
    z-index: 10;
  }

  .sidebar.dragging {
    transition: none;
    pointer-events: none; /* Prevent events during drag */
  }

  .sidebar.collapsed {
    border-right: 1px solid var(--border-primary); /* Keep border even when collapsed */
  }

  .sidebar-content {
    height: 100%;
    width: 100%;
    overflow: hidden;
  }

  /* Resizer */
  .sidebar-resizer {
    width: 12px; /* Touch target size */
    margin-left: -6px; /* Center over border */
    position: relative;
    z-index: 20;
    cursor: col-resize;
    flex-shrink: 0;
  }

  .sidebar-resizer::after {
    content: '';
    position: absolute;
    left: 6px;
    top: 0;
    bottom: 0;
    width: 1px;
    background: transparent;
    transition: background 0.2s;
  }

  .sidebar-resizer:hover::after,
  .sidebar-resizer.dragging::after {
    background: var(--accent-primary);
    width: 2px;
  }

  /* Main Content Area */
  .main-content {
    flex: 1;
    min-width: 0; /* Important for flex child to shrink */
    display: flex;
    flex-direction: column;
    overflow: hidden;
    background-color: var(--bg-primary);
    position: relative;
  }

  /* Header / Tab Bar Area */
  .tab-bar-container {
    height: 38px;
    background: var(--bg-secondary); /* Same as sidebar for unified header look */
    border-bottom: 1px solid var(--border-primary);
    display: flex;
    align-items: center;
    padding: 0 8px;
    flex-shrink: 0;
  }

  .header-spacer {
    flex: 1;
    -webkit-app-region: drag; /* Allow dragging window from empty header space */
    height: 100%;
  }

  /* Buttons */
  .expand-btn,
  .devtools-toggle-btn,
  .theme-toggle-btn {
    width: 28px;
    height: 28px;
    border: none;
    background: transparent;
    color: var(--text-secondary);
    border-radius: 4px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
    flex-shrink: 0;
  }

  .expand-btn:hover,
  .devtools-toggle-btn:hover,
  .theme-toggle-btn:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }

  .devtools-toggle-btn.active {
    color: var(--accent-primary);
    background: var(--bg-active);
  }

  .expand-btn.floating {
    position: absolute;
    top: 8px;
    left: 8px;
    z-index: 50;
    background: var(--bg-secondary);
    border: 1px solid var(--border-primary);
    box-shadow: var(--shadow-sm);
  }

  /* Terminal Area */
  .terminal-area {
    flex: 1;
    position: relative;
    overflow: hidden;
    background-color: var(--bg-primary);
  }

  .terminal-wrapper {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    padding: 0; /* Terminals usually need full bleeding edge */
  }

  /* Welcome Screen */
  .welcome {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -60%);
    text-align: center;
    color: var(--text-secondary);
  }

  .welcome h1 {
    font-size: 24px;
    font-weight: 500;
    color: var(--text-primary);
    margin-bottom: 12px;
    letter-spacing: -0.5px;
  }

  .welcome p {
    font-size: 14px;
    opacity: 0.8;
  }
</style>
