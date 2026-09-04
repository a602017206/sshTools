<script>
  import { createEventDispatcher, onDestroy } from 'svelte';
  import { databaseNavigationStore } from '../stores.js';
  import { databaseSidebarCategories, defaultDatabaseObjectCategory } from '../lib/databaseObjectTree.js';
  import { tableOpenEvents } from '../lib/databaseObjectActions.js';
  import { buildCopyTableStatements, buildDropTableSQL } from '../lib/tableObjectMutations.js';
  import { formatColumnLength, formatColumnType } from '../lib/tableStructureMetadata.js';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  import InputDialog from './ui/InputDialog.svelte';
  import { portalToBody, resolveContextMenuPoint } from '../lib/contextMenu.js';
  import { copilotStore } from '../stores/copilot.js';

  export let sessionId = null;
  export let dbConfig = null;

  let objects = {};
  let errors = {};
  let loadingCategory = '';
  let activeCategoryId = defaultDatabaseObjectCategory();
  let searchText = '';
  let selectionSeen = '';
  let selectedTable = '';
  let selectedTableSchema = null;
  let selectedTableStructureLoading = false;
  let selectedTableStructureError = '';
  let selectedTableStructureRequest = '';
  let selectedTableDDL = '';
  let selectedTableDDLLoading = false;
  let selectedTableDDLError = '';
  let selectedTableDDLRequest = '';
  let detailsVisible = true;
  let detailsMode = 'info';
  let detailsWidth = 300;
  let detailsResizeStartX = 0;
  let detailsResizeStartWidth = 0;
  let contextMenu = null;
  let pendingDropTable = '';
  let pendingCopy = null;
  let copyDialogOpen = false;
  let actionMessage = '';
  const dispatch = createEventDispatcher();

  $: selected = $databaseNavigationStore[sessionId] || { databaseName: dbConfig?.metadata?.database || '', schemaName: '' };
  $: databaseType = dbConfig?.metadata?.db_type || dbConfig?.dbType || '';
  $: categories = databaseSidebarCategories(databaseType);
  $: selectionKey = `${sessionId || ''}:${selected.databaseName}:${selected.schemaName}`;
  $: activeCategory = categories.find(category => category.id === activeCategoryId) || categories[0];
  $: activeObjects = objects[activeCategoryId] || [];
  $: filteredObjects = activeObjects.filter(name => name.toLowerCase().includes(searchText.trim().toLowerCase()));
  $: if (sessionId && selected.databaseName && selectionKey !== selectionSeen) {
    selectionSeen = selectionKey;
    activeCategoryId = defaultDatabaseObjectCategory();
    searchText = '';
    objects = {};
    errors = {};
    selectedTable = '';
    selectedTableSchema = null;
    selectedTableStructureError = '';
    selectedTableDDL = '';
    selectedTableDDLError = '';
    selectedTableDDLRequest = '';
    loadCategory(defaultDatabaseObjectCategory());
  }

  async function loadCategory(categoryID = activeCategoryId, force = false) {
    if (!sessionId || !selected.databaseName) return;
    if (!force && objects[categoryID]) return;

    const category = categories.find(item => item.id === categoryID);
    if (!category) return;
    loadingCategory = categoryID;
    errors = { ...errors, [categoryID]: '' };
    try {
      const names = category.types
        ? await window.wailsBindings.ListDatabaseObjects(sessionId, selected.databaseName, selected.schemaName, category.types)
        : await window.wailsBindings.ListDatabaseRoutines(sessionId, selected.databaseName, selected.schemaName, category.functions);
      objects = { ...objects, [categoryID]: names || [] };
    } catch (error) {
      objects = { ...objects, [categoryID]: [] };
      errors = { ...errors, [categoryID]: error?.message || String(error || `加载${category.label}失败`) };
    } finally {
      if (loadingCategory === categoryID) loadingCategory = '';
    }
  }

  function selectCategory(categoryID) {
    activeCategoryId = categoryID;
    searchText = '';
    loadCategory(categoryID);
  }

  function refresh() {
    loadCategory(activeCategoryId, true);
  }

  $: mutationSupported = ['mysql', 'postgresql', 'kingbase'].includes(String(databaseType).toLowerCase());
  $: if (sessionId) {
    copilotStore.setWorkspaceFocus(sessionId, {
      database: selected.databaseName || '',
      schema: selected.schemaName || '',
      objectKind: selectedTable ? 'table' : '',
      objectName: selectedTable || '',
      objectParent: ''
    });
  }
  $: selectedTableStructureKey = `${sessionId || ''}:${selected.databaseName || ''}:${selected.schemaName || ''}:${selectedTable || ''}`;
  $: if (!selectedTable) {
    selectedTableSchema = null;
    selectedTableStructureError = '';
    selectedTableDDL = '';
    selectedTableDDLError = '';
  } else if (selectedTableStructureKey !== selectedTableStructureRequest) {
    selectedTableStructureRequest = selectedTableStructureKey;
    loadSelectedTableStructure(selectedTableStructureKey);
  }

  $: selectedTableDDLKey = `${sessionId || ''}:${selected.databaseName || ''}:${selected.schemaName || ''}:${selectedTable || ''}`;
  $: if (detailsMode === 'ddl' && selectedTable && selectedTableDDLKey !== selectedTableDDLRequest) {
    selectedTableDDLRequest = selectedTableDDLKey;
    loadSelectedTableDDL(selectedTableDDLKey);
  }

  function tableDetail(tableName) {
    return {
      sessionId,
      databaseName: selected.databaseName,
      schemaName: selected.schemaName,
      tableName
    };
  }

  function openQuery(initialQuery = '') {
    dispatch('database:new-query', { sessionId, databaseName: selected.databaseName, schemaName: selected.schemaName, initialQuery });
  }

  function openTableDesign() {
    if (selectedTable) dispatch(tableOpenEvents.click, tableDetail(selectedTable));
  }

  function createTable() {
    dispatch('open-table-designer', {
      sessionId,
      databaseName: selected.databaseName,
      schemaName: selected.schemaName,
      tableName: 'new_table',
      mode: 'create'
    });
  }

  function openTableData(tableName) {
    dispatch(tableOpenEvents.doubleClick, tableDetail(tableName));
  }

  async function loadSelectedTableStructure(requestKey) {
    if (!window.wailsBindings?.GetTableSchemaInSchema || !sessionId || !selectedTable) return;
    selectedTableStructureLoading = true;
    selectedTableStructureError = '';
    try {
      const schema = await window.wailsBindings.GetTableSchemaInSchema(sessionId, selected.databaseName, selected.schemaName, selectedTable);
      if (requestKey === selectedTableStructureRequest) selectedTableSchema = schema || null;
    } catch (error) {
      if (requestKey === selectedTableStructureRequest) {
        selectedTableSchema = null;
        selectedTableStructureError = error?.message || String(error || '加载表结构失败');
      }
    } finally {
      if (requestKey === selectedTableStructureRequest) selectedTableStructureLoading = false;
    }
  }

  async function loadSelectedTableDDL(requestKey) {
    if (!sessionId || !selectedTable) return;
    const getDDL = selected.schemaName
      ? window.wailsBindings?.GetTableDDLInSchema
      : window.wailsBindings?.GetTableDDL;
    if (!getDDL) return;

    selectedTableDDLLoading = true;
    selectedTableDDLError = '';
    try {
      const ddl = selected.schemaName
        ? await getDDL(sessionId, selected.databaseName, selected.schemaName, selectedTable)
        : await getDDL(sessionId, selected.databaseName, selectedTable);
      if (requestKey === selectedTableDDLRequest) selectedTableDDL = ddl?.ddl || '';
    } catch (error) {
      if (requestKey === selectedTableDDLRequest) {
        selectedTableDDL = '';
        selectedTableDDLError = error?.message || String(error || '加载 DDL 失败');
      }
    } finally {
      if (requestKey === selectedTableDDLRequest) selectedTableDDLLoading = false;
    }
  }

  function selectDetailsMode(mode) {
    detailsMode = mode;
    if (mode === 'ddl' && selectedTable && selectedTableDDLKey !== selectedTableDDLRequest) {
      selectedTableDDLRequest = selectedTableDDLKey;
      loadSelectedTableDDL(selectedTableDDLKey);
    }
  }

  function startDetailsResize(event) {
    event.preventDefault();
    detailsResizeStartX = event.clientX;
    detailsResizeStartWidth = detailsWidth;
    window.addEventListener('pointermove', resizeDetails);
    window.addEventListener('pointerup', stopDetailsResize, { once: true });
  }

  function resizeDetails(event) {
    detailsWidth = Math.min(560, Math.max(220, detailsResizeStartWidth - (event.clientX - detailsResizeStartX)));
  }

  function stopDetailsResize() {
    window.removeEventListener('pointermove', resizeDetails);
  }

  onDestroy(stopDetailsResize);

  function openTableContextMenu(event, tableName) {
    selectedTable = tableName;
    actionMessage = '';
    contextMenu = { tableName, ...resolveContextMenuPoint(event, { menuWidth: 220, menuHeight: 180 }) };
  }

  function closeContextMenu() {
    contextMenu = null;
  }

  function designTable(tableName) {
    closeContextMenu();
    dispatch(tableOpenEvents.click, tableDetail(tableName));
  }

  function requestDropTable(tableName) {
    closeContextMenu();
    pendingDropTable = tableName;
  }

  async function dropTable() {
    const tableName = pendingDropTable;
    pendingDropTable = '';
    const sql = buildDropTableSQL({ databaseType, databaseName: selected.databaseName, schemaName: selected.schemaName, tableName });
    if (!sql || !window.wailsBindings?.ExecuteDatabaseQuery) return;
    actionMessage = '';
    try {
      await window.wailsBindings.ExecuteDatabaseQuery(sessionId, sql);
      if (selectedTable === tableName) selectedTable = '';
      actionMessage = `已删除表 ${tableName}`;
      await loadCategory('tables', true);
    } catch (error) {
      actionMessage = `删除表失败：${error?.message || String(error || '未知错误')}`;
    }
  }

  function requestCopyTable(tableName, includeData) {
    closeContextMenu();
    pendingCopy = { tableName, includeData };
    copyDialogOpen = true;
  }

  async function copyTable(targetTable) {
    const copy = pendingCopy;
    copyDialogOpen = false;
    pendingCopy = null;
    if (!copy) return;
    const statements = buildCopyTableStatements({
      databaseType,
      databaseName: selected.databaseName,
      schemaName: selected.schemaName,
      sourceTable: copy.tableName,
      targetTable,
      includeData: copy.includeData
    });
    if (!statements.length || !window.wailsBindings?.ExecuteDatabaseQuery) return;
    actionMessage = '';
    try {
      for (const statement of statements) await window.wailsBindings.ExecuteDatabaseQuery(sessionId, statement);
      selectedTable = targetTable;
      actionMessage = `已复制表 ${copy.tableName} 到 ${targetTable}`;
      await loadCategory('tables', true);
    } catch (error) {
      actionMessage = `复制表失败：${error?.message || String(error || '未知错误')}`;
    }
  }
</script>

<div class="object-browser" on:click={closeContextMenu}>
  <header class="object-browser__header">
    <div class="object-browser__location">
      <span class="object-browser__eyebrow">对象</span>
      <strong>{selected.databaseName || '请选择数据库'}</strong>
      {#if selected.schemaName}<span class="object-browser__schema">/ {selected.schemaName}</span>{/if}
    </div>
    <div class="object-browser__tabs" aria-label="数据库对象类型">
      {#each categories as category}
        <button
          type="button"
          class:object-browser__tab--active={activeCategoryId === category.id}
          class="object-browser__tab"
          on:click={() => selectCategory(category.id)}
        >
          <span aria-hidden="true">{category.icon}</span>{category.label}
        </button>
      {/each}
    </div>
  </header>

  <div class="object-browser__body">
    <main class="object-browser__main">
      <div class="object-browser__toolbar">
        <button type="button" class="object-browser__icon-button" title="新增查询" on:click={() => openQuery()}>＋</button>
        <button type="button" class="object-browser__icon-button" title="设计表" on:click={openTableDesign} disabled={!selectedTable}>▤</button>
        <button type="button" class="object-browser__icon-button" title="新建表" on:click={createTable} disabled={!selected.databaseName}>▦</button>
        <button type="button" class="object-browser__icon-button" title="刷新" on:click={refresh} disabled={loadingCategory === activeCategoryId}>↻</button>
        <span class="object-browser__toolbar-divider"></span>
        <span class="object-browser__toolbar-title">{activeCategory?.label}</span>
        <label class="object-browser__search">
          <span aria-hidden="true">⌕</span>
          <input bind:value={searchText} placeholder={`搜索${activeCategory?.label || '对象'}`} aria-label={`搜索${activeCategory?.label || '对象'}`} />
        </label>
        <button
          type="button"
          class:object-browser__icon-button--active={detailsVisible}
          class="object-browser__icon-button object-browser__details-toggle"
          title={detailsVisible ? '隐藏查看栏' : '显示查看栏'}
          aria-label={detailsVisible ? '隐藏查看栏' : '显示查看栏'}
          aria-pressed={detailsVisible}
          on:click={() => detailsVisible = !detailsVisible}
        >▥</button>
      </div>

      <div class="object-browser__table" role="table" aria-label={`${activeCategory?.label || '对象'}列表`}>
        <div class="object-browser__table-head" role="row">
          <span role="columnheader">名称</span>
          <span role="columnheader">类型</span>
        </div>
        {#if !selected.databaseName}
          <div class="object-browser__empty">请在左侧选择数据库</div>
        {:else if loadingCategory === activeCategoryId}
          <div class="object-browser__empty">正在加载{activeCategory.label}...</div>
        {:else if errors[activeCategoryId]}
          <div class="object-browser__error">加载失败：{errors[activeCategoryId]}</div>
        {:else if !filteredObjects.length}
          <div class="object-browser__empty">{searchText ? '没有匹配的对象' : `暂无${activeCategory.label}`}</div>
        {:else}
          {#each filteredObjects as name}
            {#if activeCategoryId === 'tables'}
              <button type="button" class:object-browser__row--selected={selectedTable === name} class="object-browser__row object-browser__row--button" role="row" on:click={() => selectedTable = name} on:dblclick|preventDefault={() => openTableData(name)} on:contextmenu={(event) => openTableContextMenu(event, name)}>
                <span role="cell"><span class="object-browser__item-icon" aria-hidden="true">▦</span>{name}</span>
                <span role="cell">表</span>
              </button>
            {:else}
              <div class="object-browser__row" role="row">
                <span role="cell"><span class="object-browser__item-icon" aria-hidden="true">{activeCategory.icon}</span>{name}</span>
                <span role="cell">{activeCategory.label}</span>
              </div>
            {/if}
          {/each}
        {/if}
      </div>
      <footer class="object-browser__status">{actionMessage || `${filteredObjects.length} 个${activeCategory?.label || '对象'}`}</footer>
    </main>

    {#if detailsVisible}
      <div class="object-browser__details-resizer" role="separator" aria-orientation="vertical" aria-label="调整查看栏宽度" on:pointerdown={startDetailsResize}></div>
      <aside class="object-browser__details" style={`width: ${detailsWidth}px;`}>
        <div class="object-browser__details-tabs" aria-label="查看内容">
          <button type="button" class:object-browser__details-tab--active={detailsMode === 'info'} class="object-browser__details-tab" title="对象信息" aria-label="对象信息" aria-pressed={detailsMode === 'info'} on:click={() => selectDetailsMode('info')}>i</button>
          <button type="button" class:object-browser__details-tab--active={detailsMode === 'ddl'} class="object-browser__details-tab object-browser__details-tab--ddl" title="DDL" aria-label="DDL" aria-pressed={detailsMode === 'ddl'} on:click={() => selectDetailsMode('ddl')}>DDL</button>
        </div>
        {#if detailsMode === 'info'}
          <div class="object-browser__database-mark" aria-hidden="true">▰</div>
          <h2>{selected.databaseName || '未选择数据库'}</h2>
          <p>数据库</p>
          <dl>
            <div><dt>类型</dt><dd>{String(databaseType || 'JDBC').toUpperCase()}</dd></div>
            <div><dt>主机</dt><dd>{dbConfig?.host || dbConfig?.metadata?.host || '当前连接'}</dd></div>
            {#if dbConfig?.port}<div><dt>端口</dt><dd>{dbConfig.port}</dd></div>{/if}
            {#if selected.schemaName}<div><dt>Schema</dt><dd>{selected.schemaName}</dd></div>{/if}
          </dl>
          {#if selectedTable}
            <section class="object-browser__table-structure" aria-label={`${selectedTable} 表结构`}>
              <div class="object-browser__table-structure-head"><h3>表结构</h3><span>{selectedTable}</span></div>
              {#if selectedTableStructureLoading}
                <p class="object-browser__table-structure-empty">正在加载字段...</p>
              {:else if selectedTableStructureError}
                <p class="object-browser__table-structure-error">加载失败：{selectedTableStructureError}</p>
              {:else if selectedTableSchema?.columns?.length}
                <div class="object-browser__column-list">
                  {#each selectedTableSchema.columns as column}
                    <div class="object-browser__column-row">
                      <strong title={column.name}>{column.name}</strong>
                      <span>{formatColumnType(column)}{formatColumnLength(column) !== '-' ? ` · ${formatColumnLength(column)}` : ''}</span>
                      {#if column.is_primary_key || column.primary_key || column.isPrimaryKey}<em>主键</em>{:else if !column.nullable}<em>非空</em>{/if}
                    </div>
                  {/each}
                </div>
              {:else}
                <p class="object-browser__table-structure-empty">暂无字段信息</p>
              {/if}
            </section>
          {/if}
        {:else if !selectedTable}
          <p class="object-browser__ddl-empty">请选择一张表以查看 DDL</p>
        {:else if selectedTableDDLLoading}
          <p class="object-browser__ddl-empty">正在加载 DDL...</p>
        {:else if selectedTableDDLError}
          <p class="object-browser__ddl-error">加载失败：{selectedTableDDLError}</p>
        {:else if selectedTableDDL}
          <section class="object-browser__ddl" aria-label={`${selectedTable} DDL`}>
            <div class="object-browser__ddl-head"><h2>{selectedTable}</h2><span>DDL</span></div>
            <pre>{selectedTableDDL}</pre>
          </section>
        {:else}
          <p class="object-browser__ddl-empty">暂无 DDL</p>
        {/if}
      </aside>
    {/if}
  </div>

  {#if contextMenu}
    <div class="object-browser__context-menu" style={`left: ${contextMenu.x}px; top: ${contextMenu.y}px;`} use:portalToBody on:click|stopPropagation>
      <button type="button" on:click={() => { closeContextMenu(); openTableData(contextMenu.tableName); }}>打开表</button>
      <button type="button" on:click={() => designTable(contextMenu.tableName)}>设计表</button>
      <button type="button" on:click={createTable}>新建表</button>
      <div class="object-browser__context-divider"></div>
      <button type="button" on:click={() => requestCopyTable(contextMenu.tableName, true)} disabled={!mutationSupported}>复制表（结构和数据）</button>
      <button type="button" on:click={() => requestCopyTable(contextMenu.tableName, false)} disabled={!mutationSupported}>复制表（仅结构）</button>
      <div class="object-browser__context-divider"></div>
      <button type="button" class="object-browser__context-danger" on:click={() => requestDropTable(contextMenu.tableName)} disabled={!mutationSupported}>删除表</button>
    </div>
  {/if}

  <ConfirmDialog
    isOpen={Boolean(pendingDropTable)}
    title="删除表"
    message={`确定删除表 ${pendingDropTable} 吗？此操作不可恢复。`}
    confirmText="删除表"
    type="danger"
    onConfirm={dropTable}
    onCancel={() => pendingDropTable = ''}
  />
  <InputDialog
    isOpen={copyDialogOpen}
    title={pendingCopy?.includeData ? '复制表（结构和数据）' : '复制表（仅结构）'}
    message={`请输入 ${pendingCopy?.tableName || ''} 的新表名。`}
    defaultValue={pendingCopy ? `${pendingCopy.tableName}_copy` : ''}
    placeholder="新表名"
    confirmText="复制"
    onConfirm={copyTable}
    onCancel={() => { copyDialogOpen = false; pendingCopy = null; }}
  />
</div>

<style>
  .object-browser { height: 100%; display: flex; flex-direction: column; background: var(--bg-primary); color: var(--text-primary); }
  .object-browser__header { display: flex; align-items: end; gap: 24px; min-height: 88px; padding: 16px 24px 0; border-bottom: 1px solid var(--border-primary); }
  .object-browser__location { min-width: 160px; padding-bottom: 15px; display: flex; flex-direction: column; gap: 3px; }
  .object-browser__location strong { font-size: 16px; font-weight: 650; }
  .object-browser__eyebrow { font-size: 12px; color: var(--text-secondary); }
  .object-browser__schema { font-size: 12px; color: var(--text-secondary); }
  .object-browser__tabs { display: flex; gap: 2px; align-self: stretch; }
  .object-browser__tab { min-width: 72px; height: 56px; border: 0; border-bottom: 3px solid transparent; background: transparent; color: var(--text-secondary); font-size: 13px; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; gap: 6px; }
  .object-browser__tab:hover { color: var(--text-primary); background: var(--bg-secondary); }
  .object-browser__tab--active { color: #1687d4; border-bottom-color: #1687d4; font-weight: 600; }
  .object-browser__body { min-height: 0; flex: 1; display: flex; overflow: hidden; }
  .object-browser__main { min-width: 0; min-height: 0; flex: 1 1 0; display: flex; flex-direction: column; }
  .object-browser__toolbar { min-height: 54px; padding: 0 18px; display: flex; align-items: center; gap: 12px; border-bottom: 1px solid var(--border-primary); }
  .object-browser__icon-button { width: 28px; height: 28px; border: 0; background: transparent; color: var(--text-secondary); font-size: 21px; line-height: 1; cursor: pointer; }
  .object-browser__icon-button:hover:not(:disabled) { color: #1687d4; background: var(--bg-secondary); }
  .object-browser__icon-button:disabled { opacity: .45; cursor: wait; }
  .object-browser__icon-button--active { color: #1687d4; background: color-mix(in srgb, #1687d4 10%, transparent); }
  .object-browser__details-toggle { margin-left: 2px; font-size: 18px; }
  .object-browser__toolbar-divider { width: 1px; height: 20px; background: var(--border-primary); }
  .object-browser__toolbar-title { font-size: 14px; font-weight: 600; }
  .object-browser__search { margin-left: auto; width: min(300px, 46%); height: 32px; padding: 0 9px; display: flex; align-items: center; gap: 7px; border: 1px solid var(--border-primary); color: var(--text-secondary); }
  .object-browser__search:focus-within { border-color: #1687d4; }
  .object-browser__search input { width: 100%; border: 0; outline: 0; background: transparent; color: inherit; font-size: 13px; }
  .object-browser__table { min-height: 0; overflow-y: auto; overflow-x: auto; flex: 1; scrollbar-gutter: stable; }
  .object-browser__table::-webkit-scrollbar { width: 12px; height: 12px; }
  .object-browser__table::-webkit-scrollbar-track { background: var(--bg-secondary); border-left: 1px solid var(--border-primary); }
  .object-browser__table::-webkit-scrollbar-thumb { min-height: 40px; background: #a7b4c7; border: 3px solid var(--bg-secondary); border-radius: 8px; }
  .object-browser__table::-webkit-scrollbar-thumb:hover { background: #7f8da3; }
  .object-browser__table-head, .object-browser__row { display: grid; grid-template-columns: minmax(240px, 1fr) 140px; align-items: center; min-height: 38px; }
  .object-browser__table-head { position: sticky; top: 0; z-index: 1; color: var(--text-secondary); background: var(--bg-secondary); border-bottom: 1px solid var(--border-primary); font-size: 12px; }
  .object-browser__table-head span, .object-browser__row span { padding: 0 18px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .object-browser__row { border-bottom: 1px solid var(--border-primary); font-size: 14px; user-select: none; -webkit-user-select: none; }
  .object-browser__row--button { width: 100%; border-left: 0; border-right: 0; border-top: 0; background: transparent; color: inherit; text-align: left; cursor: pointer; }
  .object-browser__row:hover { background: color-mix(in srgb, #1687d4 7%, transparent); }
  .object-browser__row--selected { background: color-mix(in srgb, #1687d4 12%, transparent); }
  .object-browser__item-icon { display: inline-block; width: 25px; padding: 0 !important; color: #1687d4; }
  .object-browser__empty, .object-browser__error { padding: 28px 18px; font-size: 13px; color: var(--text-secondary); }
  .object-browser__error { color: #dc2626; }
  .object-browser__status { min-height: 28px; padding: 6px 18px; border-top: 1px solid var(--border-primary); color: var(--text-secondary); font-size: 12px; }
  .object-browser__context-menu {
    position: fixed;
    z-index: 120;
    width: 200px;
    padding: 5px;
    border: 1px solid var(--glass-border);
    border-radius: 12px;
    background: var(--glass-bg-strong);
    box-shadow: var(--shadow-lg), var(--shadow-glass);
    backdrop-filter: blur(calc(var(--glass-blur) + 6px)) saturate(var(--glass-saturate));
    -webkit-backdrop-filter: blur(calc(var(--glass-blur) + 6px)) saturate(var(--glass-saturate));
  }
  .object-browser__context-menu button { display: block; width: 100%; min-height: 30px; padding: 0 9px; border: 0; border-radius: 3px; background: transparent; color: var(--text-primary); cursor: pointer; font: inherit; text-align: left; }
  .object-browser__context-menu button:hover:not(:disabled) { background: color-mix(in srgb, #1687d4 14%, transparent); }
  .object-browser__context-menu button:disabled { color: var(--text-secondary); cursor: not-allowed; }
  .object-browser__context-danger { color: #b42318 !important; }
  .object-browser__context-divider { height: 1px; margin: 5px 4px; background: var(--border-primary); }
  .object-browser__details-resizer { width: 6px; flex: 0 0 6px; cursor: col-resize; border-left: 1px solid var(--border-primary); background: transparent; touch-action: none; }
  .object-browser__details-resizer:hover { background: color-mix(in srgb, #1687d4 28%, transparent); }
  .object-browser__details { box-sizing: border-box; flex: 0 0 auto; padding: 0 22px 28px; background: var(--bg-secondary); overflow: auto; }
  .object-browser__details-tabs { position: sticky; top: 0; z-index: 2; height: 48px; margin: 0 -22px 24px; padding: 0 15px; display: flex; align-items: center; justify-content: flex-end; gap: 5px; border-bottom: 1px solid var(--border-primary); background: var(--bg-secondary); }
  .object-browser__details-tab { width: 29px; height: 29px; border: 1px solid transparent; border-radius: 4px; background: transparent; color: var(--text-secondary); font-size: 16px; font-weight: 650; cursor: pointer; }
  .object-browser__details-tab:hover { color: #1687d4; background: color-mix(in srgb, #1687d4 8%, transparent); }
  .object-browser__details-tab--active { color: #1687d4; border-color: #1687d4; background: color-mix(in srgb, #1687d4 10%, transparent); }
  .object-browser__details-tab--ddl { font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 10px; }
  .object-browser__database-mark { color: #0a9d5b; font-size: 52px; line-height: 1; }
  .object-browser__details h2 { margin: 16px 0 4px; font-size: 18px; overflow-wrap: anywhere; }
  .object-browser__details > p { margin: 0 0 24px; color: var(--text-secondary); font-size: 13px; }
  .object-browser__details dl { margin: 0; display: grid; gap: 17px; }
  .object-browser__details dl div { display: grid; gap: 3px; }
  .object-browser__details dt { font-size: 12px; color: var(--text-secondary); }
  .object-browser__details dd { margin: 0; font-size: 13px; overflow-wrap: anywhere; }
  .object-browser__table-structure { margin-top: 28px; padding-top: 18px; border-top: 1px solid var(--border-primary); }
  .object-browser__table-structure-head { display: grid; gap: 3px; margin-bottom: 10px; }
  .object-browser__table-structure h3 { margin: 0; font-size: 14px; }
  .object-browser__table-structure-head span { color: var(--text-secondary); font-size: 12px; overflow-wrap: anywhere; }
  .object-browser__column-list { display: grid; border-top: 1px solid var(--border-primary); }
  .object-browser__column-row { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 3px 8px; padding: 8px 0; border-bottom: 1px solid var(--border-primary); }
  .object-browser__column-row strong { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; }
  .object-browser__column-row span { grid-column: 1; color: var(--text-secondary); font-size: 11px; }
  .object-browser__column-row em { grid-column: 2; grid-row: 1 / span 2; align-self: center; color: #0e6674; font-size: 11px; font-style: normal; white-space: nowrap; }
  .object-browser__table-structure-empty, .object-browser__table-structure-error { margin: 0; color: var(--text-secondary); font-size: 12px; }
  .object-browser__table-structure-error { color: #b42318; }
  .object-browser__ddl { min-width: 0; }
  .object-browser__ddl-head { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; margin-bottom: 12px; }
  .object-browser__ddl-head h2 { margin: 0; }
  .object-browser__ddl-head span { color: var(--text-secondary); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 11px; }
  .object-browser__ddl pre { margin: 0; padding: 12px; overflow: auto; border: 1px solid var(--border-primary); background: var(--bg-primary); color: var(--text-primary); font: 12px/1.6 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; white-space: pre-wrap; overflow-wrap: anywhere; }
  .object-browser__ddl-empty, .object-browser__ddl-error { margin: 0; color: var(--text-secondary); font-size: 13px; }
  .object-browser__ddl-error { color: #b42318; }
  @media (max-width: 900px) { .object-browser__details-resizer, .object-browser__details { display: none; } .object-browser__header { gap: 8px; padding-left: 16px; } .object-browser__location { display: none; } .object-browser__tab { min-width: 58px; } }
</style>
