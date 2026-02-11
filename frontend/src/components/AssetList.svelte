<script>
  import { assetsStore, groupedAssetsStore } from '../stores.js';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  import Dialog from './ui/Dialog.svelte';
  import InputDialog from './ui/InputDialog.svelte';

  export let onConnect;
  export let onAddClick;
  export let onDelete;
  export let onEdit;

  let searchTerm = '';
  let expandedGroups = new Set(['生产环境', '开发环境']);
  let showExportMenu = false;
  let showDeleteConfirm = false;
  let pendingDeleteAsset = null;
  let showImportConfirm = false;
  let pendingImportPath = '';
  let showExportSelect = false;
  let exportMode = 'machine';
  let selectedExportIds = new Set();
  let showPassphraseInput = false;
  let showPassphraseConfirm = false;
  let passphraseValue = '';
  let showImportPassphraseInput = false;

  // Close dropdown when clicking outside
  function handleClickOutside(event) {
    showExportMenu = false;
    event.stopPropagation();
  }

  function openExportSelection(mode, event) {
    event.stopPropagation();
    showExportMenu = false;

    if (!window.wailsBindings) {
      alert('Wails 绑定未加载，请确保在 wails dev 模式下运行');
      return;
    }

    exportMode = mode;
    selectedExportIds = new Set($assetsStore.map(asset => asset.id));
    showExportSelect = true;
  }

  async function exportSelected(encryptPasswords, passphrase = '') {
    const { ExportConnectionsByIDs, ExportConnectionsByIDsWithPassphrase, SaveBinaryFile } = window.wailsBindings;

    if (encryptPasswords === 'passphrase') {
      if (typeof ExportConnectionsByIDsWithPassphrase !== 'function') {
        throw new Error('导出功能不可用，请升级应用');
      }
      return ExportConnectionsByIDsWithPassphrase(Array.from(selectedExportIds), passphrase)
        .then(jsonData => SaveBinaryFile('ssh-connections.json', btoa(unescape(encodeURIComponent(jsonData)))));
    }

    if (typeof ExportConnectionsByIDs !== 'function') {
      throw new Error('导出功能不可用，请升级应用');
    }

    const jsonData = await ExportConnectionsByIDs(Array.from(selectedExportIds), encryptPasswords);
    const base64Data = btoa(unescape(encodeURIComponent(jsonData)));
    return SaveBinaryFile('ssh-connections.json', base64Data);
  }

  function handleExportSelected() {
    if (!window.wailsBindings) {
      alert('Wails 绑定未加载，请确保在 wails dev 模式下运行');
      return;
    }

    if (selectedExportIds.size === 0) {
      alert('请至少选择一个连接');
      return;
    }

    if (exportMode === 'passphrase') {
      showPassphraseInput = true;
      return;
    }

    const encryptPasswords = exportMode === 'machine';
    exportSelected(encryptPasswords)
      .then(() => {
        showExportSelect = false;
      })
      .catch(error => {
        console.error('导出失败:', error);
        alert('导出失败: ' + error.message);
      });
  }

  function cancelExportSelection() {
    showExportSelect = false;
  }

  function handlePassphraseInputConfirm(value) {
    passphraseValue = value;
    showPassphraseInput = false;
    showPassphraseConfirm = true;
  }

  function handlePassphraseInputCancel() {
    passphraseValue = '';
    showPassphraseInput = false;
  }

  function handlePassphraseConfirmConfirm(value) {
    showPassphraseConfirm = false;
    if (value !== passphraseValue) {
      alert('两次输入的密码不一致');
      passphraseValue = '';
      showPassphraseInput = true;
      return;
    }

    exportSelected('passphrase', passphraseValue)
      .then(() => {
        showExportSelect = false;
        passphraseValue = '';
      })
      .catch(error => {
        console.error('导出失败:', error);
        alert('导出失败: ' + error.message);
      });
  }

  function handlePassphraseConfirmCancel() {
    passphraseValue = '';
    showPassphraseConfirm = false;
  }

  function toggleAssetSelection(id) {
    const next = new Set(selectedExportIds);
    if (next.has(id)) {
      next.delete(id);
    } else {
      next.add(id);
    }
    selectedExportIds = next;
  }

  function toggleGroupSelection(groupAssets) {
    const next = new Set(selectedExportIds);
    const allSelected = groupAssets.every(asset => next.has(asset.id));

    if (allSelected) {
      groupAssets.forEach(asset => next.delete(asset.id));
    } else {
      groupAssets.forEach(asset => next.add(asset.id));
    }

    selectedExportIds = next;
  }

  function toggleAllSelection() {
    if (selectedExportIds.size === $assetsStore.length) {
      selectedExportIds = new Set();
      return;
    }

    selectedExportIds = new Set($assetsStore.map(asset => asset.id));
  }

  function isGroupSelected(groupAssets) {
    return groupAssets.length > 0 && groupAssets.every(asset => selectedExportIds.has(asset.id));
  }

  async function handleImport(event) {
    event.stopPropagation();
    showExportMenu = false;

    if (!window.wailsBindings) {
      alert('Wails 绑定未加载，请确保在 wails dev 模式下运行');
      return;
    }

    try {
      const filePath = await window.wailsBindings.SelectImportFile();

      if (!filePath) return;

      pendingImportPath = filePath;
      showImportConfirm = true;
    } catch (error) {
      console.error('导入失败:', error);
      alert('导入失败: ' + error.message);
    }
  }

  async function confirmImportConnections() {
    if (!pendingImportPath) {
      showImportConfirm = false;
      return;
    }

    try {
      const { ImportConnectionsFromFileWithPassphrase, ImportConnectionsFromFile } = window.wailsBindings;
      let count = 0;

      if (typeof ImportConnectionsFromFileWithPassphrase === 'function') {
        try {
          count = await ImportConnectionsFromFileWithPassphrase(pendingImportPath, '');
        } catch (error) {
          const message = error?.message || String(error);
          if (message.includes('passphrase required') || message.includes('invalid passphrase')) {
            showImportConfirm = false;
            showImportPassphraseInput = true;
            return;
          }
          throw error;
        }
      } else if (typeof ImportConnectionsFromFile === 'function') {
        count = await ImportConnectionsFromFile(pendingImportPath);
      } else {
        throw new Error('导入功能不可用');
      }

      alert(`成功导入 ${count} 个连接`);

      window.dispatchEvent(new CustomEvent('assets-changed'));
    } catch (error) {
      console.error('导入失败:', error);
      alert('导入失败: ' + (error?.message || error));
    } finally {
      showImportConfirm = false;
      pendingImportPath = '';
    }
  }

  async function handleImportPassphraseConfirm(value) {
    if (!pendingImportPath) {
      showImportPassphraseInput = false;
      return;
    }

    try {
      const { ImportConnectionsFromFileWithPassphrase } = window.wailsBindings;
      if (typeof ImportConnectionsFromFileWithPassphrase !== 'function') {
        throw new Error('导入功能不可用');
      }

      const count = await ImportConnectionsFromFileWithPassphrase(pendingImportPath, value);
      alert(`成功导入 ${count} 个连接`);
      window.dispatchEvent(new CustomEvent('assets-changed'));
      showImportPassphraseInput = false;
      pendingImportPath = '';
    } catch (error) {
      const message = error?.message || String(error);
      if (message.includes('invalid passphrase')) {
        alert('导入密码不正确，请重试');
        showImportPassphraseInput = true;
        return;
      }
      console.error('导入失败:', error);
      alert('导入失败: ' + message);
      showImportPassphraseInput = false;
      pendingImportPath = '';
    }
  }

  function handleImportPassphraseCancel() {
    showImportPassphraseInput = false;
    pendingImportPath = '';
  }

  function cancelImportConnections() {
    showImportConfirm = false;
    pendingImportPath = '';
  }

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

    pendingDeleteAsset = asset;
    showDeleteConfirm = true;
  }

  async function confirmDeleteAsset() {
    if (!pendingDeleteAsset) return;

    try {
      await window.wailsBindings.RemoveConnection(pendingDeleteAsset.id);
      assetsStore.update(assets => assets.filter(a => a.id !== pendingDeleteAsset.id));
    } catch (error) {
      console.error('Failed to delete asset:', error);
      alert('删除连接失败: ' + (error?.message || error));
    } finally {
      showDeleteConfirm = false;
      pendingDeleteAsset = null;
    }
  }

  function cancelDeleteAsset() {
    showDeleteConfirm = false;
    pendingDeleteAsset = null;
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

<div class="h-full flex flex-col bg-white dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700" on:click={handleClickOutside}>
  <!-- 头部 -->
  <div class="p-4 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between mb-3">
      <h2 class="text-sm font-semibold text-gray-900 dark:text-white">服务器资产</h2>
      <div class="flex gap-1">
        <button
          on:click={onAddClick}
          class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
          title="添加连接"
        >
          <svg class="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
        </button>
        <div class="relative">
          <button
            on:click={(e) => { e.stopPropagation(); showExportMenu = !showExportMenu; }}
            class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            title="导出/导入"
          >
            <svg class="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
            </svg>
          </button>
          {#if showExportMenu}
            <div class="absolute right-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10 py-1 min-w-[160px]">
              <button
                on:click={(e) => openExportSelection('machine', e)}
                class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
              >
                <svg class="w-4 h-4 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <span>导出(加密)</span>
              </button>
              <button
                on:click={(e) => openExportSelection('passphrase', e)}
                class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
              >
                <svg class="w-4 h-4 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
                <span>导出(跨设备加密)</span>
              </button>
              <button
                on:click={(e) => openExportSelection('plain', e)}
                class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
              >
                <svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <span>导出(明文)</span>
              </button>
              <div class="border-t border-gray-200 dark:border-gray-700 my-1"></div>
              <button
                on:click={handleImport}
                class="w-full px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center gap-2"
              >
                <svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 13h6m-3-3v6m5 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                <span>导入连接</span>
              </button>
            </div>
          {/if}
        </div>
      </div>
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

<ConfirmDialog
  bind:isOpen={showDeleteConfirm}
  title="删除连接"
  message={pendingDeleteAsset ? `确定要删除连接 "${pendingDeleteAsset.name}" 吗？` : '确定要删除该连接吗？'}
  type="danger"
  confirmText="删除"
  cancelText="取消"
  onConfirm={confirmDeleteAsset}
  onCancel={cancelDeleteAsset}
/>

<ConfirmDialog
  bind:isOpen={showImportConfirm}
  title="导入连接"
  message="确定要导入连接配置吗？导入后会合并到现有连接中。"
  type="warning"
  confirmText="导入"
  cancelText="取消"
  onConfirm={confirmImportConnections}
  onCancel={cancelImportConnections}
/>

<Dialog
  bind:isOpen={showExportSelect}
  onClose={cancelExportSelection}
  title={exportMode === 'plain' ? '导出连接（明文）' : exportMode === 'passphrase' ? '导出连接（跨设备加密）' : '导出连接（加密）'}
  size="md"
>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div class="text-xs text-gray-500 dark:text-gray-400">
        已选 {selectedExportIds.size} / {$assetsStore.length}
      </div>
      <button
        type="button"
        on:click={toggleAllSelection}
        class="text-xs text-purple-600 dark:text-purple-400 hover:underline"
      >
        {selectedExportIds.size === $assetsStore.length ? '取消全选' : '全选'}
      </button>
    </div>

    <div class="max-h-64 overflow-y-auto space-y-2">
      {#each Object.entries($groupedAssetsStore) as [group, groupAssets]}
        <div class="rounded-lg border border-gray-200 dark:border-gray-700">
          <div class="flex items-center gap-2 px-3 py-2 bg-gray-50 dark:bg-gray-700/40">
            <input
              type="checkbox"
              checked={isGroupSelected(groupAssets)}
              on:change={() => toggleGroupSelection(groupAssets)}
              class="w-4 h-4 text-purple-600"
            />
            <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{group}</span>
            <span class="text-xs text-gray-400 dark:text-gray-500">({groupAssets.length})</span>
          </div>
          <div class="px-3 py-2 space-y-1">
            {#each groupAssets as asset (asset.id)}
              <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
                <input
                  type="checkbox"
                  checked={selectedExportIds.has(asset.id)}
                  on:change={() => toggleAssetSelection(asset.id)}
                  class="w-4 h-4 text-purple-600"
                />
                <span class="truncate">{asset.name}</span>
                <span class="text-xs text-gray-400 dark:text-gray-500 truncate">
                  {asset.username}@{asset.host}
                </span>
              </label>
            {/each}
          </div>
        </div>
      {/each}
    </div>

    <div class="flex gap-2 justify-end">
      <button
        type="button"
        on:click={cancelExportSelection}
        class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-md text-xs font-medium transition-colors"
      >
        取消
      </button>
      <button
        type="button"
        on:click={handleExportSelected}
        disabled={selectedExportIds.size === 0}
        class="px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm bg-purple-600 hover:bg-purple-700 text-white disabled:opacity-50 disabled:cursor-not-allowed"
      >
        导出
      </button>
    </div>
  </div>
</Dialog>

<InputDialog
  bind:isOpen={showPassphraseInput}
  title="导出加密"
  message="请输入导出密码（用于跨设备解密）"
  placeholder="导出密码"
  inputType="password"
  allowEmpty={false}
  trimValue={false}
  confirmText="下一步"
  cancelText="取消"
  onConfirm={handlePassphraseInputConfirm}
  onCancel={handlePassphraseInputCancel}
/>

<InputDialog
  bind:isOpen={showPassphraseConfirm}
  title="确认导出密码"
  message="请再次输入导出密码"
  placeholder="确认导出密码"
  inputType="password"
  allowEmpty={false}
  trimValue={false}
  confirmText="导出"
  cancelText="取消"
  onConfirm={handlePassphraseConfirmConfirm}
  onCancel={handlePassphraseConfirmCancel}
/>

<InputDialog
  bind:isOpen={showImportPassphraseInput}
  title="导入解密"
  message="该文件需要导入密码，请输入"
  placeholder="导入密码"
  inputType="password"
  allowEmpty={false}
  trimValue={false}
  confirmText="导入"
  cancelText="取消"
  onConfirm={handleImportPassphraseConfirm}
  onCancel={handleImportPassphraseCancel}
/>
