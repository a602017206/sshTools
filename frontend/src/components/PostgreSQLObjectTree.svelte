<script>
  import { onMount } from 'svelte';
  import { buildPostgreSQLObjectQuery, buildPostgreSQLSchemaQuery, databaseObjectCategories, queryFirstColumn } from '../lib/databaseObjectTree.js';

  export let sessionId = null;
  export let dbConfig = null;

  let schemas = [];
  let expandedSchemas = new Set();
  let expandedCategories = new Set();
  let objects = {};
  let loading = false;
  let errorMessage = '';
  $: databaseType = dbConfig?.metadata?.db_type || dbConfig?.dbType || 'postgresql';
  $: databaseName = dbConfig?.metadata?.database || '当前数据库';
  $: categories = databaseObjectCategories(databaseType);

  async function queryNames(sql) {
    const response = await window.wailsBindings.ExecuteDatabaseQuery(sessionId, sql);
    return queryFirstColumn(response);
  }

  async function loadSchemas() {
    if (!window.wailsBindings || !sessionId) return;
    loading = true;
    errorMessage = '';
    try {
      schemas = await queryNames(buildPostgreSQLSchemaQuery());
    } catch (error) {
      errorMessage = `加载 Schema 失败: ${error?.message || String(error || '未知错误')}`;
    } finally { loading = false; }
  }

  function key(schema, category) { return `${schema}:${category}`; }

  async function toggleSchema(schema) {
    const next = new Set(expandedSchemas);
    next.has(schema) ? next.delete(schema) : next.add(schema);
    expandedSchemas = next;
  }

  async function toggleCategory(schema, category) {
    const nodeKey = key(schema, category);
    const next = new Set(expandedCategories);
    if (next.has(nodeKey)) { next.delete(nodeKey); expandedCategories = next; return; }
    next.add(nodeKey); expandedCategories = next;
    if (objects[nodeKey]) return;
    try {
      objects = { ...objects, [nodeKey]: await queryNames(buildPostgreSQLObjectQuery(schema, category)) };
    } catch (error) {
      errorMessage = `加载${category}失败: ${error?.message || String(error || '未知错误')}`;
    }
  }

  function openObject(schema, category, name) {
    if (category !== 'tables') return;
    window.dispatchEvent(new CustomEvent('database:table-structure', { detail: { sessionId, databaseName, tableName: name, schemaName: schema } }));
  }

  function openTableData(schema, category, name) {
    if (category !== 'tables') return;
    window.dispatchEvent(new CustomEvent('database:table-select', { detail: { sessionId, databaseName, tableName: name, schemaName: schema } }));
  }

  onMount(loadSchemas);
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800">
  <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
    <div><div class="text-sm font-semibold text-gray-900 dark:text-white">{databaseName}</div><div class="text-xs text-gray-500">{String(databaseType).toUpperCase()} 对象浏览</div></div>
    <button class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700" on:click={loadSchemas}>刷新</button>
  </div>
  {#if errorMessage}<div class="px-4 py-2 text-xs text-red-600 bg-red-50">{errorMessage}</div>{/if}
  <div class="flex-1 overflow-auto p-3 text-sm">
    {#if loading}<div class="text-gray-500">加载 Schema...</div>
    {:else}{#each schemas as schema}
      <div>
        <button class="w-full text-left px-2 py-1 hover:bg-gray-100 dark:hover:bg-gray-700" on:click={() => toggleSchema(schema)}>{expandedSchemas.has(schema) ? '⌄' : '›'}  ▱ {schema}</button>
        {#if expandedSchemas.has(schema)}<div class="ml-5">
          {#each categories as category}
            {@const nodeKey = key(schema, category.id)}
            <button class="w-full text-left px-2 py-1 hover:bg-gray-100 dark:hover:bg-gray-700" on:click={() => toggleCategory(schema, category.id)}>{expandedCategories.has(nodeKey) ? '⌄' : '›'}  {category.icon} {category.label}{objects[nodeKey] ? ` (${objects[nodeKey].length})` : ''}</button>
            {#if expandedCategories.has(nodeKey)}<div class="ml-5">{#each objects[nodeKey] || [] as name}<button class="w-full text-left px-2 py-1 hover:bg-blue-50 dark:hover:bg-blue-900/30" on:click={() => openObject(schema, category.id, name)} on:dblclick={() => openTableData(schema, category.id, name)}>{category.icon} {name}</button>{/each}</div>{/if}
          {/each}
        </div>{/if}
      </div>
    {/each}{/if}
  </div>
</div>
