<script>
  import { onDestroy, onMount } from 'svelte';
  import {
    GetJDBCAgentLogTail,
    GetJDBCAgentStatus,
    GetJDBCRuntimeStatus,
    ImportJDBCDriverPackage,
    ImportJDBCRuntimeArchive,
    InstallJDBCDriver,
    InstallJDBCManagedRuntime,
    ListJDBCDrivers,
    RemoveJDBCDriver,
    RestartJDBCAgent,
    SelectJDBCDriverPackage,
    SelectJDBCJavaExecutable,
    SelectJDBCRuntimeArchive,
    SetJDBCRuntimeMode,
    ValidateJDBCDriver
  } from '../../wailsjs/go/main/App.js';

  let drivers = [];
  let runtimeStatus = null;
  let agentStatus = null;
  let selectedDriverId = '';
  let selectedProfileId = '';
  let search = '';
  let filter = 'all';
  let isBusy = false;
  let activeTaskMessage = '';
  let errorMessage = '';
  let errorCode = '';
  let rawError = '';
  let showRawError = false;
  let resourcePath = '';
  let statusTimer = null;
  let statusPollError = '';
  const statusPollInterval = 2000;
  let showAgentLog = false;
  let agentLog = null;
  let agentLogBusy = false;
  let agentLogError = '';
  let agentLogCopyStatus = '';

  $: filteredDrivers = drivers.filter((driver) => {
    const text = `${driver.name || ''} ${driver.id || ''}`.toLowerCase();
    const matchesSearch = text.includes(search.trim().toLowerCase());
    const matchesFilter =
      filter === 'all' ||
      (filter === 'installed' && driver.installed) ||
      (filter === 'missing' && !driver.installed);
    return matchesSearch && matchesFilter;
  });

  $: selectedDriver =
    drivers.find((driver) => driver.id === selectedDriverId) ||
    filteredDrivers[0] ||
    drivers[0] ||
    null;

  $: profiles = selectedDriver?.profiles || [];
  $: selectedProfile =
    profiles.find((profile) => profile.id === selectedProfileId) ||
    profiles.find((profile) => profile.version === selectedDriver?.recommendedVersion) ||
    profiles[0] ||
    null;

  $: agentStatusLabel = activeTaskMessage ? activeTaskMessage : agentStateLabel(agentStatus?.state);

  onMount(() => {
    loadData();
    statusTimer = setInterval(pollJDBCStatus, statusPollInterval);
  });

  onDestroy(() => {
    if (statusTimer) clearInterval(statusTimer);
  });

  async function pollJDBCStatus() {
    if (typeof document !== 'undefined' && document.hidden) return;
    try {
      const [runtime, agent] = await Promise.all([GetJDBCRuntimeStatus(), GetJDBCAgentStatus()]);
      runtimeStatus = runtime;
      agentStatus = agent;
      statusPollError = '';
    } catch (error) {
      statusPollError = error?.message || String(error);
    }
  }

  async function loadData() {
    await runTask('正在刷新驱动状态', async () => {
      const [driverList, runtime, agent] = await Promise.all([
        ListJDBCDrivers(),
        GetJDBCRuntimeStatus(),
        GetJDBCAgentStatus()
      ]);
      drivers = driverList || [];
      runtimeStatus = runtime;
      agentStatus = agent;
      if (!selectedDriverId && drivers.length > 0) {
        selectedDriverId = drivers[0].id;
      }
    });
  }

  function selectDriver(driver) {
    selectedDriverId = driver.id;
    selectedProfileId = '';
    errorMessage = '';
  }

  async function runTask(message, task) {
    isBusy = true;
    activeTaskMessage = message;
    errorMessage = '';
    errorCode = '';
    rawError = '';
    showRawError = false;
    resourcePath = '';
    try {
      await task();
    } catch (error) {
      rawError = error?.message || String(error);
      errorCode = parseJDBCErrorCode(rawError);
      errorMessage = jdbcErrorLabel(errorCode);
    } finally {
      isBusy = false;
      activeTaskMessage = '';
    }
  }

  async function installSelected() {
    if (!selectedDriver || !selectedProfile) return;
    await runTask('正在安装推荐驱动', async () => {
      await InstallJDBCDriver(selectedDriver.id, selectedProfile.version);
      await loadData();
    });
  }

  async function importOfflinePackage() {
    const path = await SelectJDBCDriverPackage();
    if (!path) return;
    await runTask('正在导入离线驱动包', async () => {
      await ImportJDBCDriverPackage(path);
      await loadData();
    });
  }

  async function validateSelected() {
    if (!selectedDriver || !selectedProfile) return;
    await runTask('正在校验驱动文件', async () => {
      await ValidateJDBCDriver(selectedDriver.id, selectedProfile.version);
      activeTaskMessage = '校验完成';
    });
  }

  async function removeSelected() {
    if (!selectedDriver || !selectedProfile) return;
    const confirmed = window.confirm(`卸载 ${selectedDriver.name} ${selectedProfile.version}？`);
    if (!confirmed) return;
    await runTask('正在卸载驱动', async () => {
      await RemoveJDBCDriver(selectedDriver.id, selectedProfile.version);
      await loadData();
    });
  }

  async function restartAgent() {
    await runTask('正在重启 agent', async () => {
      try {
        await RestartJDBCAgent();
      } finally {
        agentStatus = await GetJDBCAgentStatus();
      }
    });
  }

  function parseJDBCErrorCode(message) {
    const knownCodes = [
      'RUNTIME_MISSING',
      'DRIVER_MISSING',
      'DRIVER_INVALID',
      'AGENT_UNAVAILABLE',
      'DB_CONNECT_FAILED'
    ];
    const upper = String(message || '').toUpperCase();
    const exact = knownCodes.find((code) => upper.includes(code));
    if (exact) return exact;
    if (upper.includes('JAVA') || upper.includes('运行时') || upper.includes('JRE')) return 'RUNTIME_MISSING';
    return 'DB_CONNECT_FAILED';
  }

  function jdbcErrorLabel(code) {
    switch (code) {
      case 'RUNTIME_MISSING':
        return '未找到可用 Java 运行时';
      case 'DRIVER_MISSING':
        return '所需 JDBC 驱动尚未安装';
      case 'DRIVER_INVALID':
        return 'JDBC 驱动文件校验失败';
      case 'AGENT_UNAVAILABLE':
        return 'JDBC agent 当前不可用';
      default:
        return '数据库连接失败';
    }
  }

  async function useManagedRuntime() {
    await runTask('正在安装托管 JRE', async () => {
      applyActivationResult(await InstallJDBCManagedRuntime());
    });
  }

  async function importRuntimeArchive() {
    const archivePath = await SelectJDBCRuntimeArchive();
    if (!archivePath) return;
    await runTask('正在导入 Java 运行时', async () => {
      applyActivationResult(await ImportJDBCRuntimeArchive(archivePath));
    });
  }

  async function chooseJavaRuntime() {
    const javaPath = await SelectJDBCJavaExecutable();
    if (!javaPath) return;
    await runTask('正在更新 Java 运行时', async () => {
      applyActivationResult(await SetJDBCRuntimeMode('system', javaPath));
    });
  }

  function applyActivationResult(result) {
    if (!result) return;
    runtimeStatus = result.runtime || runtimeStatus;
    agentStatus = result.agent || agentStatus;
  }

  async function openAgentLog() {
    showAgentLog = true;
    await refreshAgentLog();
  }

  function closeAgentLog() {
    showAgentLog = false;
    agentLogError = '';
    agentLogCopyStatus = '';
  }

  async function refreshAgentLog() {
    agentLogBusy = true;
    agentLogError = '';
    agentLogCopyStatus = '';
    try {
      agentLog = await GetJDBCAgentLogTail(65536);
    } catch (error) {
      agentLogError = error?.message || String(error);
    } finally {
      agentLogBusy = false;
    }
  }

  async function copyAgentLog() {
    if (!agentLog?.content) return;
    try {
      await navigator.clipboard.writeText(agentLog.content);
      agentLogCopyStatus = '已复制';
    } catch (error) {
      agentLogCopyStatus = '';
      agentLogError = `复制日志失败：${error?.message || String(error)}`;
    }
  }

  function formatLogSize(size) {
    if (!size) return '0 B';
    if (size < 1024) return `${size} B`;
    if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KiB`;
    return `${(size / (1024 * 1024)).toFixed(1)} MiB`;
  }

  function showResource(path) {
    resourcePath = path;
  }

  function runtimeLabel(kind) {
    switch (kind) {
      case 'managed':
        return '托管 JRE';
      case 'system':
        return '系统 Java';
      case 'missing':
        return '未配置';
      default:
        return '未知';
    }
  }

  function agentStateLabel(state) {
    switch (state) {
      case 'starting':
        return '启动中';
      case 'running':
        return '运行中';
      case 'failed':
        return '启动失败';
      case 'stopped':
        return '已停止';
      default:
        return '未知';
    }
  }
</script>

<div class="jdbc-manager">
  <header class="jdbc-manager__status">
    <div>
      <div class="jdbc-manager__eyebrow">JRE</div>
      <div class="jdbc-manager__status-title">{runtimeLabel(runtimeStatus?.kind)}</div>
      <div class="jdbc-manager__status-meta">{runtimeStatus?.javaPath || '未找到可用 Java 运行时'}</div>
    </div>
    <div>
      <div class="jdbc-manager__eyebrow">Agent</div>
      <div class="jdbc-manager__status-title">{agentStatusLabel}</div>
      <div class="jdbc-manager__status-meta">{agentStatus?.lastError || '本地 gRPC 子进程'}</div>
      {#if statusPollError}
        <div class="jdbc-manager__status-warning">状态刷新失败</div>
      {/if}
    </div>
    <div class="jdbc-manager__status-actions">
      <button type="button" class="jdbc-manager__button" disabled={isBusy} on:click={useManagedRuntime}>托管 JRE</button>
      <button type="button" class="jdbc-manager__button" disabled={isBusy} on:click={importRuntimeArchive}>导入 JRE</button>
      <button type="button" class="jdbc-manager__button" disabled={isBusy} on:click={chooseJavaRuntime}>系统 Java</button>
      <button type="button" class="jdbc-manager__button" disabled={isBusy} on:click={loadData}>刷新</button>
      <button type="button" class="jdbc-manager__button jdbc-manager__button--primary" disabled={isBusy} on:click={restartAgent}>重启 agent</button>
    </div>
  </header>

  {#if errorMessage}
    <div class="jdbc-manager__error" role="alert">
      <strong>{errorMessage}</strong>
      <div class="jdbc-manager__error-actions">
        {#if errorCode === 'RUNTIME_MISSING'}
          <button type="button" disabled={isBusy} on:click={useManagedRuntime}>安装 JRE</button>
          <button type="button" disabled={isBusy} on:click={importRuntimeArchive}>导入 JRE</button>
          <button type="button" disabled={isBusy} on:click={chooseJavaRuntime}>选择系统 Java</button>
        {:else if errorCode === 'DRIVER_MISSING'}
          <button type="button" disabled={isBusy} on:click={installSelected}>安装推荐驱动</button>
          <button type="button" disabled={isBusy} on:click={importOfflinePackage}>导入离线包</button>
        {:else if errorCode === 'DRIVER_INVALID'}
          <button type="button" disabled={isBusy} on:click={installSelected}>重新安装</button>
          <button type="button" on:click={() => showResource(selectedProfile?.installPath || '~/.sshtools/drivers')}>查看文件</button>
          <button type="button" disabled={isBusy} on:click={removeSelected}>删除</button>
        {:else if errorCode === 'AGENT_UNAVAILABLE'}
          <button type="button" disabled={isBusy} on:click={restartAgent}>重启 agent</button>
          <button type="button" on:click={openAgentLog}>查看日志</button>
        {:else}
          <button type="button" on:click={() => (showRawError = !showRawError)}>查看原始错误</button>
        {/if}
      </div>
      {#if resourcePath}
        <code class="jdbc-manager__error-detail">{resourcePath}</code>
      {/if}
      {#if showRawError}
        <pre class="jdbc-manager__error-detail">{rawError}</pre>
      {/if}
    </div>
  {/if}

  <div class="jdbc-manager__body">
    <aside class="jdbc-manager__list">
      <div class="jdbc-manager__filters">
        <input bind:value={search} class="jdbc-manager__search" placeholder="搜索驱动" />
        <div class="jdbc-manager__segments">
          <button type="button" class:active={filter === 'all'} on:click={() => (filter = 'all')}>全部</button>
          <button type="button" class:active={filter === 'installed'} on:click={() => (filter = 'installed')}>已安装</button>
          <button type="button" class:active={filter === 'missing'} on:click={() => (filter = 'missing')}>未安装</button>
        </div>
      </div>

      <div class="jdbc-manager__driver-list">
        {#each filteredDrivers as driver}
          <button
            type="button"
            class="jdbc-manager__driver"
            class:active={driver.id === selectedDriver?.id}
            on:click={() => selectDriver(driver)}
          >
            <span>
              <strong>{driver.name}</strong>
              <small>{driver.id} · 推荐 {driver.recommendedVersion || '未指定'}</small>
            </span>
            <em class:installed={driver.installed}>{driver.installed ? '已安装' : '未安装'}</em>
          </button>
        {/each}
      </div>
    </aside>

    <section class="jdbc-manager__detail">
      {#if selectedDriver}
        <div class="jdbc-manager__detail-head">
          <div>
            <div class="jdbc-manager__eyebrow">驱动详情</div>
            <h3>{selectedDriver.name}</h3>
            <p>{selectedDriver.id} · 推荐版本 {selectedDriver.recommendedVersion || '未指定'}</p>
          </div>
          <div class="jdbc-manager__detail-actions">
            {#if selectedDriver.installed}
              <button type="button" class="jdbc-manager__button" disabled={isBusy} on:click={validateSelected}>校验</button>
              <button type="button" class="jdbc-manager__button" disabled={isBusy} on:click={installSelected}>重新安装</button>
              <button type="button" class="jdbc-manager__button jdbc-manager__button--danger" disabled={isBusy} on:click={removeSelected}>卸载</button>
            {:else}
              <button type="button" class="jdbc-manager__button jdbc-manager__button--primary" disabled={isBusy} on:click={installSelected}>安装</button>
              <button type="button" class="jdbc-manager__button" disabled={isBusy} on:click={importOfflinePackage}>导入离线包</button>
            {/if}
          </div>
        </div>

        <div class="jdbc-manager__profile-row">
          {#each profiles as profile}
            <button
              type="button"
              class="jdbc-manager__profile"
              class:active={profile.id === selectedProfile?.id}
              on:click={() => (selectedProfileId = profile.id)}
            >
              <span>{profile.version}</span>
              <small>{profile.installed ? '已安装' : '可用 profile'}</small>
            </button>
          {/each}
        </div>

        {#if selectedProfile}
          <div class="jdbc-manager__grid">
            <div class="jdbc-manager__field">
              <span>Driver class</span>
              <code>{selectedProfile.driverClass}</code>
            </div>
            <div class="jdbc-manager__field">
              <span>URL template</span>
              <code>{selectedProfile.urlTemplate}</code>
            </div>
            <div class="jdbc-manager__field">
              <span>默认端口</span>
              <strong>{selectedProfile.defaultPort}</strong>
            </div>
            <div class="jdbc-manager__field">
              <span>JRE 要求</span>
              <strong>{selectedProfile.jre}</strong>
            </div>
          </div>

          <div class="jdbc-manager__section">
            <h4>Jar 文件</h4>
            {#each selectedProfile.jars || [] as jar}
              <div class="jdbc-manager__jar">
                <span>{jar.name}</span>
                <code>{jar.sha256 || '未提供 checksum'}</code>
              </div>
            {/each}
          </div>

          <div class="jdbc-manager__section">
            <h4>高级配置</h4>
            {#if selectedProfile.properties?.length}
              {#each selectedProfile.properties as prop}
                <div class="jdbc-manager__prop">
                  <span>{prop.name}</span>
                  <small>{prop.required ? '必填' : '可选'} · 默认 {prop.defaultValue || '无'}</small>
                </div>
              {/each}
            {:else}
              <div class="jdbc-manager__empty">当前 profile 没有额外属性。</div>
            {/if}
          </div>
        {/if}
      {:else}
        <div class="jdbc-manager__empty">没有可显示的 JDBC 驱动。</div>
      {/if}
    </section>
  </div>

  {#if showAgentLog}
    <div class="jdbc-manager__log-dialog" role="dialog" aria-modal="true" aria-labelledby="jdbc-agent-log-title">
      <div class="jdbc-manager__log-panel">
        <header class="jdbc-manager__log-header">
          <div>
            <h3 id="jdbc-agent-log-title">JDBC Agent 日志</h3>
            <p>
              {formatLogSize(agentLog?.size || 0)}
              {#if agentLog?.truncated} · 仅显示最近 64 KiB{/if}
            </p>
          </div>
          <div class="jdbc-manager__log-actions">
            <button type="button" disabled={agentLogBusy} on:click={refreshAgentLog}>刷新</button>
            <button type="button" disabled={!agentLog?.content || agentLogBusy} on:click={copyAgentLog}>复制</button>
            <button type="button" on:click={closeAgentLog}>关闭</button>
          </div>
        </header>
        {#if agentLogError}
          <div class="jdbc-manager__log-error" role="alert">{agentLogError}</div>
        {:else if agentLogBusy}
          <div class="jdbc-manager__log-empty">正在读取日志...</div>
        {:else if agentLog?.content}
          <pre class="jdbc-manager__log-content">{agentLog.content}</pre>
        {:else}
          <div class="jdbc-manager__log-empty">暂无 JDBC Agent 日志。</div>
        {/if}
        {#if agentLogCopyStatus}
          <div class="jdbc-manager__log-copy-status">{agentLogCopyStatus}</div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .jdbc-manager {
    display: flex;
    flex-direction: column;
    gap: 14px;
    color: var(--text-primary);
  }

  .jdbc-manager__status,
  .jdbc-manager__body,
  .jdbc-manager__detail,
  .jdbc-manager__list {
    border: 1px solid var(--border-primary);
    background: var(--bg-secondary);
    border-radius: 8px;
  }

  .jdbc-manager__status {
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
    gap: 16px;
    padding: 14px;
    align-items: center;
  }

  .jdbc-manager__eyebrow {
    color: var(--text-tertiary);
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
  }

  .jdbc-manager__status-title {
    color: var(--text-primary);
    font-size: 14px;
    font-weight: 700;
  }

  .jdbc-manager__status-meta {
    color: var(--text-tertiary);
    font-size: 12px;
    overflow-wrap: anywhere;
  }

  .jdbc-manager__status-warning {
    color: #b45309;
    font-size: 11px;
    margin-top: 2px;
  }

  .jdbc-manager__status-actions,
  .jdbc-manager__detail-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    justify-content: flex-end;
  }

  .jdbc-manager__button {
    min-height: 32px;
    padding: 0 12px;
    border: 1px solid var(--border-primary);
    border-radius: 7px;
    background: var(--bg-tertiary);
    color: var(--text-primary);
    font-size: 12px;
    font-weight: 700;
  }

  .jdbc-manager__button:disabled {
    cursor: not-allowed;
    opacity: 0.55;
  }

  .jdbc-manager__button--primary {
    border-color: var(--accent-primary);
    background: var(--accent-primary);
    color: white;
  }

  .jdbc-manager__button--danger {
    border-color: rgba(220, 38, 38, 0.45);
    background: rgba(220, 38, 38, 0.1);
    color: #dc2626;
  }

  .jdbc-manager__error {
    border: 1px solid rgba(220, 38, 38, 0.35);
    border-radius: 8px;
    background: rgba(220, 38, 38, 0.08);
    color: #dc2626;
    padding: 10px 12px;
    font-size: 12px;
  }

  .jdbc-manager__error strong {
    display: block;
  }

  .jdbc-manager__error-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 7px;
    margin-top: 8px;
  }

  .jdbc-manager__error-actions button {
    min-height: 28px;
    border: 1px solid rgba(220, 38, 38, 0.35);
    border-radius: 6px;
    background: var(--bg-secondary);
    color: #dc2626;
    padding: 0 9px;
    font-size: 11px;
    font-weight: 700;
  }

  .jdbc-manager__error-actions button:disabled {
    opacity: 0.55;
  }

  .jdbc-manager__error-detail {
    display: block;
    max-height: 96px;
    overflow: auto;
    margin: 8px 0 0;
    border: 1px solid rgba(220, 38, 38, 0.2);
    border-radius: 6px;
    background: var(--bg-input);
    padding: 8px;
    color: var(--text-primary);
    font-family: Menlo, Monaco, 'Courier New', monospace;
    font-size: 11px;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .jdbc-manager__body {
    display: grid;
    grid-template-columns: 280px minmax(0, 1fr);
    min-height: 520px;
    overflow: hidden;
  }

  .jdbc-manager__list {
    border-width: 0 1px 0 0;
    border-radius: 0;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  .jdbc-manager__filters {
    border-bottom: 1px solid var(--border-primary);
    padding: 12px;
  }

  .jdbc-manager__search {
    width: 100%;
    min-height: 34px;
    border: 1px solid var(--border-primary);
    border-radius: 7px;
    background: var(--bg-input);
    color: var(--text-primary);
    padding: 0 10px;
    font-size: 12px;
  }

  .jdbc-manager__segments {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 4px;
    margin-top: 8px;
  }

  .jdbc-manager__segments button {
    min-height: 28px;
    border: 1px solid transparent;
    border-radius: 6px;
    background: var(--bg-tertiary);
    color: var(--text-secondary);
    font-size: 11px;
    font-weight: 700;
  }

  .jdbc-manager__segments button.active {
    border-color: var(--accent-primary);
    color: var(--accent-primary);
    background: var(--accent-subtle);
  }

  .jdbc-manager__driver-list {
    overflow-y: auto;
    padding: 8px;
  }

  .jdbc-manager__driver {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    border: 1px solid transparent;
    border-radius: 7px;
    background: transparent;
    color: var(--text-primary);
    padding: 10px;
    text-align: left;
  }

  .jdbc-manager__driver.active {
    border-color: var(--accent-primary);
    background: var(--accent-subtle);
  }

  .jdbc-manager__driver strong,
  .jdbc-manager__driver small,
  .jdbc-manager__driver em {
    display: block;
  }

  .jdbc-manager__driver small,
  .jdbc-manager__driver em {
    color: var(--text-tertiary);
    font-size: 11px;
    font-style: normal;
  }

  .jdbc-manager__driver em.installed {
    color: var(--accent-primary);
  }

  .jdbc-manager__detail {
    border: 0;
    border-radius: 0;
    min-width: 0;
    padding: 16px;
  }

  .jdbc-manager__detail-head {
    display: flex;
    justify-content: space-between;
    gap: 16px;
    border-bottom: 1px solid var(--border-primary);
    padding-bottom: 14px;
  }

  .jdbc-manager__detail h3,
  .jdbc-manager__section h4 {
    margin: 0;
    color: var(--text-primary);
  }

  .jdbc-manager__detail p {
    margin: 2px 0 0;
    color: var(--text-tertiary);
    font-size: 12px;
  }

  .jdbc-manager__profile-row {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    padding: 14px 0;
  }

  .jdbc-manager__profile {
    border: 1px solid var(--border-primary);
    border-radius: 7px;
    background: var(--bg-tertiary);
    color: var(--text-primary);
    padding: 8px 10px;
    text-align: left;
  }

  .jdbc-manager__profile.active {
    border-color: var(--accent-primary);
    background: var(--accent-subtle);
  }

  .jdbc-manager__profile span,
  .jdbc-manager__profile small {
    display: block;
  }

  .jdbc-manager__profile small {
    color: var(--text-tertiary);
    font-size: 11px;
  }

  .jdbc-manager__grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
  }

  .jdbc-manager__field,
  .jdbc-manager__section {
    border: 1px solid var(--border-primary);
    border-radius: 8px;
    background: var(--bg-tertiary);
    padding: 12px;
  }

  .jdbc-manager__field span,
  .jdbc-manager__jar code,
  .jdbc-manager__prop small {
    color: var(--text-tertiary);
    font-size: 11px;
  }

  .jdbc-manager__field code,
  .jdbc-manager__field strong {
    display: block;
    margin-top: 4px;
    color: var(--text-primary);
    font-size: 12px;
    overflow-wrap: anywhere;
  }

  .jdbc-manager__section {
    margin-top: 12px;
  }

  .jdbc-manager__jar,
  .jdbc-manager__prop {
    display: flex;
    justify-content: space-between;
    gap: 10px;
    border-top: 1px solid var(--border-secondary);
    margin-top: 8px;
    padding-top: 8px;
    font-size: 12px;
  }

  .jdbc-manager__jar code {
    max-width: 58%;
    overflow-wrap: anywhere;
    text-align: right;
  }

  .jdbc-manager__empty {
    color: var(--text-tertiary);
    font-size: 12px;
    padding: 12px;
  }

  .jdbc-manager__log-dialog {
    position: fixed;
    inset: 0;
    z-index: 90;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.48);
    padding: 24px;
  }

  .jdbc-manager__log-panel {
    width: min(860px, 100%);
    max-height: min(680px, calc(100vh - 48px));
    display: flex;
    flex-direction: column;
    border: 1px solid var(--border-primary);
    border-radius: 8px;
    background: var(--bg-secondary);
    box-shadow: 0 18px 48px rgba(0, 0, 0, 0.28);
    padding: 16px;
  }

  .jdbc-manager__log-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    border-bottom: 1px solid var(--border-primary);
    padding-bottom: 12px;
  }

  .jdbc-manager__log-header h3,
  .jdbc-manager__log-header p {
    margin: 0;
  }

  .jdbc-manager__log-header p {
    color: var(--text-tertiary);
    font-size: 11px;
    margin-top: 3px;
  }

  .jdbc-manager__log-actions {
    display: flex;
    gap: 8px;
  }

  .jdbc-manager__log-actions button {
    min-height: 30px;
    border: 1px solid var(--border-primary);
    border-radius: 6px;
    background: var(--bg-tertiary);
    color: var(--text-primary);
    padding: 0 10px;
    font-size: 12px;
  }

  .jdbc-manager__log-actions button:disabled {
    opacity: 0.5;
  }

  .jdbc-manager__log-content {
    flex: 1;
    min-height: 260px;
    overflow: auto;
    margin: 12px 0 0;
    border: 1px solid var(--border-primary);
    border-radius: 6px;
    background: var(--bg-input);
    color: var(--text-primary);
    padding: 12px;
    font-family: Menlo, Monaco, 'Courier New', monospace;
    font-size: 11px;
    line-height: 1.5;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .jdbc-manager__log-empty,
  .jdbc-manager__log-error {
    margin-top: 12px;
    border: 1px solid var(--border-primary);
    border-radius: 6px;
    padding: 20px;
    color: var(--text-tertiary);
    text-align: center;
    font-size: 12px;
  }

  .jdbc-manager__log-error {
    border-color: rgba(220, 38, 38, 0.35);
    color: #dc2626;
  }

  .jdbc-manager__log-copy-status {
    color: var(--accent-primary);
    font-size: 11px;
    margin-top: 8px;
    text-align: right;
  }

  @media (max-width: 900px) {
    .jdbc-manager__status,
    .jdbc-manager__body,
    .jdbc-manager__grid {
      grid-template-columns: 1fr;
    }

    .jdbc-manager__list {
      border-width: 0 0 1px 0;
    }

    .jdbc-manager__detail-head {
      flex-direction: column;
    }

    .jdbc-manager__log-header {
      flex-direction: column;
    }
  }
</style>
