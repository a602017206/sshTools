<script>
  import { onMount } from 'svelte';
  import AssetList from './components/AssetList.svelte';
  import TerminalPanel from './components/TerminalPanel.svelte';
  import FileManager from './components/FileManager.svelte';
  import ServerMonitor from './components/ServerMonitor.svelte';
  import DevToolsPanel from './components/DevToolsPanel.svelte';
  import AddAssetDialog from './components/AddAssetDialog.svelte';
  import { assetsStore, connectionsStore, themeStore, uiStore, setSidebarWidth, setRightPanelWidth, setFileManagerHeight } from './stores.js';

  let isDevToolsOpen = false;
  let isAddDialogOpen = false;
  let isSidebarCollapsed = false;
  let editingAsset = null;
  let terminalPanelRef;

  $: connectionsArray = $connectionsStore ? Array.from($connectionsStore.values()) : [];
  $: themeClass = $themeStore === 'dark' ? 'dark' : '';

  // 从 store 获取面板尺寸
  $: sidebarWidth = $uiStore.sidebarWidth;
  $: rightPanelWidth = $uiStore.rightPanelWidth;
  $: fileManagerHeight = $uiStore.fileManagerHeight;

  // Resize state
  let isResizingSidebar = false;
  let isResizingRightPanel = false;
  let isResizingFileManager = false;

  function toggleTheme() {
    themeStore.update(t => t === 'light' ? 'dark' : 'light');
  }

  function toggleDevTools() {
    if (isAddDialogOpen) return;
    isDevToolsOpen = !isDevToolsOpen;
  }

  function toggleSidebar() {
    if (isAddDialogOpen) return;
    isSidebarCollapsed = !isSidebarCollapsed;
  }

  // Sidebar resize handlers
  function startSidebarResize(e) {
    e.preventDefault();
    if (isSidebarCollapsed || isAddDialogOpen) return;
    isResizingSidebar = true;
    document.addEventListener('mousemove', handleSidebarResize);
    document.addEventListener('mouseup', stopSidebarResize);
  }

  function handleSidebarResize(e) {
    if (!isResizingSidebar) return;
    const newWidth = Math.max(200, Math.min(500, e.clientX));
    setSidebarWidth(newWidth);
  }

  function stopSidebarResize() {
    isResizingSidebar = false;
    document.removeEventListener('mousemove', handleSidebarResize);
    document.removeEventListener('mouseup', stopSidebarResize);
  }

  // Right panel resize handlers
  function startRightPanelResize(e) {
    e.preventDefault();
    if (isAddDialogOpen) return;
    isResizingRightPanel = true;
    document.addEventListener('mousemove', handleRightPanelResize);
    document.addEventListener('mouseup', stopRightPanelResize);
  }

  function handleRightPanelResize(e) {
    if (!isResizingRightPanel) return;
    const containerWidth = window.innerWidth;
    const newWidth = Math.max(300, Math.min(600, containerWidth - e.clientX));
    setRightPanelWidth(newWidth);
  }

  function stopRightPanelResize() {
    isResizingRightPanel = false;
    document.removeEventListener('mousemove', handleRightPanelResize);
    document.removeEventListener('mouseup', stopRightPanelResize);
  }

  // File manager resize handlers
  function startFileManagerResize(e) {
    e.preventDefault();
    if (isAddDialogOpen) return;
    isResizingFileManager = true;
    const rightPanel = document.querySelector('[data-right-panel]');
    if (rightPanel) {
      const rect = rightPanel.getBoundingClientRect();
      fileManagerInitialY = e.clientY;
      fileManagerInitialHeight = fileManagerHeight;
      rightPanelHeight = rect.height;
    }
    document.addEventListener('mousemove', handleFileManagerResize);
    document.addEventListener('mouseup', stopFileManagerResize);
  }

  let fileManagerInitialY = 0;
  let fileManagerInitialHeight = 50;
  let rightPanelHeight = 0;

  function handleFileManagerResize(e) {
    if (!isResizingFileManager) return;
    const deltaY = e.clientY - fileManagerInitialY;
    const deltaPercent = (deltaY / rightPanelHeight) * 100;
    const newHeightPercent = Math.max(20, Math.min(80, fileManagerInitialHeight + deltaPercent));
    setFileManagerHeight(newHeightPercent);
  }

  function stopFileManagerResize() {
    isResizingFileManager = false;
    document.removeEventListener('mousemove', handleFileManagerResize);
    document.removeEventListener('mouseup', stopFileManagerResize);
  }

  // 连接处理 - 转发给 TerminalPanel
  function handleConnect(asset) {
    if (terminalPanelRef && typeof terminalPanelRef.handleConnect === 'function') {
      terminalPanelRef.handleConnect(asset);
    } else {
      console.error('TerminalPanel not available');
      alert('终端面板未初始化');
    }
  }

  async function handleAddAsset(connectionData) {
    if (!window.wailsBindings) {
      console.error('Wails bindings not loaded');
      return;
    }

    try {
      await window.wailsBindings.AddConnection(connectionData);

      const asset = {
        id: connectionData.id,
        name: connectionData.name,
        host: connectionData.host,
        port: connectionData.port,
        username: connectionData.user,
        group: connectionData.tags?.[0] || '默认分组',
        status: 'online',
        type: connectionData.type || 'ssh',
        auth_type: connectionData.auth_type,
        key_path: connectionData.key_path,
        tags: connectionData.tags || []
      };

      assetsStore.update(assets => [...assets, asset]);
    } catch (error) {
      console.error('Failed to add asset:', error);
      throw error;
    }
  }

  function handleEditAsset(asset) {
    editingAsset = asset;
    isAddDialogOpen = true;
  }

  async function handleUpdateAsset(connectionData) {
    if (!window.wailsBindings) {
      console.error('Wails bindings not loaded');
      return;
    }

    try {
      // Update connection is already called in the dialog
      // Just update the local store
      assetsStore.update(assets => {
        return assets.map(asset => {
          if (asset.id === connectionData.id) {
            return {
              ...asset,
              id: connectionData.id,
              name: connectionData.name,
              host: connectionData.host,
              port: connectionData.port,
              username: connectionData.user,
              group: connectionData.tags?.[0] || '默认分组',
              type: connectionData.type || 'ssh',
              auth_type: connectionData.auth_type,
              key_path: connectionData.key_path,
              tags: connectionData.tags || []
            };
          }
          return asset;
        });
      });
    } catch (error) {
      console.error('Failed to update asset:', error);
      throw error;
    }
  }

  async function loadAssetsFromBackend() {
    if (!window.wailsBindings) {
      console.warn('Wails bindings not loaded yet');
      return;
    }

    try {
      const connections = await window.wailsBindings.GetConnections();
      console.log('Loaded connections from backend:', connections);

      const assets = connections.map(conn => ({
        id: conn.id,
        name: conn.name,
        host: conn.host,
        port: conn.port,
        username: conn.user,
        group: conn.tags?.[0] || '默认分组',
        status: 'online',
        type: 'ssh',
        auth_type: conn.auth_type,
        key_path: conn.key_path,
        tags: conn.tags || []
      }));

      assetsStore.set(assets);
    } catch (error) {
      console.error('Failed to load connections:', error);
    }
  }

  onMount(async () => {
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      themeStore.set('dark');
    }

    try {
      const wails = await import('../wailsjs/go/main/App.js');
      window.wailsBindings = wails;
      window.dispatchEvent(new CustomEvent('wails-bindings-loaded', {
        detail: Object.keys(wails.default || wails).join(', ')
      }));

      await loadAssetsFromBackend();
    } catch (error) {
      console.warn('Wails bindings not available yet:', error.message);
    }

    // 确保容器适配窗口大小
    const ensureFullHeight = () => {
      const appElement = document.getElementById('app');
      if (appElement) {
        appElement.style.height = '100vh';
        appElement.style.width = '100vw';
      }
    };

    ensureFullHeight();
    window.addEventListener('resize', ensureFullHeight);

    return () => {
      window.removeEventListener('resize', ensureFullHeight);
    };
  });
</script>

<div class="h-screen w-full flex flex-col {themeClass} {$themeStore === 'dark' ? 'bg-gray-900 text-gray-100' : 'bg-gray-50 text-gray-900'}">
  <!-- 顶部标题栏 -->
  <header class="h-14 flex-shrink-0 {$themeStore === 'dark' ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'} border-b flex items-center px-6 shadow-sm" style="pointer-events: {isAddDialogOpen ? 'none' : 'auto'};">
    <div class="flex items-center gap-3">
      <div class="w-8 h-8 bg-gradient-to-br from-purple-600 to-blue-600 rounded-lg flex items-center justify-center font-bold text-sm text-white shadow-md">
        SSH
      </div>
      <div>
        <div class="font-semibold text-sm {$themeStore === 'dark' ? 'text-white' : 'text-gray-900'}">跨平台 SSH 连接工具</div>
        <div class="text-xs {$themeStore === 'dark' ? 'text-gray-400' : 'text-gray-500'}">Cross-Platform SSH Manager</div>
      </div>
    </div>
    
    <div class="ml-auto">
      <button
        on:click={toggleDevTools}
        disabled={isAddDialogOpen}
        class="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white rounded-lg font-medium transition-all shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31 .826 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31 -2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-1.066 2.573c-.94-1.543.826-3.31 -2.37a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31 .826-2.37a1.724 1.724 0 00-1.065-2.572-1.065c-.426 1.756-2.924 0 3.35a1.724 1.724 0 001.066-2.573c.996.608 2.296.07 2.572-1.065z"></path>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
        </svg>
        <span class="text-sm">开发工具</span>
      </button>
    </div>
  </header>

  <!-- 主内容区域 -->
  <div class="flex-1 flex overflow-hidden min-h-0">
    <!-- 左侧：资产列表 -->
    <div
      class="flex-shrink-0 transition-all duration-200 {$themeStore === 'dark' ? 'border-gray-700 bg-gray-800' : 'border-gray-200 bg-white'} border-r overflow-hidden"
      class:collapsed={isSidebarCollapsed}
      style="width: {isSidebarCollapsed ? '0' : sidebarWidth}px; min-width: {isSidebarCollapsed ? '0' : sidebarWidth}px;"
    >
      <AssetList
        onConnect={handleConnect}
        onAddClick={() => {
          editingAsset = null;
          isAddDialogOpen = true;
        }}
        onDelete={async (asset) => {
          if (!window.wailsBindings) {
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
            alert('删除连接失败: ' + error);
          }
        }}
        onEdit={handleEditAsset}
      />
    </div>

    <!-- 侧边栏调整手柄 -->
    {#if !isSidebarCollapsed}
    <div
      class="resize-handle-horizontal flex-shrink-0 relative"
      style="cursor: {isAddDialogOpen ? 'default' : 'col-resize'}; height: 100%; padding: 0 2px; pointer-events: {isAddDialogOpen ? 'none' : 'auto'};"
      on:mousedown={startSidebarResize}
    >
      <div class="h-full w-full rounded"></div>
    </div>
    {/if}

    <!-- 中间：终端面板 -->
    <div class="flex-1 min-w-0 min-h-0 flex flex-col {$themeStore === 'dark' ? 'bg-gray-900' : 'bg-gray-50'}">
      <TerminalPanel bind:this={terminalPanelRef} />
    </div>

    <!-- 右侧面板调整手柄 -->
    <div
      class="resize-handle-horizontal flex-shrink-0 relative"
      style="cursor: {isAddDialogOpen ? 'default' : 'col-resize'}; height: 100%; padding: 0 2px; pointer-events: {isAddDialogOpen ? 'none' : 'auto'};"
      on:mousedown={startRightPanelResize}
    >
      <div class="h-full w-full rounded"></div>
    </div>

    <!-- 右侧：文件管理和服务器监控 -->
    <div
      data-right-panel="true"
      class="flex-shrink-0 flex flex-col overflow-hidden {$themeStore === 'dark' ? 'border-gray-700 bg-gray-800' : 'border-gray-200 bg-white'} border-l shadow-sm"
      style="width: {rightPanelWidth}px; min-width: 300px; max-width: 600px;"
    >
      <!-- 文件管理 -->
      <div class="flex-1 overflow-hidden" style="height: {fileManagerHeight}%">
        <FileManager />
      </div>

      <!-- 文件管理/监控调整手柄 -->
      <div
        class="resize-handle-vertical flex-shrink-0 relative"
        style="cursor: {isAddDialogOpen ? 'default' : 'row-resize'}; width: 100%; padding: 2px 0; pointer-events: {isAddDialogOpen ? 'none' : 'auto'};"
        on:mousedown={startFileManagerResize}
      >
        <div class="w-full h-full rounded"></div>
      </div>

      <!-- 服务器监控 -->
      <div class="flex-1 overflow-hidden" style="height: {100 - fileManagerHeight}%">
        <ServerMonitor />
      </div>
    </div>
  </div>

  <!-- 对话框 -->
  <AddAssetDialog
    bind:isOpen={isAddDialogOpen}
    bind:editingAsset={editingAsset}
    onAdd={handleAddAsset}
    onUpdate={handleUpdateAsset}
  />

  <DevToolsPanel bind:isOpen={isDevToolsOpen} />
</div>

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    overflow: hidden;
    height: 100vh;
  }

  :global(html) {
    height: 100%;
  }

  #app {
    height: 100vh;
    width: 100vw;
  }

  .resize-handle-horizontal {
    width: auto;
    height: 100%;
    transition: background-color 0.2s ease;
    z-index: 100;
    background-color: transparent;
    position: relative;
  }

  .resize-handle-horizontal > div {
    width: 1px;
    height: 100%;
    background-color: transparent;
    border-radius: 0;
    transition: background-color 0.2s ease, width 0.2s ease;
    opacity: 0;
  }

  .resize-handle-horizontal:hover > div {
    background-color: #e5e7eb;
    width: 2px;
    opacity: 1;
  }

  :global(.dark) .resize-handle-horizontal:hover > div {
    background-color: #374151;
  }

  .resize-handle-vertical {
    height: auto;
    width: 100%;
    transition: background-color 0.2s ease;
    z-index: 100;
    background-color: transparent;
    position: relative;
  }

  .resize-handle-vertical > div {
    height: 1px;
    width: 100%;
    background-color: transparent;
    border-radius: 0;
    transition: background-color 0.2s ease, height 0.2s ease;
    opacity: 0;
  }

  .resize-handle-vertical:hover > div {
    background-color: #e5e7eb;
    height: 2px;
    opacity: 1;
  }

  :global(.dark) .resize-handle-vertical:hover > div {
    background-color: #374151;
  }
</style>
