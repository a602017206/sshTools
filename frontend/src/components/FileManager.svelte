<script>
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import {
    ListFiles,
    UploadFiles,
    DownloadFiles,
    SelectUploadFiles,
    SelectDownloadDirectory,
  } from '../../wailsjs/go/main/App.js';
  import { activeSessionIdStore, connectionsStore } from '../stores.js';

  let currentPath = '/';
  let expandedDirs = new Set([]);
  let files = [];
  let isLoading = false;
  let error = null;

  // Transfer progress unsubscribers
  let progressUnsubscribers = [];

  // Get current session object
  $: currentSession = $activeSessionIdStore ? $connectionsStore.get($activeSessionIdStore) : null;
  $: isSessionConnected = currentSession?.connected || false;
  
  async function loadDirectory(path) {
    if (!$activeSessionIdStore || !isSessionConnected) return;

    isLoading = true;
    error = null;

    try {
      const fileList = await ListFiles($activeSessionIdStore, path);
      files = fileList || [];
      currentPath = path;
    } catch (err) {
      console.error('Failed to load directory:', err);
      error = err.message || '加载目录失败';
    } finally {
      isLoading = false;
    }
  }

  function toggleDirectory(dirName) {
    const newExpanded = new Set(expandedDirs);
    const fullPath = `${currentPath}/${dirName}`;
    if (newExpanded.has(fullPath)) {
      newExpanded.delete(fullPath);
    } else {
      newExpanded.add(fullPath);
    }
    expandedDirs = newExpanded;
  }

  // Convert API response format to display format
  function formatFileSize(size) {
    if (!size) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    let i = 0;
    while (size >= 1024 && i < units.length - 1) {
      size /= 1024;
      i++;
    }
    return `${size.toFixed(1)} ${units[i]}`;
  }

  async function handleRefresh() {
    await loadDirectory(currentPath);
  }

  async function handleUpload() {
    if (!$activeSessionIdStore || !isSessionConnected) return;

    try {
      const localPaths = await SelectUploadFiles();
      if (!localPaths || localPaths.length === 0) return;

      const transferIDs = await UploadFiles($activeSessionIdStore, localPaths, currentPath);
      transferIDs.forEach((id) => subscribeToTransfer(id));

      // Refresh directory after a short delay
      setTimeout(() => loadDirectory(currentPath), 2000);
    } catch (err) {
      console.error('Upload failed:', err);
      error = err.message || '上传失败';
    }
  }

  async function handleDownload(file) {
    if (!$activeSessionIdStore || !isSessionConnected || file.is_dir) return;

    try {
      const localDir = await SelectDownloadDirectory();
      if (!localDir) return;

      const transferIDs = await DownloadFiles($activeSessionIdStore, [file.path], localDir);
      transferIDs.forEach((id) => subscribeToTransfer(id));
    } catch (err) {
      console.error('Download failed:', err);
      error = err.message || '下载失败';
    }
  }

  // Transfer progress subscription
  function subscribeToTransfer(transferID) {
    const eventName = `sftp:progress:${transferID}`;
    const unsubscriber = EventsOn(eventName, (progress) => {
      // Auto-cleanup completed transfers after 3 seconds
      if (progress.status === 'completed' || progress.status === 'failed') {
        setTimeout(() => {
          // Could show transfer status in UI if needed
        }, 3000);
      }
    });

    progressUnsubscribers.push(unsubscriber);
  }

  function navigateTo(path) {
    loadDirectory(path);
  }

  // Breadcrumb navigation
  $: pathParts = currentPath.split('/').filter((p) => p);

  function handleBreadcrumbClick(index) {
    if (index === -1) {
      navigateTo('/');
    } else {
      const path = '/' + pathParts.slice(0, index + 1).join('/');
      navigateTo(path);
    }
  }

  // React to active session changes - only load when session is connected
  $: if ($activeSessionIdStore && isSessionConnected) {
    loadDirectory(currentPath);
  }

  onMount(() => {
    // No need for event listeners - we use reactive store subscription
  });

  onDestroy(() => {
    progressUnsubscribers.forEach((unsub) => unsub());
  });
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
  <!-- 头部工具栏 -->
  <div class="p-3 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-center justify-between mb-2">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">文件管理</h3>
      <div class="flex items-center gap-1">
        <button 
          on:click={handleRefresh}
          class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors" 
          title="刷新"
        >
          <svg class="w-3.5 h-3.5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
        <button 
          on:click={handleUpload}
          class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors" 
          title="上传"
        >
          <svg class="w-3.5 h-3.5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
          </svg>
        </button>
        <button 
          class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors" 
          title="下载"
        >
          <svg class="w-3.5 h-3.5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0L8 8m4-4V4" />
          </svg>
        </button>
      </div>
    </div>
    
    <!-- 路径导航 -->
    <div class="flex items-center gap-1 text-xs bg-gray-50 dark:bg-gray-700 rounded-lg px-3 py-2">
      <button
        on:click={() => handleBreadcrumbClick(-1)}
        class="p-0.5 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors"
        title="根目录"
      >
        <svg class="w-3 h-3 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
        </svg>
      </button>
      {#each pathParts as part, i}
        <span class="text-gray-400 dark:text-gray-500">/</span>
        <button
          on:click={() => handleBreadcrumbClick(i)}
          class="text-purple-600 dark:text-purple-400 font-medium hover:bg-gray-200 dark:hover:bg-gray-600 px-1 rounded transition-colors"
        >
          {part}
        </button>
      {/each}
    </div>
  </div>

  <!-- 文件列表 -->
  <div class="flex-1 overflow-y-auto scrollbar-thin text-xs">
    {#if isLoading}
      <div class="flex flex-col items-center justify-center h-40 text-gray-500 dark:text-gray-400 gap-2">
        <svg class="animate-spin w-6 h-6" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <span>加载中...</span>
      </div>
    {:else if error}
      <div class="flex flex-col items-center justify-center h-40 text-red-500 dark:text-red-400 gap-2">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span class="text-center px-4">{error}</span>
        <button on:click={handleRefresh} class="text-blue-500 hover:text-blue-600 dark:text-blue-400 dark:hover:text-blue-300">
          重试
        </button>
      </div>
    {:else if files.length === 0}
      <div class="flex flex-col items-center justify-center h-40 text-gray-500 dark:text-gray-400 gap-2">
        <svg class="w-8 h-8 opacity-50" fill="currentColor" viewBox="0 0 24 24">
          <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
        </svg>
        <span>目录为空</span>
      </div>
    {:else}
      {#each files as file, index (file.path)}
        <div
          class="group flex items-center gap-2 px-3 py-2 hover:bg-purple-50 dark:hover:bg-purple-900/20 cursor-pointer transition-colors mx-2 my-0.5 rounded-lg"
          on:click={() => file.is_dir && navigateTo(file.path)}
          on:dblclick={() => { if (!file.is_dir) handleDownload(file); }}
        >
          {#if file.is_dir}
            {#if expandedDirs.has(file.path)}
              <svg class="w-3 h-3 text-gray-500 dark:text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            {:else}
              <svg class="w-3 h-3 text-gray-500 dark:text-gray-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            {/if}
            <svg class="w-3.5 h-3.5 text-amber-500 flex-shrink-0" fill="currentColor" viewBox="0 0 24 24">
              <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
          {:else}
            <div class="w-3"></div>
            <svg class="w-3.5 h-3.5 text-blue-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          {/if}
          <div class="flex-1 min-w-0">
            <div class="text-gray-900 dark:text-white font-medium truncate">{file.name}</div>
            {#if !file.is_dir}
              <div class="text-gray-500 dark:text-gray-400 flex items-center gap-2">
                <span>{formatFileSize(file.size)}</span>
                <span>•</span>
                <span>{file.modified}</span>
              </div>
            {/if}
          </div>
          <span class="text-gray-400 dark:text-gray-500 font-mono text-[10px]">{file.permissions}</span>
        </div>
      {/each}
    {/if}
  </div>
</div>
