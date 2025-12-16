<script>
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { WebLinksAddon } from '@xterm/addon-web-links';
  import '@xterm/xterm/css/xterm.css';
  import { themeStore } from '../stores/theme.js';

  export let sessionId = null;
  export let onData = null;
  export let onResize = null;

  let terminalElement;
  let terminal;
  let fitAddon;
  let currentTheme = 'dark';

  // 深色主题配置
  const darkTheme = {
    background: '#1e1e1e',
    foreground: '#d4d4d4',
    cursor: '#d4d4d4',
    black: '#000000',
    red: '#cd3131',
    green: '#0dbc79',
    yellow: '#e5e510',
    blue: '#2472c8',
    magenta: '#bc3fbc',
    cyan: '#11a8cd',
    white: '#e5e5e5',
    brightBlack: '#666666',
    brightRed: '#f14c4c',
    brightGreen: '#23d18b',
    brightYellow: '#f5f543',
    brightBlue: '#3b8eea',
    brightMagenta: '#d670d6',
    brightCyan: '#29b8db',
    brightWhite: '#e5e5e5'
  };

  // 浅色主题配置
  const lightTheme = {
    background: '#ffffff',
    foreground: '#333333',
    cursor: '#333333',
    black: '#000000',
    red: '#cd3131',
    green: '#00bc00',
    yellow: '#949800',
    blue: '#0451a5',
    magenta: '#bc05bc',
    cyan: '#0598bc',
    white: '#555555',
    brightBlack: '#666666',
    brightRed: '#cd3131',
    brightGreen: '#14ce14',
    brightYellow: '#b5ba00',
    brightBlue: '#0451a5',
    brightMagenta: '#bc05bc',
    brightCyan: '#0598bc',
    brightWhite: '#a5a5a5'
  };

  // 订阅主题变化
  const unsubscribe = themeStore.subscribe(state => {
    currentTheme = state.theme;
    if (terminal) {
      // 动态更新终端主题
      terminal.options.theme = currentTheme === 'light' ? lightTheme : darkTheme;
    }
  });

  onMount(() => {
    // Create terminal instance with dynamic theme
    terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: currentTheme === 'light' ? lightTheme : darkTheme,
      allowProposedApi: true,
      scrollback: 1000,
      convertEol: true
    });

    // Add addons
    fitAddon = new FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.loadAddon(new WebLinksAddon());

    // Open terminal in DOM
    terminal.open(terminalElement);

    // Fit terminal to container
    fitAddon.fit();

    // Handle user input
    terminal.onData((data) => {
      if (onData && sessionId) {
        onData(sessionId, data);
      }
    });

    // Handle terminal resize
    terminal.onResize(({ cols, rows }) => {
      if (onResize && sessionId) {
        onResize(sessionId, cols, rows);
      }
    });

    // Handle window resize
    const resizeObserver = new ResizeObserver(() => {
      fitAddon.fit();
    });
    resizeObserver.observe(terminalElement);

    // Return cleanup function
    return () => {
      resizeObserver.disconnect();
    };
  });

  onDestroy(() => {
    // 清理订阅
    if (unsubscribe) {
      unsubscribe();
    }
    if (terminal) {
      terminal.dispose();
    }
  });

  // Expose methods for parent component
  export function write(data) {
    if (terminal) {
      terminal.write(data);
    }
  }

  export function writeln(data) {
    if (terminal) {
      terminal.writeln(data);
    }
  }

  export function clear() {
    if (terminal) {
      terminal.clear();
    }
  }

  export function focus() {
    if (terminal) {
      terminal.focus();
    }
  }

  export function getSize() {
    if (terminal) {
      return {
        cols: terminal.cols,
        rows: terminal.rows
      };
    }
    return { cols: 80, rows: 24 };
  }
</script>

<div class="terminal-container" bind:this={terminalElement}></div>

<style>
  .terminal-container {
    width: 100%;
    height: 100%;
    background-color: var(--bg-primary);
  }

  :global(.xterm) {
    height: 100%;
    padding: 10px;
    box-sizing: border-box;
  }

  :global(.xterm-viewport) {
    overflow-y: auto;
  }
</style>
