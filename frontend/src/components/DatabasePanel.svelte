<script>
  import { onMount, onDestroy } from 'svelte';
  import { buildQualifiedTableName, buildTableBrowseSQL } from '../lib/tableQueryBuilder.js';
  import { formatConnectionError } from '../lib/formatConnectionError.js';

  export let sessionId = null;
  export let dbConfig = null;

  let query = '';
  let resultData = null;
  let tables = [];
  let isLoading = false;
  let errorMessage = '';
  let errorCode = '';
  let rawError = '';
  let showRawError = false;
  let resourcePath = '';
  let queryHistory = [];
  let dbTypeLabel = '';
  let editorRatio = 0.58;
  let isResizingSplit = false;
  let splitShell;

  const historyLimit = 50;

  $: dbTypeLabel = dbConfig?.metadata?.db_type ? dbConfig.metadata.db_type.toUpperCase() : '';
  $: databaseType = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toLowerCase();
  $: errorActions = getErrorActions(errorCode);

  onMount(async () => {
    if (!sessionId) return;
    await loadTables();

    const handleTableSelect = (event) => {
      if (!event?.detail || event.detail.sessionId !== sessionId) return;
      const { databaseName, schemaName = '', tableName } = event.detail;
      if (!tableName) return;
      const qualifiedName = buildQualifiedTableName({ databaseType, databaseName, schemaName, tableName });
      query = buildTableBrowseSQL({ fromSQL: qualifiedName, databaseType, limit: 10 });
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
    clearError();

    try {
      const sql = query.trim().replace(/;+\s*$/g, '');
      const result = await window.wailsBindings.ExecuteDatabaseQuery(sessionId, sql);
      const data = JSON.parse(result);
      resultData = data;
      addToHistory(query.trim());
    } catch (error) {
      console.error('Query execution failed:', error);
      setJDBCError('查询执行失败', error);
    } finally {
      isLoading = false;
    }
  }

  async function loadTables() {
    if (!window.wailsBindings || !sessionId) return;

    isLoading = true;
    clearError();

    try {
      const result = await window.wailsBindings.ListDatabaseTables(sessionId);
      tables = (result || []).slice().sort();
    } catch (error) {
      console.error('Failed to load tables:', error);
      setJDBCError('加载数据库表失败', error);
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
    query = buildTableBrowseSQL({ fromSQL: table, databaseType, limit: 10 });
  }

  function handleTableDoubleClick(table) {
    query = buildTableBrowseSQL({ fromSQL: table, databaseType, limit: 10 });
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

  function clearError() {
    errorMessage = '';
    errorCode = '';
    rawError = '';
    showRawError = false;
    resourcePath = '';
  }

  function setJDBCError(prefix, error) {
    rawError = formatConnectionError(error, '未知错误');
    errorCode = parseJDBCErrorCode(rawError);
    errorMessage = `${prefix}: ${jdbcErrorLabel(errorCode)}`;
    showRawError = false;
    resourcePath = '';
  }

  function parseJDBCErrorCode(message) {
    const knownCodes = [
      'RUNTIME_MISSING',
      'DRIVER_MISSING',
      'DRIVER_INVALID',
      'AGENT_UNAVAILABLE',
      'QUERY_TIMEOUT',
      'QUERY_FAILED',
      'DB_CONNECT_FAILED'
    ];
    const upper = String(message || '').toUpperCase();
    return knownCodes.find((code) => upper.includes(code)) || 'DB_CONNECT_FAILED';
  }

  function jdbcErrorLabel(code) {
    switch (code) {
      case 'RUNTIME_MISSING':
        return '未找到可用 Java 运行时';
      case 'DRIVER_MISSING':
        return '当前数据库驱动未安装';
      case 'DRIVER_INVALID':
        return '当前数据库驱动文件无效';
      case 'AGENT_UNAVAILABLE':
        return 'JDBC agent 当前不可用';
      case 'QUERY_TIMEOUT':
        return '查询执行超时';
      case 'QUERY_FAILED':
        return 'SQL 执行失败';
      default:
        return '数据库连接或查询失败';
    }
  }

  function getErrorActions(code) {
    switch (code) {
      case 'RUNTIME_MISSING':
        return [
          { id: 'install-runtime', label: '安装 JRE' },
          { id: 'import-runtime', label: '导入 JRE' },
          { id: 'system-runtime', label: '选择系统 Java' }
        ];
      case 'DRIVER_MISSING':
        return [
          { id: 'install-driver', label: '安装推荐驱动' },
          { id: 'import-driver', label: '导入离线包' }
        ];
      case 'DRIVER_INVALID':
        return [
          { id: 'install-driver', label: '重新安装' },
          { id: 'view-driver', label: '查看文件' },
          { id: 'remove-driver', label: '删除' }
        ];
      case 'AGENT_UNAVAILABLE':
        return [
          { id: 'restart-agent', label: '重启 agent' },
          { id: 'view-agent-log', label: '查看日志' }
        ];
      case 'DB_CONNECT_FAILED':
        return [
          { id: 'edit-connection', label: '编辑连接' },
          { id: 'raw-error', label: '查看原始错误' }
        ];
      default:
        return [];
    }
  }

  async function handleErrorAction(action) {
    try {
      switch (action) {
        case 'install-runtime':
          await window.wailsBindings.InstallJDBCManagedRuntime();
          await loadTables();
          break;
        case 'import-runtime': {
          const archivePath = await window.wailsBindings.SelectJDBCRuntimeArchive();
          if (!archivePath) return;
          await window.wailsBindings.ImportJDBCRuntimeArchive(archivePath);
          await loadTables();
          break;
        }
        case 'system-runtime': {
          const javaPath = await window.wailsBindings.SelectJDBCJavaExecutable();
          if (!javaPath) return;
          await window.wailsBindings.SetJDBCRuntimeMode('system', javaPath);
          await loadTables();
          break;
        }
        case 'install-driver':
          await installCurrentDriver();
          await loadTables();
          break;
        case 'import-driver': {
          const packagePath = await window.wailsBindings.SelectJDBCDriverPackage();
          if (!packagePath) return;
          await window.wailsBindings.ImportJDBCDriverPackage(packagePath);
          await loadTables();
          break;
        }
        case 'view-driver':
          await showCurrentDriverPath();
          break;
        case 'remove-driver':
          await removeCurrentDriver();
          break;
        case 'restart-agent':
          await window.wailsBindings.RestartJDBCAgent();
          await loadTables();
          break;
        case 'view-agent-log':
          resourcePath = '~/.sshtools/logs/jdbc-agent.log';
          break;
        case 'edit-connection':
          window.dispatchEvent(new CustomEvent('database:edit-connection', { detail: dbConfig }));
          break;
        case 'raw-error':
          showRawError = !showRawError;
          break;
      }
    } catch (error) {
      setJDBCError('操作失败', error);
    }
  }

  async function getCurrentDriver() {
    const driverID = dbConfig?.metadata?.db_type;
    const drivers = await window.wailsBindings.ListJDBCDrivers();
    return (drivers || []).find((driver) => driver.id === driverID);
  }

  function recommendedProfile(driver) {
    return driver?.profiles?.find((profile) => profile.version === driver.recommendedVersion) || driver?.profiles?.[0];
  }

  async function installCurrentDriver() {
    const driver = await getCurrentDriver();
    const profile = recommendedProfile(driver);
    if (!driver || !profile) throw new Error('DRIVER_MISSING: 未找到当前数据库的驱动 profile');
    await window.wailsBindings.InstallJDBCDriver(driver.id, profile.version);
  }

  async function showCurrentDriverPath() {
    const profile = recommendedProfile(await getCurrentDriver());
    resourcePath = profile?.installPath || '~/.sshtools/drivers';
  }

  async function removeCurrentDriver() {
    const driver = await getCurrentDriver();
    const profile = recommendedProfile(driver);
    if (!driver || !profile) throw new Error('DRIVER_MISSING: 未找到当前数据库的驱动 profile');
    if (!window.confirm(`删除 ${driver.name} ${profile.version}？`)) return;
    await window.wailsBindings.RemoveJDBCDriver(driver.id, profile.version);
    resourcePath = '';
  }

  function formatCell(cell) {
    if (cell === null || cell === undefined) return { kind: 'null', text: 'NULL' };
    if (cell === '') return { kind: 'empty', text: '∅' };
    return { kind: 'value', text: String(cell) };
  }

  function startSplitResize(event) {
    event.preventDefault();
    isResizingSplit = true;
    const onMove = (moveEvent) => {
      if (!isResizingSplit || !splitShell) return;
      const rect = splitShell.getBoundingClientRect();
      const next = (moveEvent.clientY - rect.top) / rect.height;
      editorRatio = Math.max(0.3, Math.min(0.75, next));
    };
    const onUp = () => {
      isResizingSplit = false;
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
    };
    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
  }

  function resetSplitRatio() {
    editorRatio = 0.58;
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
          执行中
        {:else}
          运行
        {/if}
      </button>
      <button class="toolbar-btn btn-secondary" on:click={clearQuery}>清空</button>
      {#if resultData}
        <button class="toolbar-btn btn-secondary" on:click={exportResults}>导出 CSV</button>
      {/if}
      {#if isLoading}
        <span class="exec-signal" aria-live="polite">查询执行中</span>
      {/if}
    </div>

    <div class="db-split" bind:this={splitShell}>
      <div class="sql-editor-shell db-editor" style={`flex: ${editorRatio} 1 0;`}>
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
      </div>

      <div
        class="ops-split-handle"
        role="separator"
        aria-orientation="horizontal"
        title="拖拽调整比例，双击复位 58/42"
        on:mousedown={startSplitResize}
        on:dblclick={resetSplitRatio}
      ></div>

      <div class="results-wrapper" style={`flex: ${1 - editorRatio} 1 0;`}>
        {#if errorMessage}
          <div class="error-message" role="alert">
            <div class="error-message__title">{errorMessage}</div>
            <div class="error-message__actions">
              {#each errorActions as action}
                <button type="button" on:click={() => handleErrorAction(action.id)}>{action.label}</button>
              {/each}
            </div>
            {#if resourcePath}
              <code class="error-message__detail">{resourcePath}</code>
            {/if}
            {#if showRawError}
              <pre class="error-message__detail">{rawError}</pre>
            {/if}
          </div>
        {/if}

        {#if resultData && resultData.columns && resultData.columns.length > 0}
          <div class="results-header">查询结果</div>
          <div class="table-container">
            <table class="result-table">
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
                      {@const formatted = formatCell(cell)}
                      <td class:cell-null={formatted.kind !== 'value'}>
                        {formatted.text}
                      </td>
                    {/each}
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
          <div class="status-bar">
            <span class="info-text">共 {resultData.rows ? resultData.rows.length : 0} 行</span>
          </div>
        {:else if !errorMessage}
          <div class="empty-state">
            <p>运行查询以显示结果</p>
            <p class="hint-text">提示: Ctrl+Enter 快速运行</p>
          </div>
        {/if}
      </div>
    </div>
  </section>
</div>

<style>
  .db-panel {
    display: flex;
    height: 100%;
    min-height: 0;
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .db-sidebar {
    width: 220px;
    min-width: 200px;
    display: flex;
    flex-direction: column;
    background: var(--bg-secondary);
    box-shadow: inset -1px 0 0 var(--border-secondary);
  }

  .sidebar-section {
    display: flex;
    flex-direction: column;
    min-height: 0;
    border-bottom: 1px solid var(--border-secondary);
  }

  .section-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 10px;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    background: transparent;
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
    border-radius: var(--radius-sm, 4px);
    font-size: 12px;
    color: var(--text-primary);
    transition: background var(--trans-fast, 140ms ease);
  }

  .table-item:hover,
  .history-item:hover {
    background: var(--bg-hover);
  }

  .empty-text {
    font-size: 12px;
    color: var(--text-tertiary);
    padding: 8px 4px;
  }

  .icon-btn {
    border: none;
    background: transparent;
    font-size: 12px;
    cursor: pointer;
    color: var(--text-secondary);
  }

  .db-main {
    flex: 1;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    padding: 10px 12px;
    gap: 8px;
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
    border-radius: var(--radius-md, 8px);
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: background var(--trans-fast, 140ms ease), border-color var(--trans-fast, 140ms ease);
  }

  .btn-primary {
    background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));
    color: #fff;
    border: none;
  }

  .btn-primary:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: var(--glass-bg);
    color: var(--text-primary);
    border: 1px solid var(--border-primary);
  }

  .btn-secondary:hover {
    background: var(--bg-hover);
  }

  .exec-signal {
    margin-left: 4px;
    padding-left: 10px;
    border-left: 3px solid var(--ops-blue);
    color: var(--text-secondary);
    font-size: 12px;
  }

  .db-split {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .db-editor {
    flex: 1 1 auto;
    min-height: 120px;
    display: flex;
    border-radius: var(--radius-md, 8px);
    background: var(--ops-terminal-bg);
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.04);
  }

  .query-textarea {
    flex: 1;
    width: 100%;
    min-height: 0;
    resize: none;
    border: none;
    outline: none;
    padding: 12px;
    background: transparent;
    color: #d4d4d8;
    font-family: var(--terminal-font-family);
    font-size: 13px;
    line-height: 1.55;
  }

  .results-wrapper {
    flex: 1 1 auto;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    border-radius: var(--radius-md, 8px);
    background: var(--bg-secondary);
  }

  .results-header,
  .status-bar {
    padding: 6px 10px;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    background: transparent;
    border-bottom: 1px solid var(--border-secondary);
  }

  .status-bar {
    border-bottom: none;
    border-top: 1px solid var(--border-secondary);
    font-weight: 500;
  }

  .table-container {
    flex: 1;
    overflow: auto;
  }

  .info-text {
    color: var(--text-tertiary);
    font-size: 12px;
  }

  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: var(--text-tertiary);
    gap: 6px;
  }

  .hint-text {
    font-size: 11px;
    color: var(--text-tertiary);
  }

  .error-message {
    margin: 8px;
    padding: 10px;
    border-radius: var(--radius-md, 8px);
    background: color-mix(in srgb, var(--ops-alert) 12%, transparent);
    color: var(--ops-alert);
    border-left: 3px solid var(--ops-alert);
  }

  .error-message__title {
    font-size: 12px;
    font-weight: 600;
  }

  .error-message__actions {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 8px;
  }

  .error-message__actions button {
    min-height: 28px;
    border: 1px solid color-mix(in srgb, var(--ops-alert) 35%, transparent);
    border-radius: var(--radius-sm, 4px);
    background: var(--bg-elevated);
    color: var(--ops-alert);
    padding: 0 9px;
    font-size: 11px;
    font-weight: 600;
  }

  .error-message__detail {
    display: block;
    max-height: 96px;
    overflow: auto;
    margin: 8px 0 0;
    border-radius: var(--radius-sm, 4px);
    background: color-mix(in srgb, var(--bg-primary) 70%, transparent);
    padding: 7px;
    color: var(--text-secondary);
    font-family: var(--terminal-font-family);
    font-size: 11px;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .loading-spinner {
    animation: spin 1s linear infinite;
    border: 2px solid color-mix(in srgb, #fff 35%, transparent);
    border-top: 2px solid #fff;
    border-radius: 50%;
    width: 14px;
    height: 14px;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
</style>
