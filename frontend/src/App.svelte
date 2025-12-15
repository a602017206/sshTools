<script>
  import Terminal from './components/Terminal.svelte';
  import ConnectionManagerSimple from './components/ConnectionManagerSimple.svelte';
  import { onMount, onDestroy } from 'svelte';
  import { ConnectSSH, SendSSHData, ResizeSSH, CloseSSH } from '../wailsjs/go/main/App.js';
  import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime.js';

  let sessions = new Map(); // sessionId -> { terminal, connection }
  let activeSessionId = null;
  let terminalRef;
  let eventUnsubscribers = [];

  async function handleConnect(connection, authValue, passphrase = '') {
    const sessionId = `session_${Date.now()}`;

    console.log('Connecting to:', connection);

    // Create a new session
    sessions.set(sessionId, {
      connection,
      authValue,
      passphrase,
      connected: false
    });

    activeSessionId = sessionId;

    // Show connecting message
    if (terminalRef) {
      const authType = connection.auth_type === 'key' ? 'SSH key' : 'password';
      terminalRef.writeln(`Connecting to ${connection.user}@${connection.host}:${connection.port} using ${authType}...`);
      terminalRef.writeln('');
    }

    try {
      // Get terminal size
      const size = terminalRef ? terminalRef.getSize() : { cols: 80, rows: 24 };

      // Subscribe to SSH output events for this session
      const eventName = `ssh:output:${sessionId}`;
      const unsubscribe = EventsOn(eventName, (data) => {
        if (terminalRef && activeSessionId === sessionId) {
          terminalRef.write(data);
        }
      });
      eventUnsubscribers.push({ name: eventName, unsubscribe });

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
      const session = sessions.get(sessionId);
      session.connected = true;
      sessions.set(sessionId, session);

      console.log('SSH connection established:', sessionId);

    } catch (error) {
      console.error('Failed to connect:', error);
      if (terminalRef) {
        terminalRef.writeln(`\r\nConnection failed: ${error}`);
        terminalRef.writeln('');
      }

      // Clean up failed session
      sessions.delete(sessionId);
      if (activeSessionId === sessionId) {
        activeSessionId = null;
      }
    }
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

  async function handleCloseSession(sessionId) {
    try {
      await CloseSSH(sessionId);
    } catch (error) {
      console.error('Failed to close session:', error);
    }

    sessions.delete(sessionId);
    if (activeSessionId === sessionId) {
      activeSessionId = null;
    }
  }

  onMount(() => {
    // Focus terminal on mount
    if (terminalRef) {
      setTimeout(() => terminalRef.focus(), 100);
    }
  });

  onDestroy(() => {
    // Clean up event listeners
    eventUnsubscribers.forEach(({ name, unsubscribe }) => {
      unsubscribe();
    });

    // Close all sessions
    sessions.forEach((session, sessionId) => {
      handleCloseSession(sessionId);
    });
  });
</script>

<main>
  <div class="app-container">
    <aside class="sidebar">
      <ConnectionManagerSimple onConnect={handleConnect} />
    </aside>
    <div class="main-content">
      {#if activeSessionId}
        <div class="terminal-area">
          <Terminal
            bind:this={terminalRef}
            sessionId={activeSessionId}
            onData={handleTerminalData}
            onResize={handleTerminalResize}
          />
        </div>
      {:else}
        <div class="welcome">
          <h1>SSH Tools</h1>
          <p>选择一个连接开始使用</p>
        </div>
      {/if}
    </div>
  </div>
</main>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen',
      'Ubuntu', 'Cantarell', 'Fira Sans', 'Droid Sans', 'Helvetica Neue',
      sans-serif;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  main {
    width: 100vw;
    height: 100vh;
    overflow: hidden;
  }

  .app-container {
    display: flex;
    height: 100%;
    background-color: #1e1e1e;
  }

  .sidebar {
    width: 300px;
    min-width: 300px;
    border-right: 1px solid #3c3c3c;
    overflow-y: auto;
    -webkit-app-region: no-drag !important;
  }

  .main-content {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .terminal-area {
    flex: 1;
    overflow: hidden;
  }

  .welcome {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: flex-start;
    padding: 40px;
    color: #858585;
  }

  .welcome h1 {
    font-size: 32px;
    margin-bottom: 10px;
    color: #cccccc;
  }

  .welcome p {
    font-size: 16px;
  }
</style>
