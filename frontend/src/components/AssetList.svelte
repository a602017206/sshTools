<script>
  import { assetsStore, groupedAssetsStore } from '../stores.js';

  export let onConnect;
  export let onAddClick;
  export let onDelete;
  export let onEdit;

  let searchTerm = '';
  let expandedGroups = new Set(['生产环境', '开发环境']);

  $: filteredAssets = $assetsStore.filter(asset => {
    const searchLower = searchTerm.toLowerCase();
    return (
      asset.name.toLowerCase().includes(searchLower) ||
      asset.host.toLowerCase().includes(searchLower) ||
      asset.group.toLowerCase().includes(searchLower) ||
      asset.username.toLowerCase().includes(searchLower)
    );
  });

  $: groupedFilteredAssets = filteredAssets.reduce((acc, asset) => {
    if (!acc[asset.group]) {
      acc[asset.group] = [];
    }
    acc[asset.group].push(asset);
    return acc;
  }, {});

  function toggleGroup(group) {
    const newExpanded = new Set(expandedGroups);
    if (newExpanded.has(group)) {
      newExpanded.delete(group);
    } else {
      newExpanded.add(group);
    }
    expandedGroups = newExpanded;
  }

  async function handleDelete(asset, event) {
    event.stopPropagation();

    if (!window.wailsBindings) {
      console.error('Wails bindings not loaded');
      assetsStore.update(assets => assets.filter(a => a.id !== asset.id));
      return;
    }

    if (!confirm(`确定要删除连接 "${asset.name}" 吗？`)) {
      return;
    }

    try {
      await window.wailsBindings.RemoveConnection(asset.id);
      assetsStore.update(assets => assets.filter(a => a.id !== asset.id));
    } catch (error) {
      console.error('Failed to delete asset:', error);
    }
  }

  function getAssetIcon(type) {
    switch (type) {
      case 'database':
        return `<svg class="w-4 h-4 text-blue-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
        </svg>`;
      case 'docker':
        return `<svg class="w-4 h-4 text-cyan-600 flex-shrink-0" fill="currentColor" viewBox="0 0 24 24">
          <path d="M13.983 11.078h2.119a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186h-2.119a.185.185 0 00-.185.185v1.888c0 .102.083.185.185.185m-2.954-3.333h2.118a.186.186 0 00.186-.186V5.671a.186.186 0 00-.186-.185h-2.118a.185.185 0 00-.185.185v1.888c0 .102.082.185.185.185m-2.954 3.333h2.118a.186.186 0 00.186-.185V9.006a.186.186 0 00-.186-.186H8.075a.186.186 0 00-.186.186v1.888c0 .102.083.185.186.185m-2.954-3.333h2.119a.186.186 0 00.185-.186V5.671a.185.185 0 00-.185-.185H5.12a.186.186 0 00-.186.185v1.888c0 .102.084.185.186.185m-2.93 3.333h2.12a.185.185 0 00.185-.185V9.006a.185.185 0 00-.186-.186h-2.12a.185.185 0 00-.184.186v1.888c0 .102.083.185.185.185M20.69 6.662c.057.16.09.331.09.51v7.9c0 3.058-2.724 4.928-8.78 4.928-6.055 0-8.779-1.87-8.779-4.928v-7.9c0-.179.033-.35.09-.51C1.536 7.396 0 9.522 0 12.072v3.639c0 4.072 3.608 6.789 12 6.789 8.391 0 12-2.717 12-6.79v-3.638c0-2.55-1.536-4.677-4.31-6.41" />
        </svg>`;
      default:
        return `<svg class="w-4 h-4 text-purple-600 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
        </svg>`;
    }
  }
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
  <!-- 头部 -->
  <div class="p-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-sm font-semibold text-gray-900 dark:text-white">服务器资产</h2>
      <button 
        on:click={onAddClick}
        class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
      >
        <svg class="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
      </button>
    </div>
    
    <!-- 搜索框 -->
    <div class="relative">
      <svg class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
      <input
        type="text"
        placeholder="搜索服务器..."
        bind:value={searchTerm}
        class="w-full pl-9 pr-3 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm text-gray-900 dark:text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
      />
    </div>
  </div>

  <!-- 资产列表 -->
  <div class="flex-1 overflow-y-auto scrollbar-thin">
    {#each Object.entries(groupedFilteredAssets) as [group, groupAssets]}
      <div class="mb-1">
        <!-- 分组头部 -->
        <div
          on:click={() => toggleGroup(group)}
          class="flex items-center gap-2 px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer transition-colors"
        >
          {#if expandedGroups.has(group)}
            <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          {:else}
            <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          {/if}
          <svg class="w-4 h-4 text-amber-500" fill="currentColor" viewBox="0 0 24 24">
            <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{group}</span>
          <span class="ml-auto text-xs text-gray-400 dark:text-gray-500">({groupAssets.length})</span>
        </div>

        <!-- 分组内的服务器 -->
        {#if expandedGroups.has(group)}
          <div class="ml-4">
            {#each groupAssets as asset (asset.id)}
              <div
                on:click={() => onConnect(asset)}
                class="group flex items-center gap-2 px-3 py-2.5 hover:bg-purple-50 dark:hover:bg-purple-900/20 rounded-lg mx-2 cursor-pointer transition-all"
              >
                <div class="flex-shrink-0">
                  {@html getAssetIcon(asset.type)}
                </div>
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-medium text-gray-900 dark:text-white truncate">{asset.name}</span>
                    <div class={`w-2 h-2 rounded-full flex-shrink-0 ${
                      asset.status === 'online' ? 'bg-green-500' : 'bg-gray-300 dark:bg-gray-600'
                    }`} />
                  </div>
                  <div class="text-xs text-gray-500 dark:text-gray-400 truncate">
                    {asset.username}@{asset.host}:{asset.port}
                    {#if asset.type === 'database' && asset.dbType}
                      • {asset.dbType.toUpperCase()}
                    {/if}
                  </div>
                </div>
                <div class="opacity-0 group-hover:opacity-100 flex gap-1 transition-opacity">
                  <button
                    class="p-1 hover:bg-green-100 dark:hover:bg-green-800 rounded"
                    on:click|stopPropagation={() => onConnect(asset)}
                    title="连接"
                  >
                    <svg class="w-3 h-3 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                    </svg>
                  </button>
                  <button class="p-1 hover:bg-purple-100 dark:hover:bg-purple-800 rounded" on:click|stopPropagation={() => onEdit(asset)}>
                    <svg class="w-3 h-3 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                    </svg>
                  </button>
                  <button class="p-1 hover:bg-red-100 dark:hover:bg-red-800 rounded" on:click={(e) => handleDelete(asset, e)}>
                    <svg class="w-3 h-3 text-red-600 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>
    {/each}
  </div>
</div>
