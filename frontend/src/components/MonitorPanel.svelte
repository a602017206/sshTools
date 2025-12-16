<script>
  import { onMount, onDestroy } from 'svelte';
  import { monitorStore } from '../stores/monitor.js';
  import { GetMonitoringData } from '../../wailsjs/go/main/App.js';

  export let activeSessionId = null;

  let collapsed = true;
  let width = 350;
  let refreshInterval = 2;
  let monitoringData = null;
  let isLoading = false;
  let error = null;
  let intervalId = null;
  let cpuHistory = [];

  // Subscribe to store
  monitorStore.subscribe(state => {
    collapsed = state.collapsed;
    width = state.width;
    refreshInterval = state.refreshInterval;
  });

  // Fetch monitoring data
  async function fetchMetrics() {
    if (!activeSessionId || collapsed) return;

    isLoading = true;
    error = null;

    try {
      const data = await GetMonitoringData(activeSessionId);
      monitoringData = data;

      // Update CPU history (keep last 60 samples)
      cpuHistory = [...cpuHistory, data.cpu.overall].slice(-60);
    } catch (err) {
      console.error('Failed to fetch monitoring data:', err);
      error = err.message;
    } finally {
      isLoading = false;
    }
  }

  // Toggle collapsed state
  async function toggleCollapsed() {
    await monitorStore.setCollapsed(!collapsed);
  }

  // Dragging logic
  let isDragging = false;
  let startX = 0;
  let startWidth = 0;

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
    const delta = startX - event.clientX; // Reversed
    const newWidth = Math.max(300, Math.min(600, startWidth + delta));
    width = newWidth;
  }

  async function handleDragEnd() {
    isDragging = false;
    document.removeEventListener('mousemove', handleDragMove);
    document.removeEventListener('mouseup', handleDragEnd);
    document.body.style.userSelect = '';
    document.body.style.cursor = '';
    await monitorStore.setWidth(width);
  }

  // Polling logic
  $: if (activeSessionId && !collapsed) {
    fetchMetrics();
    if (intervalId) clearInterval(intervalId);
    intervalId = setInterval(fetchMetrics, refreshInterval * 1000);
  } else {
    if (intervalId) {
      clearInterval(intervalId);
      intervalId = null;
    }
  }

  onMount(async () => {
    await monitorStore.init();
  });

  onDestroy(() => {
    if (intervalId) clearInterval(intervalId);
  });

  // Format bytes
  function formatBytes(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
  }

  // Format rate
  function formatRate(bytesPerSec) {
    return formatBytes(bytesPerSec) + '/s';
  }

  // Get color based on percentage
  function getColor(percent) {
    if (percent < 60) return 'var(--accent-success)';
    if (percent < 80) return '#f9a825';
    return 'var(--accent-error)';
  }
</script>

{#if collapsed}
  <!-- Collapsed State -->
  <div class="monitor-panel collapsed" style="width: 60px;">
    <div class="collapsed-content">
      <div class="metric">
        <div class="label">CPU</div>
        <div class="value">{monitoringData?.cpu?.overall.toFixed(0) || '--'}%</div>
      </div>

      <div class="metric">
        <div class="label">MEM</div>
        <div class="value">{monitoringData?.memory?.used_percent.toFixed(0) || '--'}%</div>
      </div>

      <div class="metric">
        <div class="label">↑</div>
        <div class="value small">{formatBytes(monitoringData?.network?.tx_rate || 0)}</div>
      </div>

      <div class="metric">
        <div class="label">↓</div>
        <div class="value small">{formatBytes(monitoringData?.network?.rx_rate || 0)}</div>
      </div>

      <button class="expand-btn" on:click={toggleCollapsed} title="展开监控面板">
        ☰
      </button>
    </div>
  </div>
{:else}
  <!-- Expanded State -->
  <div
    class="monitor-resizer"
    class:dragging={isDragging}
    on:mousedown={handleDragStart}
  ></div>

  <div class="monitor-panel expanded" style="width: {width}px;">
    <div class="header">
      <span class="title">
        {monitoringData?.system?.username || ''}@{monitoringData?.system?.hostname || '服务器监控'}
      </span>
      <button class="collapse-btn" on:click={toggleCollapsed}>─</button>
    </div>

    <div class="content">
      {#if error}
        <div class="error-message">{error}</div>
      {:else if !monitoringData && isLoading}
        <div class="loading">加载中...</div>
      {:else if monitoringData}
        <!-- System Info -->
        <section>
          <h3>系统信息</h3>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">运行时间:</span>
              <span class="info-value">{monitoringData.system.uptime}</span>
            </div>
            <div class="info-item">
              <span class="info-label">系统:</span>
              <span class="info-value">{monitoringData.system.os}</span>
            </div>
          </div>
        </section>

        <!-- CPU -->
        <section>
          <h3>CPU</h3>
          <div class="progress-item">
            <div class="progress-header">
              <span>总体使用率</span>
              <span>{monitoringData.cpu.overall.toFixed(1)}%</span>
            </div>
            <div class="progress-bar">
              <div
                class="progress-fill"
                style="width: {monitoringData.cpu.overall}%; background: {getColor(monitoringData.cpu.overall)};"
              ></div>
            </div>
          </div>
          <div class="cpu-details">
            <span>用户: {monitoringData.cpu.user.toFixed(1)}%</span>
            <span>系统: {monitoringData.cpu.system.toFixed(1)}%</span>
            <span>IO等待: {monitoringData.cpu.iowait.toFixed(1)}%</span>
          </div>
        </section>

        <!-- Memory -->
        <section>
          <h3>内存</h3>
          <div class="progress-item">
            <div class="progress-header">
              <span>物理内存</span>
              <span>{formatBytes(monitoringData.memory.used)} / {formatBytes(monitoringData.memory.total)}</span>
            </div>
            <div class="progress-bar">
              <div
                class="progress-fill"
                style="width: {monitoringData.memory.used_percent}%; background: {getColor(monitoringData.memory.used_percent)};"
              ></div>
            </div>
          </div>
        </section>

        <!-- Network -->
        <section>
          <h3>网络</h3>
          <div class="info-grid">
            <div class="info-item">
              <span class="info-label">总上行:</span>
              <span class="info-value">{formatBytes(monitoringData.network.total_tx_bytes)}</span>
            </div>
            <div class="info-item">
              <span class="info-label">总下行:</span>
              <span class="info-value">{formatBytes(monitoringData.network.total_rx_bytes)}</span>
            </div>
            <div class="info-item">
              <span class="info-label">实时上行:</span>
              <span class="info-value">{formatRate(monitoringData.network.tx_rate)}</span>
            </div>
            <div class="info-item">
              <span class="info-label">实时下行:</span>
              <span class="info-value">{formatRate(monitoringData.network.rx_rate)}</span>
            </div>
          </div>
        </section>

        <!-- Disk -->
        <section>
          <h3>磁盘</h3>
          {#each monitoringData.disk.partitions.slice(0, 5) as partition}
            <div class="progress-item">
              <div class="progress-header">
                <span>{partition.mount_point}</span>
                <span>{formatBytes(partition.used)} / {formatBytes(partition.total)}</span>
              </div>
              <div class="progress-bar">
                <div
                  class="progress-fill"
                  style="width: {partition.used_percent}%; background: {getColor(partition.used_percent)};"
                ></div>
              </div>
            </div>
          {/each}
        </section>
      {/if}
    </div>

    <div class="footer">
      <label>
        刷新:
        <select bind:value={refreshInterval} on:change={() => monitorStore.setRefreshInterval(refreshInterval)}>
          <option value={1}>1秒</option>
          <option value={2}>2秒</option>
          <option value={3}>3秒</option>
          <option value={5}>5秒</option>
          <option value={10}>10秒</option>
        </select>
      </label>
    </div>
  </div>
{/if}

<style>
  .monitor-panel {
    background: var(--bg-secondary);
    border-left: 1px solid var(--border-primary);
    overflow-y: auto;
    flex-shrink: 0;
    -webkit-app-region: no-drag;
  }

  .monitor-panel.collapsed {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 16px 8px;
  }

  .collapsed-content {
    display: flex;
    flex-direction: column;
    gap: 20px;
    width: 100%;
  }

  .metric {
    text-align: center;
  }

  .metric .label {
    font-size: 11px;
    color: var(--text-secondary);
    margin-bottom: 4px;
  }

  .metric .value {
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .metric .value.small {
    font-size: 11px;
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

  .monitor-resizer {
    width: 4px;
    background: transparent;
    cursor: col-resize;
    flex-shrink: 0;
    position: relative;
    -webkit-app-region: no-drag;
  }

  .monitor-resizer:hover,
  .monitor-resizer.dragging {
    background: var(--accent-primary);
  }

  .monitor-panel.expanded {
    display: flex;
    flex-direction: column;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    border-bottom: 1px solid var(--border-primary);
  }

  .title {
    font-size: 13px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .collapse-btn {
    padding: 4px 8px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 16px;
  }

  .collapse-btn:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
  }

  section {
    margin-bottom: 20px;
  }

  section:last-child {
    margin-bottom: 0;
  }

  h3 {
    margin: 0 0 12px 0;
    font-size: 12px;
    font-weight: 600;
    color: var(--text-secondary);
    text-transform: uppercase;
  }

  .info-grid {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .info-item {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
  }

  .info-label {
    color: var(--text-secondary);
  }

  .info-value {
    color: var(--text-primary);
  }

  .progress-item {
    margin-bottom: 12px;
  }

  .progress-item:last-child {
    margin-bottom: 0;
  }

  .progress-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 4px;
    font-size: 11px;
    color: var(--text-primary);
  }

  .progress-bar {
    height: 6px;
    background: var(--bg-tertiary);
    border-radius: 3px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    transition: width 0.3s ease;
  }

  .cpu-details {
    display: flex;
    gap: 12px;
    margin-top: 8px;
    font-size: 11px;
    color: var(--text-secondary);
  }

  .footer {
    padding: 12px 16px;
    border-top: 1px solid var(--border-primary);
  }

  .footer label {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--text-primary);
  }

  .footer select {
    padding: 4px 8px;
    background: var(--bg-input);
    border: 1px solid var(--border-primary);
    color: var(--text-primary);
    border-radius: 4px;
    font-size: 12px;
  }

  .loading,
  .error-message {
    padding: 20px;
    text-align: center;
    color: var(--text-secondary);
  }

  .error-message {
    color: var(--accent-error);
  }
</style>
