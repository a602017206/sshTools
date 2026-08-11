<script>
  import FileManager from './FileManager.svelte';
  import ServerMonitor from './ServerMonitor.svelte';
  import { SSH_TOOL_TABS, resolveSshToolTab } from '../lib/workspaceTabs.js';

  export let activeTab = 'files';
  export let boundSessionName = '';
  export let hasBoundSession = false;
  export let onSelectTab = () => {};
  export let onConnectHint = () => {};

  $: toolTab = resolveSshToolTab(activeTab);

  function selectTab(id) {
    onSelectTab(resolveSshToolTab(id));
  }
</script>

<aside class="session-tool-dock" aria-label="会话工具">
  <header class="dock-header">
    <div>
      <p class="dock-kicker">会话工具</p>
      <h2>{hasBoundSession ? `绑定 · ${boundSessionName}` : '未绑定会话'}</h2>
    </div>
    <div class="dock-tabs" role="tablist" aria-label="工具分段">
      {#each SSH_TOOL_TABS as tab}
        <button
          type="button"
          role="tab"
          class:active={toolTab === tab.id}
          aria-selected={toolTab === tab.id}
          on:click={() => selectTab(tab.id)}
        >
          {tab.label}
        </button>
      {/each}
    </div>
  </header>

  <div class="dock-body">
    {#if !hasBoundSession}
      <div class="dock-empty" role="status">
        <strong>先连接一台主机</strong>
        <span>文件与性能都会跟随当前 SSH 会话。</span>
        <button type="button" on:click={onConnectHint}>打开资源树</button>
      </div>
    {:else if toolTab === 'performance'}
      <div class="dock-panel">
        <ServerMonitor />
      </div>
    {:else}
      <div class="dock-panel">
        <FileManager />
      </div>
    {/if}
  </div>
</aside>

<style>
  .session-tool-dock {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    background: transparent;
    color: var(--text-primary);
  }

  .dock-header {
    flex-shrink: 0;
    padding: 12px 14px 10px;
    border-bottom: 1px solid var(--glass-border);
    display: grid;
    gap: 10px;
    background: color-mix(in srgb, var(--glass-bg) 70%, transparent);
  }

  .dock-kicker {
    margin: 0;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--ops-signal);
    font-family: var(--terminal-font-family);
  }

  h2 {
    margin: 4px 0 0;
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .dock-tabs {
    display: inline-flex;
    gap: 2px;
    padding: 2px;
    border-radius: 999px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    width: fit-content;
  }

  .dock-tabs button {
    appearance: none;
    border: 0;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    font: inherit;
    font-size: 12px;
    font-weight: 600;
    min-height: 28px;
    padding: 0 12px;
    border-radius: 999px;
    transition: background-color var(--trans-fast), color var(--trans-fast);
  }

  .dock-tabs button.active {
    color: var(--ops-signal);
    background: color-mix(in srgb, var(--glass-bg-strong) 80%, var(--accent-subtle));
  }

  .dock-tabs button:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 1px;
  }

  .dock-body {
    flex: 1;
    min-height: 0;
    overflow: hidden;
  }

  .dock-panel {
    height: 100%;
    min-height: 0;
  }

  .dock-empty {
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 24px;
    text-align: center;
    color: var(--text-secondary);
  }

  .dock-empty span {
    font-size: 12px;
    color: var(--text-tertiary);
    max-width: 220px;
  }

  .dock-empty button {
    margin-top: 6px;
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-size: 12px;
    font-weight: 600;
    min-height: 36px;
    padding: 0 14px;
    border-radius: 10px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg-strong);
    color: var(--text-primary);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    transition: border-color var(--trans-fast), color var(--trans-fast), transform var(--trans-fast);
  }

  .dock-empty button:hover {
    border-color: var(--ops-signal);
    color: var(--ops-signal);
  }

  .dock-empty button:active {
    transform: scale(0.98);
  }
</style>
