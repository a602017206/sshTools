<script>
  import { onMount } from 'svelte';
  import {
    containerField,
    dockerTransportLabel,
    formatContainerCreated,
    isContainerRunning,
    usedCliFallback,
    visibleContainers
  } from '../lib/dockerWorkspace.js';

  export let sessionId = '';
  export let connection = null;
  /** (connected: boolean, error: string) => void，供 TerminalPanel 同步页签状态 */
  export let onStatus = () => {};

  let status = 'idle'; // idle | connecting | connected | error
  let errorMessage = '';
  let loading = false;
  let onlyRunning = false;
  let containers = [];
  let destroyed = false;

  $: transportLabel = dockerTransportLabel(connection);
  $: rows = visibleContainers(containers, onlyRunning);
  $: cliFallback = usedCliFallback(containers);
  $: connected = status === 'connected';

  function bindings() {
    return window.wailsBindings || {};
  }

  function errText(error) {
    return String(error?.message || error || '未知错误');
  }

  async function connect() {
    const { ConnectDocker } = bindings();
    if (typeof ConnectDocker !== 'function') {
      status = 'error';
      errorMessage = 'Docker 连接功能不可用';
      onStatus(false, errorMessage);
      return;
    }
    status = 'connecting';
    errorMessage = '';
    try {
      await ConnectDocker(sessionId, connection);
      if (destroyed) return;
      status = 'connected';
      onStatus(true, '');
      await refresh();
    } catch (error) {
      if (destroyed) return;
      status = 'error';
      errorMessage = errText(error);
      containers = [];
      onStatus(false, errorMessage);
    }
  }

  async function refresh() {
    const { ListDockerContainers } = bindings();
    if (!connected || typeof ListDockerContainers !== 'function') return;
    loading = true;
    try {
      // all=true 取全部容器，「仅运行中」在前端筛选，切换无需再请求
      const list = await ListDockerContainers(sessionId, true);
      if (destroyed) return;
      containers = Array.isArray(list) ? list : [];
      errorMessage = '';
    } catch (error) {
      if (destroyed) return;
      errorMessage = errText(error);
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    connect();
    return () => {
      destroyed = true;
    };
  });
</script>

<div class="docker-workspace">
  <div class="docker-toolbar">
    <div class="docker-title">
      <span class="docker-name">{connection?.name || 'Docker'}</span>
      <span class="docker-badge">{transportLabel}</span>
      {#if cliFallback}
        <span class="docker-hint">CLI 回退</span>
      {/if}
    </div>
    <div class="docker-actions">
      <label class="docker-filter">
        <input type="checkbox" bind:checked={onlyRunning} />
        仅运行中
      </label>
      <button
        type="button"
        class="docker-btn"
        disabled={status === 'connecting' || loading}
        on:click={() => (connected ? refresh() : connect())}
      >
        {connected ? (loading ? '刷新中…' : '刷新') : '重新连接'}
      </button>
    </div>
  </div>

  <div class="docker-body">
    {#if status === 'connecting' || status === 'idle'}
      <div class="docker-empty" role="status">
        <p class="kicker">容器</p>
        <h2>正在连接 Docker…</h2>
      </div>
    {:else if status === 'error'}
      <div class="docker-empty" role="alert">
        <p class="kicker">容器</p>
        <h2>未连接到 Docker</h2>
        <p class="detail">{errorMessage}</p>
      </div>
    {:else if rows.length === 0}
      <div class="docker-empty" role="status">
        <p class="kicker">容器</p>
        <h2>{containers.length === 0 ? '没有容器' : '没有运行中的容器'}</h2>
        <p class="detail">
          {containers.length === 0 ? '该 Docker 引擎上还没有容器。' : '取消「仅运行中」可查看已停止的容器。'}
        </p>
        {#if errorMessage}<p class="detail docker-error">{errorMessage}</p>{/if}
      </div>
    {:else}
      {#if errorMessage}<div class="docker-banner docker-error">{errorMessage}</div>{/if}
      <div class="docker-table-wrap">
        <table class="docker-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>镜像</th>
              <th>状态</th>
              <th>端口</th>
              <th>ID</th>
              <th>创建时间</th>
            </tr>
          </thead>
          <tbody>
            {#each rows as container (containerField(container, 'ID') || containerField(container, 'Name'))}
              <tr>
                <td class="docker-cell-name">
                  <span class="ops-status-dot docker-dot" class:docker-dot--up={isContainerRunning(container)}></span>
                  {containerField(container, 'Name') || '-'}
                </td>
                <td>{containerField(container, 'Image') || '-'}</td>
                <td>{containerField(container, 'Status') || '-'}</td>
                <td class="docker-mono">{containerField(container, 'Ports') || '-'}</td>
                <td class="docker-mono">{String(containerField(container, 'ID')).slice(0, 12) || '-'}</td>
                <td>{formatContainerCreated(containerField(container, 'Created'))}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</div>

<style>
  .docker-workspace {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
    color: var(--text-primary);
  }

  .docker-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 14px;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-title {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .docker-name {
    font-size: 14px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .docker-badge {
    flex-shrink: 0;
    padding: 1px 8px;
    border-radius: 999px;
    border: 1px solid var(--glass-border);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.06em;
    color: var(--ops-signal);
    font-family: var(--terminal-font-family);
  }

  .docker-hint {
    font-size: 10px;
    color: var(--text-secondary);
  }

  .docker-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
  }

  .docker-filter {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--text-secondary);
    cursor: pointer;
  }

  .docker-btn {
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-size: 12px;
    font-weight: 600;
    min-height: 30px;
    padding: 0 14px;
    border-radius: 10px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    color: var(--text-primary);
    box-shadow: var(--shadow-glass);
    transition: border-color var(--trans-fast), color var(--trans-fast);
  }

  .docker-btn:hover:not(:disabled) {
    border-color: var(--ops-signal);
    color: var(--ops-signal);
  }

  .docker-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .docker-btn:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 2px;
  }

  .docker-body {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .docker-empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 32px;
    text-align: center;
  }

  .kicker {
    margin: 0;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--ops-signal);
    font-family: var(--terminal-font-family);
  }

  h2 {
    margin: 0;
    font-size: 20px;
    letter-spacing: -0.03em;
  }

  .detail {
    margin: 0;
    max-width: 480px;
    font-size: 13px;
    color: var(--text-secondary);
    word-break: break-word;
  }

  .docker-error {
    color: var(--ops-danger, #ef4444);
  }

  .docker-banner {
    padding: 6px 14px;
    font-size: 12px;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-table-wrap {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }

  .docker-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
  }

  .docker-table th {
    position: sticky;
    top: 0;
    z-index: 1;
    text-align: left;
    padding: 8px 14px;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
    color: var(--text-secondary);
    background: var(--glass-bg-strong);
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-table td {
    padding: 7px 14px;
    border-bottom: 1px solid var(--glass-border);
    max-width: 280px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .docker-table tbody tr:hover {
    background: var(--glass-bg);
  }

  .docker-cell-name {
    font-weight: 600;
  }

  .docker-dot {
    display: inline-block;
    margin-right: 8px;
    background: var(--text-secondary);
    opacity: 0.5;
  }

  .docker-dot--up {
    background: #22c55e;
    opacity: 1;
  }

  .docker-mono {
    font-family: var(--terminal-font-family);
  }
</style>
