<script>
  import { onDestroy, onMount } from 'svelte';
  import {
    buildQualifiedTableName as buildQualifiedTableSQL,
    buildTableBrowseSQL,
    operationNeedsValue,
    operationUsesList,
    tableFilterOperations
  } from '../lib/tableQueryBuilder.js';
  import { formatColumnDescription, formatColumnLength, formatColumnType } from '../lib/tableStructureMetadata.js';
  import { buildGridTemplateColumns, clampColumnWidth, getInitialColumnWidth } from '../lib/tableGridColumns.js';
  import { buildDeleteSQL, buildInsertSQL, buildUpdateSQL } from '../lib/tableDataMutations.js';
  import { formatConnectionError } from '../lib/formatConnectionError.js';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';

  export let sessionId = null;
  export let dbConfig = null;
  export let databaseName = '';
  export let schemaName = '';
  export let tableName = '';
  export let initialQuery = '';

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
  let queryBuilderOpen = false;
  let filterRules = [];
  let sortRules = [];
  let columnMetadata = {};
  let columnWidths = {};
  let resizingColumn = '';
  let resizeStartX = 0;
  let resizeStartWidth = 0;
  let pageSize = 100;
  let currentPage = 1;
  let editedCells = {};
  let contextMenu = null;
  let isMutating = false;
  let pendingDelete = null;

  const historyLimit = 50;

  function buildQualifiedTableName() {
    return buildQualifiedTableSQL({
      databaseType: String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toLowerCase(),
      databaseName,
      schemaName,
      tableName
    });
  }

  function buildDefaultQuery(page = 1) {
    const qualifiedName = buildQualifiedTableName();
    if (!qualifiedName) return '';
    const size = Number(pageSize) > 0 ? Number(pageSize) : 100;
    return buildTableBrowseSQL({
      fromSQL: qualifiedName,
      databaseType: String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toLowerCase(),
      filters: [],
      sorters: [],
      limit: size,
      offset: Math.max(0, (page - 1) * size)
    });
  }

  function createFilterRule() {
    return {
      connector: 'AND',
      field: availableColumns[0] || '',
      operation: 'contains',
      value: ''
    };
  }

  function createSortRule() {
    return {
      field: availableColumns[0] || '',
      direction: 'ASC'
    };
  }

  function addFilterRule() {
    filterRules = [...filterRules, createFilterRule()];
  }

  function updateFilterRule(index, changes) {
    filterRules = filterRules.map((rule, ruleIndex) => ruleIndex === index ? { ...rule, ...changes } : rule);
  }

  function removeFilterRule(index) {
    filterRules = filterRules.filter((_, ruleIndex) => ruleIndex !== index);
  }

  function updateFilterOperation(index, operation) {
    updateFilterRule(index, { operation, value: operationNeedsValue(operation) ? filterRules[index]?.value || '' : '' });
  }

  function addSortRule() {
    sortRules = [...sortRules, createSortRule()];
  }

  function updateSortRule(index, changes) {
    sortRules = sortRules.map((rule, ruleIndex) => ruleIndex === index ? { ...rule, ...changes } : rule);
  }

  function removeSortRule(index) {
    sortRules = sortRules.filter((_, ruleIndex) => ruleIndex !== index);
  }

  async function applyQueryBuilder() {
    const qualifiedName = buildQualifiedTableName();
    if (!qualifiedName) return;
    const size = Number(pageSize) > 0 ? Number(pageSize) : 100;
    query = buildTableBrowseSQL({
      fromSQL: qualifiedName,
      databaseType,
      filters: filterRules,
      sorters: sortRules,
      limit: size,
      offset: (currentPage - 1) * size
    });
    sortState = { key: '', direction: 'desc' };
    await executeQuery();
  }

  async function resetQueryBuilder() {
    filterRules = [];
    sortRules = [];
    sortState = { key: '', direction: 'desc' };
    await runDefaultQuery();
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

  function getColumnWidth(column) {
    return columnWidths[column] ?? getInitialColumnWidth(column, columnMetadata[column]);
  }

  function resizeColumn(event) {
    if (!resizingColumn) return;
    const nextWidth = clampColumnWidth(resizeStartWidth + event.clientX - resizeStartX);
    columnWidths = { ...columnWidths, [resizingColumn]: nextWidth };
  }

  function stopColumnResize() {
    window.removeEventListener('pointermove', resizeColumn);
    window.removeEventListener('pointerup', stopColumnResize);
    resizingColumn = '';
  }

  function startColumnResize(event, column) {
    event.preventDefault();
    resizingColumn = column;
    resizeStartX = event.clientX;
    resizeStartWidth = getColumnWidth(column);
    window.addEventListener('pointermove', resizeColumn);
    window.addEventListener('pointerup', stopColumnResize);
  }

  async function executeQuery() {
    if (!query.trim() || !window.wailsBindings || !sessionId) return;
    isLoading = true;
    errorMessage = '';
    warningMessage = '';
    try {
      const sql = query.trim().replace(/;+\s*$/g, '');
      resultData = JSON.parse(await window.wailsBindings.ExecuteDatabaseQuery(sessionId, sql));
      selectedCell = { row: -1, column: -1 };
      addToHistory(query.trim());
      if (resultData?.rows?.length && !hasOrderBy(sql) && /^\s*select\b/i.test(sql)) {
        warningMessage = '当前查询未包含 ORDER BY，结果行顺序可能不稳定。';
      }
      activeMode = 'data';
    } catch (error) {
      errorMessage = `查询执行失败: ${formatConnectionError(error, '未知错误')}`;
    } finally {
      isLoading = false;
    }
  }

  async function runDefaultQuery(page = 1) {
    currentPage = page;
    query = buildDefaultQuery(page);
    await executeQuery();
  }

  async function loadColumnMetadata() {
    if (!window.wailsBindings || !sessionId || !tableName) return;
    const dbType = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toLowerCase();
    // Oracle/达梦在无 schema 时全库扫列元数据极慢，且会与查询争用同一连接。
    if (['oracle', 'dm'].includes(dbType) && !String(schemaName || '').trim()) {
      columnMetadata = {};
      return;
    }
    try {
      const schema = await window.wailsBindings.GetTableSchemaInSchema(sessionId, databaseName, schemaName, tableName);
      columnMetadata = Object.fromEntries((schema?.columns || []).map(column => [column.name, column]));
    } catch (error) {
      console.warn('Failed to load table column metadata:', error);
      columnMetadata = {};
    }
  }

  function clearResult() {
    resultData = null;
    errorMessage = '';
    warningMessage = '';
    selectedCell = { row: -1, column: -1 };
  }

  function editedValue(rowIndex, columnIndex, value) {
    return editedCells[`${rowIndex}:${columnIndex}`] ?? value;
  }

  function editCell(rowIndex, columnIndex, value) {
    editedCells = { ...editedCells, [`${rowIndex}:${columnIndex}`]: value };
  }

  function rowChanges(rowIndex, row) {
    return Object.fromEntries(resultData.columns.map((column, columnIndex) => [column, editedValue(rowIndex, columnIndex, row[columnIndex])]).filter(([, value], columnIndex) => value !== row[columnIndex]));
  }

  function primaryKeys() {
    return Object.values(columnMetadata)
      .filter(column => column.is_primary_key || column.primary_key || column.isPrimaryKey)
      .map(column => column.name);
  }

  async function copyText(value) {
    await navigator.clipboard?.writeText(String(value ?? ''));
    contextMenu = null;
  }

  function openContextMenu(event, rowIndex, row) {
    event.preventDefault();
    contextMenu = { x: event.clientX, y: event.clientY, rowIndex, row };
  }

  function mutationInput(row, changes = {}) {
    return { databaseType, table: titleName, columns: resultData.columns, row, primaryKeys: primaryKeys(), changes };
  }

  function requestDelete(row) {
    const sql = buildDeleteSQL(mutationInput(row));
    const message = hasPrimaryKey
      ? '删除当前记录？此操作无法撤销。'
      : '当前表未识别到主键，将按整行原始值删除；重复行可能会同时受影响。是否继续？';
    if (!sql) {
      errorMessage = '无法为当前记录生成删除条件。';
      return;
    }
    pendingDelete = { sql, message };
    contextMenu = null;
  }

  async function deleteRow() {
    const sql = pendingDelete?.sql;
    if (!sql) return;
    isMutating = true;
    try {
      const result = JSON.parse(await window.wailsBindings.ExecuteDatabaseQuery(sessionId, sql));
      await runDefaultQuery(currentPage);
      warningMessage = Number(result?.affected || 0) > 0
        ? `删除结果：已删除 ${result.affected} 条记录。`
        : '删除结果：未匹配到记录，数据未发生变化。';
    } catch (error) {
      errorMessage = `删除记录失败: ${formatConnectionError(error, '未知错误')}`;
    } finally {
      isMutating = false;
      pendingDelete = null;
    }
  }

  async function saveChanges() {
    const groups = Object.entries(editedCells).reduce((acc, [key, value]) => { const [rowIndex, columnIndex] = key.split(':').map(Number); (acc[rowIndex] ||= {})[resultData.columns[columnIndex]] = value; return acc; }, {});
    const statements = Object.entries(groups).map(([rowIndex, changes]) => buildUpdateSQL(mutationInput(filteredRows[Number(rowIndex)], changes))).filter(Boolean);
    if (!statements.length) return;
    isMutating = true;
    try { for (const statement of statements) await window.wailsBindings.ExecuteDatabaseQuery(sessionId, statement); editedCells = {}; await runDefaultQuery(currentPage); } catch (error) { errorMessage = `保存失败: ${formatConnectionError(error, '未知错误')}`; } finally { isMutating = false; }
  }

  function discardChanges() { editedCells = {}; }

  function useHistory(statement) {
    query = statement;
    activeMode = 'sql';
  }

  onMount(async () => {
    // JDBC Connection 非线程安全：必须先完成元数据再跑查询，避免 Oracle 等驱动并发挂死超时。
    await loadColumnMetadata();
    if (tableName) await runDefaultQuery();
    else {
      query = initialQuery;
      activeMode = 'sql';
    }
  });

  onDestroy(stopColumnResize);

  $: titleName = buildQualifiedTableName();
  $: databaseType = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toLowerCase();
  $: dbTypeLabel = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toUpperCase();
  $: availableColumns = resultData?.columns || [];
  $: if (availableColumns.length && filterRules.some(rule => !rule.field)) {
    filterRules = filterRules.map(rule => rule.field ? rule : { ...rule, field: availableColumns[0] });
  }
  $: if (availableColumns.length && sortRules.some(rule => !rule.field)) {
    sortRules = sortRules.map(rule => rule.field ? rule : { ...rule, field: availableColumns[0] });
  }
  $: sortColumnIndex = resultData?.columns?.findIndex(column => column === sortState.key) ?? -1;
  $: sortedRows = !resultData?.rows ? [] : sortColumnIndex < 0
    ? resultData.rows
    : resultData.rows.slice().sort((leftRow, rightRow) => {
      const difference = compareValues(toSortableValue(leftRow?.[sortColumnIndex]), toSortableValue(rightRow?.[sortColumnIndex]));
      return sortState.direction === 'asc' ? difference : -difference;
    });
  $: filteredRows = !filterText.trim() ? sortedRows : sortedRows.filter(row => row.some(cell => String(cell ?? '').toLowerCase().includes(filterText.trim().toLowerCase())));
  $: selectedValue = selectedCell.row >= 0 ? filteredRows[selectedCell.row]?.[selectedCell.column] : null;
  $: gridTemplateColumns = buildGridTemplateColumns(resultData?.columns || [], columnWidths, columnMetadata);
  $: hasPrimaryKey = primaryKeys().length > 0;
  $: hasEdits = Object.keys(editedCells).length > 0;
</script>

<div class="table-workspace">
  <header class="table-workspace__header table-workspace__context">
    <div class="table-workspace__identity">
      <span class="table-workspace__table-icon" aria-hidden="true">▦</span>
      <div>
        <strong>{tableName || 'SQL 查询'}</strong>
        <span>{databaseName || '当前数据库'}{schemaName ? ` / ${schemaName}` : ''}</span>
      </div>
    </div>
    <nav class="table-workspace__modes" aria-label="表工作区模式">
      <button type="button" class:table-workspace__mode--active={activeMode === 'data'} class="table-workspace__mode" on:click={() => activeMode = 'data'}>数据</button>
      <button type="button" class:table-workspace__mode--active={activeMode === 'sql'} class="table-workspace__mode" on:click={() => activeMode = 'sql'}>SQL</button>
    </nav>
    <div class="table-workspace__header-meta">{dbTypeLabel || 'JDBC'}{tableName ? ` · ${pageSize} 行/页` : ''}</div>
  </header>

  <div class="table-workspace__toolbar table-workspace__query-strip">
    <button type="button" class="table-workspace__tool table-workspace__tool--primary" title="运行" on:click={executeQuery} disabled={isLoading}>▶</button>
    <button type="button" class="table-workspace__tool" title="重新加载当前页" on:click={() => runDefaultQuery(currentPage)} disabled={isLoading || !tableName}>↻</button>
    <button type="button" class:table-workspace__tool--active={queryBuilderOpen} class="table-workspace__tool" title="筛选与排序" on:click={() => queryBuilderOpen = !queryBuilderOpen} disabled={!tableName}>⏷</button>
    <button type="button" class="table-workspace__tool" title="清空结果" on:click={clearResult}>⌫</button>
    <button type="button" class="table-workspace__tool" title={hasEdits ? '保存' : '修改单元格后可保存'} disabled={!hasEdits || isMutating} on:click={saveChanges}>✓</button>
    <button type="button" class="table-workspace__tool" title={hasEdits ? '放弃修改' : '当前没有待放弃的修改'} disabled={!hasEdits || isMutating} on:click={discardChanges}>↶</button>
    <span class="table-workspace__divider"></span>
    <button type="button" class="table-workspace__tool" title="打开 SQL 编辑器" on:click={() => activeMode = 'sql'}>{'</>'}</button>
    <span class="table-workspace__toolbar-title">{isLoading ? '正在执行...' : titleName || '未选择表'}</span>
    <label class="table-workspace__filter">
      <span aria-hidden="true">⌕</span>
      <input bind:value={filterText} placeholder="在当前结果中查找" aria-label="在当前结果中查找" />
    </label>
  </div>

  {#if queryBuilderOpen}
    <section class="table-workspace__query-builder" aria-label="筛选与排序">
      <div class="table-workspace__query-builder-head">
        <div>
          <strong>筛选与排序</strong>
          <span>应用后重新读取前 100 条数据</span>
        </div>
        <div class="table-workspace__query-builder-actions">
          <button type="button" on:click={resetQueryBuilder} disabled={isLoading}>重置</button>
          <button type="button" class="table-workspace__query-builder-apply" on:click={applyQueryBuilder} disabled={isLoading || !titleName}>应用</button>
        </div>
      </div>

      <div class="table-workspace__query-builder-section">
        <div class="table-workspace__query-builder-section-head">
          <strong>筛选条件</strong>
          <button type="button" on:click={addFilterRule} disabled={!availableColumns.length}>添加条件</button>
        </div>
        {#if !filterRules.length}
          <p class="table-workspace__query-builder-empty">暂无筛选条件</p>
        {:else}
          <div class="table-workspace__query-builder-rules">
            {#each filterRules as rule, index}
              <div class="table-workspace__filter-rule">
                {#if index === 0}
                  <span class="table-workspace__rule-connector">WHERE</span>
                {:else}
                  <select aria-label="条件连接符" value={rule.connector} on:change={(event) => updateFilterRule(index, { connector: event.currentTarget.value })}>
                    <option value="AND">AND</option>
                    <option value="OR">OR</option>
                  </select>
                {/if}
                <select aria-label="筛选字段" value={rule.field} on:change={(event) => updateFilterRule(index, { field: event.currentTarget.value })} disabled={!availableColumns.length}>
                  {#each availableColumns as column}<option value={column}>{column}</option>{/each}
                </select>
                <select aria-label="比较方式" value={rule.operation} on:change={(event) => updateFilterOperation(index, event.currentTarget.value)}>
                  {#each tableFilterOperations as operation}<option value={operation.value}>{operation.label}</option>{/each}
                </select>
                {#if operationNeedsValue(rule.operation)}
                  {#if operationUsesList(rule.operation)}
                    <textarea aria-label="筛选值" value={rule.value} placeholder="用逗号或换行分隔" on:input={(event) => updateFilterRule(index, { value: event.currentTarget.value })}></textarea>
                  {:else}
                    <input aria-label="筛选值" value={rule.value} placeholder="输入值" on:input={(event) => updateFilterRule(index, { value: event.currentTarget.value })} />
                  {/if}
                {:else}
                  <span class="table-workspace__rule-null-value">无需输入值</span>
                {/if}
                <button type="button" class="table-workspace__rule-remove" title="移除条件" on:click={() => removeFilterRule(index)}>×</button>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <div class="table-workspace__query-builder-section">
        <div class="table-workspace__query-builder-section-head">
          <strong>排序规则</strong>
          <button type="button" on:click={addSortRule} disabled={!availableColumns.length}>添加排序</button>
        </div>
        {#if !sortRules.length}
          <p class="table-workspace__query-builder-empty">暂无排序规则</p>
        {:else}
          <div class="table-workspace__query-builder-rules">
            {#each sortRules as rule, index}
              <div class="table-workspace__sort-rule">
                <span>ORDER BY</span>
                <select aria-label="排序字段" value={rule.field} on:change={(event) => updateSortRule(index, { field: event.currentTarget.value })} disabled={!availableColumns.length}>
                  {#each availableColumns as column}<option value={column}>{column}</option>{/each}
                </select>
                <select aria-label="排序方向" value={rule.direction} on:change={(event) => updateSortRule(index, { direction: event.currentTarget.value })}>
                  <option value="ASC">升序</option>
                  <option value="DESC">降序</option>
                </select>
                <button type="button" class="table-workspace__rule-remove" title="移除排序" on:click={() => removeSortRule(index)}>×</button>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    </section>
  {/if}

  {#if activeMode === 'sql'}
    <section class="table-workspace__sql-panel" aria-label="SQL 编辑器">
      <textarea bind:value={query} placeholder="输入 SQL，Ctrl+Enter 运行" on:keydown={(event) => { if (event.ctrlKey && event.key === 'Enter') { event.preventDefault(); executeQuery(); } }}></textarea>
      <div><span>Ctrl+Enter 运行</span><button type="button" on:click={executeQuery} disabled={isLoading}>运行</button></div>
    </section>
  {/if}

  {#if errorMessage}<div class="table-workspace__notice table-workspace__notice--error">{errorMessage}</div>{/if}
  {#if warningMessage}<div class="table-workspace__notice table-workspace__notice--warning">{warningMessage}</div>{/if}

  <div class="table-workspace__content">
    <main class="table-workspace__grid-wrap">
      <div class="table-workspace__grid" role="table" aria-label="表数据结果">
        {#if resultData?.columns?.length}
          <div class="table-workspace__grid-head" role="row" style={`grid-template-columns: ${gridTemplateColumns};`}>
            <span class="table-workspace__row-number" role="columnheader">#</span>
            {#each resultData.columns as column}
              {@const metadata = columnMetadata[column]}
              <div class:table-workspace__column--sorted={sortState.key === column} class="table-workspace__column" role="columnheader">
                <button type="button" class="table-workspace__column-sort" on:click={() => handleSort(column)}>
                  <span class="table-workspace__column-name">{column}<span aria-hidden="true">{sortState.key === column ? (sortState.direction === 'asc' ? ' ↑' : ' ↓') : ''}</span></span>
                  {#if metadata}
                    <span class="table-workspace__column-metadata" title={`${formatColumnType(metadata)} · 长度 ${formatColumnLength(metadata)} · ${formatColumnDescription(metadata)}`}>
                      <span>{formatColumnType(metadata)}</span>
                      <span>长度 {formatColumnLength(metadata)}</span>
                    <span>{formatColumnDescription(metadata)}</span>
                  </span>
                {/if}
                  <span class="table-workspace__column-accent" aria-hidden="true"></span>
                </button>
                <button type="button" class="table-workspace__column-resizer" aria-label={`调整 ${column} 列宽`} title={`调整 ${column} 列宽`} on:pointerdown={(event) => startColumnResize(event, column)}></button>
              </div>
            {/each}
          </div>
          {#each filteredRows as row, rowIndex}
            <div class="table-workspace__grid-row" class:table-workspace__grid-row--selected={selectedCell.row === rowIndex} role="row" style={`grid-template-columns: ${gridTemplateColumns};`}>
              <span class="table-workspace__row-number" role="cell">{rowIndex + 1}</span>
              {#each row as cell, columnIndex}
                <input
                  class:table-workspace__cell--active={selectedCell.row === rowIndex && selectedCell.column === columnIndex}
                  class:table-workspace__cell--edited={editedValue(rowIndex, columnIndex, cell) !== cell}
                  class:table-workspace__cell--null={cell === null || cell === undefined}
                  class:table-workspace__cell--empty={cell === ''}
                  class="table-workspace__cell"
                  role="cell"
                  value={editedValue(rowIndex, columnIndex, cell) ?? ''}
                  placeholder={cell === null || cell === undefined ? 'NULL' : cell === '' ? '∅' : ''}
                  title={cell === null || cell === undefined ? 'NULL' : String(cell)}
                  on:focus={() => selectedCell = { row: rowIndex, column: columnIndex }}
                  on:input={(event) => editCell(rowIndex, columnIndex, event.currentTarget.value)}
                  on:contextmenu={(event) => openContextMenu(event, rowIndex, row)}
                />
              {/each}
            </div>
          {/each}
        {:else}
          <div class="table-workspace__empty">{isLoading ? '正在加载表数据...' : '暂无结果，运行查询后显示数据'}</div>
        {/if}
      </div>
    </main>

    <aside class="table-workspace__details table-workspace__inspector">
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
    <div class="table-workspace__pagination"><button type="button" title="上一页" disabled={currentPage <= 1 || isLoading} on:click={() => runDefaultQuery(currentPage - 1)}>‹</button><span>{currentPage}</span><button type="button" title="下一页" disabled={filteredRows.length < pageSize || isLoading} on:click={() => runDefaultQuery(currentPage + 1)}>›</button><select aria-label="每页显示条数" bind:value={pageSize} on:change={() => runDefaultQuery(1)}><option value={25}>25 / 页</option><option value={50}>50 / 页</option><option value={100}>100 / 页</option><option value={200}>200 / 页</option></select></div>
    <span>{selectedValue === null || selectedValue === undefined ? '未选择单元格' : `当前值：${String(selectedValue)}`}</span>
  </footer>

  {#if contextMenu}
    <div class="table-workspace__context-menu" style={`left:${contextMenu.x}px; top:${contextMenu.y}px;`} role="menu"><button type="button" on:click={() => copyText(contextMenu.row.join('\t'))}>复制行</button><button type="button" on:click={() => copyText(buildInsertSQL(mutationInput(contextMenu.row)))}>复制为 INSERT</button><button type="button" disabled={isMutating} on:click={() => requestDelete(contextMenu.row)}>删除记录</button></div>
  {/if}

  <ConfirmDialog
    isOpen={Boolean(pendingDelete)}
    title="删除记录"
    message={pendingDelete?.message || ''}
    confirmText="删除"
    type="danger"
    onConfirm={deleteRow}
    onCancel={() => pendingDelete = null}
  />
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
  .table-workspace__tool--active { color: #1586d1; border-color: #1586d1; background: color-mix(in srgb, #1586d1 8%, transparent); }
  .table-workspace__tool:disabled { opacity: .5; cursor: not-allowed; }
  .table-workspace__divider { width: 1px; height: 20px; margin: 0 5px; background: var(--border-primary); }
  .table-workspace__toolbar-title { margin-left: 5px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-secondary); font-size: 12px; }
  .table-workspace__filter { margin-left: auto; width: min(260px, 35%); height: 28px; padding: 0 8px; display: flex; align-items: center; gap: 6px; border: 1px solid var(--border-primary); background: var(--bg-primary); color: var(--text-secondary); }
  .table-workspace__filter:focus-within { border-color: #1586d1; }
  .table-workspace__filter input { min-width: 0; flex: 1; border: 0; outline: 0; background: transparent; color: inherit; font-size: 12px; }
  .table-workspace__query-builder { padding: 12px; display: grid; gap: 14px; border-bottom: 1px solid var(--border-primary); background: var(--bg-secondary); }
  .table-workspace__query-builder-head, .table-workspace__query-builder-section-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
  .table-workspace__query-builder-head > div:first-child { display: grid; gap: 3px; }
  .table-workspace__query-builder-head strong, .table-workspace__query-builder-section-head strong { font-size: 13px; }
  .table-workspace__query-builder-head span { color: var(--text-secondary); font-size: 11px; }
  .table-workspace__query-builder-actions, .table-workspace__query-builder-section-head { display: flex; align-items: center; gap: 8px; }
  .table-workspace__query-builder button { min-height: 26px; padding: 0 9px; border: 1px solid var(--border-primary); border-radius: 3px; background: var(--bg-primary); color: var(--text-primary); cursor: pointer; font-size: 12px; }
  .table-workspace__query-builder button:hover:not(:disabled) { border-color: #1586d1; color: #1586d1; }
  .table-workspace__query-builder button:disabled { opacity: .5; cursor: not-allowed; }
  .table-workspace__query-builder .table-workspace__query-builder-apply { border-color: #1687d4; background: #1687d4; color: #fff; }
  .table-workspace__query-builder-section { display: grid; gap: 8px; }
  .table-workspace__query-builder-rules { display: grid; gap: 6px; }
  .table-workspace__filter-rule, .table-workspace__sort-rule { display: grid; align-items: center; gap: 7px; }
  .table-workspace__filter-rule { grid-template-columns: 72px minmax(120px, 1fr) 128px minmax(160px, 2fr) 28px; }
  .table-workspace__sort-rule { grid-template-columns: 72px minmax(120px, 1fr) 100px 28px; max-width: 500px; }
  .table-workspace__filter-rule select, .table-workspace__filter-rule input, .table-workspace__filter-rule textarea, .table-workspace__sort-rule select { box-sizing: border-box; width: 100%; min-width: 0; min-height: 28px; padding: 4px 7px; border: 1px solid var(--border-primary); border-radius: 3px; outline: 0; background: var(--bg-primary); color: var(--text-primary); font-size: 12px; }
  .table-workspace__filter-rule textarea { height: 46px; resize: vertical; }
  .table-workspace__filter-rule select:focus, .table-workspace__filter-rule input:focus, .table-workspace__filter-rule textarea:focus, .table-workspace__sort-rule select:focus { border-color: #1586d1; }
  .table-workspace__rule-connector, .table-workspace__sort-rule > span, .table-workspace__rule-null-value { color: var(--text-secondary); font-size: 11px; }
  .table-workspace__rule-remove { width: 28px; padding: 0 !important; font-size: 17px !important; line-height: 1; }
  .table-workspace__query-builder-empty { margin: 0; padding: 8px 9px; border: 1px dashed var(--border-primary); color: var(--text-secondary); font-size: 12px; }
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
  .table-workspace__grid-head, .table-workspace__grid-row { display: grid; min-height: 32px; }
  .table-workspace__grid-head { position: sticky; top: 0; z-index: 2; background: var(--bg-secondary); border-bottom: 1px solid var(--border-primary); }
  .table-workspace__grid-row { border-bottom: 1px solid var(--border-primary); }
  .table-workspace__grid-row:hover { background: var(--bg-hover); }
  .table-workspace__grid-row--selected { background: var(--bg-active); }
  .table-workspace__row-number { position: sticky; left: 0; z-index: 1; padding: 8px 9px; text-align: right; color: var(--text-secondary); background: var(--bg-secondary); border-right: 1px solid var(--border-primary); font: 11px ui-monospace, SFMono-Regular, Menlo, monospace; }
  .table-workspace__grid-head .table-workspace__row-number { z-index: 3; }
  .table-workspace__column { position: relative; min-width: 0; min-height: 54px; border-right: 1px solid var(--border-primary); }
  .table-workspace__column-sort { width: 100%; min-height: 54px; padding: 6px 10px; display: grid; align-content: center; gap: 3px; overflow: hidden; border: 0; background: transparent; color: var(--text-secondary); text-align: left; cursor: pointer; font-size: 12px; font-weight: 650; }
  .table-workspace__column-name { overflow: hidden; color: var(--text-primary); text-overflow: ellipsis; white-space: nowrap; }
  .table-workspace__column-metadata { display: flex; gap: 6px; overflow: hidden; color: var(--text-secondary); font-size: 10px; font-weight: 500; line-height: 1.25; white-space: nowrap; }
  .table-workspace__column-metadata span { overflow: hidden; text-overflow: ellipsis; }
  .table-workspace__column:hover .table-workspace__column-sort, .table-workspace__column--sorted .table-workspace__column-sort { color: #1586d1; background: color-mix(in srgb, #1586d1 6%, transparent); }
  .table-workspace__column-resizer { position: absolute; top: 0; right: -5px; z-index: 4; width: 9px; height: 100%; padding: 0; border: 0; background: transparent; cursor: col-resize; }
  .table-workspace__column-resizer:hover, .table-workspace__column-resizer:focus-visible { background: color-mix(in srgb, #1586d1 42%, transparent); outline: 0; }
  .table-workspace__cell { min-width: 0; overflow: hidden; padding: 8px 10px; border: 0; border-right: 1px solid var(--border-primary); background: transparent; color: var(--text-primary); text-align: left; text-overflow: ellipsis; white-space: nowrap; cursor: cell; font-size: 12px; }
  .table-workspace__cell--active { outline: 2px solid var(--accent-primary); outline-offset: -2px; background: var(--bg-active); }
  .table-workspace__cell--null::placeholder,
  .table-workspace__cell--empty::placeholder {
    color: var(--text-tertiary);
    font-style: italic;
    opacity: 0.7;
  }
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
  @media (max-width: 900px) { .table-workspace__content { grid-template-columns: 1fr; } .table-workspace__details { display: none; } .table-workspace__header { padding: 0 12px; gap: 8px; } .table-workspace__identity { min-width: 0; } .table-workspace__header-meta { display: none; } .table-workspace__filter-rule { grid-template-columns: 64px minmax(100px, 1fr) 110px minmax(130px, 2fr) 28px; } }
  @media (max-width: 640px) { .table-workspace__toolbar-title { display: none; } .table-workspace__filter { width: 190px; } .table-workspace__filter-rule { grid-template-columns: 1fr 1fr 28px; } .table-workspace__filter-rule > :first-child { grid-column: 1; } .table-workspace__filter-rule > :nth-child(2) { grid-column: 2; } .table-workspace__filter-rule > :nth-child(3) { grid-column: 1; } .table-workspace__filter-rule > :nth-child(4) { grid-column: 2; } .table-workspace__filter-rule > :last-child { grid-column: 3; grid-row: 1 / span 2; } .table-workspace__sort-rule { grid-template-columns: 1fr 1fr 28px; } }

  /* 查询台账：让上下文、操作与数据承担明确的层级。 */
  .table-workspace { background: #f7f8f5; color: #1d2935; font-family: "PingFang SC", "Hiragino Sans GB", -apple-system, BlinkMacSystemFont, sans-serif; }
  .table-workspace__context { min-height: 58px; padding: 0 20px; gap: 18px; background: #fff; border-bottom-color: #d9e0e4; }
  .table-workspace__identity { min-width: 260px; }
  .table-workspace__table-icon { color: #0e6674; font-size: 20px; }
  .table-workspace__identity strong { font-size: 14px; letter-spacing: 0; }
  .table-workspace__identity span, .table-workspace__header-meta { color: #6d7783; font-size: 11px; }
  .table-workspace__modes { gap: 2px; }
  .table-workspace__mode { min-width: 54px; font-size: 12px; }
  .table-workspace__mode--active { color: #0e6674; border-bottom-color: #0e6674; }
  .table-workspace__query-strip { min-height: 48px; padding: 0 16px; gap: 6px; background: #fff; border-bottom-color: #d9e0e4; }
  .table-workspace__tool { width: 30px; height: 30px; border-radius: 4px; color: #52606d; }
  .table-workspace__tool:hover:not(:disabled) { color: #0e6674; border-color: #bdd1d4; background: #eff6f5; }
  .table-workspace__tool--primary { background: #0e6674; border-color: #0e6674; }
  .table-workspace__tool--active { color: #0e6674; border-color: #9fc5c8; background: #eff6f5; }
  .table-workspace__filter { height: 30px; border-radius: 4px; border-color: #d9e0e4; background: #f7f8f5; }
  .table-workspace__content { grid-template-columns: minmax(0, 1fr) 250px; background: #fff; }
  .table-workspace__grid-head { background: #f4f6f5; border-bottom-color: #cfd8dc; }
  .table-workspace__row-number { background: #f7f8f5; border-right-color: #d9e0e4; color: #84909b; }
  .table-workspace__column { min-height: 64px; border-right-color: #d9e0e4; }
  .table-workspace__column-sort { min-height: 64px; padding: 8px 12px 7px; gap: 4px; }
  .table-workspace__column-name { font-size: 12px; }
  .table-workspace__column-metadata { color: #6d7783; font-size: 10px; }
  .table-workspace__column-accent { position: absolute; right: 0; bottom: 0; left: 0; height: 2px; background: #2f8994; }
  .table-workspace__column:nth-child(3n) .table-workspace__column-accent { background: #b58024; }
  .table-workspace__column:nth-child(4n) .table-workspace__column-accent { background: #8168a8; }
  .table-workspace__grid-row { border-bottom-color: #e1e6e8; }
  .table-workspace__grid-row:hover { background: #f1f7f6; }
  .table-workspace__cell { padding: 9px 12px; border-right-color: #e1e6e8; font: 12px "SFMono-Regular", Menlo, Consolas, monospace; }
  .table-workspace__cell--active { outline-color: #0e6674; background: #e9f4f3; }
  .table-workspace__cell { width: 100%; outline: 0; }
  .table-workspace__cell--edited { background: #fff8e8; color: #8b5e12; }
  .table-workspace__pagination { display: flex; align-items: center; gap: 5px; }
  .table-workspace__pagination button, .table-workspace__pagination select { min-width: 26px; height: 24px; border: 1px solid #d9e0e4; border-radius: 3px; background: #fff; color: #31414d; font: inherit; }
  .table-workspace__pagination select { padding: 0 5px; }
  .table-workspace__context-menu {
    position: fixed;
    z-index: 20;
    min-width: 146px;
    padding: 4px;
    border: 1px solid var(--glass-border);
    border-radius: 12px;
    background: var(--glass-bg-strong);
    box-shadow: var(--shadow-lg), var(--shadow-glass);
    backdrop-filter: blur(calc(var(--glass-blur) + 6px)) saturate(var(--glass-saturate));
    -webkit-backdrop-filter: blur(calc(var(--glass-blur) + 6px)) saturate(var(--glass-saturate));
  }
  .table-workspace__context-menu button {
    display: block;
    width: 100%;
    padding: 7px 9px;
    border: 0;
    background: transparent;
    color: var(--text-primary);
    text-align: left;
    cursor: pointer;
    font-size: 12px;
    border-radius: 8px;
  }
  .table-workspace__context-menu button:hover:not(:disabled) {
    background: var(--accent-subtle);
    color: var(--ops-signal);
  }
  .table-workspace__inspector { padding: 18px 16px; border-left-color: #d9e0e4; background: #f7f8f5; }
  .table-workspace__details h2 { margin-bottom: 10px; color: #31414d; font-size: 11px; letter-spacing: .04em; }
  .table-workspace__details section + section { margin-top: 28px; }
  .table-workspace__details dl { gap: 10px; }
  .table-workspace__details dt { color: #7b8791; }
  .table-workspace__details dd { color: #31414d; }
  .table-workspace__status { min-height: 30px; padding: 7px 14px; background: #fff; border-top-color: #d9e0e4; color: #6d7783; font: 11px "SFMono-Regular", Menlo, Consolas, monospace; }
</style>
