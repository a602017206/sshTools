<script>
  import { onMount } from 'svelte';

  export let sessionId = null;
  export let dbConfig = null;
  export let databaseName = '';
  export let tableName = '';

  let query = '';
  let resultData = null;
  let isLoading = false;
  let errorMessage = '';
  let warningMessage = '';
  let queryHistory = [];
  let sortState = { key: '', direction: 'desc' };

  const historyLimit = 50;

  function buildQualifiedTableName() {
    if (!tableName) return '';
    return databaseName ? `${databaseName}.${tableName}` : tableName;
  }

  function buildDefaultQuery() {
    const qualifiedName = buildQualifiedTableName();
    if (!qualifiedName) return '';
    return `SELECT * FROM ${qualifiedName} LIMIT 100;`;
  }

  function hasOrderBy(sql) {
    return /\border\s+by\b/i.test(sql);
  }

  function addToHistory(statement) {
    const normalized = statement.trim();
    if (!normalized) return;
    queryHistory = [normalized, ...queryHistory.filter(item => item !== normalized)].slice(0, historyLimit);
  }

  function toSortableValue(value) {
    if (value === null || value === undefined || value === '') return null;
    if (typeof value === 'number') return value;

    const text = String(value).trim();
    if (!text) return null;

    const numeric = Number(text);
    if (!Number.isNaN(numeric) && /^[-+]?\d+(\.\d+)?$/.test(text)) {
      return numeric;
    }

    const ts = Date.parse(text);
    if (!Number.isNaN(ts) && /[-/:T]/.test(text)) {
      return ts;
    }

    return text.toLowerCase();
  }

  function compareValues(left, right) {
    if (left === null && right === null) return 0;
    if (left === null) return 1;
    if (right === null) return -1;

    if (typeof left === 'number' && typeof right === 'number') {
      return left - right;
    }

    return String(left).localeCompare(String(right), undefined, { numeric: true, sensitivity: 'base' });
  }

  function handleSort(nextKey) {
    if (sortState.key === nextKey) {
      sortState = {
        key: sortState.key,
        direction: sortState.direction === 'asc' ? 'desc' : 'asc'
      };
      return;
    }
    sortState = { key: nextKey, direction: 'desc' };
  }

  async function executeQuery() {
    if (!query.trim()) return;
    if (!window.wailsBindings || !sessionId) return;

    isLoading = true;
    errorMessage = '';
    warningMessage = '';

    try {
      const sql = query.trim();
      const result = await window.wailsBindings.ExecuteDatabaseQuery(sessionId, sql);
      const data = JSON.parse(result);
      resultData = data;
      addToHistory(sql);

      if (data?.rows?.length > 0 && !hasOrderBy(sql) && /^\s*select\b/i.test(sql)) {
        warningMessage = '提示：当前查询未包含 ORDER BY，结果行顺序可能不稳定。';
      }
    } catch (error) {
      console.error('Query execution failed:', error);
      errorMessage = `查询执行失败: ${error.message || '未知错误'}`;
    } finally {
      isLoading = false;
    }
  }

  function clearQuery() {
    query = '';
    resultData = null;
    errorMessage = '';
    warningMessage = '';
  }

  function useHistory(statement) {
    query = statement;
  }

  async function runDefaultQuery() {
    query = buildDefaultQuery();
    if (query) {
      await executeQuery();
    }
  }

  onMount(async () => {
    await runDefaultQuery();
  });

  $: titleName = buildQualifiedTableName();
  $: dbTypeLabel = dbConfig?.metadata?.db_type ? dbConfig.metadata.db_type.toUpperCase() : '';
  $: sortKey = sortState.key;
  $: sortDirection = sortState.direction;
  $: sortColumnIndex = resultData?.columns?.findIndex(col => col === sortKey) ?? -1;
  $: sortedRows = !resultData?.rows
    ? []
    : sortColumnIndex < 0
      ? resultData.rows
      : resultData.rows.slice().sort((a, b) => {
          const left = toSortableValue(a?.[sortColumnIndex]);
          const right = toSortableValue(b?.[sortColumnIndex]);
          const diff = compareValues(left, right);
          return sortDirection === 'asc' ? diff : -diff;
        });

  // Reactive derived values for sort indicator
  $: currentSortKey = sortState.key;
  $: currentSortDirection = sortState.direction;
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800">
  <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between gap-3">
    <div class="min-w-0">
      <div class="text-sm font-semibold text-gray-900 dark:text-white truncate">{titleName || '表数据'}</div>
      <div class="text-xs text-gray-500 dark:text-gray-400">{dbTypeLabel ? `${dbTypeLabel} · ` : ''}默认查询 LIMIT 100</div>
    </div>
    <div class="flex items-center gap-2">
      <button class="px-3 py-1.5 text-xs rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50" on:click={executeQuery} disabled={isLoading}>
        {isLoading ? '执行中...' : '执行'}
      </button>
      <button class="px-3 py-1.5 text-xs rounded bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200" on:click={clearQuery}>
        清空
      </button>
      <button class="px-3 py-1.5 text-xs rounded bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200" on:click={runDefaultQuery}>
        重置100条
      </button>
    </div>
  </div>

  <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-900/40">
    <textarea
      class="w-full h-28 rounded border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 text-sm font-mono px-3 py-2 text-gray-900 dark:text-gray-100"
      bind:value={query}
      placeholder="输入 SQL，Ctrl+Enter 执行"
      on:keydown={(e) => {
        if (e.ctrlKey && e.key === 'Enter') {
          e.preventDefault();
          executeQuery();
        }
      }}
    ></textarea>
  </div>

  {#if errorMessage}
    <div class="px-4 py-2 text-xs text-red-600 bg-red-50 dark:bg-red-900/20 border-b border-red-100 dark:border-red-800">{errorMessage}</div>
  {/if}

  {#if warningMessage}
    <div class="px-4 py-2 text-xs text-amber-700 bg-amber-50 dark:bg-amber-900/20 border-b border-amber-100 dark:border-amber-800">{warningMessage}</div>
  {/if}

  <div class="flex-1 min-h-0 flex">
    <div class="flex-1 min-w-0 overflow-auto">
      {#if resultData?.columns?.length}
        <table class="w-full text-xs border-collapse">
          <thead class="sticky top-0 z-10 bg-gray-100 dark:bg-gray-700">
              <tr>
                {#each resultData.columns as col}
                  <th class="text-left px-3 py-2 border-b border-gray-200 dark:border-gray-600 font-semibold text-gray-700 dark:text-gray-200">
                    <button
                      type="button"
                      class="inline-flex items-center gap-1 hover:text-blue-600 dark:hover:text-blue-300 {currentSortKey === col ? 'text-blue-600 dark:text-blue-300' : ''}"
                      on:click={() => handleSort(col)}
                    >
                      <span>{col}</span>
                      {#if currentSortKey === col}
                        <span aria-hidden="true">{currentSortDirection === 'asc' ? '↑' : '↓'}</span>
                      {/if}
                    </button>
                  </th>
                {/each}
              </tr>
            </thead>
          <tbody>
            {#each sortedRows as row}
              <tr class="hover:bg-gray-50 dark:hover:bg-gray-700/40">
                {#each row as cell}
                  <td class="px-3 py-2 border-b border-gray-100 dark:border-gray-700 text-gray-800 dark:text-gray-100 align-top">{cell}</td>
                {/each}
              </tr>
            {/each}
          </tbody>
        </table>
      {:else}
        <div class="h-full flex items-center justify-center text-sm text-gray-500 dark:text-gray-400">
          暂无结果，执行查询后显示数据
        </div>
      {/if}
    </div>

    <aside class="w-64 border-l border-gray-200 dark:border-gray-700 overflow-y-auto p-3 hidden lg:block">
      <div class="text-xs font-semibold text-gray-600 dark:text-gray-300 mb-2">查询历史</div>
      {#if queryHistory.length === 0}
        <div class="text-xs text-gray-500 dark:text-gray-400">暂无历史</div>
      {:else}
        {#each queryHistory as item}
          <button
            type="button"
            class="w-full text-left mb-1 px-2 py-1.5 rounded text-xs bg-gray-50 dark:bg-gray-700 hover:bg-gray-100 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200"
            on:click={() => useHistory(item)}
          >
            {item}
          </button>
        {/each}
      {/if}
    </aside>
  </div>

  <div class="px-4 py-2 border-t border-gray-200 dark:border-gray-700 text-xs text-gray-500 dark:text-gray-400">
    行数：{resultData?.rows?.length || 0}
  </div>
</div>
