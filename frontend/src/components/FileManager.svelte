<script>
  import { onMount, onDestroy } from 'svelte';
  import { fileManagerStore } from '../stores/fileManager.js';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import {
    ListFiles,
    UploadFiles,
    DownloadFiles,
    DeleteFiles,
    RenameFile,
    CreateDirectory,
    SelectUploadFiles,
    SelectDownloadDirectory,
    CancelTransfer,
    ShowQuestionDialog,
    ShowErrorDialog
  } from '../../wailsjs/go/main/App.js';
  import FileListItem from './FileListItem.svelte';
  import TransferProgressBar from './TransferProgressBar.svelte';
  import FileOperationModal from './FileOperationModal.svelte';

  export let activeSessionId = null;

  let collapsed = true;
  let width = 400;
  let currentPath = '/';
  let files = [];
  let selectedFiles = new Set();
  let isLoading = false;
  let error = null;
  let transfers = {};

  // Drag & drop state
  let isDraggingOver = false;

  // Modal states
  let showRenameModal = false;
  let showCreateDirModal = false;
  let renameTarget = null;
  let newName = '';

  // Resize state
  let isDragging = false;
  let startX = 0;
  let startWidth = 0;

  // Progress unsubscribers
  let progressUnsubscribers = [];

  // Subscribe to store
  const unsubscribe = fileManagerStore.subscribe(state => {
    collapsed = state.collapsed;
    width = state.width;
    transfers = state.transfers;

    if (activeSessionId && state.sessionStates[activeSessionId]) {
      const sessionState = state.sessionStates[activeSessionId];
      currentPath = sessionState.currentPath || '/';
      files = sessionState.files || [];
      selectedFiles = new Set(sessionState.selectedFiles || []);
    }
  });

  $: if (activeSessionId && !collapsed) {
    loadDirectory(currentPath);
  }

  $: sortedFiles = sortFiles(files);

  function sortFiles(fileList) {
    if (!fileList || fileList.length === 0) return [];

    return [...fileList].sort((a, b) => {
      // Directories first
      if (a.is_dir && !b.is_dir) return -1;
      if (!a.is_dir && b.is_dir) return 1;

      // Then sort by name
      return a.name.localeCompare(b.name);
    });
  }

  async function loadDirectory(path) {
    if (!activeSessionId) return;

    isLoading = true;
    error = null;

    try {
      const fileList = await ListFiles(activeSessionId, path);
      files = fileList || [];
      currentPath = path;

      fileManagerStore.setSessionState(activeSessionId, {
        currentPath: path,
        files: files,
        selectedFiles: []
      });
      selectedFiles.clear();
    } catch (err) {
      console.error('Failed to load directory:', err);
      error = err.message || '加载目录失败';
    } finally {
      isLoading = false;
    }
  }

  async function toggleCollapsed() {
    await fileManagerStore.setCollapsed(!collapsed);
  }

  // Drag resize
  function handleDragStart(event) {
    isDragging = true;
    startX = event.clientX;
    startWidth = width;

    document.addEventListener('mousemove', handleDragMove);
    document.addEventListener('mouseup', handleDragEnd);

    document.body.style.userSelect = 'none';
    document.body.style.cursor = 'col-resize';
  }

  function handleDragMove(event) {
    if (!isDragging) return;
    const delta = startX - event.clientX; // Reversed for right panel
    const newWidth = Math.max(350, Math.min(800, startWidth + delta));
    width = newWidth;
  }

  async function handleDragEnd() {
    isDragging = false;
    document.removeEventListener('mousemove', handleDragMove);
    document.removeEventListener('mouseup', handleDragEnd);
    document.body.style.userSelect = '';
    document.body.style.cursor = '';
    await fileManagerStore.setWidth(width);
  }

  // File operations
  function handleFileClick(event) {
    const { file, event: mouseEvent } = event.detail;

    if (mouseEvent.ctrlKey || mouseEvent.metaKey) {
      // Toggle selection
      if (selectedFiles.has(file.path)) {
        selectedFiles.delete(file.path);
      } else {
        selectedFiles.add(file.path);
      }
      selectedFiles = selectedFiles;
    } else {
      // Single selection
      selectedFiles.clear();
      selectedFiles.add(file.path);
      selectedFiles = selectedFiles;
    }

    fileManagerStore.setSessionState(activeSessionId, {
      selectedFiles: Array.from(selectedFiles)
    });
  }

  function handleFileDoubleClick(event) {
    const file = event.detail;
    if (file.is_dir) {
      loadDirectory(file.path);
    }
  }

  async function handleUpload() {
    try {
      const localPaths = await SelectUploadFiles();
      if (!localPaths || localPaths.length === 0) return;

      const transferIDs = await UploadFiles(activeSessionId, localPaths, currentPath);
      transferIDs.forEach(id => subscribeToTransfer(id));

      // Refresh directory after a short delay
      setTimeout(() => loadDirectory(currentPath), 2000);
    } catch (err) {
      console.error('Upload failed:', err);
      await ShowErrorDialog('上传失败', err.message || '上传文件时发生错误');
    }
  }

  async function handleDownload() {
    if (selectedFiles.size === 0) return;

    try {
      const localDir = await SelectDownloadDirectory();
      if (!localDir) return;

      const remotePaths = Array.from(selectedFiles);
      const transferIDs = await DownloadFiles(activeSessionId, remotePaths, localDir);
      transferIDs.forEach(id => subscribeToTransfer(id));
    } catch (err) {
      console.error('Download failed:', err);
      await ShowErrorDialog('下载失败', err.message || '下载文件时发生错误');
    }
  }

  async function handleDelete() {
    if (selectedFiles.size === 0) return;

    try {
      const confirmed = await ShowQuestionDialog(
        '确认删除',
        `确定要删除选中的 ${selectedFiles.size} 个项目吗？`
      );

      if (!confirmed) return;

      await DeleteFiles(activeSessionId, Array.from(selectedFiles));
      selectedFiles.clear();
      await loadDirectory(currentPath);
    } catch (err) {
      console.error('Delete failed:', err);
      await ShowErrorDialog('删除失败', err.message || '删除文件时发生错误');
    }
  }

  function handleRename() {
    if (selectedFiles.size !== 1) return;

    const filePath = Array.from(selectedFiles)[0];
    const file = files.find(f => f.path === filePath);
    if (!file) return;

    renameTarget = file;
    newName = file.name;
    showRenameModal = true;
  }

  async function handleRenameConfirm(event) {
    const name = event.detail;
    if (!renameTarget || !name) return;

    try {
      const dirPath = currentPath.endsWith('/') ? currentPath : currentPath + '/';
      const newPath = dirPath + name;

      await RenameFile(activeSessionId, renameTarget.path, newPath);
      await loadDirectory(currentPath);
    } catch (err) {
      console.error('Rename failed:', err);
      await ShowErrorDialog('重命名失败', err.message || '重命名时发生错误');
    }

    renameTarget = null;
    newName = '';
  }

  async function handleCreateDir(event) {
    const name = event.detail;
    if (!name) return;

    try {
      const dirPath = currentPath.endsWith('/') ? currentPath : currentPath + '/';
      const newPath = dirPath + name;

      await CreateDirectory(activeSessionId, newPath);
      await loadDirectory(currentPath);
    } catch (err) {
      console.error('Create directory failed:', err);
      await ShowErrorDialog('创建文件夹失败', err.message || '创建文件夹时发生错误');
    }
  }

  async function handleRefresh() {
    await loadDirectory(currentPath);
  }

  function navigateTo(path) {
    loadDirectory(path);
  }

  // Breadcrumb navigation
  $: pathParts = currentPath.split('/').filter(p => p);

  function handleBreadcrumbClick(index) {
    if (index === -1) {
      navigateTo('/');
    } else {
      const path = '/' + pathParts.slice(0, index + 1).join('/');
      navigateTo(path);
    }
  }

  // Drag & drop
  function handleDragOver(event) {
    event.preventDefault();
    isDraggingOver = true;
  }

  function handleDragLeave(event) {
    isDraggingOver = false;
  }

  async function handleDrop(event) {
    event.preventDefault();
    isDraggingOver = false;

    const items = event.dataTransfer.items;
    if (!items || items.length === 0) return;

    const filePaths = [];
    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      if (item.kind === 'file') {
        const file = item.getAsFile();
        if (file && file.path) {
          filePaths.push(file.path);
        }
      }
    }

    if (filePaths.length === 0) return;

    try {
      const transferIDs = await UploadFiles(activeSessionId, filePaths, currentPath);
      transferIDs.forEach(id => subscribeToTransfer(id));
      setTimeout(() => loadDirectory(currentPath), 2000);
    } catch (err) {
      console.error('Drop upload failed:', err);
      await ShowErrorDialog('上传失败', err.message || '上传文件时发生错误');
    }
  }

  // Transfer progress subscription
  function subscribeToTransfer(transferID) {
    const eventName = `sftp:progress:${transferID}`;
    const unsubscriber = EventsOn(eventName, (progress) => {
      fileManagerStore.updateTransfer(transferID, progress);

      // Auto-cleanup completed transfers after 3 seconds
      if (progress.status === 'completed' || progress.status === 'failed') {
        setTimeout(() => {
          fileManagerStore.removeTransfer(transferID);
        }, 3000);
      }
    });

    progressUnsubscribers.push(unsubscriber);
  }

  async function handleCancelTransfer(event) {
    const transferID = event.detail;
    try {
      await CancelTransfer(transferID);
    } catch (err) {
      console.error('Cancel transfer failed:', err);
    }
  }

  onMount(async () => {
    await fileManagerStore.init();
  });

  onDestroy(() => {
    unsubscribe();
    progressUnsubscribers.forEach(unsub => unsub());
  });
</script>

{#if collapsed}
  <!-- Collapsed State -->
  <div class="file-panel collapsed" style="width: 60px;">
    <div class="collapsed-content">
      <div class="icon">📁</div>
      <button class="expand-btn" on:click={toggleCollapsed} title="展开文件管理器">
        ☰
      </button>
    </div>
  </div>
{:else}
  <!-- Expanded State -->
  <div
    class="file-resizer"
    class:dragging={isDragging}
    on:mousedown={handleDragStart}
  ></div>

  <div class="file-panel expanded" style="width: {width}px;">
    <!-- Header -->
    <div class="header">
      <div class="breadcrumb">
        <button class="breadcrumb-item" on:click={() => handleBreadcrumbClick(-1)}>
          🏠
        </button>
        {#each pathParts as part, i}
          <span class="separator">/</span>
          <button class="breadcrumb-item" on:click={() => handleBreadcrumbClick(i)}>
            {part}
          </button>
        {/each}
      </div>
      <button class="collapse-btn" on:click={toggleCollapsed}>─</button>
    </div>

    <!-- Toolbar -->
    <div class="toolbar">
      <button on:click={handleUpload} title="上传文件" class="toolbar-btn">
        ↑
      </button>
      <button
        on:click={handleDownload}
        disabled={selectedFiles.size === 0}
        title="下载文件"
        class="toolbar-btn"
      >
        ↓
      </button>
      <button on:click={() => showCreateDirModal = true} title="新建文件夹" class="toolbar-btn">
        +📁
      </button>
      <button
        on:click={handleRename}
        disabled={selectedFiles.size !== 1}
        title="重命名"
        class="toolbar-btn"
      >
        ✎
      </button>
      <button on:click={handleRefresh} title="刷新" class="toolbar-btn">
        ⟳
      </button>
      <button
        on:click={handleDelete}
        disabled={selectedFiles.size === 0}
        title="删除"
        class="toolbar-btn toolbar-btn-danger"
      >
        🗑
      </button>
    </div>

    <!-- File list with drag-drop -->
    <div
      class="file-list-container"
      class:drag-over={isDraggingOver}
      on:drop={handleDrop}
      on:dragover={handleDragOver}
      on:dragleave={handleDragLeave}
    >
      {#if isLoading}
        <div class="loading">加载中...</div>
      {:else if error}
        <div class="error">{error}</div>
      {:else if files.length === 0}
        <div class="empty">此目录为空</div>
      {:else}
        <div class="file-list">
          {#each sortedFiles as file (file.path)}
            <FileListItem
              {file}
              selected={selectedFiles.has(file.path)}
              on:click={handleFileClick}
              on:dblclick={handleFileDoubleClick}
            />
          {/each}
        </div>
      {/if}

      {#if isDraggingOver}
        <div class="drop-overlay">
          <div class="drop-message">📁 拖放文件到此处上传</div>
        </div>
      {/if}
    </div>

    <!-- Transfer progress section -->
    {#if Object.keys(transfers).length > 0}
      <div class="transfers">
        <h4>传输进度</h4>
        <div class="transfer-list">
          {#each Object.values(transfers) as transfer (transfer.transfer_id)}
            <TransferProgressBar {transfer} on:cancel={handleCancelTransfer} />
          {/each}
        </div>
      </div>
    {/if}
  </div>
{/if}

<!-- Modals -->
<FileOperationModal
  bind:visible={showRenameModal}
  title="重命名"
  label="新名称："
  bind:value={newName}
  placeholder="输入新名称"
  on:confirm={handleRenameConfirm}
/>

<FileOperationModal
  bind:visible={showCreateDirModal}
  title="新建文件夹"
  label="文件夹名称："
  bind:value={newName}
  placeholder="输入文件夹名称"
  on:confirm={handleCreateDir}
/>

<style>
  /* Collapsed state */
  .file-panel.collapsed {
    width: 60px;
    background: var(--bg-secondary);
    border-left: 1px solid var(--border-primary);
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 16px 8px;
    flex-shrink: 0;
    -webkit-app-region: no-drag;
  }

  .collapsed-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
    width: 100%;
  }

  .icon {
    font-size: 24px;
  }

  .expand-btn {
    margin-top: auto;
    padding: 8px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 16px;
  }

  .expand-btn:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }

  /* Resizer */
  .file-resizer {
    width: 4px;
    background: transparent;
    cursor: col-resize;
    flex-shrink: 0;
    position: relative;
    -webkit-app-region: no-drag;
  }

  .file-resizer:hover,
  .file-resizer.dragging {
    background: var(--accent-primary);
  }

  /* Expanded state */
  .file-panel.expanded {
    background: var(--bg-secondary);
    border-left: 1px solid var(--border-primary);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    overflow: hidden;
    -webkit-app-region: no-drag;
  }

  /* Header */
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-primary);
    min-height: 48px;
  }

  .breadcrumb {
    display: flex;
    align-items: center;
    flex: 1;
    min-width: 0;
    overflow-x: auto;
    gap: 4px;
  }

  .breadcrumb::-webkit-scrollbar {
    height: 4px;
  }

  .breadcrumb-item {
    padding: 4px 8px;
    background: transparent;
    border: none;
    color: var(--text-primary);
    cursor: pointer;
    font-size: 12px;
    white-space: nowrap;
    border-radius: 4px;
  }

  .breadcrumb-item:hover {
    background: var(--bg-hover);
  }

  .separator {
    color: var(--text-secondary);
    font-size: 12px;
  }

  .collapse-btn {
    padding: 4px 8px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 16px;
    margin-left: 8px;
  }

  .collapse-btn:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }

  /* Toolbar */
  .toolbar {
    display: flex;
    gap: 4px;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-primary);
    background: var(--bg-tertiary);
  }

  .toolbar-btn {
    padding: 6px 10px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-primary);
    border-radius: 4px;
    color: var(--text-primary);
    cursor: pointer;
    font-size: 14px;
    transition: all 0.2s;
  }

  .toolbar-btn:hover:not(:disabled) {
    background: var(--bg-hover);
    border-color: var(--accent-primary);
  }

  .toolbar-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .toolbar-btn-danger:hover:not(:disabled) {
    border-color: var(--accent-error);
    color: var(--accent-error);
  }

  /* File list container */
  .file-list-container {
    flex: 1;
    overflow-y: auto;
    position: relative;
  }

  .file-list-container.drag-over {
    background: rgba(14, 99, 156, 0.1);
  }

  .file-list {
    padding: 8px;
  }

  .loading,
  .error,
  .empty {
    padding: 40px 20px;
    text-align: center;
    color: var(--text-secondary);
    font-size: 13px;
  }

  .error {
    color: var(--accent-error);
  }

  .drop-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(14, 99, 156, 0.2);
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none;
    border: 2px dashed var(--accent-primary);
    margin: 8px;
    border-radius: 8px;
  }

  .drop-message {
    font-size: 16px;
    color: var(--accent-primary);
    font-weight: 500;
  }

  /* Transfers */
  .transfers {
    border-top: 1px solid var(--border-primary);
    background: var(--bg-primary);
    max-height: 300px;
    overflow-y: auto;
  }

  .transfers h4 {
    margin: 0;
    padding: 12px 16px;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
    border-bottom: 1px solid var(--border-primary);
  }

  .transfer-list {
    padding: 12px;
  }
</style>
