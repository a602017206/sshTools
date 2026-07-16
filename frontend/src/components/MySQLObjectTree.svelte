<script>
  import { onMount } from 'svelte';
  import { buildMySQLObjectQuery, databaseObjectCategories, queryFirstColumn } from '../lib/databaseObjectTree.js';

  export let sessionId = null;
  export let dbConfig = null;

  let databases = [];
  let expandedDatabases = new Set();
  let expandedCategories = new Set();
  let objects = {};
  let loading = false;
  let errorMessage = '';
  const categories = databaseObjectCategories('mysql');

  async function queryNames(sql) {
    const response = await window.wailsBindings.ExecuteDatabaseQuery(sessionId, sql);
    return queryFirstColumn(response);
  }

  async function loadDatabases() {
    if (!window.wailsBindings || !sessionId) return;
    loading = true;
    errorMessage = '';
    try {
      databases = (await window.wailsBindings.ListDatabases(sessionId) || []).slice().sort();
      expandedDatabases = new Set();
      expandedCategories = new Set();
      objects = {};
    } catch (error) {
      errorMessage = `加载数据库失败: ${error?.message || String(error || '未知错误')}`;
    } finally {
      loading = false;
    }
  }

  function key(database, category) { return `${database}:${category}`; }

  function toggleDatabase(database) {
    const next = new Set(expandedDatabases);
    next.has(database) ? next.delete(database) : next.add(database);
    expandedDatabases = next;
  }

  async function toggleCategory(database, category) {
    const nodeKey = key(database, category);
    const next = new Set(expandedCategories);
    if (next.has(nodeKey)) {
      next.delete(nodeKey);
      expandedCategories = next;
      return;
    }
    next.add(nodeKey);
    expandedCategories = next;
    if (objects[nodeKey]) return;
    try {
      objects = { ...objects, [nodeKey]: await queryNames(buildMySQLObjectQuery(database, category)) };
    } catch (error) {
      errorMessage = `加载对象失败: ${error?.message || String(error || '未知错误')}`;
    }
  }

  function showStructure(databaseName, tableName) {
    window.dispatchEvent(new CustomEvent('database:table-structure', { detail: { sessionId, databaseName, tableName } }));
  }

  function openTableData(databaseName, tableName) {
    window.dispatchEvent(new CustomEvent('database:table-select', { detail: { sessionId, databaseName, tableName } }));
  }

  onMount(loadDatabases);
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800">
  <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
    <div><div class="text-sm font-semibold text-gray-900 dark:text-white">{dbConfig?.name || 'MySQL'}</div><div class="text-xs text-gray-500">MYSQL 对象浏览</div></div>
    <button class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700" on:click={loadDatabases}>刷新</button>
  </div>
  {#if errorMessage}<div class="px-4 py-2 text-xs text-red-600 bg-red-50">{errorMessage}</div>{/if}
  <div class="flex-1 overflow-auto p-3 text-sm">
    {#if loading}<div class="text-gray-500">加载数据库...</div>
    {:else if databases.length === 0}<div class="text-gray-500">暂无数据库</div>
    {:else}{#each databases as database}
      <div>
        <button class="w-full text-left px-2 py-1 hover:bg-gray-100 dark:hover:bg-gray-700" on:click={() => toggleDatabase(database)}>{expandedDatabases.has(database) ? '⌄' : '›'}  ▱ {database}</button>
        {#if expandedDatabases.has(database)}<div class="ml-5">
          {#each categories as category}
            {@const nodeKey = key(database, category.id)}
            <button class="w-full text-left px-2 py-1 hover:bg-gray-100 dark:hover:bg-gray-700" on:click={() => toggleCategory(database, category.id)}>{expandedCategories.has(nodeKey) ? '⌄' : '›'}  {category.icon} {category.label}{objects[nodeKey] ? ` (${objects[nodeKey].length})` : ''}</button>
            {#if expandedCategories.has(nodeKey)}<div class="ml-5">{#each objects[nodeKey] || [] as name}<button class="w-full text-left px-2 py-1 hover:bg-blue-50 dark:hover:bg-blue-900/30" on:click={() => category.id === 'tables' && showStructure(database, name)} on:dblclick={() => category.id === 'tables' && openTableData(database, name)}>{category.icon} {name}</button>{/each}</div>{/if}
          {/each}
        </div>{/if}
      </div>
    {/each}{/if}
  </div>
</div>
