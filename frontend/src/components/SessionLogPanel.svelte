<script>
  import { formatFileSize } from '../stores/uploadStore.js';

  export let connectionId = '';
  export let panelVisible = true;

  let logs = [];
  let hits = [];
  let query = '';
  let loading = false;
  let searching = false;
  let busy = false;
  let error = '';
  let status = '';
  let lastLoadedId = '';

  $: if (panelVisible && connectionId && connectionId !== lastLoadedId) {
    lastLoadedId = connectionId;
    loadLogs();
  }

  $: if (!connectionId) {
    logs = [];
    hits = [];
    query = '';
    error = '';
    status = '';
    lastLoadedId = '';
  }

  function api() {
    return window.wailsBindings || window.go?.main?.App || {};
  }

  function logId(log) {
    return log?.ID || log?.id || '';
  }

  function logSessionId(log) {
    return log?.SessionID || log?.sessionID || log?.session_id || '';
  }

  function logSize(log) {
    return Number(log?.Size ?? log?.size ?? 0);
  }

  function logModTime(log) {
    const raw = log?.ModTime || log?.modTime || log?.mod_time;
    if (!raw) return '';
    const date = raw instanceof Date ? raw : new Date(raw);
    if (Number.isNaN(date.getTime())) return String(raw);
    const pad = (n) => String(n).padStart(2, '0');
    return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}`;
  }

  function logLabel(log) {
    const id = logId(log);
    const base = id.includes('/') ? id.split('/').pop() : id;
    return base || logSessionId(log) || '未命名日志';
  }

  async function loadLogs() {
    if (!connectionId) {
      logs = [];
      return;
    }
    const { ListSessionLogs } = api();
    if (typeof ListSessionLogs !== 'function') {
      error = '会话日志接口不可用';
      return;
    }
    loading = true;
    error = '';
    status = '';
    try {
      const rows = await ListSessionLogs(connectionId);
      logs = Array.isArray(rows) ? rows : [];
    } catch (err) {
      console.error('ListSessionLogs failed:', err);
      error = '加载日志列表失败';
      logs = [];
    } finally {
      loading = false;
    }
  }

  async function handleSearch() {
    const q = String(query || '').trim();
    if (!connectionId) return;
    if (!q) {
      hits = [];
      status = '';
      await loadLogs();
      return;
    }
    const { SearchSessionLogs } = api();
    if (typeof SearchSessionLogs !== 'function') {
      error = '搜索接口不可用';
      return;
    }
    searching = true;
    error = '';
    status = '';
    try {
      const rows = await SearchSessionLogs(connectionId, q, 100);
      hits = Array.isArray(rows) ? rows : [];
      status = hits.length ? `找到 ${hits.length} 条匹配` : '无匹配结果';
    } catch (err) {
      console.error('SearchSessionLogs failed:', err);
      error = '搜索失败';
      hits = [];
    } finally {
      searching = false;
    }
  }

  async function handleExport(log) {
    const id = logId(log);
    if (!id || busy) return;
    const { ExportSessionLog } = api();
    if (typeof ExportSessionLog !== 'function') {
      error = '导出接口不可用';
      return;
    }
    busy = true;
    error = '';
    status = '';
    try {
      const path = await ExportSessionLog(id);
      status = path ? `已导出到 ${path}` : '已取消导出';
    } catch (err) {
      console.error('ExportSessionLog failed:', err);
      error = '导出失败';
    } finally {
      busy = false;
    }
  }

  async function handleDelete(log) {
    const id = logId(log);
    if (!id || busy) return;
    if (typeof window !== 'undefined' && !window.confirm('确定删除该会话日志？')) {
      return;
    }
    const { DeleteSessionLog } = api();
    if (typeof DeleteSessionLog !== 'function') {
      error = '删除接口不可用';
      return;
    }
    busy = true;
    error = '';
    status = '';
    try {
      await DeleteSessionLog(id);
      status = '已删除';
      hits = hits.filter((hit) => (hit?.LogID || hit?.logID || hit?.log_id) !== id);
      await loadLogs();
    } catch (err) {
      console.error('DeleteSessionLog failed:', err);
      error = '删除失败';
    } finally {
      busy = false;
    }
  }

  async function handlePurge() {
    if (busy) return;
    if (typeof window !== 'undefined' && !window.confirm('清理超过保留天数的会话日志？')) {
      return;
    }
    const { PurgeExpiredSessionLogs } = api();
    if (typeof PurgeExpiredSessionLogs !== 'function') {
      error = '清理接口不可用';
      return;
    }
    busy = true;
    error = '';
    status = '';
    try {
      const removed = await PurgeExpiredSessionLogs();
      status = `已清理 ${Number(removed) || 0} 个过期日志`;
      await loadLogs();
    } catch (err) {
      console.error('PurgeExpiredSessionLogs failed:', err);
      error = '清理失败';
    } finally {
      busy = false;
    }
  }

  function hitLogId(hit) {
    return hit?.LogID || hit?.logID || hit?.log_id || '';
  }

  function hitLine(hit) {
    return hit?.Line ?? hit?.line ?? '';
  }

  function hitText(hit) {
    return hit?.Text || hit?.text || '';
  }
</script>

<section class="session-log-panel" aria-label="会话日志">
  <header class="toolbar">
    <div class="search-row">
      <input
        type="search"
        bind:value={query}
        placeholder="搜索日志内容…"
        disabled={!connectionId || searching || busy}
        on:keydown={(e) => e.key === 'Enter' && handleSearch()}
      />
      <button type="button" disabled={!connectionId || searching || busy} on:click={handleSearch}>
        {searching ? '搜索中…' : '搜索'}
      </button>
    </div>
    <div class="actions">
      <button type="button" disabled={!connectionId || loading || busy} on:click={loadLogs}>刷新</button>
      <button type="button" disabled={busy} on:click={handlePurge}>清理过期</button>
    </div>
  </header>

  {#if error}
    <p class="msg error" role="alert">{error}</p>
  {:else if status}
    <p class="msg status" role="status">{status}</p>
  {/if}

  {#if !connectionId}
    <div class="empty">当前会话未绑定连接</div>
  {:else if loading}
    <div class="empty">加载中…</div>
  {:else if hits.length > 0}
    <ul class="hit-list">
      {#each hits as hit (hitLogId(hit) + ':' + hitLine(hit) + ':' + hitText(hit))}
        <li>
          <div class="hit-meta">
            <span class="mono">{hitLogId(hit).split('/').pop()}</span>
            <span>L{hitLine(hit)}</span>
          </div>
          <pre>{hitText(hit)}</pre>
        </li>
      {/each}
    </ul>
  {:else if logs.length === 0}
    <div class="empty">暂无会话日志</div>
  {:else}
    <ul class="log-list">
      {#each logs as log (logId(log))}
        <li>
          <div class="log-main">
            <strong title={logId(log)}>{logLabel(log)}</strong>
            <span class="meta">{logModTime(log)} · {formatFileSize(logSize(log))}</span>
          </div>
          <div class="log-actions">
            <button type="button" disabled={busy} on:click={() => handleExport(log)}>导出</button>
            <button type="button" class="danger" disabled={busy} on:click={() => handleDelete(log)}>删除</button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .session-log-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    padding: 10px 12px;
    gap: 8px;
    color: var(--text-primary);
  }

  .toolbar {
    display: grid;
    gap: 8px;
    flex-shrink: 0;
  }

  .search-row {
    display: flex;
    gap: 6px;
  }

  .search-row input {
    flex: 1;
    min-width: 0;
    min-height: 32px;
    padding: 0 10px;
    border-radius: 8px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    color: var(--text-primary);
    font: inherit;
    font-size: 12px;
  }

  .actions {
    display: flex;
    gap: 6px;
  }

  button {
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-size: 12px;
    font-weight: 600;
    min-height: 30px;
    padding: 0 10px;
    border-radius: 8px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg-strong);
    color: var(--text-primary);
  }

  button:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  button:hover:not(:disabled) {
    border-color: var(--ops-signal);
    color: var(--ops-signal);
  }

  button.danger:hover:not(:disabled) {
    border-color: var(--ops-alert, #ef4444);
    color: var(--ops-alert, #ef4444);
  }

  .msg {
    margin: 0;
    font-size: 11px;
  }

  .msg.error {
    color: var(--ops-alert, #ef4444);
  }

  .msg.status {
    color: var(--text-secondary);
  }

  .empty {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-tertiary);
    font-size: 12px;
  }

  .log-list,
  .hit-list {
    list-style: none;
    margin: 0;
    padding: 0;
    overflow: auto;
    flex: 1;
    min-height: 0;
    display: grid;
    gap: 6px;
  }

  .log-list li,
  .hit-list li {
    border: 1px solid var(--glass-border);
    border-radius: 10px;
    background: color-mix(in srgb, var(--glass-bg) 80%, transparent);
    padding: 8px 10px;
  }

  .log-list li {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px;
  }

  .log-main {
    min-width: 0;
    display: grid;
    gap: 2px;
  }

  .log-main strong {
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .meta {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .log-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  .log-actions button {
    min-height: 26px;
    padding: 0 8px;
  }

  .hit-meta {
    display: flex;
    justify-content: space-between;
    gap: 8px;
    font-size: 11px;
    color: var(--text-secondary);
    margin-bottom: 4px;
  }

  .mono {
    font-family: var(--terminal-font-family);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .hit-list pre {
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
    font-family: var(--terminal-font-family);
    font-size: 11px;
    color: var(--text-primary);
  }
</style>
