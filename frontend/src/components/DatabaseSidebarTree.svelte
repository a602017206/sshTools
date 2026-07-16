<script>
  import { onMount } from 'svelte';
  import { databaseObjectCategories, isPostgreSQLCompatible } from '../lib/databaseObjectTree.js';
  import { selectDatabaseNavigation } from '../stores.js';

  export let asset;

  let databases = [];
  let schemas = {};
  let objects = {};
  let expandedDatabases = new Set();
  let expandedSchemas = new Set();
  let expandedCategories = new Set();
  let loading = false;
  let errorMessage = '';
  $: sessionId = asset?.dbSessionId;
  $: databaseType = String(asset?.metadata?.db_type || asset?.dbType || '').toLowerCase();
  $: currentDatabase = asset?.metadata?.database || '';
  $: categories = databaseObjectCategories(databaseType);

  const key = (...parts) => parts.join(':');

  async function loadDatabases() {
    if (!sessionId || !window.wailsBindings) return;
    loading = true;
    errorMessage = '';
    try {
      const names = await window.wailsBindings.ListDatabases(sessionId) || [];
      databases = names.length ? names.slice().sort() : (currentDatabase ? [currentDatabase] : []);
    } catch (error) {
      errorMessage = error?.message || '加载数据库失败';
      databases = currentDatabase ? [currentDatabase] : [];
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

  onMount(loadDatabases);
</script>

<div class="database-sidebar-tree">
  <div class="database-sidebar-tree__header"><span>数据库</span><button type="button" on:click={loadDatabases}>刷新</button></div>
  {#if loading}<div class="database-sidebar-tree__hint">加载中...</div>
  {:else if errorMessage}<div class="database-sidebar-tree__error">{errorMessage}</div>
  {:else}{#each databases as database}
    <div class="database-sidebar-tree__node">
      <button type="button" class="database-sidebar-tree__row" on:click={() => toggleDatabase(database)}>{expandedDatabases.has(database) ? '⌄' : '›'} ▱ {database}</button>
      {#if expandedDatabases.has(database)}
        {#if databaseType === 'mysql'}
          {#each categories as category}{@const nodeKey = key(database, '', category.id)}
            <button type="button" class="database-sidebar-tree__row database-sidebar-tree__indent" on:click={() => toggleCategory(database, '', category)}>{expandedCategories.has(nodeKey) ? '⌄' : '›'} {category.icon} {category.label}</button>
            {#if expandedCategories.has(nodeKey)}{#each objects[nodeKey] || [] as name}<div class="database-sidebar-tree__object">{category.icon} {name}</div>{/each}{/if}
          {/each}
        {:else}
          {#each schemas[database] || [''] as schema}{@const schemaKey = key(database, schema)}
            <button type="button" class="database-sidebar-tree__row database-sidebar-tree__indent" on:click={() => toggleSchema(database, schema)}>{expandedSchemas.has(schemaKey) ? '⌄' : '›'} ▱ {schema || '默认 Schema'}</button>
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
</style>
