<script>
  import { onMount, onDestroy } from 'svelte';
  import { activeSessionIdStore, connectionsStore } from '../stores.js';

  let cpuData = [];
  let memoryData = [];
  let cpuPerCore = []; // Per-core CPU usage array
  let currentStats = {
    cpu: 45,
    memory: 62,
    disk: 78,
    network: { in: 125.5, out: 89.3 }
  };

  let systemInfo = {
    os: 'Ubuntu 22.04 LTS',
    kernel: '5.15.0-91-generic',
    uptime: '15天 7小时 32分',
    processes: 187
  };

  let diskInfo = {
    used: '156 GB',
    total: '200 GB'
  };

  let dataInterval = null;

  // Get current session object
  $: currentSession = $activeSessionIdStore ? $connectionsStore.get($activeSessionIdStore) : null;
  $: isSessionConnected = currentSession?.connected || false;
  $: isLocalSession = currentSession?.type === 'local';
  $: canUseMonitor = isSessionConnected && !isLocalSession;

  function getStatusColor(value) {
    if (value < 50) return '#10b981'; // green
    if (value < 80) return '#f59e0b'; // amber
    return '#ef4444'; // red
  }

  // Convert bytes to appropriate unit
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

  // Normalize backend monitoring data to component's expected format
  function normalizeMonitoringData(data) {
    if (!data) return null;

    // Handle per-core CPU data
    if (data.cpu?.per_core && Array.isArray(data.cpu.per_core)) {
      cpuPerCore = data.cpu.per_core;
    }

    // Backend returns detailed structure, normalize to simple format
    const normalized = {
      cpu: typeof data.cpu === 'number' ? data.cpu : (data.cpu?.overall || 0),
      memory: typeof data.memory === 'number' ? data.memory : (data.memory?.used_percent || 0),
      disk: typeof data.disk === 'number' ? data.disk : (data.disk?.partitions?.[0]?.used_percent || 0),
      network: {
        in: typeof data.network?.in === 'number' ? data.network.in * (1024 * 1024) : // Convert MB to bytes
              (typeof data.network?.rx_rate === 'number' ? data.network.rx_rate : 0), // Already in bytes/s
        out: typeof data.network?.out === 'number' ? data.network.out * (1024 * 1024) : // Convert MB to bytes
               (typeof data.network?.tx_rate === 'number' ? data.network.tx_rate : 0) // Already in bytes/s
      }
    };

    // Update system info if available
    if (data.system) {
      systemInfo = {
        os: data.system.os || systemInfo.os,
        kernel: data.system.kernel || systemInfo.kernel,
        uptime: data.system.uptime || systemInfo.uptime,
        processes: systemInfo.processes // Backend doesn't provide process count
      };
    }

    // Update disk info if available
    if (data.disk?.partitions?.[0]) {
      const partition = data.disk.partitions[0];
      const usedGB = (partition.used / (1024 * 1024 * 1024)).toFixed(1);
      const totalGB = (partition.total / (1024 * 1024 * 1024)).toFixed(1);
      diskInfo = {
        used: `${usedGB} GB`,
        total: `${totalGB} GB`
      };
    }

    return normalized;
  }

  // 获取监控数据
  async function fetchMonitoringData() {
    if (!$activeSessionIdStore || !isSessionConnected || isLocalSession) return;

    const { GetMonitoringData } = window.wailsBindings || {};
    if (typeof GetMonitoringData !== 'function') return;

    try {
      const data = await GetMonitoringData($activeSessionIdStore);
      const normalizedData = normalizeMonitoringData(data);
      console.log('data:', data);
      if (normalizedData) {
        currentStats = normalizedData;

        // Update CPU history array (keep last 60 data points)
        cpuData = [...cpuData, normalizedData.cpu].slice(-60);

        // Update memory history array (keep last 60 data points)
        memoryData = [...memoryData, normalizedData.memory].slice(-60);
      }
    } catch (error) {
      console.error('Failed to fetch monitoring data:', error);
    }
  }

  // React to active session changes - only fetch when session is connected and can use monitor
  $: if ($activeSessionIdStore && canUseMonitor) {
    fetchMonitoringData();
  }

  onMount(() => {
    // 开始定时更新数据
    dataInterval = setInterval(fetchMonitoringData, 2000);

    return () => {
      clearInterval(dataInterval);
    };
  });

  onDestroy(() => {
    if (dataInterval) {
      clearInterval(dataInterval);
    }
  });
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800 overflow-y-auto scrollbar-thin">
  <!-- 头部 -->
   <div class="p-3 border-b border-gray-200 dark:border-gray-700">
    <h3 class="text-sm font-semibold text-gray-900 dark:text-white">服务器监控</h3>
    <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
      {#if isLocalSession}
        本地终端不支持服务器监控
      {:else if isSessionConnected}
        实时数据获取中...
      {:else}
        请先连接到服务器
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
    <!-- CPU 使用率 -->
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
          {currentStats.cpu}%
        </span>
      </div>
      <div class="h-[80px] bg-white dark:bg-gray-800 rounded-lg flex items-center justify-center text-xs text-gray-400">
        CPU 图表待实现
      </div>
    </div>

    <!-- 内存使用率 -->
    <div class="bg-gradient-to-br from-emerald-50 to-teal-50 dark:from-emerald-900/20 dark:to-teal-900/20 rounded-xl p-3 shadow-sm border border-emerald-100 dark:border-emerald-800">
      <div class="flex items-center justify-between mb-2">
        <div class="flex items-center gap-2">
          <div class="p-1.5 bg-emerald-100 dark:bg-emerald-900 rounded-lg">
            <svg class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 002-2h2a2 2 0 002-2z" />
            </svg>
          </div>
          <span class="text-xs font-semibold text-gray-900 dark:text-white">内存</span>
        </div>
        <span class="text-xs font-bold px-2 py-1 rounded-lg bg-white dark:bg-gray-800" style="color: {getStatusColor(currentStats.memory)}">
          {currentStats.memory.toFixed(2)}%
        </span>
      </div>
      <div class="h-[80px] bg-white dark:bg-gray-800 rounded-lg flex items-center justify-center text-xs text-gray-400">
        内存图表待实现
      </div>
    </div>

    <!-- 磁盘使用 -->
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

    <!-- 网络流量 -->
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

    <!-- 系统信息 -->
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
