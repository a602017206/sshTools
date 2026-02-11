<script>
  import { onMount } from 'svelte';
  import AssetList from './components/AssetList.svelte';
  import TerminalPanel from './components/TerminalPanel.svelte';
  import FileManager from './components/FileManager.svelte';
  import ServerMonitor from './components/ServerMonitor.svelte';
  import DevToolsPanel from './components/DevToolsPanel.svelte';
  import AddAssetDialog from './components/AddAssetDialog.svelte';
  import AboutDialog from './components/AboutDialog.svelte';
  import { assetsStore, connectionsStore, themeStore, uiStore, setSidebarWidth, setRightPanelWidth, setFileManagerHeight, setTheme } from './stores.js';
  import { uploadStore, activeTransfers, completedTransfers } from './stores/uploadStore.js';
  import { formatFileSize, formatSpeed, getTransferPercentage } from './stores/uploadStore.js';
  import { CancelTransfer } from '../wailsjs/go/main/App.js';

  let isDevToolsOpen = false;
  let isAddDialogOpen = false;
  let isAboutDialogOpen = false;
  let isSidebarCollapsed = false;
  let isRightPanelCollapsed = false;
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
    const newTheme = $themeStore === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
  }

  function toggleDevTools() {
    if (isAddDialogOpen) return;
    isDevToolsOpen = !isDevToolsOpen;
  }

  function toggleSidebar() {
    if (isAddDialogOpen) return;
    isSidebarCollapsed = !isSidebarCollapsed;
  }

  function toggleRightPanel() {
    if (isAddDialogOpen) return;
    isRightPanelCollapsed = !isRightPanelCollapsed;
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

  function ensureFullHeight() {
    const appElement = document.getElementById('app');
    if (appElement) {
      appElement.style.height = '100vh';
      appElement.style.width = '100vw';
    }
  }

   onMount(async () => {
    const savedTheme = localStorage.getItem('ssh-tools-theme');
    if (!savedTheme) {
      const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
      if (prefersDark) {
        setTheme('dark');
      }
    } else {
      setTheme(savedTheme);
    }

    let cleanupEvents = null;

    try {
      const wails = await import('../wailsjs/go/main/App.js');
      window.wailsBindings = wails;
      window.dispatchEvent(new CustomEvent('wails-bindings-loaded', {
        detail: Object.keys(wails.default || wails).join(', ')
      }));

      await loadAssetsFromBackend();

      // Listen for about dialog event from backend
      const runtime = await import('../wailsjs/runtime/runtime.js');
      cleanupEvents = runtime.EventsOn('app:show-about', () => {
        isAboutDialogOpen = true;
      });

      // Listen for assets changed event (from import)
      window.addEventListener('assets-changed', loadAssetsFromBackend);
    } catch (error) {
      console.warn('Wails bindings not available yet:', error.message);
    }

    ensureFullHeight();
    window.addEventListener('resize', ensureFullHeight);

    return () => {
      window.removeEventListener('resize', ensureFullHeight);
      window.removeEventListener('assets-changed', loadAssetsFromBackend);
      if (cleanupEvents) {
        cleanupEvents();
      }
    };
  });
</script>

<div class="h-screen w-full flex flex-col {themeClass} {$themeStore === 'dark' ? 'bg-gray-900 text-gray-100' : 'bg-gray-50 text-gray-900'}">
  <!-- 顶部标题栏 -->
  <header class="h-14 flex-shrink-0 {$themeStore === 'dark' ? 'bg-gray-800 border-gray-700' : 'bg-white border-gray-200'} border-b flex items-center px-6 shadow-sm" style="pointer-events: {isAddDialogOpen ? 'none' : 'auto'};">
    <div class="flex items-center gap-3">
      <div class="w-8 h-8 bg-gradient-to-br from-purple-600 to-blue-600 rounded-lg flex items-center justify-center font-bold text-sm text-white shadow-md">
        哈
      </div>
      <div>
        <div class="font-semibold text-sm {$themeStore === 'dark' ? 'text-white' : 'text-gray-900'}">啊哈 SSH 连接工具</div>
        <div class="text-xs {$themeStore === 'dark' ? 'text-gray-400' : 'text-gray-500'}">AHa SSH Manager</div>
      </div>
    </div>

    <div class="ml-auto flex items-center gap-3">
      <div class="relative">
        <button
          on:click={() => uploadStore.togglePanel()}
          class="flex items-center justify-center w-9 h-9 rounded-lg transition-all shadow-sm {$themeStore === 'dark' ? 'bg-gray-700 hover:bg-gray-600 text-blue-400' : 'bg-gray-100 hover:bg-gray-200 text-gray-600'}"
          title="上传任务"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
        </button>
        {#if $activeTransfers.length > 0}
          <span class="absolute -top-1 -right-1 flex items-center justify-center w-4 h-4 text-[10px] font-bold text-white bg-red-500 rounded-full">
            {$activeTransfers.length}
          </span>
        {/if}
      </div>

      <!-- 主题切换按钮 -->
      <button
        on:click={toggleTheme}
        disabled={isAddDialogOpen}
        class="flex items-center justify-center w-9 h-9 rounded-lg transition-all shadow-sm disabled:opacity-50 disabled:cursor-not-allowed {$themeStore === 'dark' ? 'bg-gray-700 hover:bg-gray-600 text-yellow-400' : 'bg-gray-100 hover:bg-gray-200 text-gray-600'}"
        title={$themeStore === 'dark' ? '切换到浅色模式' : '切换到深色模式'}
      >
        {#if $themeStore === 'dark'}
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path>
          </svg>
        {:else}
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"></path>
          </svg>
        {/if}
      </button>

      <button
        on:click={toggleDevTools}
        disabled={isAddDialogOpen}
        class="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 text-white rounded-lg font-medium transition-all shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path>
        </svg>
        <span class="text-sm">开发工具</span>
      </button>
    </div>
  </header>

  <!-- 主内容区域 -->
  <div class="flex-1 flex overflow-hidden min-h-0 relative">
    <!-- 左侧展开按钮（折叠时显示） -->
    {#if isSidebarCollapsed}
    <button
      on:click={toggleSidebar}
      disabled={isAddDialogOpen}
      class="absolute left-0 top-1/2 -translate-y-1/2 z-50 flex items-center justify-center w-8 h-12 rounded-r-lg transition-all shadow-md disabled:opacity-50 disabled:cursor-not-allowed opacity-0 hover:opacity-100 {$themeStore === 'dark' ? 'bg-gray-800 hover:bg-gray-700 text-gray-300 border border-gray-700' : 'bg-white hover:bg-gray-100 text-gray-600 border border-gray-200'}"
      title="展开资产列表"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
      </svg>
    </button>
    {/if}

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
        onEdit={handleEditAsset}
      />
    </div>

    <!-- 侧边栏调整手柄 -->
    {#if !isSidebarCollapsed}
    <div
      class="resize-handle-horizontal flex-shrink-0 relative group"
      role="separator"
      aria-hidden="true"
      style="cursor: {isAddDialogOpen ? 'default' : 'col-resize'}; height: 100%; padding: 0 2px; pointer-events: {isAddDialogOpen ? 'none' : 'auto'};"
      on:mousedown={startSidebarResize}
    >
      <div class="h-full w-full rounded"></div>
      <button
        on:click={toggleSidebar}
        disabled={isAddDialogOpen}
        class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 flex items-center justify-center w-7 h-7 rounded transition-all shadow-md disabled:opacity-50 disabled:cursor-not-allowed opacity-0 group-hover:opacity-100 {$themeStore === 'dark' ? 'bg-gray-700 hover:bg-gray-600 text-gray-300' : 'bg-gray-100 hover:bg-gray-200 text-gray-600'}"
        title="折叠资产列表"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
        </svg>
      </button>
    </div>
    {/if}

    <!-- 中间：终端面板 -->
    <div class="flex-1 min-w-0 min-h-0 flex flex-col {$themeStore === 'dark' ? 'bg-gray-900' : 'bg-gray-50'}">
      <TerminalPanel bind:this={terminalPanelRef} />
    </div>

    <!-- 右侧面板调整手柄 -->
    {#if !isRightPanelCollapsed}
    <div
      class="resize-handle-horizontal flex-shrink-0 relative group"
      role="separator"
      aria-hidden="true"
      style="cursor: {isAddDialogOpen ? 'default' : 'col-resize'}; height: 100%; padding: 0 2px; pointer-events: {isAddDialogOpen ? 'none' : 'auto'};"
      on:mousedown={startRightPanelResize}
    >
      <div class="h-full w-full rounded"></div>
      <button
        on:click={toggleRightPanel}
        disabled={isAddDialogOpen}
        class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 flex items-center justify-center w-7 h-7 rounded transition-all shadow-md disabled:opacity-50 disabled:cursor-not-allowed opacity-0 group-hover:opacity-100 {$themeStore === 'dark' ? 'bg-gray-700 hover:bg-gray-600 text-gray-300' : 'bg-gray-100 hover:bg-gray-200 text-gray-600'}"
        title="折叠右侧面板"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
        </svg>
      </button>
    </div>
    {/if}

    <!-- 右侧：文件管理和服务器监控 -->
    <div
      data-right-panel="true"
      class="flex-shrink-0 flex flex-col overflow-hidden {$themeStore === 'dark' ? 'border-gray-700 bg-gray-800' : 'border-gray-200 bg-white'} border-l shadow-sm"
      class:collapsed={isRightPanelCollapsed}
      style="width: {isRightPanelCollapsed ? '0' : rightPanelWidth}px; min-width: {isRightPanelCollapsed ? '0' : '300px'}; max-width: 600px;"
    >
      <!-- 文件管理 -->
      <div class="overflow-hidden" style="height: {fileManagerHeight}%; min-height: 0;">
        <FileManager />
      </div>

      <!-- 文件管理/监控调整手柄 -->
      <div
        class="resize-handle-vertical flex-shrink-0 relative"
        role="separator"
        aria-hidden="true"
        style="cursor: {isAddDialogOpen ? 'default' : 'row-resize'}; width: 100%; padding: 2px 0; pointer-events: {isAddDialogOpen ? 'none' : 'auto'};"
        on:mousedown={startFileManagerResize}
      >
        <div class="w-full h-full rounded"></div>
      </div>

      <!-- 服务器监控 -->
      <div class="overflow-hidden" style="height: {100 - fileManagerHeight}%; min-height: 0;">
        <ServerMonitor />
      </div>
    </div>

    <!-- 右侧展开按钮（折叠时显示） -->
    {#if isRightPanelCollapsed}
    <button
      on:click={toggleRightPanel}
      disabled={isAddDialogOpen}
      class="absolute right-0 top-1/2 -translate-y-1/2 z-50 flex items-center justify-center w-8 h-12 rounded-l-lg transition-all shadow-md disabled:opacity-50 disabled:cursor-not-allowed opacity-0 hover:opacity-100 {$themeStore === 'dark' ? 'bg-gray-800 hover:bg-gray-700 text-gray-300 border border-gray-700' : 'bg-white hover:bg-gray-100 text-gray-600 border border-gray-200'}"
      title="展开右侧面板"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
      </svg>
    </button>
    {/if}
  </div>

  <!-- 对话框 -->
  <AddAssetDialog
    bind:isOpen={isAddDialogOpen}
    bind:editingAsset={editingAsset}
    onAdd={handleAddAsset}
    onUpdate={handleUpdateAsset}
  />

  <DevToolsPanel bind:isOpen={isDevToolsOpen} {themeStore} />

  <AboutDialog
    bind:isOpen={isAboutDialogOpen}
    onClose={() => isAboutDialogOpen = false}
    themeStore={themeStore}
  />

  {#if $uploadStore.isPanelOpen}
    <div
      class="fixed top-14 right-0 z-50 w-96 max-h-[600px] {$themeStore === 'dark' ? 'bg-gray-800 border-gray-700 text-white' : 'bg-white border-gray-200 text-gray-900'} border-l shadow-xl flex flex-col"
      style="height: calc(100vh - 3.5rem);"
    >
      <div class="p-4 border-b {$themeStore === 'dark' ? 'border-gray-700' : 'border-gray-200'}">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-semibold text-sm">上传任务</h3>
          <button
            on:click={() => uploadStore.setPanelOpen(false)}
            class="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="flex gap-2">
          <button
            on:click={() => uploadStore.setActiveTab('active')}
            class="flex-1 py-1.5 px-3 text-xs font-medium rounded-lg transition-colors {$uploadStore.activeTab === 'active'
              ? $themeStore === 'dark' ? 'bg-blue-600 text-white' : 'bg-blue-600 text-white'
              : $themeStore === 'dark' ? 'text-gray-400 hover:bg-gray-700' : 'text-gray-600 hover:bg-gray-100'}"
          >
             进行中 ({$activeTransfers.length})
          </button>
          <button
            on:click={() => uploadStore.setActiveTab('history')}
            class="flex-1 py-1.5 px-3 text-xs font-medium rounded-lg transition-colors {$uploadStore.activeTab === 'history'
              ? $themeStore === 'dark' ? 'bg-blue-600 text-white' : 'bg-blue-600 text-white'
              : $themeStore === 'dark' ? 'text-gray-400 hover:bg-gray-700' : 'text-gray-600 hover:bg-gray-100'}"
          >
             历史 ({$completedTransfers.length})
          </button>
        </div>
      </div>

       {#if $uploadStore.activeTab === 'active' && $activeTransfers.length === 0}
        <div class="flex-1 flex flex-col items-center justify-center text-gray-500 dark:text-gray-400 gap-3">
          <svg class="w-12 h-12 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
          <span class="text-sm">暂无上传任务</span>
        </div>
       {:else if $uploadStore.activeTab === 'active'}
         <div class="flex-1 overflow-y-auto p-4 space-y-3">
           {#each $activeTransfers as transfer (transfer.id)}
            <div class="rounded-lg {$themeStore === 'dark' ? 'bg-gray-700/50 border-gray-600' : 'bg-gray-50 border-gray-200'} border p-3">
              <div class="flex items-center justify-between mb-2">
                <div class="flex-1 min-w-0">
                  <div class="text-sm font-medium truncate" title={transfer.filename}>{transfer.filename}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                    {formatFileSize(transfer.bytesSent)} / {formatFileSize(transfer.totalBytes)}
                    {#if transfer.speed}
                      • {formatSpeed(transfer.speed)}
                    {/if}
                  </div>
                </div>
                <div class="flex items-center gap-2 ml-3">
                  <span class="text-xs font-medium text-blue-600 dark:text-blue-400">
                    {Math.round(getTransferPercentage(transfer))}%
                  </span>
                  <button
                    on:click={async () => {
                      await CancelTransfer(transfer.id);
                      uploadStore.cancelTransfer(transfer.id);
                    }}
                    class="p-1.5 rounded hover:bg-red-100 dark:hover:bg-red-900/30 text-red-500 transition-colors"
                    title="取消上传"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </div>
              <div class="h-2 {$themeStore === 'dark' ? 'bg-gray-600' : 'bg-gray-200'} rounded-full overflow-hidden">
                <div
                  class="h-full bg-blue-500 transition-all duration-300"
                  style={`width: ${Math.min(100, Math.max(0, getTransferPercentage(transfer)))}%`}
                ></div>
              </div>
            </div>
          {/each}
        </div>
       {:else if $uploadStore.activeTab === 'history' && $completedTransfers.length === 0}
        <div class="flex-1 flex flex-col items-center justify-center text-gray-500 dark:text-gray-400 gap-3">
          <svg class="w-12 h-12 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="text-sm">暂无历史记录</span>
        </div>
       {:else if $uploadStore.activeTab === 'history'}
         <div class="flex-1 flex flex-col">
           <div class="p-3 border-b {$themeStore === 'dark' ? 'border-gray-700' : 'border-gray-200'}">
             <button
               on:click={() => uploadStore.clearCompleted()}
               class="w-full py-2 px-3 text-xs font-medium rounded-lg transition-colors {$themeStore === 'dark' ? 'hover:bg-red-900/30 text-red-400' : 'hover:bg-red-50 text-red-600'}"
             >
               清空历史记录
             </button>
           </div>
           <div class="flex-1 overflow-y-auto p-4 space-y-3">
             {#each $completedTransfers as transfer (transfer.id)}
              <div class="rounded-lg {$themeStore === 'dark' ? 'bg-gray-700/50 border-gray-600' : 'bg-gray-50 border-gray-200'} border p-3">
                <div class="flex items-center justify-between">
                  <div class="flex-1 min-w-0">
                    <div class="text-sm font-medium truncate" title={transfer.filename}>{transfer.filename}</div>
                    <div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5 flex items-center gap-2">
                      {#if transfer.status === 'completed'}
                        <span class="text-green-500">完成</span>
                      {:else if transfer.status === 'failed'}
                        <span class="text-red-500">失败</span>
                      {:else if transfer.status === 'cancelled'}
                        <span class="text-gray-500">已取消</span>
                      {/if}
                      <span>•</span>
                      <span>{formatFileSize(transfer.totalBytes)}</span>
                    </div>
                    {#if transfer.status === 'failed' && transfer.error}
                      <div class="text-xs text-red-500 dark:text-red-400 mt-1 truncate" title={transfer.error}>
                        {transfer.error}
                      </div>
                    {/if}
                  </div>
                  <button
                    on:click={() => uploadStore.removeTransfer(transfer.id)}
                    class="p-1.5 rounded hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-500 dark:text-gray-400 transition-colors ml-3"
                    title="删除记录"
                  >
                    <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  {/if}
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
