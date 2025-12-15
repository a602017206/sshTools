<script>
  import { onMount, onDestroy } from 'svelte';
  import { Terminal } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import { WebLinksAddon } from '@xterm/addon-web-links';
  import '@xterm/xterm/css/xterm.css';

  export let sessionId = null;
  export let onData = null;
  export let onResize = null;

  let terminalElement;
  let terminal;
  let fitAddon;

  onMount(() => {
    // Create terminal instance
    terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
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
      },
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
    background-color: #1e1e1e;
  }

  :global(.xterm) {
    height: 100%;
    padding: 10px;
  }

  :global(.xterm-viewport) {
    overflow-y: auto;
  }
</style>
