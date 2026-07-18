<script>
  import { onMount } from 'svelte';

  export let sessionId = null;
  export let dbConfig = null;
  export let databaseName = '';
  export let schemaName = '';
  export let tableName = '';

  let query = '';
  let resultData = null;
  let isLoading = false;
  let errorMessage = '';
  let warningMessage = '';
  let queryHistory = [];
  let sortState = { key: '', direction: 'desc' };
  let activeMode = 'data';
  let filterText = '';
  let selectedCell = { row: -1, column: -1 };

  const historyLimit = 50;

  function quoteIdentifier(name, quote) {
    return `${quote}${String(name).replaceAll(quote, `${quote}${quote}`)}${quote}`;
  }

  function buildQualifiedTableName() {
    if (!tableName) return '';
    const dbType = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toLowerCase();
    const schemaScoped = ['postgresql', 'kingbase', 'opengauss'].includes(dbType);
    const quote = dbType === 'mysql' ? '`' : '"';
    const parts = schemaScoped ? [schemaName, tableName] : [databaseName, tableName];
    return parts.filter(Boolean).map(part => quoteIdentifier(part, quote)).join('.');
  }

  function buildDefaultQuery() {
    const qualifiedName = buildQualifiedTableName();
    return qualifiedName ? `SELECT * FROM ${qualifiedName} LIMIT 100;` : '';
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
    if (!Number.isNaN(numeric) && /^[-+]?\d+(\.\d+)?$/.test(text)) return numeric;
    const timestamp = Date.parse(text);
    if (!Number.isNaN(timestamp) && /[-/:T]/.test(text)) return timestamp;
    return text.toLowerCase();
  }

  function compareValues(left, right) {
    if (left === null && right === null) return 0;
    if (left === null) return 1;
    if (right === null) return -1;
    if (typeof left === 'number' && typeof right === 'number') return left - right;
    return String(left).localeCompare(String(right), undefined, { numeric: true, sensitivity: 'base' });
  }

  function handleSort(nextKey) {
    sortState = sortState.key === nextKey
      ? { key: nextKey, direction: sortState.direction === 'asc' ? 'desc' : 'asc' }
      : { key: nextKey, direction: 'desc' };
  }

  async function executeQuery() {
    if (!query.trim() || !window.wailsBindings || !sessionId) return;
    isLoading = true;
    errorMessage = '';
    warningMessage = '';
    try {
      const sql = query.trim();
      resultData = JSON.parse(await window.wailsBindings.ExecuteDatabaseQuery(sessionId, sql));
      selectedCell = { row: -1, column: -1 };
      addToHistory(sql);
      if (resultData?.rows?.length && !hasOrderBy(sql) && /^\s*select\b/i.test(sql)) {
        warningMessage = '当前查询未包含 ORDER BY，结果行顺序可能不稳定。';
      }
      activeMode = 'data';
    } catch (error) {
      errorMessage = `查询执行失败: ${error.message || '未知错误'}`;
    } finally {
      isLoading = false;
    }
  }

  async function runDefaultQuery() {
    query = buildDefaultQuery();
    await executeQuery();
  }

  function clearResult() {
    resultData = null;
    errorMessage = '';
    warningMessage = '';
    selectedCell = { row: -1, column: -1 };
  }

  function useHistory(statement) {
    query = statement;
    activeMode = 'sql';
  }

  onMount(runDefaultQuery);

  $: titleName = buildQualifiedTableName();
  $: dbTypeLabel = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toUpperCase();
  $: sortColumnIndex = resultData?.columns?.findIndex(column => column === sortState.key) ?? -1;
  $: sortedRows = !resultData?.rows ? [] : sortColumnIndex < 0
    ? resultData.rows
    : resultData.rows.slice().sort((leftRow, rightRow) => {
      const difference = compareValues(toSortableValue(leftRow?.[sortColumnIndex]), toSortableValue(rightRow?.[sortColumnIndex]));
      return sortState.direction === 'asc' ? difference : -difference;
    });
  $: filteredRows = !filterText.trim() ? sortedRows : sortedRows.filter(row => row.some(cell => String(cell ?? '').toLowerCase().includes(filterText.trim().toLowerCase())));
  $: selectedValue = selectedCell.row >= 0 ? filteredRows[selectedCell.row]?.[selectedCell.column] : null;
</script>

<div class="table-workspace">
  <header class="table-workspace__header">
    <div class="table-workspace__identity">
      <span class="table-workspace__table-icon" aria-hidden="true">▦</span>
      <div>
        <strong>{tableName || '表数据'}</strong>
        <span>{databaseName || '当前数据库'}{schemaName ? ` / ${schemaName}` : ''}</span>
      </div>
    </div>
    <nav class="table-workspace__modes" aria-label="表工作区模式">
      <button type="button" class:table-workspace__mode--active={activeMode === 'data'} class="table-workspace__mode" on:click={() => activeMode = 'data'}>数据</button>
      <button type="button" class:table-workspace__mode--active={activeMode === 'sql'} class="table-workspace__mode" on:click={() => activeMode = 'sql'}>SQL</button>
    </nav>
    <div class="table-workspace__header-meta">{dbTypeLabel || 'JDBC'} · LIMIT 100</div>
  </header>

  <div class="table-workspace__toolbar">
    <button type="button" class="table-workspace__tool table-workspace__tool--primary" title="执行 SQL" on:click={executeQuery} disabled={isLoading}>▶</button>
    <button type="button" class="table-workspace__tool" title="重新加载前 100 条" on:click={runDefaultQuery} disabled={isLoading}>↻</button>
    <button type="button" class="table-workspace__tool" title="清空结果" on:click={clearResult}>⌫</button>
    <span class="table-workspace__divider"></span>
    <button type="button" class="table-workspace__tool" title="打开 SQL 编辑器" on:click={() => activeMode = 'sql'}>{'</>'}</button>
    <span class="table-workspace__toolbar-title">{isLoading ? '正在执行...' : titleName || '未选择表'}</span>
    <label class="table-workspace__filter">
      <span aria-hidden="true">⌕</span>
      <input bind:value={filterText} placeholder="筛选结果" aria-label="筛选结果" />
    </label>
  </div>

  {#if activeMode === 'sql'}
    <section class="table-workspace__sql-panel" aria-label="SQL 编辑器">
      <textarea bind:value={query} placeholder="输入 SQL，Ctrl+Enter 执行" on:keydown={(event) => { if (event.ctrlKey && event.key === 'Enter') { event.preventDefault(); executeQuery(); } }}></textarea>
      <div><span>Ctrl+Enter 执行</span><button type="button" on:click={executeQuery} disabled={isLoading}>执行 SQL</button></div>
    </section>
  {/if}

  {#if errorMessage}<div class="table-workspace__notice table-workspace__notice--error">{errorMessage}</div>{/if}
  {#if warningMessage}<div class="table-workspace__notice table-workspace__notice--warning">{warningMessage}</div>{/if}

  <div class="table-workspace__content">
    <main class="table-workspace__grid-wrap">
      <div class="table-workspace__grid" role="table" aria-label="表数据结果">
        {#if resultData?.columns?.length}
          <div class="table-workspace__grid-head" role="row">
            <span class="table-workspace__row-number" role="columnheader">#</span>
            {#each resultData.columns as column}
              <button type="button" class:table-workspace__column--sorted={sortState.key === column} class="table-workspace__column" role="columnheader" on:click={() => handleSort(column)}>
                {column}<span aria-hidden="true">{sortState.key === column ? (sortState.direction === 'asc' ? ' ↑' : ' ↓') : ''}</span>
              </button>
            {/each}
          </div>
          {#each filteredRows as row, rowIndex}
            <div class="table-workspace__grid-row" class:table-workspace__grid-row--selected={selectedCell.row === rowIndex} role="row">
              <span class="table-workspace__row-number" role="cell">{rowIndex + 1}</span>
              {#each row as cell, columnIndex}
                <button type="button" class:table-workspace__cell--active={selectedCell.row === rowIndex && selectedCell.column === columnIndex} class="table-workspace__cell" role="cell" title={String(cell ?? '')} on:click={() => selectedCell = { row: rowIndex, column: columnIndex }}>{cell ?? 'NULL'}</button>
              {/each}
            </div>
          {/each}
        {:else}
          <div class="table-workspace__empty">{isLoading ? '正在加载表数据...' : '暂无结果，执行查询后显示数据'}</div>
        {/if}
      </div>
    </main>

    <aside class="table-workspace__details">
      <section>
        <h2>对象信息</h2>
        <dl>
          <div><dt>表</dt><dd>{tableName || '-'}</dd></div>
          <div><dt>数据库</dt><dd>{databaseName || '-'}</dd></div>
          {#if schemaName}<div><dt>Schema</dt><dd>{schemaName}</dd></div>{/if}
          <div><dt>返回行</dt><dd>{filteredRows.length}</dd></div>
        </dl>
      </section>
      <section>
        <h2>查询历史</h2>
        {#if queryHistory.length === 0}<p class="table-workspace__muted">暂无历史</p>{:else}
          {#each queryHistory as item}
            <button type="button" class="table-workspace__history" on:click={() => useHistory(item)}>{item}</button>
          {/each}
        {/if}
      </section>
    </aside>
  </div>

  <footer class="table-workspace__status">
    <span>{filteredRows.length} / {resultData?.rows?.length || 0} 行</span>
    <span>{selectedValue === null || selectedValue === undefined ? '未选择单元格' : `当前值：${String(selectedValue)}`}</span>
  </footer>
</div>

<style>
  .table-workspace { height: 100%; min-height: 0; display: flex; flex-direction: column; background: var(--bg-primary); color: var(--text-primary); }
  .table-workspace__header { min-height: 64px; padding: 0 18px; display: flex; align-items: stretch; gap: 22px; border-bottom: 1px solid var(--border-primary); }
  .table-workspace__identity { min-width: 210px; display: flex; align-items: center; gap: 10px; }
  .table-workspace__table-icon { color: #1586d1; font-size: 24px; }
  .table-workspace__identity div { min-width: 0; display: grid; gap: 2px; }
  .table-workspace__identity strong { font-size: 15px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .table-workspace__identity span, .table-workspace__header-meta { color: var(--text-secondary); font-size: 12px; }
  .table-workspace__modes { display: flex; align-items: stretch; }
  .table-workspace__mode { min-width: 64px; border: 0; border-bottom: 3px solid transparent; background: transparent; color: var(--text-secondary); cursor: pointer; font-size: 13px; }
  .table-workspace__mode:hover { color: var(--text-primary); background: var(--bg-secondary); }
  .table-workspace__mode--active { color: #1586d1; border-bottom-color: #1586d1; font-weight: 650; }
  .table-workspace__header-meta { margin-left: auto; display: flex; align-items: center; }
  .table-workspace__toolbar { min-height: 44px; padding: 0 12px; display: flex; align-items: center; gap: 5px; border-bottom: 1px solid var(--border-primary); background: var(--bg-secondary); }
  .table-workspace__tool { width: 28px; height: 28px; border: 1px solid transparent; background: transparent; color: var(--text-secondary); cursor: pointer; font-size: 15px; }
  .table-workspace__tool:hover:not(:disabled) { color: #1586d1; border-color: var(--border-primary); background: var(--bg-primary); }
  .table-workspace__tool--primary { color: #fff; background: #1687d4; border-color: #1687d4; }
  .table-workspace__tool:disabled { opacity: .5; cursor: wait; }
  .table-workspace__divider { width: 1px; height: 20px; margin: 0 5px; background: var(--border-primary); }
  .table-workspace__toolbar-title { margin-left: 5px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-secondary); font-size: 12px; }
  .table-workspace__filter { margin-left: auto; width: min(260px, 35%); height: 28px; padding: 0 8px; display: flex; align-items: center; gap: 6px; border: 1px solid var(--border-primary); background: var(--bg-primary); color: var(--text-secondary); }
  .table-workspace__filter:focus-within { border-color: #1586d1; }
  .table-workspace__filter input { min-width: 0; flex: 1; border: 0; outline: 0; background: transparent; color: inherit; font-size: 12px; }
  .table-workspace__sql-panel { padding: 10px 12px; border-bottom: 1px solid var(--border-primary); background: var(--bg-secondary); }
  .table-workspace__sql-panel textarea { box-sizing: border-box; width: 100%; height: 94px; resize: vertical; padding: 8px 10px; border: 1px solid var(--border-primary); outline: 0; background: var(--bg-primary); color: var(--text-primary); font: 12px ui-monospace, SFMono-Regular, Menlo, monospace; }
  .table-workspace__sql-panel textarea:focus { border-color: #1586d1; }
  .table-workspace__sql-panel div { padding-top: 7px; display: flex; justify-content: space-between; color: var(--text-secondary); font-size: 11px; }
  .table-workspace__sql-panel button { border: 0; background: transparent; color: #1586d1; cursor: pointer; font-size: 12px; }
  .table-workspace__notice { padding: 7px 12px; border-bottom: 1px solid var(--border-primary); font-size: 12px; }
  .table-workspace__notice--error { color: #c43832; background: #fff1f0; }
  .table-workspace__notice--warning { color: #945b00; background: #fff8e5; }
  .table-workspace__content { min-height: 0; flex: 1; display: grid; grid-template-columns: minmax(0, 1fr) 236px; overflow: hidden; }
  .table-workspace__grid-wrap { min-width: 0; min-height: 0; overflow: auto; scrollbar-gutter: stable; }
  .table-workspace__grid-wrap::-webkit-scrollbar { width: 12px; height: 12px; }
  .table-workspace__grid-wrap::-webkit-scrollbar-track { background: var(--bg-secondary); }
  .table-workspace__grid-wrap::-webkit-scrollbar-thumb { background: #a7b4c7; border: 3px solid var(--bg-secondary); border-radius: 8px; }
  .table-workspace__grid { min-width: max-content; }
  .table-workspace__grid-head, .table-workspace__grid-row { display: grid; grid-auto-flow: column; grid-auto-columns: minmax(150px, 1fr); grid-template-columns: 48px; min-height: 32px; }
  .table-workspace__grid-head { position: sticky; top: 0; z-index: 2; background: var(--bg-secondary); border-bottom: 1px solid var(--border-primary); }
  .table-workspace__grid-row { border-bottom: 1px solid var(--border-primary); }
  .table-workspace__grid-row:hover { background: color-mix(in srgb, #1586d1 5%, transparent); }
  .table-workspace__grid-row--selected { background: color-mix(in srgb, #1586d1 8%, transparent); }
  .table-workspace__row-number { position: sticky; left: 0; z-index: 1; padding: 8px 9px; text-align: right; color: var(--text-secondary); background: var(--bg-secondary); border-right: 1px solid var(--border-primary); font: 11px ui-monospace, SFMono-Regular, Menlo, monospace; }
  .table-workspace__grid-head .table-workspace__row-number { z-index: 3; }
  .table-workspace__column { min-width: 150px; padding: 0 10px; border: 0; border-right: 1px solid var(--border-primary); background: transparent; color: var(--text-secondary); text-align: left; cursor: pointer; font-size: 12px; font-weight: 650; }
  .table-workspace__column:hover, .table-workspace__column--sorted { color: #1586d1; background: color-mix(in srgb, #1586d1 6%, transparent); }
  .table-workspace__cell { min-width: 150px; overflow: hidden; padding: 8px 10px; border: 0; border-right: 1px solid var(--border-primary); background: transparent; color: var(--text-primary); text-align: left; text-overflow: ellipsis; white-space: nowrap; cursor: cell; font-size: 12px; }
  .table-workspace__cell--active { outline: 2px solid #1586d1; outline-offset: -2px; background: color-mix(in srgb, #1586d1 12%, transparent); }
  .table-workspace__empty { min-height: 180px; display: flex; align-items: center; justify-content: center; color: var(--text-secondary); font-size: 13px; }
  .table-workspace__details { padding: 15px; border-left: 1px solid var(--border-primary); background: var(--bg-secondary); overflow-y: auto; }
  .table-workspace__details section + section { margin-top: 24px; }
  .table-workspace__details h2 { margin: 0 0 12px; font-size: 12px; font-weight: 700; }
  .table-workspace__details dl { margin: 0; display: grid; gap: 12px; }
  .table-workspace__details dl div { display: grid; gap: 3px; }
  .table-workspace__details dt { color: var(--text-secondary); font-size: 11px; }
  .table-workspace__details dd { margin: 0; overflow-wrap: anywhere; font-size: 12px; }
  .table-workspace__history { display: block; width: 100%; margin-bottom: 5px; padding: 7px; overflow: hidden; border: 1px solid transparent; background: transparent; color: var(--text-secondary); text-align: left; text-overflow: ellipsis; white-space: nowrap; cursor: pointer; font: 11px ui-monospace, SFMono-Regular, Menlo, monospace; }
  .table-workspace__history:hover { border-color: var(--border-primary); background: var(--bg-primary); color: var(--text-primary); }
  .table-workspace__muted { color: var(--text-secondary); font-size: 12px; }
  .table-workspace__status { min-height: 28px; padding: 0 12px; display: flex; align-items: center; justify-content: space-between; gap: 16px; border-top: 1px solid var(--border-primary); color: var(--text-secondary); background: var(--bg-secondary); font-size: 11px; }
  @media (max-width: 900px) { .table-workspace__content { grid-template-columns: 1fr; } .table-workspace__details { display: none; } .table-workspace__header { padding: 0 12px; gap: 8px; } .table-workspace__identity { min-width: 0; } .table-workspace__header-meta { display: none; } }
</style>
