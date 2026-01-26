<script>
  import { onMount } from 'svelte';

  let currentPath = '/home/user';
  let expandedDirs = new Set(['/home/user']);
  let activeSessionId = null;
  
  // 模拟文件数据（实际应从 API 获取）
  let files = [
    { name: 'Documents', type: 'directory', modified: '2024-01-20 10:30', permissions: 'drwxr-xr-x' },
    { name: 'Downloads', type: 'directory', modified: '2024-01-20 09:15', permissions: 'drwxr-xr-x' },
    { name: 'Projects', type: 'directory', modified: '2024-01-22 14:20', permissions: 'drwxr-xr-x' },
    { name: 'config.json', type: 'file', size: '2.4 KB', modified: '2024-01-22 11:45', permissions: '-rw-r--r--' },
    { name: 'deploy.sh', type: 'file', size: '1.8 KB', modified: '2024-01-21 16:30', permissions: '-rwxr-xr-x' },
    { name: 'README.md', type: 'file', size: '4.2 KB', modified: '2024-01-20 08:00', permissions: '-rw-r--r--' },
    { name: 'package.json', type: 'file', size: '1.2 KB', modified: '2024-01-19 15:20', permissions: '-rw-r--r--' },
  ];
  
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
  
  async function handleRefresh() {
    if (!activeSessionId) return;
    
    const { ListFiles } = window.wailsBindings || {};
    if (typeof ListFiles !== 'function') return;

    try {
      const fileList = await ListFiles(activeSessionId, currentPath);
      if (fileList) {
        files = fileList;
      }
    } catch (error) {
      console.error('Failed to refresh files:', error);
    }
  }
  
  async function handleUpload() {
    if (!activeSessionId) return;
    
    const { SelectUploadFiles, UploadFile } = window.wailsBindings || {};
    if (typeof SelectUploadFiles !== 'function') return;

    try {
      const selectedFiles = await SelectUploadFiles();
      if (selectedFiles && selectedFiles.length > 0) {
        // 显示待实现提示
        console.log('Selected files to upload:', selectedFiles);
      }
    } catch (error) {
      console.error('Failed to select files:', error);
    }
  }
  
  async function handleDownload(file) {
    if (!activeSessionId || file.type === 'directory') return;
    
    const { SelectDownloadDirectory, DownloadFile } = window.wailsBindings || {};
    if (typeof SelectDownloadDirectory !== 'function') return;

    try {
      const downloadPath = await SelectDownloadDirectory();
      if (downloadPath) {
        await DownloadFile(activeSessionId, currentPath, file.name, downloadPath);
      }
    } catch (error) {
      console.error('Failed to download file:', error);
    }
  }

  onMount(() => {
    // 监听活动会话变化
    const handleSessionChange = (event) => {
      activeSessionId = event.detail;
      if (activeSessionId) {
        handleRefresh();
      }
    };

    window.addEventListener('active-session-changed', handleSessionChange);

    return () => {
      window.removeEventListener('active-session-changed', handleSessionChange);
    };
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
      <button class="p-0.5 hover:bg-gray-200 dark:hover:bg-gray-600 rounded transition-colors">
        <svg class="w-3 h-3 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
        </svg>
      </button>
      <span class="text-gray-400 dark:text-gray-500">/</span>
      <span class="text-purple-600 dark:text-purple-400 font-medium">{currentPath.split('/').filter(Boolean).join(' / ')}</span>
    </div>
  </div>

  <!-- 文件列表 -->
  <div class="flex-1 overflow-y-auto scrollbar-thin text-xs">
    {#each files as file, index (index)}
      <div
        class="group flex items-center gap-2 px-3 py-2 hover:bg-purple-50 dark:hover:bg-purple-900/20 cursor-pointer transition-colors mx-2 my-0.5 rounded-lg"
        on:click={() => file.type === 'directory' && toggleDirectory(file.name)}
      >
        {#if file.type === 'directory'}
          {#if expandedDirs.has(`${currentPath}/${file.name}`)}
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
          {#if file.type === 'file'}
            <div class="text-gray-500 dark:text-gray-400 flex items-center gap-2">
              <span>{file.size}</span>
              <span>•</span>
              <span>{file.modified}</span>
            </div>
          {/if}
        </div>
        <span class="text-gray-400 dark:text-gray-500 font-mono text-[10px]">{file.permissions}</span>
      </div>
    {/each}
  </div>
</div>
