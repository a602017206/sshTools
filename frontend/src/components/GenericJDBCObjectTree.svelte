<script>
  import { onMount } from 'svelte';

  export let sessionId = null;
  export let dbConfig = null;

  let schemas = [];
  let expanded = new Set();
  let objects = {};
  let loading = false;
  let errorMessage = '';
  $: databaseName = dbConfig?.metadata?.database || '';
  $: databaseType = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || 'jdbc').toUpperCase();
  const categories = [
    { id: 'tables', label: '表', icon: '▦', types: ['TABLE'] },
    { id: 'views', label: '视图', icon: '◉', types: ['VIEW'] },
    { id: 'system-tables', label: '系统表', icon: '▦', types: ['SYSTEM TABLE'] },
    { id: 'procedures', label: '存储过程', icon: '▤', functions: false },
    { id: 'functions', label: '函数', icon: 'ƒ', functions: true }
  ];

  async function loadSchemas() {
    if (!window.wailsBindings || !sessionId) return;
    loading = true;
    errorMessage = '';
    try {
      const names = await window.wailsBindings.ListDatabaseSchemas(sessionId, databaseName) || [];
      schemas = names.length ? names : [''];
      expanded = new Set();
      objects = {};
    } catch (error) {
      errorMessage = `加载 Schema 失败: ${error?.message || String(error || '未知错误')}`;
    } finally { loading = false; }
  }

  async function toggleSchema(schema) {
    const next = new Set(expanded);
    if (next.has(schema)) { next.delete(schema); expanded = next; return; }
    try {
      next.add(schema);
      expanded = next;
    } catch (error) {
      errorMessage = `加载表失败: ${error?.message || String(error || '未知错误')}`;
    }
  }

  function key(schema, category) { return `${schema}:${category}`; }

  async function toggleCategory(schema, category) {
    const nodeKey = key(schema, category.id);
    if (objects[nodeKey]) {
      objects = { ...objects, [nodeKey]: null };
      return;
    }
    try {
      const names = category.types
        ? await window.wailsBindings.ListDatabaseObjects(sessionId, databaseName, schema, category.types)
        : await window.wailsBindings.ListDatabaseRoutines(sessionId, databaseName, schema, category.functions);
      objects = { ...objects, [nodeKey]: names || [] };
    } catch (error) {
      errorMessage = `加载${category.label}失败: ${error?.message || String(error || '未知错误')}`;
    }
  }

  function showStructure(schema, tableName) {
    window.dispatchEvent(new CustomEvent('database:table-structure', {
      detail: { sessionId, databaseName, schemaName: schema, tableName }
    }));
  }

  onMount(loadSchemas);
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800">
  <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
    <div><div class="text-sm font-semibold text-gray-900 dark:text-white">{dbConfig?.name || databaseName || databaseType}</div><div class="text-xs text-gray-500">{databaseType} 对象浏览</div></div>
    <button class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700" on:click={loadSchemas}>刷新</button>
  </div>
  {#if errorMessage}<div class="px-4 py-2 text-xs text-red-600 bg-red-50">{errorMessage}</div>{/if}
  <div class="flex-1 overflow-auto p-3 text-sm">
    {#if loading}<div class="text-gray-500">加载 Schema...</div>
    {:else}{#each schemas as schema}
      <div>
        <button class="w-full text-left px-2 py-1 hover:bg-gray-100 dark:hover:bg-gray-700" on:click={() => toggleSchema(schema)}>{expanded.has(schema) ? '⌄' : '›'}  ▱ {schema || '默认 Schema'}</button>
        {#if expanded.has(schema)}<div class="ml-5">{#each categories as category}{@const nodeKey = key(schema, category.id)}<button class="w-full text-left px-2 py-1 hover:bg-gray-100 dark:hover:bg-gray-700" on:click={() => toggleCategory(schema, category)}>{objects[nodeKey] ? '⌄' : '›'} {category.icon} {category.label}{objects[nodeKey] ? ` (${objects[nodeKey].length})` : ''}</button>{#if objects[nodeKey]}<div class="ml-5">{#each objects[nodeKey] as object}{#if category.id === 'tables'}<button class="w-full text-left px-2 py-1 hover:bg-blue-50 dark:hover:bg-blue-900/30" on:click={() => showStructure(schema, object)}>{category.icon} {object}</button>{:else}<div class="px-2 py-1 hover:bg-blue-50 dark:hover:bg-blue-900/30">{category.icon} {object}</div>{/if}{/each}</div>{/if}{/each}</div>{/if}
      </div>
    {/each}{/if}
  </div>
</div>
