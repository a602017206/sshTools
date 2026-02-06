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
    GetCurrentPath,
    GetFileManagerSettings,
    UpdateFileManagerSettings,
    SearchDirectories,
  } from '../../wailsjs/go/main/App.js';
  import { activeSessionIdStore, connectionsStore, fileManagerConfigStore } from '../stores.js';
  import { uploadStore, formatFileSize, formatSpeed, getTransferPercentage } from '../stores/uploadStore.js';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  import InputDialog from './ui/InputDialog.svelte';

  // Svelte action to focus element on mount
  function focus(node) {
    node.focus();
    return {
      destroy() {}
    };
  }

  // Per-session path storage
  let sessionPaths = new Map();

  // File manager configuration state
  let fileManagerConfig = {
    directoryTracking: false,
    historyEnabled: true,
    historyLimit: 5,
    history: [],
  };

  // Session-level temporary tracking toggle (not saved to config)
  let sessionDirectoryTracking = new Map();
  let cwdEventUnsubscriber = null;
  let isManualNavigation = false; // Flag to track if navigation was manual (should pause tracking)

  // Settings dialog state
  let showSettingsDialog = false;

  // History dropdown state
  let showHistoryDropdown = false;
  let historyFilter = '';

  // Directory fuzzy search state
  let isDirSearchOpen = false;
  let dirSearchQuery = '';
  let dirSearchResults = [];
  let dirSearchLoading = false;
  let dirSearchSelectedIndex = -1;
  let dirSearchAbortController = null;

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

  // Editable path
  let isEditingPath = false;
  let editPathValue = '';

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

  // Get connection ID for file manager config
  function getConnectionId() {
    if (!currentSession) return null;
    return currentSession.connection?.id || null;
  }

  // Load file manager config from backend
  $: if ($activeSessionIdStore && currentConnectionId) {
    loadFileManagerConfig(currentConnectionId);
  }

  async function loadFileManagerConfig(connectionId) {
    try {
      const { GetFileManagerSettings } = window.wailsBindings;
      if (typeof GetFileManagerSettings !== 'function') return;

      const config = await GetFileManagerSettings(connectionId);
      // Map snake_case backend response to camelCase frontend state
      fileManagerConfig = {
        directoryTracking: config?.directory_tracking ?? false,
        historyEnabled: config?.history_enabled ?? true,
        historyLimit: config?.history_limit ?? 5,
        history: config?.history ?? [],
      };

      // Initialize session tracking state from server config only if not already set
      if (!sessionDirectoryTracking.has(connectionId)) {
        const nextTracking = new Map(sessionDirectoryTracking);
        nextTracking.set(connectionId, fileManagerConfig.directoryTracking || false);
        sessionDirectoryTracking = nextTracking;
      }
    } catch (error) {
      console.error('Failed to load file manager config:', error);
    }
  }

  async function handleSaveSettings() {
    try {
      const { UpdateFileManagerSettings } = window.wailsBindings;
      if (typeof UpdateFileManagerSettings !== 'function') return;

      if (!currentConnectionId) return;

      await UpdateFileManagerSettings(currentConnectionId, {
        directory_tracking: fileManagerConfig.directoryTracking,
        history_enabled: fileManagerConfig.historyEnabled,
        history_limit: fileManagerConfig.historyLimit,
        history: fileManagerConfig.history,
      });

      // Note: Don't update sessionDirectoryTracking here
      // It reflects the temporary pause state, not the feature toggle
    } catch (error) {
      console.error('Failed to save file manager settings:', error);
    }
  }

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
    navigateTo(file.path, true);
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

  function navigateTo(path, isManual = false) {
    if (isManual && currentTrackingEnabled && currentConnectionId) {
      const nextTracking = new Map(sessionDirectoryTracking);
      nextTracking.set(currentConnectionId, false);
      sessionDirectoryTracking = nextTracking;
      isManualNavigation = true;
    }

    if (fileManagerConfig.historyEnabled && path !== currentPath) {
      const maxHistory = fileManagerConfig.historyLimit || 5;
      const newHistory = [...(fileManagerConfig.history || [])];

      const existingIndex = newHistory.indexOf(path);
      if (existingIndex !== -1) {
        newHistory.splice(existingIndex, 1);
      }
      newHistory.unshift(path);

      const limitedHistory = newHistory.slice(0, maxHistory);
      fileManagerConfig.history = limitedHistory;
      handleSaveSettings();
    }

    loadDirectory(path);
  }

  // Sync current directory from server when re-enabling tracking
  async function syncCurrentDirectoryFromServer() {
    if (!$activeSessionIdStore) return;
    
    try {
      const { GetCurrentPath } = window.wailsBindings;
      if (typeof GetCurrentPath !== 'function') return;
      
      const serverCwd = await GetCurrentPath($activeSessionIdStore);
      if (serverCwd && serverCwd !== currentPath) {
        // Navigate without triggering manual navigation pause
        isManualNavigation = false;
        navigateTo(serverCwd, false);
      }
    } catch (error) {
      console.error('Failed to sync directory from server:', error);
    }
  }

  // Start directory tracking via event listener
  function startCWDTracking() {
    if (!fileManagerConfig.directoryTracking || !currentTrackingEnabled) return;
    if (cwdEventUnsubscriber) return; // Already listening

    const sessionId = $activeSessionIdStore;
    if (!sessionId) return;

    // Subscribe to CWD change events from backend
    cwdEventUnsubscriber = EventsOn(`ssh:cwd:${sessionId}`, (cwd) => {
      // Only auto-navigate if not in manual navigation mode
      if (cwd && cwd !== currentPath && !isManualNavigation) {
        navigateTo(cwd, false);
      }
    });
  }

  // Stop directory tracking
  function stopCWDTracking() {
    if (cwdEventUnsubscriber) {
      cwdEventUnsubscriber();
      cwdEventUnsubscriber = null;
    }
  }

  // Breadcrumb navigation
  $: pathParts = currentPath.split('/').filter((p) => p);

  // Filter history for search
  $: filteredHistory = fileManagerConfig.historyEnabled && historyFilter
    ? fileManagerConfig.history.filter(path =>
        path.toLowerCase().includes(historyFilter.toLowerCase())
      )
    : fileManagerConfig.history || [];

  // Get current session tracking state - use currentSession directly so Svelte can track dependencies
  $: currentConnectionId = currentSession?.connection?.id || null;
  $: currentTrackingEnabled = currentConnectionId && sessionDirectoryTracking.get(currentConnectionId) || false;

  function handleBreadcrumbClick(index) {
    if (index === -1) {
      navigateTo('/', true);
    } else {
      const path = '/' + pathParts.slice(0, index + 1).join('/');
      navigateTo(path, true);
    }
  }

  function handleStartEditPath() {
    editPathValue = currentPath;
    isEditingPath = true;
  }

  function handleSaveEditPath() {
    if (editPathValue.trim()) {
      navigateTo(editPathValue.trim(), true);
    }
    isEditingPath = false;
  }

  function handleCancelEditPath() {
    isEditingPath = false;
    editPathValue = '';
  }

  // Fuzzy search directories
  async function handleDirSearch() {
    if (!dirSearchQuery.trim() || !$activeSessionIdStore) return;

    // Cancel previous search
    if (dirSearchAbortController) {
      dirSearchAbortController.abort();
    }
    dirSearchAbortController = new AbortController();

    dirSearchLoading = true;
    dirSearchResults = [];
    dirSearchSelectedIndex = -1;

    try {
      const results = await SearchDirectories(
        $activeSessionIdStore,
        currentPath,
        dirSearchQuery.trim(),
        3, // maxDepth
        50 // maxResults
      );

      // Check if search was aborted
      if (dirSearchAbortController.signal.aborted) return;

      dirSearchResults = results || [];
    } catch (err) {
      console.error('Directory search failed:', err);
      error = err.message || '搜索失败';
    } finally {
      dirSearchLoading = false;
    }
  }

  // Debounced search
  let dirSearchTimeout;
  function handleDirSearchInput() {
    clearTimeout(dirSearchTimeout);
    if (dirSearchQuery.trim().length >= 1) {
      dirSearchTimeout = setTimeout(() => {
        handleDirSearch();
      }, 300);
    } else {
      dirSearchResults = [];
    }
  }

  function handleDirSearchKeydown(e) {
    if (e.key === 'Enter') {
      e.preventDefault();
      if (dirSearchSelectedIndex >= 0 && dirSearchResults[dirSearchSelectedIndex]) {
        navigateTo(dirSearchResults[dirSearchSelectedIndex].path, true);
        closeDirSearch();
      } else {
        handleDirSearch();
      }
    } else if (e.key === 'Escape') {
      closeDirSearch();
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (dirSearchSelectedIndex < dirSearchResults.length - 1) {
        dirSearchSelectedIndex++;
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (dirSearchSelectedIndex > 0) {
        dirSearchSelectedIndex--;
      }
    }
  }

  function closeDirSearch() {
    isDirSearchOpen = false;
    dirSearchQuery = '';
    dirSearchResults = [];
    dirSearchSelectedIndex = -1;
    if (dirSearchAbortController) {
      dirSearchAbortController.abort();
      dirSearchAbortController = null;
    }
  }

  function openDirSearch() {
    isDirSearchOpen = true;
    dirSearchQuery = '';
    dirSearchResults = [];
    dirSearchSelectedIndex = -1;
  }

  // Highlight matched text in search results
  function highlightMatch(text, query) {
    if (!query) return text;
    const regex = new RegExp(`(${escapeRegExp(query)})`, 'gi');
    return text.replace(regex, '<mark class="bg-yellow-200 dark:bg-yellow-600 text-yellow-900 dark:text-yellow-100 px-0.5 rounded">$1</mark>');
  }

  function escapeRegExp(string) {
    return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  }

  // React to active session changes - only load when session is connected and can use file manager
  $: if ($activeSessionIdStore && canUseFileManager && currentConnectionId) {
    loadFileManagerConfig(currentConnectionId);
    loadDirectory(currentPath);
    startCWDTracking();
  } else {
    stopCWDTracking();
  }

  // Clear all file manager state when no active session or cannot use file manager
  $: if (!$activeSessionIdStore || !canUseFileManager) {
    files = [];
    error = null;
    selectedFiles = new Set();
    expandedDirs = new Set();
    currentPath = '/';
    sessionPaths = new Map();
    sessionDirectoryTracking = new Map();
    // Reset file manager config history but keep settings
    fileManagerConfig = {
      directoryTracking: false,
      historyEnabled: true,
      historyLimit: 5,
      history: [],
    };
    // Close any open UI elements
    showSettingsDialog = false;
    showHistoryDropdown = false;
    isDirSearchOpen = false;
    contextMenu = { open: false, x: 0, y: 0, file: null };
  }

  // Restart tracking when tracking settings change
  $: if ($activeSessionIdStore && canUseFileManager && fileManagerConfig.directoryTracking && currentTrackingEnabled) {
    startCWDTracking();
  } else {
    stopCWDTracking();
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
    stopCWDTracking();
    clearTimeout(dirSearchTimeout);
    if (dirSearchAbortController) {
      dirSearchAbortController.abort();
    }
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

          {#if fileManagerConfig.directoryTracking}
            <button
              on:click={async () => {
                if (currentConnectionId) {
                  const newState = !currentTrackingEnabled;
                  const nextTracking = new Map(sessionDirectoryTracking);
                  nextTracking.set(currentConnectionId, newState);
                  sessionDirectoryTracking = nextTracking;
                  
                  if (newState) {
                    isManualNavigation = false;
                    await syncCurrentDirectoryFromServer();
                  }
                }
              }}
              disabled={isLocalSession}
              class={`flex items-center gap-1 px-2 py-1 rounded-lg text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed ${
                currentTrackingEnabled
                  ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400'
                  : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
              }`}
              title="切换目录跟踪"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              <span>{currentTrackingEnabled ? '跟踪中' : '已暂停'}</span>
            </button>
          {/if}
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
          <button
            on:click={() => showSettingsDialog = true}
            class="p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            title="文件管理设置"
          >
            <svg class="w-4 h-4 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
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

       <!-- Directory fuzzy search button -->
       <button
         on:click={openDirSearch}
         class="p-2 hover:bg-amber-100 dark:hover:bg-amber-900/30 rounded-lg transition-colors"
         title="模糊搜索文件夹"
       >
         <svg class="w-3.5 h-3.5 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
           <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
         </svg>
       </button>

       <div class="relative">
         <button
           on:click={() => showHistoryDropdown = !showHistoryDropdown}
           class="p-2 hover:bg-blue-100 dark:hover:bg-blue-900/30 rounded-lg transition-colors"
           title="路径历史"
         >
           <svg class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
             <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
           </svg>
         </button>

         {#if showHistoryDropdown && fileManagerConfig.historyEnabled}
           <div class="absolute right-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg w-64 max-h-64 overflow-y-auto z-10">
             <div class="p-2 border-b border-gray-200 dark:border-gray-700">
               <input
                 type="text"
                 bind:value={historyFilter}
                 placeholder="搜索历史..."
                 class="w-full px-2 py-1 text-xs bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded focus:outline-none focus:ring-1 focus:ring-purple-500"
               />
             </div>
              {#if filteredHistory.length > 0}
                {#each filteredHistory as path, index (path)}
                  <button
                    on:click={() => {
                      navigateTo(path, true);
                      showHistoryDropdown = false;
                      historyFilter = '';
                    }}
                    class="w-full text-left px-3 py-1.5 text-xs text-gray-900 dark:text-white hover:bg-purple-50 dark:hover:bg-purple-900/30 truncate"
                  >
                    {path}
                  </button>
                {/each}
             {:else}
               <div class="px-3 py-2 text-xs text-gray-500 dark:text-gray-400">
                 {historyFilter ? '无匹配的历史' : '暂无历史记录'}
               </div>
             {/if}
           </div>
         {/if}
       </div>
    </div>

    {#if isDirSearchOpen}
      <div class="mt-2 relative">
        <div class="flex items-center gap-2">
          <div class="relative flex-1">
            <div class="relative">
              <input
                type="text"
                bind:value={dirSearchQuery}
                on:keydown={handleDirSearchKeydown}
                on:input={handleDirSearchInput}
                placeholder="搜索文件夹名称..."
                use:focus
                class="w-full px-3 py-2 pl-9 text-xs bg-white dark:bg-gray-800 border border-amber-200 dark:border-amber-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-amber-500 focus:border-transparent"
              />
              <svg class="w-4 h-4 text-amber-500 absolute left-3 top-1/2 transform -translate-y-1/2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              {#if dirSearchLoading}
                <svg class="w-4 h-4 text-amber-500 absolute right-3 top-1/2 transform -translate-y-1/2 animate-spin" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
              {/if}
            </div>
            {#if dirSearchResults.length > 0}
              <div class="absolute left-0 right-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg max-h-64 overflow-y-auto z-10">
                <div class="px-3 py-1.5 text-xs text-gray-500 dark:text-gray-400 border-b border-gray-100 dark:border-gray-700">
                  找到 {dirSearchResults.length} 个文件夹
                </div>
                {#each dirSearchResults as result, index (result.path)}
                  <button
                    on:click={() => {
                      navigateTo(result.path, true);
                      closeDirSearch();
                    }}
                    on:mouseenter={() => dirSearchSelectedIndex = index}
                    class={`w-full text-left px-3 py-2 text-xs hover:bg-amber-50 dark:hover:bg-amber-900/30 transition-colors ${
                      index === dirSearchSelectedIndex ? 'bg-amber-100 dark:bg-amber-900/50' : ''
                    }`}
                  >
                    <div class="flex items-center gap-2">
                      <svg class="w-3.5 h-3.5 text-amber-500 flex-shrink-0" fill="currentColor" viewBox="0 0 24 24">
                        <path d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
                      </svg>
                      <div class="flex-1 min-w-0">
                        <div class="text-gray-900 dark:text-white font-medium truncate">
                          {@html highlightMatch(result.name, dirSearchQuery)}
                        </div>
                        <div class="text-gray-400 dark:text-gray-500 text-[10px] truncate">
                          {result.path}
                        </div>
                      </div>
                    </div>
                  </button>
                {/each}
              </div>
            {:else if dirSearchQuery.trim() && !dirSearchLoading}
              <div class="absolute left-0 right-0 top-full mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-10">
                <div class="px-3 py-2 text-xs text-gray-500 dark:text-gray-400 text-center">
                  未找到匹配的文件夹
                </div>
              </div>
            {/if}
          </div>
          <button
            on:click={closeDirSearch}
            class="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
            title="取消"
          >
            <svg class="w-3.5 h-3.5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
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
              navigateTo(file.path, true);
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

  <!-- File Manager Settings Dialog -->
  {#if showSettingsDialog}
    <div class="fixed inset-0 z-50 flex items-center justify-center">
      <!-- Backdrop with blur -->
      <div 
        class="fixed inset-0 bg-gradient-to-br from-slate-900/40 via-slate-800/20 to-slate-700/10 dark:from-black/70 dark:via-black/55 dark:to-slate-900/40 backdrop-blur-md transition-opacity"
        on:click={() => showSettingsDialog = false}
      ></div>
      
      <!-- Dialog Container -->
      <div class="relative bg-white/95 dark:bg-slate-900/90 rounded-2xl shadow-[0_20px_70px_-30px_rgba(15,23,42,0.8)] w-[420px] max-h-[85vh] overflow-hidden flex flex-col border border-slate-200/70 dark:border-slate-700/60 ring-1 ring-slate-200/60 dark:ring-slate-700/40">
        <!-- Header -->
        <div class="px-5 py-4 bg-gradient-to-r from-slate-50 via-white to-slate-50 dark:from-slate-900 dark:via-slate-900/80 dark:to-slate-950 border-b border-slate-200/70 dark:border-slate-800/80">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="p-2 bg-gradient-to-br from-purple-500 to-indigo-600 rounded-xl shadow-lg shadow-purple-500/20">
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
              </div>
              <div>
                <h3 class="text-lg font-semibold tracking-tight bg-gradient-to-r from-slate-800 to-cyan-600 dark:from-slate-100 dark:to-cyan-300 bg-clip-text text-transparent">文件管理设置</h3>
                <div class="text-[11px] uppercase tracking-[0.15em] text-slate-400 dark:text-slate-500 mt-0.5">Preferences</div>
              </div>
            </div>
            <button
              on:click={() => showSettingsDialog = false}
              class="p-2 rounded-xl transition-all duration-200 group hover:bg-slate-200/70 dark:hover:bg-slate-800/70"
              title="关闭"
            >
              <svg class="w-5 h-5 text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-y-auto p-5 space-y-4 bg-slate-50/50 dark:bg-slate-950/30">
          <!-- Section: Features -->
          <div class="bg-white dark:bg-slate-900/80 rounded-xl border border-slate-200/60 dark:border-slate-800/70 p-4 shadow-sm">
            <div class="flex items-center gap-2 mb-4 pb-3 border-b border-slate-100 dark:border-slate-800/60">
              <svg class="w-4 h-4 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">功能设置</span>
            </div>

            <!-- Directory Tracking -->
            <div class="flex items-center justify-between py-2">
              <div class="flex items-center gap-3">
                <div class="p-1.5 bg-purple-50 dark:bg-purple-900/30 rounded-lg">
                  <svg class="w-4 h-4 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </div>
                <div>
                  <div class="text-sm font-medium text-slate-900 dark:text-white">目录跟踪</div>
                  <div class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">实时同步终端目录变化</div>
                </div>
              </div>
              <button
                on:click={() => {
                  fileManagerConfig.directoryTracking = !fileManagerConfig.directoryTracking;
                  handleSaveSettings();
                }}
                class={`relative inline-flex h-6 w-11 items-center rounded-full transition-all duration-300 ${
                  fileManagerConfig.directoryTracking ? 'bg-gradient-to-r from-purple-500 to-indigo-600 shadow-lg shadow-purple-500/30' : 'bg-slate-200 dark:bg-slate-600'
                }`}
              >
                <span
                  class={`inline-block h-4 w-4 transform rounded-full bg-white shadow-md transition-all duration-300 ${
                    fileManagerConfig.directoryTracking ? 'translate-x-6' : 'translate-x-1'
                  }`}
                ></span>
              </button>
            </div>
          </div>

          <!-- Section: History -->
          <div class="bg-white dark:bg-slate-900/80 rounded-xl border border-slate-200/60 dark:border-slate-800/70 p-4 shadow-sm">
            <div class="flex items-center gap-2 mb-4 pb-3 border-b border-slate-100 dark:border-slate-800/60">
              <svg class="w-4 h-4 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">历史记录</span>
            </div>

            <!-- History Enabled -->
            <div class="flex items-center justify-between py-2 mb-4">
              <div class="flex items-center gap-3">
                <div class="p-1.5 bg-blue-50 dark:bg-blue-900/30 rounded-lg">
                  <svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                </div>
                <div>
                  <div class="text-sm font-medium text-slate-900 dark:text-white">启用历史记录</div>
                  <div class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">自动记录访问过的目录</div>
                </div>
              </div>
              <button
                on:click={() => {
                  fileManagerConfig.historyEnabled = !fileManagerConfig.historyEnabled;
                  if (!fileManagerConfig.historyEnabled) {
                    fileManagerConfig.history = [];
                  }
                  handleSaveSettings();
                }}
                class={`relative inline-flex h-6 w-11 items-center rounded-full transition-all duration-300 ${
                  fileManagerConfig.historyEnabled ? 'bg-gradient-to-r from-blue-500 to-cyan-500 shadow-lg shadow-blue-500/30' : 'bg-slate-200 dark:bg-slate-600'
                }`}
              >
                <span
                  class={`inline-block h-4 w-4 transform rounded-full bg-white shadow-md transition-all duration-300 ${
                    fileManagerConfig.historyEnabled ? 'translate-x-6' : 'translate-x-1'
                  }`}
                ></span>
              </button>
            </div>

            <!-- History Limit -->
            {#if fileManagerConfig.historyEnabled}
              <div class="py-3 px-3 bg-slate-50/80 dark:bg-slate-800/50 rounded-xl">
                <div class="flex items-center justify-between mb-3">
                  <div class="flex items-center gap-2">
                    <svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 20l4-16m2 16l4-16M6 9h14M4 15h14" />
                    </svg>
                    <span class="text-xs font-medium text-slate-600 dark:text-slate-400">历史记录数量</span>
                  </div>
                  <span class="text-sm font-bold text-purple-600 dark:text-purple-400 min-w-[2rem] text-center">
                    {fileManagerConfig.historyLimit}
                  </span>
                </div>
                <div class="relative">
                  <input
                    type="range"
                    min="5"
                    max="50"
                    step="5"
                    bind:value={fileManagerConfig.historyLimit}
                    on:change={handleSaveSettings}
                    class="w-full h-2 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-purple-600 focus:outline-none focus:ring-2 focus:ring-purple-500/30"
                  />
                  <div class="flex justify-between text-[10px] text-slate-400 mt-1.5">
                    <span>5</span>
                    <span>50</span>
                  </div>
                </div>
              </div>
            {/if}
          </div>

          <!-- Current History Section -->
          {#if fileManagerConfig.historyEnabled && fileManagerConfig.history.length > 0}
            <div class="bg-white dark:bg-slate-900/80 rounded-xl border border-slate-200/60 dark:border-slate-800/70 p-4 shadow-sm">
              <div class="flex items-center justify-between mb-3 pb-2 border-b border-slate-100 dark:border-slate-800/60">
                <div class="flex items-center gap-2">
                  <svg class="w-4 h-4 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                  </svg>
                  <span class="text-sm font-semibold text-slate-700 dark:text-slate-200">
                    当前历史记录
                    <span class="ml-1.5 text-xs px-1.5 py-0.5 bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-400 rounded-full">
                      {fileManagerConfig.history.length}
                    </span>
                  </span>
                </div>
                <button
                  on:click={() => {
                    fileManagerConfig.history = [];
                    handleSaveSettings();
                  }}
                  class="text-xs text-red-500 hover:text-red-600 dark:text-red-400 dark:hover:text-red-300 transition-colors flex items-center gap-1"
                >
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                  清空
                </button>
              </div>
              <div class="space-y-1.5 max-h-40 overflow-y-auto pr-1 scrollbar-thin">
                {#each fileManagerConfig.history.slice().reverse() as path, index (index)}
                  <div class="group flex items-center justify-between text-xs py-2 px-3 bg-slate-50 dark:bg-slate-800/60 rounded-lg border border-slate-100 dark:border-slate-700/50 hover:border-purple-300 dark:hover:border-purple-700/50 transition-all duration-200">
                    <div class="flex items-center gap-2 flex-1 min-w-0">
                      <div class="w-5 h-5 flex items-center justify-center rounded-md bg-purple-100 dark:bg-purple-900/40 text-purple-600 dark:text-purple-400 font-bold text-[10px]">
                        {fileManagerConfig.history.length - index}
                      </div>
                      <span class="truncate text-slate-700 dark:text-slate-300 font-medium">{path}</span>
                    </div>
                    <button
                      on:click={() => {
                        fileManagerConfig.history = fileManagerConfig.history.filter(p => p !== path);
                        handleSaveSettings();
                      }}
                      class="ml-2 p-1 rounded-md text-slate-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/30 transition-all opacity-0 group-hover:opacity-100"
                      title="删除记录"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>

        <!-- Footer -->
        <div class="p-4 bg-gradient-to-r from-slate-50 via-white to-slate-50 dark:from-slate-900 dark:via-slate-900/80 dark:to-slate-950 border-t border-slate-200/70 dark:border-slate-800/80 flex justify-between items-center">
          <button
            on:click={() => {
              fileManagerConfig = {
                directoryTracking: false,
                historyEnabled: true,
                historyLimit: 5,
                history: [],
              };
              handleSaveSettings();
            }}
            class="px-3 py-2 text-xs font-medium text-slate-600 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 rounded-lg transition-all duration-200 flex items-center gap-1.5"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            恢复默认
          </button>
          <button
            on:click={() => showSettingsDialog = false}
            class="px-5 py-2 text-sm font-semibold text-white bg-gradient-to-r from-purple-600 to-indigo-600 hover:from-purple-700 hover:to-indigo-700 rounded-xl transition-all duration-200 shadow-lg shadow-purple-500/25 hover:shadow-purple-500/40"
          >
            完成
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>
