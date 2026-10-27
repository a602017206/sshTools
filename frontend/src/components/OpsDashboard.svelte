<script>
  import { onDestroy, onMount } from 'svelte';
  import { assetsStore, connectionsStore } from '../stores.js';
  import { WALL_METRIC_CATALOG, buildAssetCards, groupCards, stateLabel, summarizeCards } from '../lib/assetWall.js';
  import { sessionMatchesMode } from '../lib/workspaceTabs.js';

  export let onConnect = () => {};
  export let onCreateConnection = () => {};

  let probes = new Map();
  let metrics = new Map();
  let lastSyncAt = 0;
  let probing = false;
  let probedOnce = false;
  let refreshTimer = null;
  let wall = { enabled: true, intervalSeconds: 15, metricsEnabled: true };
  let editMode = false;

  function clampIntervalSeconds(value) {
    const seconds = Math.floor(Number(value));
    if (!Number.isFinite(seconds) || seconds <= 0) return 15;
    return Math.min(300, Math.max(5, seconds));
  }

  async function loadWallSettings() {
    try {
      const settings = await window.wailsBindings?.GetSettings?.();
      if (settings) {
        wall = {
          enabled: settings.asset_wall_enabled !== false,
          intervalSeconds: clampIntervalSeconds(settings.asset_wall_probe_interval_seconds),
          metricsEnabled: settings.asset_wall_metrics_enabled !== false
        };
      }
    } catch (error) {
      console.warn('GetSettings for status wall failed:', error);
    }
  }

  $: assets = $assetsStore || [];
  $: sessions = $connectionsStore ? Array.from($connectionsStore.values()) : [];
  $: cards = buildAssetCards(assets, sessions, probes, metrics);
  $: boardCards = cards.filter((card) => card.onBoard);
  $: offBoardCards = cards.filter((card) => !card.onBoard);
  $: groups = groupCards(boardCards);
  $: summary = summarizeCards(boardCards);
  $: lastSyncText = lastSyncAt
    ? new Date(lastSyncAt).toLocaleTimeString('zh-CN', { hour12: false })
    : '尚未同步';

  async function refresh() {
    if (probing) return;
    const bindings = window.wailsBindings || {};
    probing = true;
    try {
      if (typeof bindings.ProbeAllAssets === 'function') {
        const results = await bindings.ProbeAllAssets();
        const next = new Map();
        for (const item of results || []) {
          if (item?.id) next.set(item.id, item);
        }
        probes = next;
        if (results?.length) {
          lastSyncAt = Math.max(...results.map((item) => item.checked_at || 0));
        } else {
          lastSyncAt = Date.now();
        }
      }
      probedOnce = true;

      if (wall.metricsEnabled && typeof bindings.GetMonitoringData === 'function') {
        const boardIds = new Set(
          assets
            .filter((asset) => asset.metadata?.wall_on_board === 'true' && asset.metadata?.status_wall_excluded !== 'true')
            .map((asset) => asset.id)
        );
        const liveSessions = sessions.filter(
          (session) =>
            session?.connected &&
            sessionMatchesMode(session, 'ssh') &&
            session?.sessionId &&
            boardIds.has(session?.connection?.id)
        );
        const nextMetrics = new Map(metrics);
        await Promise.all(
          liveSessions.map(async (session) => {
            try {
              const data = await bindings.GetMonitoringData(session.sessionId);
              if (data) nextMetrics.set(session.sessionId, data);
            } catch (error) {
              console.warn('GetMonitoringData failed:', error);
            }
          })
        );
        metrics = nextMetrics;
      }
    } catch (error) {
      console.warn('ProbeAllAssets failed:', error);
    } finally {
      probing = false;
    }
  }

  async function toggleProbe(card) {
    const bindings = window.wailsBindings || {};
    if (typeof bindings.SetAssetProbeExcluded !== 'function') return;
    const excluded = !card.probeExcluded;
    try {
      await bindings.SetAssetProbeExcluded(card.id, excluded);
      assetsStore.update((items) =>
        items.map((item) => {
          if (item.id !== card.id) return item;
          const metadata = { ...(item.metadata || {}) };
          if (excluded) metadata.status_wall_excluded = 'true';
          else delete metadata.status_wall_excluded;
          return { ...item, metadata };
        })
      );
      if (excluded && probes.has(card.id)) {
        const next = new Map(probes);
        next.delete(card.id);
        probes = next;
      }
    } catch (error) {
      console.warn('SetAssetProbeExcluded failed:', error);
    }
  }

  async function persistWallPrefs(card, onBoard, metricIds) {
    const bindings = window.wailsBindings || {};
    if (typeof bindings.SetAssetWallPrefs !== 'function') return;
    try {
      await bindings.SetAssetWallPrefs(card.id, onBoard, (metricIds || []).join(','));
      assetsStore.update((items) =>
        items.map((item) => {
          if (item.id !== card.id) return item;
          const metadata = { ...(item.metadata || {}) };
          if (onBoard) metadata.wall_on_board = 'true';
          else delete metadata.wall_on_board;
          if (metricIds?.length) metadata.wall_metrics = metricIds.join(',');
          return { ...item, metadata };
        })
      );
    } catch (error) {
      console.warn('SetAssetWallPrefs failed:', error);
    }
  }

  function toggleMetric(card, metricId) {
    const selected = card.wallMetrics.includes(metricId)
      ? card.wallMetrics.filter((id) => id !== metricId)
      : [...card.wallMetrics, metricId];
    persistWallPrefs(card, true, selected);
  }

  onMount(async () => {
    await loadWallSettings();
    if (!wall.enabled) return;
    refresh();
    refreshTimer = setInterval(refresh, wall.intervalSeconds * 1000);
  });

  onDestroy(() => {
    if (refreshTimer) clearInterval(refreshTimer);
  });
</script>

<main class="ops-dashboard">
  <section class="dashboard-intro" aria-labelledby="dashboard-title">
    <div>
      <p class="eyebrow"><span></span> ASSET STATUS WALL</p>
      <h1 id="dashboard-title">多资产状态墙</h1>
      {#if wall.enabled}
        <p>看板内资产每 {wall.intervalSeconds} 秒探活一次；已连接的 SSH 会话按各卡自选参数显示。点「编辑看板」可添加 / 移出机器并逐卡勾选参数，点击卡片即可连接或切回会话。</p>
      {:else}
        <p>自动探活已在「全局设置 → 多资产状态墙」关闭，在线状态不会自动更新；可点「立即刷新」手动探活一次，或回设置开启。</p>
      {/if}
    </div>
    <div class="live-status" aria-label="同步状态">
      <span class="pulse-dot"></span>
      <strong>{probing ? '正在探活…' : `${summary.available} / ${summary.total} 个看板资产可用`}</strong>
      <small>最后同步：{lastSyncText}</small>
      <span class="live-actions">
        <button type="button" class="refresh-button" on:click={refresh} disabled={probing}>立即刷新</button>
        <button type="button" class="refresh-button" on:click={() => { editMode = !editMode; }}>{editMode ? '完成' : '编辑看板'}</button>
      </span>
    </div>
  </section>

  <section class="overview-grid" aria-label="资产概览">
    <article class="metric-card emphasis">
      <div class="metric-header"><span>看板资产</span><span class="metric-mark">01</span></div>
      <strong>{summary.total}</strong>
      <p>{summary.available} 个资产当前可用</p>
    </article>
    <article class="metric-card">
      <div class="metric-header"><span>活跃会话</span><span class="metric-mark">02</span></div>
      <strong>{summary.connected}</strong>
      <p>已建立的 SSH / 数据库会话</p>
    </article>
    <article class="metric-card">
      <div class="metric-header"><span>探活在线</span><span class="metric-mark">03</span></div>
      <strong>{summary.online}</strong>
      <p>未连接但端口可达</p>
    </article>
    <article class="metric-card">
      <div class="metric-header"><span>离线</span><span class="metric-mark">04</span></div>
      <strong>{summary.offline}</strong>
      <p>探活不可达，点卡片查看或重试连接</p>
    </article>
  </section>

  {#if !assets.length}
    <section class="empty-panel">
      <strong>从第一台服务器开始</strong>
      <span>保存 SSH 或数据库连接后，可以在编辑看板时把它加进来。</span>
      <button type="button" on:click={() => onCreateConnection()}>新建连接</button>
    </section>
  {:else if !boardCards.length && !editMode}
    <section class="empty-panel">
      <strong>看板还是空的</strong>
      <span>进入编辑模式，把关心的机器加进看板，并为每台勾选要看的参数。</span>
      <button type="button" on:click={() => { editMode = true; }}>编辑看板</button>
    </section>
  {:else}
    {#each groups as group}
      <section class="wall-group" aria-label={`分组 ${group.name}`}>
        <header class="wall-group-header">
          <h2>{group.name}</h2>
          <span>{group.cards.length} 个资产</span>
        </header>
        <div class="wall-grid">
          {#each group.cards as card}
            <article class="asset-card">
            <button type="button" class="asset-main" on:click={() => onConnect(card.asset)} title={`连接 ${card.name}`}>
              <span class="asset-card-top">
                <span class="status-dot state-{card.state}"></span>
                <strong class="asset-name">{card.name}</strong>
                <span class="state-label state-text-{card.state}">{stateLabel(card.state)}</span>
              </span>
              <span class="asset-meta">{card.host || '未配置主机'}{card.port ? `:${card.port}` : ''} · {card.username || '默认用户'}</span>
              <span class="asset-tags">
                <span class="asset-badge">{card.type === 'database' ? (card.dbType || '数据库').toUpperCase() : 'SSH'}</span>
                {#if card.latencyMs !== null && card.latencyMs !== undefined}
                  <span class="asset-latency">{card.latencyMs} ms</span>
                {:else if card.state === 'offline'}
                  <span class="asset-error">不可达</span>
                {:else if !probedOnce}
                  <span class="asset-latency">探测中…</span>
                {/if}
              </span>
              {#if card.connected && card.metricRows.length}
                <span class="asset-metrics">
                  {#each card.metricRows as row}
                    <span class="metric-line">
                      <span>{row.label}</span>
                      {#if row.percent !== null && row.percent !== undefined}
                        <i><b style="width: {Math.min(100, row.percent)}%"></b></i>
                      {:else}
                        <i class="no-bar"></i>
                      {/if}
                      <em title={row.hint || ''}>{row.text}</em>
                    </span>
                  {/each}
                </span>
              {/if}
            </button>
            {#if editMode}
              <span class="metric-picker" aria-label={`为 ${card.name} 选择显示参数`}>
                {#each WALL_METRIC_CATALOG as metric}
                  <button
                    type="button"
                    class="metric-chip"
                    class:selected={card.wallMetrics.includes(metric.id)}
                    on:click={() => toggleMetric(card, metric.id)}
                  >{metric.label}</button>
                {/each}
              </span>
              <button type="button" class="asset-toggle danger" on:click={() => persistWallPrefs(card, false, card.wallMetrics)}>移出看板</button>
            {/if}
            <button
              type="button"
              class="asset-toggle"
              on:click={() => toggleProbe(card)}
              title={card.probeExcluded ? '该资产已关闭探活，点击恢复' : '不再对该资产探活（连接本身不受影响）'}
            >{card.probeExcluded ? '开启探活' : '关闭探活'}</button>
            </article>
          {/each}
        </div>
      </section>
    {/each}
    {#if editMode}
      <section class="wall-group" aria-label="可添加的机器">
        <header class="wall-group-header">
          <h2>可添加的机器</h2>
          <span>{offBoardCards.length} 个未上看板</span>
        </header>
        {#if offBoardCards.length}
          <div class="add-list">
            {#each offBoardCards as card}
              <div class="add-row">
                <span class="status-dot state-{card.state}"></span>
                <div class="add-info"><strong>{card.name}</strong><small>{card.host || '未配置主机'}{card.port ? `:${card.port}` : ''} · {card.group}</small></div>
                <button type="button" class="add-button" on:click={() => persistWallPrefs(card, true, card.wallMetrics)}>加入看板</button>
              </div>
            {/each}
          </div>
        {:else}
          <p class="panel-note">全部资产都已经在看板上了。</p>
        {/if}
      </section>
    {/if}
  {/if}
</main>

<style>
  .ops-dashboard { --panel: color-mix(in srgb, var(--ops-panel-bg) 84%, #14243b); color: var(--text-primary); height: 100%; overflow: auto; padding: clamp(20px, 3vw, 42px); }
  .dashboard-intro, .metric-header, .wall-group-header { display: flex; align-items: center; justify-content: space-between; gap: 18px; }
  .dashboard-intro { border-bottom: 1px solid var(--border-secondary); padding: 0 0 26px; }
  .eyebrow { color: var(--ops-signal); font-family: var(--terminal-font-family); font-size: 10px; font-weight: 700; letter-spacing: .12em; margin: 0 0 8px; }
  .eyebrow span { background: var(--ops-signal); box-shadow: 0 0 14px var(--ops-signal); display: inline-block; height: 7px; margin-right: 7px; width: 7px; }
  h1 { font-size: clamp(23px, 2.2vw, 34px); letter-spacing: -.04em; line-height: 1.15; margin: 0; }
  h2 { font-size: 14px; letter-spacing: -.02em; margin: 0; }
  .dashboard-intro p:not(.eyebrow) { color: var(--text-secondary); margin: 10px 0 0; max-width: 560px; }
  .live-status { background: var(--accent-subtle); border: 1px solid color-mix(in srgb, var(--ops-signal) 22%, var(--border-primary)); border-radius: 12px; display: grid; grid-template-columns: auto 1fr; padding: 12px 15px; min-width: 200px; }
  .live-status small { color: var(--text-tertiary); grid-column: 2; margin-top: 2px; }
  .refresh-button { background: transparent; border: 1px solid var(--border-primary); border-radius: 8px; color: var(--text-secondary); cursor: pointer; font: inherit; font-size: 12px; grid-column: 2; justify-self: start; margin-top: 8px; min-height: 30px; padding: 0 11px; }
  .refresh-button:hover:not(:disabled) { border-color: var(--ops-signal); color: var(--ops-signal); }
  .refresh-button:disabled { cursor: default; opacity: .55; }
  .live-actions { display: flex; gap: 8px; grid-column: 2; }
  .live-actions .refresh-button { margin-top: 8px; }
  .metric-picker { border-top: 1px solid var(--border-secondary); display: flex; flex-wrap: wrap; gap: 5px; padding: 8px 4px 2px; }
  .metric-chip { background: transparent; border: 1px solid var(--border-primary); border-radius: 999px; color: var(--text-tertiary); cursor: pointer; font: inherit; font-size: 10px; padding: 3px 9px; }
  .metric-chip:hover { border-color: var(--ops-signal); color: var(--ops-signal); }
  .metric-chip.selected { background: var(--ops-signal); border-color: var(--ops-signal); color: #fff; }
  .asset-toggle.danger:hover { color: var(--ops-alert); }
  .metric-line i.no-bar { background: transparent; }
  .panel-note { color: var(--text-tertiary); font-size: 12px; margin: 10px 0 0; }
  .add-list { display: grid; gap: 8px; margin-top: 12px; }
  .add-row { align-items: center; background: var(--panel); border: 1px solid var(--border-primary); border-radius: 10px; display: flex; gap: 10px; padding: 10px 12px; }
  .add-info { flex: 1; min-width: 0; }
  .add-info strong, .add-info small { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .add-info strong { font-size: 12px; }
  .add-info small { color: var(--text-tertiary); font-size: 11px; margin-top: 1px; }
  .add-button { background: transparent; border: 1px solid var(--border-primary); border-radius: 8px; color: var(--text-secondary); cursor: pointer; font: inherit; font-size: 11px; min-height: 30px; padding: 0 11px; }
  .add-button:hover { border-color: var(--ops-signal); color: var(--ops-signal); }
  .pulse-dot, .status-dot { background: var(--ops-success); border-radius: 999px; height: 8px; width: 8px; flex-shrink: 0; }
  .pulse-dot { box-shadow: 0 0 0 0 color-mix(in srgb, var(--ops-success) 66%, transparent); margin: 6px 9px 0 0; animation: signal-pulse 2.2s infinite; }
  .status-dot.state-offline { background: var(--ops-alert); }
  .status-dot.state-unknown, .status-dot.state-excluded { background: var(--text-tertiary); }
  .overview-grid { display: grid; gap: 14px; grid-template-columns: repeat(4, minmax(0, 1fr)); margin: 22px 0 6px; }
  .metric-card { background: var(--panel); border: 1px solid var(--border-primary); border-radius: 14px; box-shadow: var(--shadow-sm); min-height: 128px; padding: 17px; }
  .metric-card.emphasis { border-color: color-mix(in srgb, var(--ops-signal) 36%, var(--border-primary)); }
  .metric-card strong { display: block; font-family: var(--terminal-font-family); font-size: 34px; letter-spacing: -.06em; margin-top: 16px; }
  .metric-card p { color: var(--text-tertiary); font-size: 12px; margin: 4px 0; }
  .metric-header { color: var(--text-secondary); font-size: 12px; }
  .metric-mark { color: var(--text-tertiary); font-family: var(--terminal-font-family); font-size: 10px; }
  .wall-group { margin-top: 22px; }
  .wall-group-header span { color: var(--text-tertiary); font-size: 11px; }
  .wall-grid { display: grid; gap: 12px; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); margin-top: 12px; }
  .asset-card { background: var(--panel); border: 1px solid var(--border-primary); border-radius: 12px; color: var(--text-primary); display: grid; gap: 4px; padding: 10px 11px 8px; transition: border-color 180ms ease, background-color 180ms ease; }
  .asset-card:hover { background: var(--accent-subtle); border-color: color-mix(in srgb, var(--ops-signal) 45%, var(--border-primary)); }
  .asset-main { background: transparent; border: 0; color: inherit; cursor: pointer; display: grid; font: inherit; gap: 7px; padding: 4px; text-align: left; width: 100%; }
  .asset-toggle { background: transparent; border: 0; border-top: 1px solid var(--border-secondary); color: var(--text-tertiary); cursor: pointer; font: inherit; font-size: 10px; justify-self: stretch; padding: 6px 4px 2px; text-align: right; }
  .asset-toggle:hover { color: var(--ops-signal); }
  .asset-card-top { align-items: center; display: flex; gap: 8px; }
  .asset-name { flex: 1; font-size: 13px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .state-label { font-size: 11px; color: var(--text-secondary); }
  .state-text-connected, .state-text-online { color: var(--ops-success); }
  .state-text-offline { color: var(--ops-alert); }
  .asset-meta { color: var(--text-tertiary); font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .asset-tags { align-items: center; display: flex; gap: 8px; }
  .asset-badge { border: 1px solid var(--border-primary); border-radius: 6px; color: var(--text-secondary); font-family: var(--terminal-font-family); font-size: 9px; letter-spacing: .08em; padding: 2px 6px; }
  .asset-latency { color: var(--text-secondary); font-family: var(--terminal-font-family); font-size: 11px; }
  .asset-error { color: var(--ops-alert); font-size: 11px; }
  .asset-metrics { display: grid; gap: 5px; margin-top: 2px; }
  .metric-line { align-items: center; display: grid; gap: 8px; grid-template-columns: 30px 1fr 44px; }
  .metric-line span { color: var(--text-tertiary); font-size: 10px; }
  .metric-line i { background: color-mix(in srgb, var(--border-primary) 58%, transparent); border-radius: 999px; display: block; height: 4px; overflow: hidden; }
  .metric-line b { background: var(--ops-signal); display: block; height: 100%; }
  .metric-line b.violet { background: var(--ops-violet); }
  .metric-line em { color: var(--text-secondary); font-family: var(--terminal-font-family); font-size: 10px; font-style: normal; overflow: hidden; text-align: right; text-overflow: ellipsis; white-space: nowrap; }
  .empty-panel { align-items: center; background: var(--panel); border: 1px solid var(--border-primary); border-radius: 14px; color: var(--text-secondary); display: flex; flex-direction: column; gap: 9px; justify-content: center; margin-top: 22px; min-height: 220px; text-align: center; }
  .empty-panel span { color: var(--text-tertiary); font-size: 12px; }
  .empty-panel button { background: transparent; border: 1px solid var(--border-primary); border-radius: 8px; color: var(--text-secondary); cursor: pointer; font: inherit; font-size: 12px; min-height: 36px; padding: 0 11px; }
  .empty-panel button:hover { border-color: var(--ops-signal); color: var(--ops-signal); }
  @keyframes signal-pulse { 70% { box-shadow: 0 0 0 7px transparent; } 100% { box-shadow: 0 0 0 0 transparent; } }
  @media (max-width: 900px) { .overview-grid { grid-template-columns: repeat(2, 1fr); } }
  @media (max-width: 620px) { .ops-dashboard { padding: 18px; } .dashboard-intro { align-items: flex-start; flex-direction: column; } .overview-grid { grid-template-columns: 1fr; } .live-status { width: 100%; } }
  @media (prefers-reduced-motion: reduce) { .pulse-dot { animation: none; } }
</style>
