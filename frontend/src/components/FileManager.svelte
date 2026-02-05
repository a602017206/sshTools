<script>
  import { onMount, onDestroy, tick } from 'svelte';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import {
    ListFiles,
    UploadFiles,
    DownloadFiles,
    SelectUploadFiles,
    SelectDownloadDirectory,
    CancelTransfer,
    DeleteFiles,
    RenameFile,
  } from '../../wailsjs/go/main/App.js';
  import { activeSessionIdStore, connectionsStore } from '../stores.js';
  import { uploadStore, formatFileSize, formatSpeed, getTransferPercentage } from '../stores/uploadStore.js';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  import InputDialog from './ui/InputDialog.svelte';

  // Per-session path storage
  let sessionPaths = new Map();

  let expandedDirs = new Set([]);
  let files = [];
  let isLoading = false;
  let error = null;
  let selectedFiles = new Set();

  // Transfer progress unsubscribers
  let progressUnsubscribers = [];

  // Context menu state
  let contextMenu = {
    open: false,
    x: 0,
    y: 0,
    file: null,
  };

  // Editable path and search input
  let isEditingPath = false;
  let editPathValue = '';
  let isPathInputOpen = false;
  let pathInput = '';

  // Delete confirmation dialog state
  let showDeleteConfirm = false;
  let deleteConfirmMessage = '';
  let pendingDeleteFiles = [];

  // Cancel upload confirmation dialog state
  let showCancelUploadConfirm = false;
  let pendingCancelTransfer = null;

  // Rename input dialog state
  let showRenameDialog = false;
  let renameDialogTitle = '';
  let renameDefaultName = '';
  let pendingRenameFile = null;

  // Get current session object
  $: currentSession = $activeSessionIdStore ? $connectionsStore.get($activeSessionIdStore) : null;
  $: isSessionConnected = currentSession?.connected || false;
  $: isLocalSession = currentSession?.type === 'local';
  $: canUseFileManager = isSessionConnected && !isLocalSession;

  // Current path for the active session
  let currentPath = '/';
  $: if ($activeSessionIdStore) {
    const storedPath = getSessionPath($activeSessionIdStore);
    if (storedPath && storedPath !== currentPath) {
      currentPath = storedPath;
    }
  } else {
    currentPath = '/';
  }

  // Get path for a session, default to '/'
  function getSessionPath(sessionId) {
    if (!sessionId) return '/';
    return sessionPaths.get(sessionId) || '/';
  }

  // Set path for a session
  function setSessionPath(sessionId, path) {
    if (!sessionId) return;
    const nextPaths = new Map(sessionPaths);
    nextPaths.set(sessionId, path);
    sessionPaths = nextPaths;
  }

  // Remove path for a session (cleanup)
  function removeSessionPath(sessionId) {
    if (!sessionId) return;
    const nextPaths = new Map(sessionPaths);
    nextPaths.delete(sessionId);
    sessionPaths = nextPaths;
  }

  // Sort files: directories first, then alphabetical by name
  function sortFiles(fileList) {
    if (!fileList || fileList.length === 0) return [];
    return [...fileList].sort((a, b) => {
      // Directories first
      if (a.is_dir && !b.is_dir) return -1;
      if (!a.is_dir && b.is_dir) return 1;
      // Then sort by name (case-insensitive alphabetical)
      return a.name.toLowerCase().localeCompare(b.name.toLowerCase());
    });
  }

  // Display files with parent directory entry
  $: displayFiles = (() => {
    const sortedFileList = sortFiles(files);

    if (currentPath === '/') {
      return sortedFileList;
    }

    const parentPath = currentPath.split('/').filter(Boolean).slice(0, -1).join('/') || '/';

    return [
      {
        name: '..',
        path: parentPath,
        is_dir: true,
        is_parent: true,
      },
      ...sortedFileList
    ];
  })();
  
  async function loadDirectory(path) {
    if (!$activeSessionIdStore || !isSessionConnected) return;

    if (isLocalSession) {
      return;
    }

    isLoading = true;
    error = null;

    try {
      const fileList = await ListFiles($activeSessionIdStore, path);
      files = fileList || [];
      setSessionPath($activeSessionIdStore, path);
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

  async function handleRemoveTransfer(transfer) {
    if (!transfer) return;

    if (transfer.status === 'running') {
      pendingCancelTransfer = transfer;
      showCancelUploadConfirm = true;
    } else {
      uploadStore.removeTransfer(transfer.id);
    }
  }

  async function handleConfirmCancelUpload() {
    showCancelUploadConfirm = false;
    const transfer = pendingCancelTransfer;
    pendingCancelTransfer = null;

    if (transfer) {
      try {
        await CancelTransfer(transfer.id);
      } catch (err) {
        console.error('Cancel transfer failed:', err);
      }
    }
    uploadStore.cancelTransfer(transfer.id);
  }

  function handleCancelCancelUpload() {
    showCancelUploadConfirm = false;
    pendingCancelTransfer = null;
  }

  async function handleRefresh() {
    await loadDirectory(currentPath);
  }

  async function handleUpload() {
    if (!$activeSessionIdStore || !isSessionConnected || isLocalSession) return;

    try {
      const localPaths = await SelectUploadFiles();
      if (!localPaths || localPaths.length === 0) return;

      const transferIDs = await UploadFiles($activeSessionIdStore, localPaths, currentPath);
      transferIDs.forEach((id) => subscribeToTransfer(id, 'upload'));

      // Refresh directory after a short delay
      setTimeout(() => loadDirectory(currentPath), 2000);
    } catch (err) {
      console.error('Upload failed:', err);
      error = err.message || '上传失败';
    }
  }

  async function handleDownload(file) {
    if (!$activeSessionIdStore || !isSessionConnected || isLocalSession || file.is_dir) return;

    try {
      const localDir = await SelectDownloadDirectory();
      if (!localDir) return;

      const transferIDs = await DownloadFiles($activeSessionIdStore, [file.path], localDir);
      transferIDs.forEach((id) => subscribeToTransfer(id, 'download'));
    } catch (err) {
      console.error('Download failed:', err);
      error = err.message || '下载失败';
    }
  }

  function handleSelectFile(file) {
    if (!file || file.is_parent) return;
    selectedFiles = new Set([file.path]);
  }

  function handleItemClick(file) {
    if (!file || file.is_parent) return;
    handleSelectFile(file);
  }

  function handleParentClick(file) {
    if (!file || !file.is_parent) return;
    navigateTo(file.path);
  }

  function handleContextMenu(event, file) {
    if (file?.is_parent) return;
    event.preventDefault();
    handleSelectFile(file);
    contextMenu = {
      open: true,
      x: event.clientX,
      y: event.clientY,
      file,
    };
  }

  function closeContextMenu() {
    if (!contextMenu.open) return;
    contextMenu = { ...contextMenu, open: false };
  }

  async function handleDeleteSelection() {
    if (!$activeSessionIdStore || !isSessionConnected || isLocalSession) return;
    if (selectedFiles.size === 0) return;

    // Show confirmation dialog
    pendingDeleteFiles = Array.from(selectedFiles);
    const count = pendingDeleteFiles.length;
    deleteConfirmMessage = count === 1
      ? `确定要删除选中的 1 个文件/文件夹吗？`
      : `确定要删除选中的 ${count} 个文件/文件夹吗？`;
    showDeleteConfirm = true;
  }

  async function handleConfirmDelete() {
    showDeleteConfirm = false;
    try {
      await DeleteFiles($activeSessionIdStore, pendingDeleteFiles);
      selectedFiles = new Set();
      pendingDeleteFiles = [];
      await loadDirectory(currentPath);
    } catch (err) {
      console.error('Delete failed:', err);
      error = err.message || '删除失败';
    }
  }

  function handleCancelDelete() {
    showDeleteConfirm = false;
    pendingDeleteFiles = [];
  }

  async function handleContextDelete() {
    const file = contextMenu.file;
    if (!file) return;

    closeContextMenu();

    // Show confirmation dialog
    pendingDeleteFiles = [file.path];
    deleteConfirmMessage = `确定要删除 "${file.name}" 吗？`;
    showDeleteConfirm = true;
  }

  async function handleContextDownload() {
    const file = contextMenu.file;
    closeContextMenu();
    if (file && !file.is_dir) {
      await handleDownload(file);
    }
  }

  async function handleContextRename() {
    const file = contextMenu.file;
    closeContextMenu();
    if (!file) return;

    // Show rename input dialog
    pendingRenameFile = file;
    renameDialogTitle = '重命名';
    renameDefaultName = file.name;
    showRenameDialog = true;
  }

  async function handleConfirmRename(newName) {
    showRenameDialog = false;
    const file = pendingRenameFile;
    pendingRenameFile = null;

    if (!file || !newName || newName.trim() === file.name) return;

    try {
      const basePath = currentPath.endsWith('/') ? currentPath : `${currentPath}/`;
      const newPath = `${basePath}${newName.trim()}`;
      await RenameFile($activeSessionIdStore, file.path, newPath);
      await loadDirectory(currentPath);
    } catch (err) {
      console.error('Rename failed:', err);
      error = err.message || '重命名失败';
    }
  }

  function handleCancelRename() {
    showRenameDialog = false;
    pendingRenameFile = null;
  }

  // Transfer progress subscription
  function subscribeToTransfer(transferID, kind) {
    const eventName = `sftp:progress:${transferID}`;
    const unsubscriber = EventsOn(eventName, (progress) => {
      if (kind === 'upload') {
        const transfer = {
          id: transferID,
          filename: progress.filename || progress.fileName || 'unknown',
          bytesSent: progress.bytes_sent ?? progress.bytesSent ?? 0,
          totalBytes: progress.total_bytes ?? progress.totalBytes ?? 0,
          percentage: progress.percentage ?? 0,
          speed: progress.speed ?? 0,
          status: progress.status || 'running',
          error: progress.error || '',
        };

        const existing = $uploadStore.transfers.find(t => t.id === transferID);
        if (existing) {
          uploadStore.updateTransfer(transferID, transfer);
        } else {
          uploadStore.addTransfer(transfer);
        }
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

  function handleStartEditPath() {
    editPathValue = currentPath;
    isEditingPath = true;
  }

  function handleSaveEditPath() {
    if (editPathValue.trim()) {
      navigateTo(editPathValue.trim());
    }
    isEditingPath = false;
  }

  function handleCancelEditPath() {
    isEditingPath = false;
    editPathValue = '';
  }

  function handlePathJump() {
    if (pathInput.trim()) {
      navigateTo(pathInput.trim());
      isPathInputOpen = false;
      pathInput = '';
    }
  }

  // React to active session changes - only load when session is connected and can use file manager
  $: if ($activeSessionIdStore && canUseFileManager) {
    loadDirectory(currentPath);
  }

  // Cleanup session paths when connections are removed
  $: if ($connectionsStore) {
    const activeSessionIds = new Set($connectionsStore.keys());
    // Remove paths for deleted sessions
    const nextPaths = new Map(sessionPaths);
    for (const sessionId of nextPaths.keys()) {
      if (!activeSessionIds.has(sessionId)) {
        nextPaths.delete(sessionId);
      }
    }
    sessionPaths = nextPaths;
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
           disabled={isLocalSession}
           class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
           title="刷新"
         >
           <svg class="w-3.5 h-3.5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
             <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
           </svg>
         </button>
          <button
            on:click={handleUpload}
            disabled={isLocalSession}
            class="p-1.5 hover:bg-blue-50 dark:hover:bg-blue-900/30 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            title="上传到服务器"
          >
            <svg class="w-4 h-4 text-blue-500 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
          </button>
          <button
            class="p-1.5 hover:bg-green-50 dark:hover:bg-green-900/30 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            title="下载到本地"
          >
            <svg class="w-4 h-4 text-green-500 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
          </button>
          <button
            on:click={handleDeleteSelection}
            disabled={isLocalSession || selectedFiles.size === 0}
            class="p-1.5 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            title="删除选中"
          >
            <svg class="w-4 h-4 text-red-500 dark:text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 7h12M9 7V5a1 1 0 011-1h4a1 1 0 011 1v2m-7 0v12a1 1 0 001 1h6a1 1 0 001-1V7" />
            </svg>
          </button>
        </div>
    </div>

    <!-- 路径导航 -->
    <div class="flex items-center gap-2">
      <div class="flex-1 flex items-center gap-1 text-xs bg-gray-50 dark:bg-gray-700 rounded-lg px-3 py-2">
        <button
          on:click={() => handleBreadcrumbClick(-1)}
          class="p-0.5 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors"
          title="根目录"
        >
          <svg class="w-3 h-3 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
        </button>
        {#if isEditingPath}
          <span class="text-gray-400 dark:text-gray-500">/</span>
          <input
            type="text"
            bind:value={editPathValue}
            on:keydown={(e) => {
              if (e.key === 'Enter') {
                handleSaveEditPath();
              } else if (e.key === 'Escape') {
                handleCancelEditPath();
              }
            }}
            on:blur={handleSaveEditPath}
            use:focus
            class="flex-1 bg-white dark:bg-gray-800 border border-purple-300 dark:border-purple-600 rounded px-2 py-1 text-purple-600 dark:text-purple-400 font-medium focus:outline-none focus:ring-2 focus:ring-purple-500"
          />
        {:else}
          <span
            on:click={handleStartEditPath}
            class="flex-1 text-purple-600 dark:text-purple-400 font-medium cursor-text hover:bg-purple-100 dark:hover:bg-purple-900/30 px-2 py-1 rounded transition-colors"
            title="点击编辑路径"
          >
            {#if currentPath === '/'}
              /
            {:else}
              {currentPath.split('/').filter(Boolean).join(' / ')}
            {/if}
          </span>
        {/if}
      </div>

      <button
        on:click={() => isPathInputOpen = !isPathInputOpen}
        class="p-2 hover:bg-purple-100 dark:hover:bg-purple-900/30 rounded-lg transition-colors"
        title="跳转到指定目录"
      >
        <svg class="w-3.5 h-3.5 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </button>
    </div>

    {#if isPathInputOpen}
      <div class="mt-2 flex items-center gap-2">
        <input
          type="text"
          bind:value={pathInput}
          on:keydown={(e) => {
            if (e.key === 'Enter') {
              handlePathJump();
            } else if (e.key === 'Escape') {
              isPathInputOpen = false;
              pathInput = '';
            }
          }}
          placeholder="输入路径,如：/var/www/html"
          use:focus
          class="flex-1 px-3 py-2 text-xs bg-white dark:bg-gray-800 border border-purple-200 dark:border-purple-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
        />
        <button
          on:click={handlePathJump}
          class="px-3 py-2 bg-purple-600 hover:bg-purple-700 text-white text-xs rounded-lg transition-colors font-medium"
        >
          跳转
        </button>
        <button
          on:click={() => {
            isPathInputOpen = false;
            pathInput = '';
          }}
          class="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
          title="取消"
        >
          <svg class="w-3.5 h-3.5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    {/if}
  </div>

  <!-- 文件列表 -->
  <div
    class="flex-1 overflow-y-auto scrollbar-thin text-xs"
    on:click={closeContextMenu}
  >
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
     {:else if displayFiles.length === 0}
       <div class="flex flex-col items-center justify-center h-40 text-gray-500 dark:text-gray-400 gap-2">
         <svg class="w-8 h-8 opacity-50" fill="currentColor" viewBox="0 0 24 24">
           <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
         </svg>
         <span>目录为空</span>
       </div>
     {:else if isLocalSession}
       <div class="flex flex-col items-center justify-center h-40 text-gray-500 dark:text-gray-400 gap-2">
         <svg class="w-8 h-8 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
           <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
         </svg>
         <span class="text-center px-4">本地终端不支持文件管理</span>
         <div class="text-xs text-gray-400 dark:text-gray-500 mt-2">
           文件管理仅适用于 SSH 远程连接
         </div>
       </div>
     {:else}
      {#each displayFiles as file, index (file.path)}
        <div
          class="group flex items-center gap-2 px-3 py-2 cursor-pointer transition-colors mx-2 my-0.5 rounded-lg {file.is_parent ? 'text-gray-500 dark:text-gray-400 italic' : ''} {selectedFiles.has(file.path) && !file.is_parent ? 'bg-purple-100 dark:bg-purple-900/40' : 'hover:bg-purple-50 dark:hover:bg-purple-900/20'}"
          on:click={() => (file.is_parent ? handleParentClick(file) : handleItemClick(file))}
          on:dblclick={() => {
            if (file.is_dir && !file.is_parent) {
              navigateTo(file.path);
            } else if (!file.is_dir) {
              handleDownload(file);
            }
          }}
          on:contextmenu={(event) => handleContextMenu(event, file)}
        >
          {#if file.is_parent}
            <svg class="w-3.5 h-3.5 text-gray-400 dark:text-gray-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
            <div class="flex-1 min-w-0">
              <div class="text-gray-600 dark:text-gray-300">{file.name}</div>
              <div class="text-gray-400 dark:text-gray-500 text-[10px]">返回上一层</div>
            </div>
          {:else if file.is_dir}
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
            <div class="flex-1 min-w-0">
              <div class="text-gray-900 dark:text-white font-medium truncate">{file.name}</div>
            </div>
          {:else}
            <div class="w-3"></div>
            <svg class="w-3.5 h-3.5 text-blue-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <div class="flex-1 min-w-0">
              <div class="text-gray-900 dark:text-white font-medium truncate">{file.name}</div>
              <div class="text-gray-500 dark:text-gray-400 flex items-center gap-2">
                <span>{formatFileSize(file.size)}</span>
                <span>•</span>
                <span>{file.modified}</span>
              </div>
            </div>
            <span class="text-gray-400 dark:text-gray-500 font-mono text-[10px]">{file.mode}</span>
          {/if}
        </div>
      {/each}
    {/if}
  </div>

  {#if contextMenu.open}
    <div
      class="fixed z-50 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-md shadow-lg text-xs py-1 min-w-[140px]"
      style={`left: ${contextMenu.x}px; top: ${contextMenu.y}px;`}
    >
      <button
        class="w-full text-left px-3 py-1.5 hover:bg-gray-100 dark:hover:bg-gray-700"
        on:click={handleContextDownload}
        disabled={contextMenu.file?.is_dir}
      >
        下载
      </button>
      <button
        class="w-full text-left px-3 py-1.5 hover:bg-gray-100 dark:hover:bg-gray-700"
        on:click={handleContextRename}
      >
        重命名
      </button>
      <button
        class="w-full text-left px-3 py-1.5 hover:bg-red-50 dark:hover:bg-red-900/30 text-red-600 dark:text-red-400"
        on:click={handleContextDelete}
      >
        删除
      </button>
    </div>
  {/if}

  <ConfirmDialog
    bind:isOpen={showDeleteConfirm}
    title="删除文件"
    message={deleteConfirmMessage}
    type="danger"
    confirmText="删除"
    cancelText="取消"
    onConfirm={handleConfirmDelete}
    onCancel={handleCancelDelete}
  />

  <ConfirmDialog
    bind:isOpen={showCancelUploadConfirm}
    title="取消上传"
    message="上传未完成，是否终止上传？"
    type="warning"
    confirmText="终止上传"
    cancelText="继续上传"
    onConfirm={handleConfirmCancelUpload}
    onCancel={handleCancelCancelUpload}
  />

  <InputDialog
    bind:isOpen={showRenameDialog}
    title={renameDialogTitle}
    message="请输入新的文件/文件夹名称"
    placeholder="新名称"
    defaultValue={renameDefaultName}
    confirmText="重命名"
    cancelText="取消"
    onConfirm={handleConfirmRename}
    onCancel={handleCancelRename}
  />
</div>
