<script>
  import { onMount } from 'svelte';

  export let sessionId = null;
  export let dbConfig = null;

  let schemas = [];
  let expanded = new Set();
  let tables = {};
  let loading = false;
  let errorMessage = '';
  $: databaseName = dbConfig?.metadata?.database || '';
  $: databaseType = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || 'jdbc').toUpperCase();

  async function loadSchemas() {
    if (!window.wailsBindings || !sessionId) return;
    loading = true;
    errorMessage = '';
    try {
      const names = await window.wailsBindings.ListDatabaseSchemas(sessionId, databaseName) || [];
      schemas = names.length ? names : [''];
      expanded = new Set();
      tables = {};
    } catch (error) {
      errorMessage = `加载 Schema 失败: ${error?.message || String(error || '未知错误')}`;
    } finally { loading = false; }
  }

  async function toggleSchema(schema) {
    const next = new Set(expanded);
    if (next.has(schema)) { next.delete(schema); expanded = next; return; }
    try {
      if (!tables[schema]) {
        tables = { ...tables, [schema]: await window.wailsBindings.ListDatabaseTablesInSchema(sessionId, databaseName, schema) || [] };
      }
      next.add(schema);
      expanded = next;
    } catch (error) {
      errorMessage = `加载表失败: ${error?.message || String(error || '未知错误')}`;
    }
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
        {#if expanded.has(schema)}<div class="ml-5"><div class="px-2 py-1 text-gray-500">▦ 表 ({tables[schema]?.length || 0})</div>{#each tables[schema] || [] as table}<div class="px-2 py-1 hover:bg-blue-50 dark:hover:bg-blue-900/30">▦ {table}</div>{/each}</div>{/if}
      </div>
    {/each}{/if}
  </div>
</div>
