<script>
  import { onDestroy, onMount } from 'svelte';
  import { databaseSidebarCategories, isPostgreSQLCompatible } from '../lib/databaseObjectTree.js';
  import { assetSupportsJdbcSidebar } from '../lib/nativeDatabaseTypes.js';
  import { selectDatabaseNavigation } from '../stores.js';
  import { databaseSchemaMenuItems } from '../lib/databaseSchemaMenu.js';
  import { portalToBody, resolveContextMenuPoint } from '../lib/contextMenu.js';

  export let asset;

  let databases = [];
  let schemas = {};
  let objects = {};
  let expandedDatabases = new Set();
  let expandedSchemas = new Set();
  let expandedCategories = new Set();
  let loading = false;
  let errorMessage = '';
  let contextMenu = null;
  $: sessionId = asset?.dbSessionId;
  $: databaseType = String(asset?.metadata?.db_type || asset?.dbType || '').toLowerCase();
  $: currentDatabase = asset?.metadata?.database || '';
  $: categories = databaseSidebarCategories(databaseType);
  $: menuItems = databaseSchemaMenuItems();
  $: jdbcSidebar = assetSupportsJdbcSidebar(asset);

  const key = (...parts) => parts.join(':');

  async function loadDatabases() {
    if (!jdbcSidebar) {
      errorMessage = '该连接类型请双击打开专用工作区，不支持库列表展开';
      databases = [];
      return;
    }
    if (!sessionId || !window.wailsBindings) return;
    loading = true;
    errorMessage = '';
    try {
      const names = await window.wailsBindings.ListDatabases(sessionId) || [];
      databases = names.length ? names.slice().sort() : (currentDatabase ? [currentDatabase] : []);
    } catch (error) {
      databases = currentDatabase ? [currentDatabase] : [];
      errorMessage = currentDatabase ? '' : (error?.message || '加载数据库失败');
    } finally { loading = false; }
  }

  async function toggleDatabase(database) {
    const next = new Set(expandedDatabases);
    if (next.has(database)) { next.delete(database); expandedDatabases = next; return; }
    next.add(database);
    expandedDatabases = next;
    selectDatabaseNavigation(sessionId, database);
    if (isPostgreSQLCompatible(databaseType) || databaseType !== 'mysql') {
      try {
        schemas = { ...schemas, [database]: schemas[database] || await window.wailsBindings.ListDatabaseSchemas(sessionId, database) || [''] };
      } catch (error) { errorMessage = error?.message || '加载 Schema 失败'; }
    }
  }

  async function toggleSchema(database, schema) {
    const nodeKey = key(database, schema);
    const next = new Set(expandedSchemas);
    next.has(nodeKey) ? next.delete(nodeKey) : next.add(nodeKey);
    expandedSchemas = next;
    selectDatabaseNavigation(sessionId, database, schema);
  }

  async function toggleCategory(database, schema, category) {
    const nodeKey = key(database, schema, category.id);
    const next = new Set(expandedCategories);
    if (next.has(nodeKey)) { next.delete(nodeKey); expandedCategories = next; return; }
    next.add(nodeKey); expandedCategories = next;
    if (objects[nodeKey]) return;
    try {
      const names = category.types
        ? await window.wailsBindings.ListDatabaseObjects(sessionId, database, schema, category.types)
        : await window.wailsBindings.ListDatabaseRoutines(sessionId, database, schema, category.functions);
      objects = { ...objects, [nodeKey]: names || [] };
    } catch (error) { errorMessage = error?.message || `加载${category.label}失败`; }
  }

  function closeContextMenu() {
    contextMenu = null;
  }

  function openNodeMenu(event, databaseName, schemaName = '') {
    selectDatabaseNavigation(sessionId, databaseName, schemaName);
    contextMenu = {
      databaseName,
      schemaName,
      ...resolveContextMenuPoint(event, { menuWidth: 200, menuHeight: 168 })
    };
  }

  function handleMenuAction(item) {
    const target = contextMenu;
    closeContextMenu();
    if (!target || !item?.id) return;
    if (item.id === 'new-query') {
      window.dispatchEvent(new CustomEvent('database:new-query', {
        detail: { sessionId, databaseName: target.databaseName, schemaName: target.schemaName, initialQuery: '' }
      }));
      return;
    }
    if (item.id === 'refresh') {
      objects = {};
      schemas = {};
      loadDatabases();
      return;
    }
    if (item.id === 'disconnect') {
      window.dispatchEvent(new CustomEvent('database:disconnect', { detail: { asset } }));
      return;
    }
    if (item.id === 'run-sql-file') {
      runSQLFile(target.databaseName, target.schemaName);
    }
  }

  async function runSQLFile(databaseName, schemaName) {
    const api = window.wailsBindings || {};
    if (typeof api.SelectSQLFile !== 'function' || typeof api.StartSQLFile !== 'function') {
      errorMessage = '运行 SQL 文件不可用';
      return;
    }
    try {
      const path = await api.SelectSQLFile();
      if (!path) return;
      await api.StartSQLFile(sessionId, path, databaseName || '', schemaName || '');
    } catch (error) {
      errorMessage = error?.message || String(error || '运行 SQL 文件失败');
    }
  }

  function handleDocumentClick() {
    closeContextMenu();
  }

  onMount(loadDatabases);
  onDestroy(closeContextMenu);
</script>

<svelte:window on:click={handleDocumentClick} on:contextmenu={closeContextMenu} />

<div class="database-sidebar-tree" on:contextmenu|stopPropagation>
  <div class="database-sidebar-tree__header"><span>数据库</span><button type="button" on:click={loadDatabases}>刷新</button></div>
  {#if loading}<div class="database-sidebar-tree__hint">加载中...</div>
  {:else if errorMessage}<div class="database-sidebar-tree__error">{errorMessage}</div>
  {:else}{#each databases as database}
    <div class="database-sidebar-tree__node">
      <button type="button" class="database-sidebar-tree__row" on:click={() => toggleDatabase(database)} on:contextmenu={(event) => openNodeMenu(event, database)}>{expandedDatabases.has(database) ? '⌄' : '›'} ▱ {database}</button>
      {#if expandedDatabases.has(database)}
        {#if databaseType === 'mysql'}
          {#each categories as category}{@const nodeKey = key(database, '', category.id)}
            <button type="button" class="database-sidebar-tree__row database-sidebar-tree__indent" on:click={() => toggleCategory(database, '', category)}>{expandedCategories.has(nodeKey) ? '⌄' : '›'} {category.icon} {category.label}</button>
            {#if expandedCategories.has(nodeKey)}{#each objects[nodeKey] || [] as name}<div class="database-sidebar-tree__object">{category.icon} {name}</div>{/each}{/if}
          {/each}
        {:else}
          {#each schemas[database] || [''] as schema}{@const schemaKey = key(database, schema)}
            <button type="button" class="database-sidebar-tree__row database-sidebar-tree__indent" on:click={() => toggleSchema(database, schema)} on:contextmenu={(event) => openNodeMenu(event, database, schema)}>{expandedSchemas.has(schemaKey) ? '⌄' : '›'} ▱ {schema || '默认 Schema'}</button>
            {#if expandedSchemas.has(schemaKey)}{#each categories as category}{@const nodeKey = key(database, schema, category.id)}
              <button type="button" class="database-sidebar-tree__row database-sidebar-tree__indent2" on:click={() => toggleCategory(database, schema, category)}>{expandedCategories.has(nodeKey) ? '⌄' : '›'} {category.icon} {category.label}</button>
              {#if expandedCategories.has(nodeKey)}{#each objects[nodeKey] || [] as name}<div class="database-sidebar-tree__object database-sidebar-tree__indent3">{category.icon} {name}</div>{/each}{/if}
            {/each}{/if}
          {/each}
        {/if}
      {/if}
    </div>
  {/each}{/if}
</div>

{#if contextMenu}
  <div class="database-sidebar-tree__menu" style={`left:${contextMenu.x}px; top:${contextMenu.y}px;`} use:portalToBody on:click|stopPropagation role="menu">
    {#each menuItems as item}
      <button type="button" class:database-sidebar-tree__menu-danger={item.danger} on:click={() => handleMenuAction(item)}>{item.label}</button>
    {/each}
  </div>
{/if}

<style>
  .database-sidebar-tree { font-size: 12px; color: var(--text-primary); }
  .database-sidebar-tree__header { display:flex; justify-content:space-between; padding:6px 8px; color:var(--text-secondary); }
  .database-sidebar-tree__header button { border:0; background:transparent; color:inherit; cursor:pointer; }
  .database-sidebar-tree__row { display:block; width:100%; border:0; background:transparent; color:inherit; text-align:left; padding:4px 6px; border-radius:4px; cursor:pointer; }
  .database-sidebar-tree__row:hover { background:var(--bg-secondary); }
  .database-sidebar-tree__indent { padding-left:18px; }
  .database-sidebar-tree__indent2 { padding-left:32px; }
  .database-sidebar-tree__indent3 { padding-left:46px; }
  .database-sidebar-tree__object { padding:3px 6px 3px 32px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
  .database-sidebar-tree__hint, .database-sidebar-tree__error { padding:6px 8px; color:var(--text-secondary); }
  .database-sidebar-tree__error { color:#dc2626; }
  .database-sidebar-tree__menu { position:fixed; z-index:130; min-width:184px; padding:6px; border-radius:12px; border:1px solid var(--glass-border); background: var(--bg-primary); box-shadow: 0 12px 32px rgba(0,0,0,.18); }
  .database-sidebar-tree__menu button { display:block; width:100%; text-align:left; border:0; background:transparent; color:inherit; padding:7px 10px; border-radius:8px; font-size:12px; cursor:pointer; }
  .database-sidebar-tree__menu button:hover { background: var(--bg-secondary); }
  .database-sidebar-tree__menu-danger { color:#dc2626; }
</style>
