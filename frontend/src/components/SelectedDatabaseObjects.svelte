<script>
  import { databaseNavigationStore } from '../stores.js';

  export let sessionId = null;
  export let dbConfig = null;

  let objects = {};
  let loading = false;
  let errorMessage = '';
  $: selected = $databaseNavigationStore[sessionId] || { databaseName: dbConfig?.metadata?.database || '', schemaName: '' };
  $: databaseType = dbConfig?.metadata?.db_type || dbConfig?.dbType || '';
  const categories = [
    { id: 'tables', label: '表', icon: '▦', types: ['TABLE'] },
    { id: 'views', label: '视图', icon: '◉', types: ['VIEW'] },
    { id: 'procedures', label: '存储过程', icon: '▤', functions: false },
    { id: 'functions', label: '函数', icon: 'ƒ', functions: true }
  ];
  $: selectionKey = `${selected.databaseName}:${selected.schemaName}`;
  $: if (sessionId && selected.databaseName && selectionKey !== loadedKey) { loadedKey = selectionKey; loadObjects(); }
  let loadedKey = '';

  async function loadObjects() {
    loading = true;
    errorMessage = '';
    try {
      const entries = await Promise.all(categories.map(async category => {
        const names = category.types
          ? await window.wailsBindings.ListDatabaseObjects(sessionId, selected.databaseName, selected.schemaName, category.types)
          : await window.wailsBindings.ListDatabaseRoutines(sessionId, selected.databaseName, selected.schemaName, category.functions);
        return [category.id, names || []];
      }));
      objects = Object.fromEntries(entries);
    } catch (error) { errorMessage = error?.message || '加载数据库对象失败'; } finally { loading = false; }
  }
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800">
  <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center"><div><div class="text-sm font-semibold">{selected.databaseName || '请选择数据库'}</div><div class="text-xs text-gray-500">{String(databaseType).toUpperCase()} 对象列表</div></div><button class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700" on:click={loadObjects}>刷新</button></div>
  {#if errorMessage}<div class="px-4 py-2 text-xs text-red-600 bg-red-50">{errorMessage}</div>{/if}
  <div class="flex-1 overflow-auto p-4 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
    {#if loading}<div class="text-sm text-gray-500">加载对象...</div>
    {:else if !selected.databaseName}<div class="text-sm text-gray-500">请在左侧选择数据库</div>
    {:else}{#each categories as category}<section><h3 class="text-sm font-semibold mb-2">{category.icon} {category.label} ({objects[category.id]?.length || 0})</h3><div class="border border-gray-200 dark:border-gray-700 rounded max-h-72 overflow-auto">{#each objects[category.id] || [] as name}<div class="px-3 py-2 text-sm border-b border-gray-100 dark:border-gray-700 last:border-b-0">{name}</div>{/each}{#if !(objects[category.id] || []).length}<div class="px-3 py-2 text-xs text-gray-500">暂无{category.label}</div>{/if}</div></section>{/each}{/if}
  </div>
</div>
