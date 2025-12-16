<script>
  import { onMount, onDestroy } from 'svelte';
  import { fade, fly } from 'svelte/transition';
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
  import Icon from './Icon.svelte';

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
  let showGoToPathModal = false;
  let renameTarget = null;
  let newName = '';
  let goToPath = '';

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

  async function handleGoToPath(event) {
    const path = event.detail;
    if (!path) return;

    try {
      // Normalize path - ensure it starts with /
      const normalizedPath = path.startsWith('/') ? path : '/' + path;

      // Try to load the directory
      await loadDirectory(normalizedPath);

      // Clear the input
      goToPath = '';
    } catch (err) {
      console.error('Go to path failed:', err);
      await ShowErrorDialog('跳转失败', err.message || '无法访问该路径');
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
    event.stopPropagation();
    isDraggingOver = true;
  }

  function handleDragLeave(event) {
    // 只在真正离开容器时触发
    if (event.target === event.currentTarget) {
      isDraggingOver = false;
    }
  }

  async function handleDrop(event) {
    event.preventDefault();
    event.stopPropagation();
    isDraggingOver = false;

    // 尝试使用拖放 API
    const files = event.dataTransfer.files;
    if (!files || files.length === 0) {
      console.warn('No files in drop event');
      await ShowErrorDialog('拖放失败', '请使用上传按钮选择文件。\n\n注意：由于 Wails 框架限制，拖放文件功能可能无法使用。');
      return;
    }

    // 检查是否可以获取文件路径（仅在某些环境下可用）
    const filePaths = [];
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      // 尝试获取路径（Electron 环境）
      if (file.path) {
        filePaths.push(file.path);
      }
    }

    // 如果无法获取路径，提示用户使用上传按钮
    if (filePaths.length === 0) {
      console.warn('Cannot access file paths from drop event in Wails');
      await ShowErrorDialog(
        '拖放功能不可用',
        '抱歉，Wails 应用暂不支持拖放上传。\n\n请点击工具栏的"上传"按钮选择文件。'
      );
      return;
    }

    // 如果成功获取路径，执行上传
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
  <div class="file-panel collapsed" transition:fly={{ x: -20, duration: 200 }}>
    <div class="collapsed-content">
      <div class="sidebar-icon" title="文件管理器">
        <Icon name="folder" size={24} color="var(--text-secondary)" />
      </div>
      <button class="expand-btn" on:click={toggleCollapsed} title="展开">
        <Icon name="menu" size={20} />
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

  <div class="file-panel expanded" style="width: {width}px;" transition:fly={{ x: -20, duration: 200 }}>
    <!-- Header -->
    <div class="header">
      <div class="breadcrumb">
        <button class="breadcrumb-item home" on:click={() => handleBreadcrumbClick(-1)} title="根目录">
          <Icon name="home" size={16} />
        </button>
        {#each pathParts as part, i}
          <Icon name="chevronRight" size={14} color="var(--text-secondary)" className="separator" />
          <button class="breadcrumb-item" on:click={() => handleBreadcrumbClick(i)}>
            {part}
          </button>
        {/each}
      </div>
      <div class="header-actions">
        <button class="action-btn" on:click={toggleCollapsed} title="折叠">
          <Icon name="chevronLeft" size={18} />
        </button>
      </div>
    </div>

    <!-- Toolbar -->
    <div class="toolbar">
      <div class="toolbar-group">
        <button on:click={handleUpload} title="上传文件" class="toolbar-btn">
          <Icon name="upload" size={16} />
          <span>上传</span>
        </button>
        <button
          on:click={handleDownload}
          disabled={selectedFiles.size === 0}
          title="下载选中文件"
          class="toolbar-btn"
        >
          <Icon name="download" size={16} />
          <span>下载</span>
        </button>
      </div>
      
      <div class="toolbar-divider"></div>
      
      <div class="toolbar-group">
        <button on:click={() => showCreateDirModal = true} title="新建文件夹" class="toolbar-btn icon-only">
          <Icon name="folder-plus" size={18} />
        </button>
        <button
          on:click={handleRename}
          disabled={selectedFiles.size !== 1}
          title="重命名"
          class="toolbar-btn icon-only"
        >
          <Icon name="edit" size={16} />
        </button>
        <button on:click={handleRefresh} title="刷新" class="toolbar-btn icon-only">
          <Icon name="refresh" size={16} />
        </button>
        <button on:click={() => showGoToPathModal = true} title="跳转到路径" class="toolbar-btn icon-only">
          <Icon name="search" size={16} />
        </button>
      </div>
      
      <div class="toolbar-spacer"></div>
      
      <div class="toolbar-group">
        <button
          on:click={handleDelete}
          disabled={selectedFiles.size === 0}
          title="删除"
          class="toolbar-btn toolbar-btn-danger icon-only"
        >
          <Icon name="trash" size={16} />
        </button>
      </div>
    </div>

    <!-- File List Header -->
    <div class="file-list-header">
      <div class="col-icon"></div>
      <div class="col-name">名称</div>
      <div class="col-size">大小</div>
      <div class="col-date">修改时间</div>
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
        <div class="state-container loading" in:fade>
          <div class="spinner"></div>
          <p>正在加载目录...</p>
        </div>
      {:else if error}
        <div class="state-container error" in:fade>
          <Icon name="close" size={32} color="var(--accent-error)" />
          <p>{error}</p>
          <button class="retry-btn" on:click={handleRefresh}>重试</button>
        </div>
      {:else if files.length === 0}
        <div class="state-container empty" in:fade>
          <Icon name="folder" size={48} color="var(--bg-input)" />
          <p>此目录为空</p>
        </div>
      {:else}
        <div class="file-list">
          {#each sortedFiles as file (file.path)}
            <FileListItem
              {file}
              selected={selectedFiles.has(file.path)}
              on:click={handleFileClick}
              on:dblclick={handleFileDoubleClick}
              on:contextmenu={(e) => {
                 // Context menu logic could go here
              }}
            />
          {/each}
        </div>
      {/if}

      {#if isDraggingOver}
        <div class="drop-overlay" transition:fade={{ duration: 150 }}>
          <div class="drop-content">
            <Icon name="upload" size={48} color="var(--accent-primary)" />
            <div class="drop-message">释放以上传文件</div>
          </div>
        </div>
      {/if}
    </div>

    <!-- Transfer progress section -->
    {#if Object.keys(transfers).length > 0}
      <div class="transfers" transition:fly={{ y: 20, duration: 200 }}>
        <div class="transfers-header">
          <h4>传输任务</h4>
          <span class="badge">{Object.keys(transfers).length}</span>
        </div>
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

<FileOperationModal
  bind:visible={showGoToPathModal}
  title="跳转到路径"
  label="目标路径："
  bind:value={goToPath}
  placeholder="例如：/home/user 或 /var/log"
  on:confirm={handleGoToPath}
/>

<style>
  /* Collapsed state */
  .file-panel.collapsed {
    width: 60px;
    height: 100%;
    background: var(--bg-secondary);
    border-left: 1px solid var(--border-primary);
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 16px 8px;
    flex-shrink: 0;
    -webkit-app-region: no-drag;
    box-sizing: border-box;
    z-index: 10;
  }

  .collapsed-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
    width: 100%;
    height: 100%;
  }

  .sidebar-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    border-radius: 8px;
    background: var(--bg-tertiary);
  }

  .expand-btn {
    margin-top: auto;
    padding: 10px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 6px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
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
    z-index: 20;
    -webkit-app-region: no-drag;
    margin-right: -2px; /* Overlap slightly */
  }

  .file-resizer:hover,
  .file-resizer.dragging {
    background: var(--accent-primary);
  }

  /* Expanded state */
  .file-panel.expanded {
    height: 100%;
    background: var(--bg-secondary);
    border-left: 1px solid var(--border-primary);
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    overflow: hidden;
    -webkit-app-region: no-drag;
    box-sizing: border-box;
    position: relative;
  }

  /* Header */
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 16px;
    border-bottom: 1px solid var(--border-primary);
    height: 48px;
    background: var(--bg-secondary);
  }

  .breadcrumb {
    display: flex;
    align-items: center;
    flex: 1;
    min-width: 0;
    overflow-x: auto;
    gap: 4px;
    padding-right: 8px;
    scrollbar-width: none; /* Firefox */
  }
  
  .breadcrumb::-webkit-scrollbar {
    display: none; /* Chrome/Safari */
  }

  .breadcrumb-item {
    padding: 4px 8px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 13px;
    white-space: nowrap;
    border-radius: 4px;
    transition: all 0.2s;
  }

  .breadcrumb-item:hover {
    background: var(--bg-hover);
    color: var(--text-primary);
  }
  
  .breadcrumb-item:last-child {
    color: var(--text-primary);
    font-weight: 500;
  }
  
  .breadcrumb-item.home {
    padding: 4px;
    display: flex;
    align-items: center;
  }

  /* Header Actions */
  .header-actions {
    display: flex;
    align-items: center;
  }

  .action-btn {
    padding: 6px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .action-btn:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }

  /* Toolbar */
  .toolbar {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-primary);
    background: var(--bg-tertiary);
    gap: 8px;
    flex-wrap: nowrap;
    overflow-x: auto;
  }
  
  .toolbar-group {
    display: flex;
    gap: 4px;
    align-items: center;
  }
  
  .toolbar-divider {
    width: 1px;
    height: 20px;
    background: var(--border-primary);
    margin: 0 4px;
  }
  
  .toolbar-spacer {
    flex: 1;
  }

  .toolbar-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 10px;
    background: transparent;
    border: 1px solid transparent;
    border-radius: 4px;
    color: var(--text-primary);
    cursor: pointer;
    font-size: 13px;
    transition: all 0.2s;
    white-space: nowrap;
  }
  
  .toolbar-btn.icon-only {
    padding: 6px;
  }

  .toolbar-btn:hover:not(:disabled) {
    background: var(--bg-hover);
    border-color: var(--border-primary);
  }
  
  .toolbar-btn:active:not(:disabled) {
    background: var(--bg-input);
  }

  .toolbar-btn:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .toolbar-btn-danger:hover:not(:disabled) {
    background: rgba(244, 67, 54, 0.1);
    color: var(--accent-error);
    border-color: rgba(244, 67, 54, 0.2);
  }
  
  /* File List Header */
  .file-list-header {
    display: grid;
    grid-template-columns: 32px 1fr 100px 140px;
    padding: 8px 12px;
    border-bottom: 1px solid var(--border-primary);
    background: var(--bg-secondary);
    font-size: 11px;
    font-weight: 600;
    color: var(--text-secondary);
    user-select: none;
  }
  
  .file-list-header .col-size,
  .file-list-header .col-date {
    text-align: right;
    padding-right: 16px;
  }
  
  .file-list-header .col-date {
    padding-right: 0;
  }

  /* File list container */
  .file-list-container {
    flex: 1;
    overflow-y: auto;
    position: relative;
    background: var(--bg-primary);
  }

  .file-list-container.drag-over {
    background: rgba(14, 99, 156, 0.05);
  }

  .file-list {
    padding-bottom: 20px;
  }

  /* States */
  .state-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    min-height: 200px;
    color: var(--text-secondary);
    gap: 16px;
  }
  
  .state-container p {
    margin: 0;
    font-size: 14px;
  }
  
  .empty {
    opacity: 0.7;
  }
  
  .error {
    color: var(--accent-error);
  }
  
  .retry-btn {
    padding: 6px 16px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-primary);
    border-radius: 4px;
    color: var(--text-primary);
    cursor: pointer;
    font-size: 13px;
    transition: all 0.2s;
  }
  
  .retry-btn:hover {
    background: var(--bg-hover);
    border-color: var(--accent-primary);
  }
  
  /* Spinner */
  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--bg-input);
    border-top-color: var(--accent-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  /* Drop Overlay */
  .drop-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(14, 99, 156, 0.15);
    backdrop-filter: blur(2px);
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none;
    z-index: 50;
    border: 2px dashed var(--accent-primary);
    margin: 10px;
    border-radius: 8px;
  }
  
  .drop-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    padding: 40px;
    background: var(--bg-secondary);
    border-radius: 12px;
    box-shadow: 0 4px 20px rgba(0,0,0,0.2);
  }

  .drop-message {
    font-size: 16px;
    color: var(--accent-primary);
    font-weight: 600;
  }

  /* Transfers */
  .transfers {
    border-top: 1px solid var(--border-primary);
    background: var(--bg-secondary);
    max-height: 250px;
    display: flex;
    flex-direction: column;
    box-shadow: 0 -4px 12px rgba(0,0,0,0.1);
  }
  
  .transfers-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 16px;
    border-bottom: 1px solid var(--border-primary);
    background: var(--bg-tertiary);
  }

  .transfers-header h4 {
    margin: 0;
    font-size: 12px;
    font-weight: 700;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  
  .badge {
    background: var(--accent-primary);
    color: white;
    font-size: 10px;
    padding: 2px 6px;
    border-radius: 10px;
    font-weight: 600;
  }

  .transfer-list {
    padding: 0;
    overflow-y: auto;
    max-height: 200px;
  }
</style>
