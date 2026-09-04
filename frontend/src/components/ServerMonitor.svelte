<script>
  import { onDestroy } from 'svelte';
  import { activeSessionIdStore, connectionsStore } from '../stores.js';
  import {
    isSessionLiveEnabled,
    setSessionLiveEnabled,
    shouldPollMonitor
  } from '../lib/monitorLiveControl.js';

  /** 性能面板当前是否可见；隐藏时即使开启实时也不轮询。 */
  export let panelVisible = true;

  const defaultStats = {
    cpu: 0,
    memory: 0,
    disk: 0,
    network: { in: 0, out: 0 }
  };

  const defaultSystemInfo = {
    os: 'Unknown',
    kernel: 'Unknown',
    uptime: 'Unknown',
    processes: 0
  };

  const defaultDiskInfo = {
    used: '0 GB',
    total: '0 GB'
  };

  const defaultLoadInfo = {
    '1min': 0,
    '5min': 0,
    '15min': 0
  };

  const defaultMemoryDetail = {
    used: '0 GB',
    total: '0 GB'
  };

  let cpuData = [];
  let memoryData = [];
  let cpuPerCore = [];
  let currentStats = { ...defaultStats };
  let systemInfo = { ...defaultSystemInfo };
  let diskInfo = { ...defaultDiskInfo };
  let loadInfo = { ...defaultLoadInfo };
  let memoryDetail = { ...defaultMemoryDetail };
  let isRefreshing = false;

  let sessionCpuData = new Map();
  let sessionMemoryData = new Map();
  let sessionCpuPerCore = new Map();
  let sessionStats = new Map();
  let sessionSystemInfo = new Map();
  let sessionDiskInfo = new Map();
  let sessionLoadInfo = new Map();
  let sessionMemoryDetail = new Map();
  let sessionLiveEnabled = new Map();

  let dataInterval = null;
  let previousSessionId = null;
  let lastPollKey = '';

  $: currentSession = $activeSessionIdStore ? $connectionsStore.get($activeSessionIdStore) : null;
  $: isSessionConnected = currentSession?.connected || false;
  $: isLocalSession = currentSession?.type === 'local';
  $: canUseMonitor = isSessionConnected && !isLocalSession;
  $: liveEnabled = isSessionLiveEnabled(sessionLiveEnabled, $activeSessionIdStore);
  $: pollingActive = shouldPollMonitor({
    liveEnabled,
    canUseMonitor,
    panelVisible
  });

  function applySessionSnapshot(sessionId) {
    if (!sessionId) {
      cpuData = [];
      memoryData = [];
      cpuPerCore = [];
      currentStats = { ...defaultStats };
      systemInfo = { ...defaultSystemInfo };
      diskInfo = { ...defaultDiskInfo };
      loadInfo = { ...defaultLoadInfo };
      memoryDetail = { ...defaultMemoryDetail };
      return;
    }

    cpuData = sessionCpuData.get(sessionId) || [];
    memoryData = sessionMemoryData.get(sessionId) || [];
    cpuPerCore = sessionCpuPerCore.get(sessionId) || [];
    currentStats = sessionStats.get(sessionId) || { ...defaultStats };
    systemInfo = sessionSystemInfo.get(sessionId) || { ...defaultSystemInfo };
    diskInfo = sessionDiskInfo.get(sessionId) || { ...defaultDiskInfo };
    loadInfo = sessionLoadInfo.get(sessionId) || { ...defaultLoadInfo };
    memoryDetail = sessionMemoryDetail.get(sessionId) || { ...defaultMemoryDetail };
  }

  function persistSessionSnapshot(sessionId) {
    if (!sessionId) return;
    sessionCpuData = new Map(sessionCpuData).set(sessionId, cpuData);
    sessionMemoryData = new Map(sessionMemoryData).set(sessionId, memoryData);
    sessionCpuPerCore = new Map(sessionCpuPerCore).set(sessionId, cpuPerCore);
    sessionStats = new Map(sessionStats).set(sessionId, currentStats);
    sessionSystemInfo = new Map(sessionSystemInfo).set(sessionId, systemInfo);
    sessionDiskInfo = new Map(sessionDiskInfo).set(sessionId, diskInfo);
    sessionLoadInfo = new Map(sessionLoadInfo).set(sessionId, loadInfo);
    sessionMemoryDetail = new Map(sessionMemoryDetail).set(sessionId, memoryDetail);
  }

  $: if (previousSessionId !== $activeSessionIdStore) {
    if (previousSessionId) {
      persistSessionSnapshot(previousSessionId);
    }
    applySessionSnapshot($activeSessionIdStore);
  }
  $: previousSessionId = $activeSessionIdStore;

  function setLiveEnabled(enabled) {
    if (!$activeSessionIdStore || !canUseMonitor) return;
    sessionLiveEnabled = setSessionLiveEnabled(sessionLiveEnabled, $activeSessionIdStore, enabled);
  }

  function getStatusColor(value) {
    if (value < 50) return '#10b981';
    if (value < 80) return '#f59e0b';
    return '#ef4444';
  }

  function formatBytesPerSecond(bytesPerSecond) {
    if (bytesPerSecond === 0) return '0 KB/s';

    const units = ['KB/s', 'MB/s', 'GB/s'];
    let unitIndex = -1;
    let value = bytesPerSecond;

    while (value >= 1024 && unitIndex < units.length - 1) {
      value /= 1024;
      unitIndex++;
    }

    return `${value.toFixed(1)} ${units[unitIndex]}`;
  }

  function formatBytes(bytes) {
    if (bytes === 0) return '0 B';

    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let unitIndex = 0;
    let value = bytes;

    while (value >= 1024 && unitIndex < units.length - 1) {
      value /= 1024;
      unitIndex++;
    }

    return `${value.toFixed(1)} ${units[unitIndex]}`;
  }

  function normalizeMonitoringData(data) {
    if (!data) return null;

    const nextCpuPerCore = data.cpu?.per_core && Array.isArray(data.cpu.per_core)
      ? data.cpu.per_core
      : null;

    const stats = {
      cpu: typeof data.cpu === 'number' ? data.cpu : (data.cpu?.overall || 0),
      memory: typeof data.memory === 'number' ? data.memory : (data.memory?.used_percent || 0),
      disk: typeof data.disk === 'number' ? data.disk : (data.disk?.partitions?.[0]?.used_percent || 0),
      network: {
        in: typeof data.network?.in === 'number' ? data.network.in * (1024 * 1024) :
              (typeof data.network?.rx_rate === 'number' ? data.network.rx_rate : 0),
        out: typeof data.network?.out === 'number' ? data.network.out * (1024 * 1024) :
               (typeof data.network?.tx_rate === 'number' ? data.network.tx_rate : 0)
      }
    };

    const nextSystemInfo = data.system ? {
      os: data.system.os || defaultSystemInfo.os,
      kernel: data.system.kernel || defaultSystemInfo.kernel,
      uptime: data.system.uptime || defaultSystemInfo.uptime,
      processes: data.system.processes || defaultSystemInfo.processes
    } : null;

    let nextDiskInfo = null;
    if (data.disk?.partitions?.[0]) {
      const partition = data.disk.partitions[0];
      const usedGB = (partition.used / (1024 * 1024 * 1024)).toFixed(1);
      const totalGB = (partition.total / (1024 * 1024 * 1024)).toFixed(1);
      nextDiskInfo = {
        used: `${usedGB} GB`,
        total: `${totalGB} GB`
      };
    }

    let nextLoadInfo = null;
    if (data.cpu?.load_average && Array.isArray(data.cpu.load_average) && data.cpu.load_average.length >= 3) {
      nextLoadInfo = {
        '1min': data.cpu.load_average[0] || 0,
        '5min': data.cpu.load_average[1] || 0,
        '15min': data.cpu.load_average[2] || 0
      };
    }

    let nextMemoryDetail = null;
    if (data.memory?.used !== undefined && data.memory?.total !== undefined) {
      const usedBytes = data.memory.used;
      const totalBytes = data.memory.total;
      nextMemoryDetail = {
        used: formatBytes(usedBytes),
        total: formatBytes(totalBytes)
      };
    }

    return {
      stats,
      cpuPerCore: nextCpuPerCore,
      systemInfo: nextSystemInfo,
      diskInfo: nextDiskInfo,
      loadInfo: nextLoadInfo,
      memoryDetail: nextMemoryDetail
    };
  }

  async function fetchMonitoringData() {
    const sessionId = $activeSessionIdStore;
    if (!sessionId || !isSessionConnected || isLocalSession) return;

    const { GetMonitoringData } = window.wailsBindings || {};
    if (typeof GetMonitoringData !== 'function') return;

    try {
      const data = await GetMonitoringData(sessionId);
      const normalizedData = normalizeMonitoringData(data);
      if (normalizedData) {
        const nextCpuData = [...(sessionCpuData.get(sessionId) || []), normalizedData.stats.cpu].slice(-60);
        const nextMemoryData = [...(sessionMemoryData.get(sessionId) || []), normalizedData.stats.memory].slice(-60);
        const nextCpuPerCore = normalizedData.cpuPerCore || sessionCpuPerCore.get(sessionId) || [];
        const nextSystemInfo = normalizedData.systemInfo || sessionSystemInfo.get(sessionId) || { ...defaultSystemInfo };
        const nextDiskInfo = normalizedData.diskInfo || sessionDiskInfo.get(sessionId) || { ...defaultDiskInfo };
        const nextLoadInfo = normalizedData.loadInfo || sessionLoadInfo.get(sessionId) || { ...defaultLoadInfo };
        const nextMemoryDetail = normalizedData.memoryDetail || sessionMemoryDetail.get(sessionId) || { ...defaultMemoryDetail };
        const nextStats = normalizedData.stats;

        sessionCpuData = new Map(sessionCpuData).set(sessionId, nextCpuData);
        sessionMemoryData = new Map(sessionMemoryData).set(sessionId, nextMemoryData);
        sessionCpuPerCore = new Map(sessionCpuPerCore).set(sessionId, nextCpuPerCore);
        sessionStats = new Map(sessionStats).set(sessionId, nextStats);
        sessionSystemInfo = new Map(sessionSystemInfo).set(sessionId, nextSystemInfo);
        sessionDiskInfo = new Map(sessionDiskInfo).set(sessionId, nextDiskInfo);
        sessionLoadInfo = new Map(sessionLoadInfo).set(sessionId, nextLoadInfo);
        sessionMemoryDetail = new Map(sessionMemoryDetail).set(sessionId, nextMemoryDetail);

        if (sessionId === $activeSessionIdStore) {
          currentStats = nextStats;
          cpuData = nextCpuData;
          memoryData = nextMemoryData;
          cpuPerCore = nextCpuPerCore;
          systemInfo = nextSystemInfo;
          diskInfo = nextDiskInfo;
          loadInfo = nextLoadInfo;
          memoryDetail = nextMemoryDetail;
        }
      }
    } catch (error) {
      console.error('Failed to fetch monitoring data:', error);
    }
  }

  async function refreshOnce() {
    if (!canUseMonitor || isRefreshing) return;
    isRefreshing = true;
    try {
      await fetchMonitoringData();
    } finally {
      isRefreshing = false;
    }
  }

  function stopPolling() {
    if (dataInterval) {
      clearInterval(dataInterval);
      dataInterval = null;
    }
  }

  function startPolling() {
    stopPolling();
    fetchMonitoringData();
    dataInterval = setInterval(fetchMonitoringData, 2000);
  }

  $: if (!$activeSessionIdStore || !canUseMonitor) {
    cpuData = [];
    memoryData = [];
    cpuPerCore = [];
    currentStats = { ...defaultStats };
    systemInfo = { ...defaultSystemInfo };
    diskInfo = { ...defaultDiskInfo };
    loadInfo = { ...defaultLoadInfo };
    memoryDetail = { ...defaultMemoryDetail };
  }

  $: pollKey = `${$activeSessionIdStore || ''}:${pollingActive ? '1' : '0'}`;
  $: if (pollKey !== lastPollKey) {
    lastPollKey = pollKey;
    if (pollingActive) {
      startPolling();
    } else {
      stopPolling();
    }
  }

  onDestroy(() => {
    stopPolling();
  });
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800 overflow-y-auto scrollbar-thin">
  <div class="p-3 border-b border-gray-200 dark:border-gray-700">
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <h3 class="text-sm font-semibold text-gray-900 dark:text-white">服务器监控</h3>
        <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
          {#if isLocalSession}
            本地终端不支持服务器监控
          {:else if !isSessionConnected}
            请先连接到服务器
          {:else if liveEnabled}
            实时查询中（约每 2 秒）
          {:else}
            实时已关闭，可手动刷新，减轻弱机压力
          {/if}
        </div>
      </div>
      {#if canUseMonitor}
        <div class="flex items-center gap-2 flex-shrink-0">
          <button
            type="button"
            class="monitor-refresh"
            title="手动刷新一次"
            disabled={isRefreshing || liveEnabled}
            on:click={refreshOnce}
          >
            {isRefreshing ? '刷新中' : '刷新'}
          </button>
          <label class="monitor-live-toggle" title="开启后持续远程采集，弱性能服务器建议关闭">
            <span>实时</span>
            <input
              type="checkbox"
              role="switch"
              aria-label="实时性能查询"
              checked={liveEnabled}
              on:change={(event) => setLiveEnabled(event.currentTarget.checked)}
            />
          </label>
        </div>
      {/if}
    </div>
  </div>

  {#if isLocalSession}
    <div class="p-3 space-y-3">
      <div class="flex flex-col items-center justify-center h-40 text-gray-500 dark:text-gray-400 gap-2">
        <svg class="w-8 h-8 opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2 2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
        </svg>
        <span class="text-center px-4">本地终端不支持服务器监控</span>
        <div class="text-xs text-gray-400 dark:text-gray-500 mt-2">
           服务器监控功能仅适用于 SSH 远程连接
        </div>
       </div>
    </div>
  {/if}

  <div class="p-3 space-y-3">
     <div class="bg-gradient-to-br from-purple-50 to-blue-50 dark:from-purple-900/20 dark:to-blue-900/20 rounded-xl p-3 shadow-sm border border-purple-100 dark:border-purple-800">
       <div class="flex items-center justify-between mb-2">
         <div class="flex items-center gap-2">
           <div class="p-1.5 bg-purple-100 dark:bg-purple-900 rounded-lg">
             <svg class="w-3.5 h-3.5 text-purple-600 dark:text-purple-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
               <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 002-2h2a2 2 0 002-2z" />
             </svg>
           </div>
           <span class="text-xs font-semibold text-gray-900 dark:text-white">CPU</span>
         </div>
         <span class="text-xs font-bold px-2 py-1 rounded-lg bg-white dark:bg-gray-800" style="color: {getStatusColor(currentStats.cpu)}">
           {currentStats.cpu.toFixed(1)}%
         </span>
       </div>
       {#if loadInfo['1min'] > 0}
         <div class="flex items-center gap-2 mb-2">
           <span class="text-[9px] text-gray-600 dark:text-gray-400">负载:</span>
           <div class="flex gap-2 text-[9px]">
             <span class="px-1.5 py-0.5 bg-purple-50 dark:bg-purple-900/30 rounded font-mono" style="color: {getStatusColor(loadInfo['1min'] * 100)}">
               {loadInfo['1min'].toFixed(2)}
             </span>
             <span class="px-1.5 py-0.5 bg-purple-50 dark:bg-purple-900/30 rounded font-mono" style="color: {getStatusColor(loadInfo['5min'] * 100)}">
               {loadInfo['5min'].toFixed(2)}
             </span>
             <span class="px-1.5 py-0.5 bg-purple-50 dark:bg-purple-900/30 rounded font-mono" style="color: {getStatusColor(loadInfo['15min'] * 100)}">
               {loadInfo['15min'].toFixed(2)}
             </span>
           </div>
         </div>
       {/if}
      <div class="h-[80px] bg-white dark:bg-gray-800 rounded-lg overflow-hidden relative">
        {#if cpuData.length > 1}
          <svg class="w-full h-full" preserveAspectRatio="none">
            <defs>
              <linearGradient id="cpuGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                <stop offset="0%" style="stop-color:rgb(139, 92, 246);stop-opacity:0.3" />
                <stop offset="100%" style="stop-color:rgb(139, 92, 246);stop-opacity:0.05" />
              </linearGradient>
            </defs>
            <line x1="0" y1="25%" x2="100%" y2="25%" stroke="currentColor" stroke-width="0.5" class="text-gray-200 dark:text-gray-700" />
            <line x1="0" y1="50%" x2="100%" y2="50%" stroke="currentColor" stroke-width="0.5" class="text-gray-200 dark:text-gray-700" />
            <line x1="0" y1="75%" x2="100%" y2="75%" stroke="currentColor" stroke-width="0.5" class="text-gray-200 dark:text-gray-700" />
            <path
              d="{cpuData.map((v, i) => {
                const x = (i / (cpuData.length - 1)) * 100;
                const y = 100 - v;
                return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
              }).join(' ')} L 100 100 L 0 100 Z"
              fill="url(#cpuGradient)"
              transform="scale(1, 0.8) translate(0, 10)"
            />
            <path
              d="{cpuData.map((v, i) => {
                const x = (i / (cpuData.length - 1)) * 100;
                const y = 100 - v;
                return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
              }).join(' ')}"
              fill="none"
              stroke="rgb(139, 92, 246)"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              transform="scale(1, 0.8) translate(0, 10)"
            />
          </svg>
          <div class="absolute left-1 top-1 text-[8px] text-gray-400 dark:text-gray-500">100%</div>
          <div class="absolute left-1 bottom-1 text-[8px] text-gray-400 dark:text-gray-500">0%</div>
        {:else}
          <div class="w-full h-full flex items-center justify-center text-xs text-gray-400">
            {liveEnabled ? '等待数据...' : '开启实时或点刷新'}
          </div>
        {/if}
      </div>
      {#if cpuPerCore.length > 0}
        <div class="mt-2 flex flex-wrap gap-1">
          {#each cpuPerCore as core, i}
            <div class="flex flex-col items-center">
              <div class="w-6 h-1 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-300"
                  style="width: {core}%; background-color: {getStatusColor(core)};"
                />
              </div>
              <span class="text-[8px] text-gray-500 dark:text-gray-400 mt-0.5">C{i + 1}</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>

     <div class="rounded-xl p-3 shadow-sm border border-emerald-100 dark:border-emerald-800" style="background: linear-gradient(135deg, var(--accent-soft), transparent);">
       <div class="flex items-center justify-between mb-2">
         <div class="flex items-center gap-2">
           <div class="p-1.5 bg-emerald-100 dark:bg-emerald-900 rounded-lg">
             <svg class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
               <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 002-2h2a2 2 0 002-2z" />
             </svg>
           </div>
           <span class="text-xs font-semibold text-gray-900 dark:text-white">内存</span>
         </div>
         <div class="flex items-center gap-2">
           <span class="text-[10px] text-gray-600 dark:text-gray-400">
             {memoryDetail.used} / {memoryDetail.total}
           </span>
           <span class="text-xs font-bold px-2 py-1 rounded-lg bg-white dark:bg-gray-800" style="color: {getStatusColor(currentStats.memory)}">
             {currentStats.memory.toFixed(2)}%
           </span>
         </div>
       </div>
      <div class="h-[80px] bg-white dark:bg-gray-800 rounded-lg overflow-hidden relative">
        {#if memoryData.length > 1}
          <svg class="w-full h-full" preserveAspectRatio="none">
            <defs>
              <linearGradient id="memoryGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                <stop offset="0%" style="stop-color:rgb(16, 185, 129);stop-opacity:0.3" />
                <stop offset="100%" style="stop-color:rgb(16, 185, 129);stop-opacity:0.05" />
              </linearGradient>
            </defs>
            <line x1="0" y1="25%" x2="100%" y2="25%" stroke="currentColor" stroke-width="0.5" class="text-gray-200 dark:text-gray-700" />
            <line x1="0" y1="50%" x2="100%" y2="50%" stroke="currentColor" stroke-width="0.5" class="text-gray-200 dark:text-gray-700" />
            <line x1="0" y1="75%" x2="100%" y2="75%" stroke="currentColor" stroke-width="0.5" class="text-gray-200 dark:text-gray-700" />
            <path
              d="{memoryData.map((v, i) => {
                const x = (i / (memoryData.length - 1)) * 100;
                const y = 100 - v;
                return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
              }).join(' ')} L 100 100 L 0 100 Z"
              fill="url(#memoryGradient)"
              transform="scale(1, 0.8) translate(0, 10)"
            />
            <path
              d="{memoryData.map((v, i) => {
                const x = (i / (memoryData.length - 1)) * 100;
                const y = 100 - v;
                return `${i === 0 ? 'M' : 'L'} ${x} ${y}`;
              }).join(' ')}"
              fill="none"
              stroke="rgb(16, 185, 129)"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              transform="scale(1, 0.8) translate(0, 10)"
            />
          </svg>
          <div class="absolute left-1 top-1 text-[8px] text-gray-400 dark:text-gray-500">100%</div>
          <div class="absolute left-1 bottom-1 text-[8px] text-gray-400 dark:text-gray-500">0%</div>
        {:else}
          <div class="w-full h-full flex items-center justify-center text-xs text-gray-400">
            {liveEnabled ? '等待数据...' : '开启实时或点刷新'}
          </div>
        {/if}
      </div>
    </div>

    <div class="bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-900/20 dark:to-orange-900/20 rounded-xl p-3 shadow-sm border border-amber-100 dark:border-amber-800">
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <div class="p-1.5 bg-amber-100 dark:bg-amber-900 rounded-lg">
            <svg class="w-3.5 h-3.5 text-amber-600 dark:text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4" />
            </svg>
          </div>
          <span class="text-xs font-semibold text-gray-900 dark:text-white">磁盘</span>
        </div>
        <span class="text-xs font-bold px-2 py-1 rounded-lg bg-white dark:bg-gray-800" style="color: {getStatusColor(currentStats.disk)}">
          {currentStats.disk}%
        </span>
      </div>
      <div class="w-full bg-gray-100 dark:bg-gray-700 rounded-full h-2 overflow-hidden">
        <div
          class="h-2 rounded-full transition-all duration-300"
          style="width: {currentStats.disk}%; background-color: {getStatusColor(currentStats.disk)};"
        />
      </div>
      <div class="flex justify-between mt-2 text-[10px] text-gray-600 dark:text-gray-400">
        <span>已用 {diskInfo.used}</span>
        <span>总计 {diskInfo.total}</span>
      </div>
    </div>

    <div class="bg-gradient-to-br from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 rounded-xl p-3 shadow-sm border border-blue-100 dark:border-blue-800">
      <div class="flex items-center gap-2 mb-3">
        <div class="p-1.5 bg-blue-100 dark:bg-blue-900 rounded-lg">
          <svg class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.141 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0" />
          </svg>
        </div>
        <span class="text-xs font-semibold text-gray-900 dark:text-white">网络流量</span>
      </div>
      <div class="space-y-2">
        <div class="flex items-center justify-between bg-white dark:bg-gray-800 rounded-lg p-2">
          <div class="flex items-center gap-2">
            <div class="w-2 h-2 rounded-full bg-emerald-500"></div>
            <span class="text-xs text-gray-700 dark:text-gray-300 font-medium">入站</span>
          </div>
          <span class="text-xs font-mono font-bold text-gray-900 dark:text-white">
            {formatBytesPerSecond(currentStats.network?.in || 0)}
          </span>
        </div>
        <div class="flex items-center justify-between bg-white dark:bg-gray-800 rounded-lg p-2">
          <div class="flex items-center gap-2">
            <div class="w-2 h-2 rounded-full bg-rose-500"></div>
            <span class="text-xs text-gray-700 dark:text-gray-300 font-medium">出站</span>
          </div>
          <span class="text-xs font-mono font-bold text-gray-900 dark:text-white">
            {formatBytesPerSecond(currentStats.network?.out || 0)}
          </span>
        </div>
      </div>
    </div>

    <div class="bg-gray-50 dark:bg-gray-700 rounded-xl p-3 shadow-sm border border-gray-200 dark:border-gray-600">
      <div class="text-xs font-semibold text-gray-900 dark:text-white mb-2">系统信息</div>
      <div class="space-y-1.5 text-[10px]">
        <div class="flex justify-between py-1">
          <span class="text-gray-600 dark:text-gray-400">操作系统</span>
          <span class="text-gray-900 dark:text-white font-medium">{systemInfo.os || 'Unknown'}</span>
        </div>
        <div class="flex justify-between py-1">
          <span class="text-gray-600 dark:text-gray-400">内核版本</span>
          <span class="text-gray-900 dark:text-white font-medium">{systemInfo.kernel || 'Unknown'}</span>
        </div>
        <div class="flex justify-between py-1">
          <span class="text-gray-600 dark:text-gray-400">运行时间</span>
          <span class="text-gray-900 dark:text-white font-medium">{systemInfo.uptime || 'Unknown'}</span>
        </div>
        <div class="flex justify-between py-1">
          <span class="text-gray-600 dark:text-gray-400">进程数</span>
          <span class="text-gray-900 dark:text-white font-medium">{systemInfo.processes || 'N/A'}</span>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .monitor-refresh {
    appearance: none;
    border: 1px solid var(--border-primary, #d9e0e4);
    border-radius: 999px;
    background: var(--bg-primary, #fff);
    color: var(--text-secondary, #52606d);
    font-size: 11px;
    font-weight: 600;
    min-height: 26px;
    padding: 0 10px;
    cursor: pointer;
  }

  .monitor-refresh:hover:not(:disabled) {
    color: #0e6674;
    border-color: #9fc5c8;
  }

  .monitor-refresh:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .monitor-live-toggle {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    font-weight: 600;
    color: var(--text-secondary, #52606d);
    cursor: pointer;
    user-select: none;
  }

  .monitor-live-toggle input {
    width: 34px;
    height: 18px;
    margin: 0;
    appearance: none;
    border-radius: 999px;
    background: #cbd5e1;
    position: relative;
    cursor: pointer;
    transition: background-color 0.15s ease;
  }

  .monitor-live-toggle input::after {
    content: '';
    position: absolute;
    top: 2px;
    left: 2px;
    width: 14px;
    height: 14px;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 2px rgba(15, 23, 42, 0.2);
    transition: transform 0.15s ease;
  }

  .monitor-live-toggle input:checked {
    background: #0e6674;
  }

  .monitor-live-toggle input:checked::after {
    transform: translateX(16px);
  }

  .monitor-live-toggle input:focus-visible {
    outline: 2px solid #0e6674;
    outline-offset: 2px;
  }
</style>
