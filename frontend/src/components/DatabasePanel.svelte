<script>
  import { onMount, onDestroy } from 'svelte';

  export let sessionId = null;
  export let dbConfig = null;

  let query = '';
  let resultData = null;
  let tables = [];
  let isLoading = false;
  let errorMessage = '';
  let queryHistory = [];
  let dbTypeLabel = '';

  const historyLimit = 50;

  $: dbTypeLabel = dbConfig?.metadata?.db_type ? dbConfig.metadata.db_type.toUpperCase() : '';

  onMount(async () => {
    if (!sessionId) return;
    await loadTables();

    const handleTableSelect = (event) => {
      if (!event?.detail || event.detail.sessionId !== sessionId) return;
      const { databaseName, tableName } = event.detail;
      if (!tableName) return;
      const qualifiedName = databaseName ? `${databaseName}.${tableName}` : tableName;
      query = `SELECT * FROM ${qualifiedName} LIMIT 10;`;
    };

    window.addEventListener('database:table-select', handleTableSelect);

    return () => {
      window.removeEventListener('database:table-select', handleTableSelect);
    };
  });

  onDestroy(() => {
    resultData = null;
  });

  async function executeQuery() {
    if (!query.trim()) return;
    if (!window.wailsBindings || !sessionId) return;

    isLoading = true;
    errorMessage = '';

    try {
      const result = await window.wailsBindings.ExecuteDatabaseQuery(sessionId, query);
      const data = JSON.parse(result);
      resultData = data;
      addToHistory(query.trim());
    } catch (error) {
      console.error('Query execution failed:', error);
      errorMessage = `查询执行失败: ${error.message || '未知错误'}`;
    } finally {
      isLoading = false;
    }
  }

  async function loadTables() {
    if (!window.wailsBindings || !sessionId) return;

    isLoading = true;
    errorMessage = '';

    try {
      const result = await window.wailsBindings.ListDatabaseTables(sessionId);
      tables = (result || []).slice().sort();
    } catch (error) {
      console.error('Failed to load tables:', error);
      errorMessage = `加载数据库表失败: ${error.message || '未知错误'}`;
    } finally {
      isLoading = false;
    }
  }

  function addToHistory(statement) {
    const normalized = statement.trim();
    if (!normalized) return;
    queryHistory = [normalized, ...queryHistory.filter(item => item !== normalized)].slice(0, historyLimit);
  }

  function clearQuery() {
    query = '';
    resultData = null;
  }

  function handleTableClick(table) {
    query = `SELECT * FROM ${table} LIMIT 10;`;
  }

  function handleTableDoubleClick(table) {
    query = `SELECT * FROM ${table} LIMIT 10;`;
    executeQuery();
  }

  function exportResults() {
    if (!resultData || !resultData.columns || !resultData.rows) return;
    const csv = [resultData.columns.join(',')].concat(
      resultData.rows.map(row => row.map(cell => `${cell}`).join(','))
    );
    downloadCSV(csv.join('\n'), 'query-results.csv');
  }

  function downloadCSV(content, filename) {
    const blob = new Blob([content], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }

  function handleHistoryClick(statement) {
    query = statement;
  }
</script>

<div class="db-panel">
  <aside class="db-sidebar">
    <div class="sidebar-section">
      <div class="section-header">
        <span>表列表</span>
        <button class="icon-btn" on:click={loadTables} title="刷新">↻</button>
      </div>
      <div class="section-body">
        {#if tables.length === 0}
          <div class="empty-text">暂无表</div>
        {:else}
          {#each tables as table}
            <button
              type="button"
              class="table-item"
              on:click={() => handleTableClick(table)}
              on:dblclick={() => handleTableDoubleClick(table)}
            >
              {table}
            </button>
          {/each}
        {/if}
      </div>
    </div>

    <div class="sidebar-section">
      <div class="section-header">
        <span>查询历史</span>
      </div>
      <div class="section-body">
        {#if queryHistory.length === 0}
          <div class="empty-text">暂无历史</div>
        {:else}
          {#each queryHistory as item}
            <button type="button" class="history-item" on:click={() => handleHistoryClick(item)}>
              {item}
            </button>
          {/each}
        {/if}
      </div>
    </div>
  </aside>

  <section class="db-main">
    <div class="query-toolbar">
      <button class="toolbar-btn btn-primary" on:click={executeQuery} disabled={isLoading}>
        {#if isLoading}
          <span class="loading-spinner"></span>
        {:else}
          ▶ 执行
        {/if}
      </button>
      <button class="toolbar-btn btn-secondary" on:click={clearQuery}>✖ 清空</button>
      {#if resultData}
        <button class="toolbar-btn btn-secondary" on:click={exportResults}>📥 导出CSV</button>
      {/if}
    </div>

    <textarea
      class="query-textarea"
      bind:value={query}
      placeholder={dbTypeLabel ? `在此输入 SQL 查询语句 (${dbTypeLabel})...` : '在此输入 SQL 查询语句...'}
      on:keydown={(e) => {
        if (e.ctrlKey && e.key === 'Enter') {
          e.preventDefault();
          executeQuery();
        }
      }}
    ></textarea>

    {#if errorMessage}
      <div class="error-message">{errorMessage}</div>
    {/if}

    <div class="results-wrapper">
      {#if resultData && resultData.columns && resultData.columns.length > 0}
        <div class="results-header">查询结果</div>
        <div class="table-container">
          <table>
            <thead>
              <tr>
                {#each resultData.columns as column}
                  <th>{column}</th>
                {/each}
              </tr>
            </thead>
            <tbody>
              {#each resultData.rows as row}
                <tr>
                  {#each row as cell}
                    <td>{cell}</td>
                  {/each}
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
        <div class="status-bar">
          <span class="info-text">共 {resultData.rows ? resultData.rows.length : 0} 行</span>
        </div>
      {:else}
        <div class="empty-state">
          <p>执行查询以显示结果</p>
          <p class="hint-text">提示: Ctrl+Enter 快速执行查询</p>
        </div>
      {/if}
    </div>
  </section>
</div>

<style>
  .db-panel {
    display: flex;
    height: 100%;
  }

  .db-sidebar {
    width: 220px;
    border-right: 1px solid #e5e7eb;
    background: #f9fafb;
    display: flex;
    flex-direction: column;
  }

  .sidebar-section {
    display: flex;
    flex-direction: column;
    border-bottom: 1px solid #e5e7eb;
    min-height: 0;
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 10px;
    font-size: 12px;
    font-weight: 600;
    color: #4b5563;
    background: #f3f4f6;
  }

  .section-body {
    padding: 6px 8px;
    overflow-y: auto;
    max-height: 200px;
  }

  .table-item,
  .history-item {
    width: 100%;
    text-align: left;
    border: none;
    background: transparent;
    padding: 6px 8px;
    cursor: pointer;
    border-radius: 4px;
    font-size: 12px;
    color: #374151;
    transition: background 0.15s ease;
  }

  .table-item:hover,
  .history-item:hover {
    background: #e5e7eb;
  }

  .empty-text {
    font-size: 12px;
    color: #9ca3af;
    padding: 8px 4px;
  }

  .icon-btn {
    border: none;
    background: transparent;
    font-size: 12px;
    cursor: pointer;
    color: #6b7280;
  }

  .db-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 12px;
    gap: 10px;
  }

  .query-toolbar {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .toolbar-btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 4px;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .toolbar-btn:hover {
    background: #f3f4f6;
  }

  .btn-primary {
    background: #2563eb;
    color: #fff;
    border: none;
  }

  .btn-secondary {
    background: #e5e7eb;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .query-textarea {
    flex: 0 0 140px;
    width: 100%;
    font-family: Menlo, Monaco, 'Courier New', monospace;
    font-size: 13px;
    background: #1e1e1e;
    color: #d4d4d8;
    border: 1px solid #e5e7eb;
    border-radius: 4px;
    padding: 10px;
    resize: vertical;
    outline: none;
    line-height: 1.5;
  }

  .results-wrapper {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .results-header {
    border-bottom: 1px solid #e5e7eb;
    padding: 6px 10px;
    background: #f9fafb;
    font-weight: 600;
    font-size: 12px;
  }

  .table-container {
    flex: 1;
    overflow: auto;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
  }

  th {
    text-align: left;
    padding: 8px 12px;
    font-weight: 600;
    color: #6b7280;
    background: #f3f4f6;
    position: sticky;
    top: 0;
    z-index: 1;
  }

  td {
    padding: 8px;
    border-bottom: 1px solid #f3f4f6;
  }

  tr:hover {
    background: #f9fafb;
  }

  .status-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 6px 10px;
    background: #f9fafb;
    border-top: 1px solid #e5e7eb;
  }

  .info-text {
    color: #6b7280;
    font-size: 12px;
  }

  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: #9ca3af;
    gap: 6px;
  }

  .hint-text {
    font-size: 11px;
    color: #9ca3af;
  }

  .error-message {
    padding: 10px;
    background: #fee2e2;
    border-radius: 4px;
    color: #dc2626;
  }

  .loading-spinner {
    animation: spin 1s linear infinite;
    border: 2px solid #f3f4f6;
    border-top: 2px solid #e5e7eb;
    border-radius: 50%;
    width: 16px;
    height: 16px;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  :global(.dark) .db-panel {
    background: #1f2937;
  }

  :global(.dark) .db-sidebar {
    background: #111827;
    border-color: #374151;
  }

  :global(.dark) .section-header {
    background: #1f2937;
    color: #d1d5db;
  }

  :global(.dark) .table-item,
  :global(.dark) .history-item {
    color: #d1d5db;
  }

  :global(.dark) .table-item:hover,
  :global(.dark) .history-item:hover {
    background: #374151;
  }

  :global(.dark) .results-header,
  :global(.dark) .status-bar {
    background: #1f2937;
    border-color: #374151;
  }

  :global(.dark) th {
    background: #1f2937;
    color: #e5e7eb;
  }

  :global(.dark) td {
    border-color: #374151;
  }

  :global(.dark) tr:hover {
    background: #1f2937;
  }
</style>
