<script>
  import { onDestroy, onMount } from 'svelte';
  import { assetsStore, connectionsStore } from '../stores.js';
  import { buildAssetCards, groupCards, stateLabel, summarizeCards } from '../lib/assetWall.js';

  export let onConnect = () => {};
  export let onCreateConnection = () => {};

  const REFRESH_INTERVAL_MS = 15000;

  let probes = new Map();
  let metrics = new Map();
  let lastSyncAt = 0;
  let probing = false;
  let probedOnce = false;
  let refreshTimer = null;

  $: assets = $assetsStore || [];
  $: sessions = $connectionsStore ? Array.from($connectionsStore.values()) : [];
  $: cards = buildAssetCards(assets, sessions, probes, metrics);
  $: groups = groupCards(cards);
  $: summary = summarizeCards(cards);
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

      if (typeof bindings.GetMonitoringData === 'function') {
        const liveSessions = sessions.filter(
          (session) => session?.connected && session?.type !== 'database' && session?.sessionId
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

  onMount(() => {
    refresh();
    refreshTimer = setInterval(refresh, REFRESH_INTERVAL_MS);
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
      <p>全部资产每 15 秒探活一次；已连接的 SSH 会话叠加真实 CPU / 内存。点击卡片即可连接或切回会话。</p>
    </div>
    <div class="live-status" aria-label="同步状态">
      <span class="pulse-dot"></span>
      <strong>{probing ? '正在探活…' : `${summary.available} / ${summary.total} 个资产可用`}</strong>
      <small>最后同步：{lastSyncText}</small>
      <button type="button" class="refresh-button" on:click={refresh} disabled={probing}>立即刷新</button>
    </div>
  </section>

  <section class="overview-grid" aria-label="资产概览">
    <article class="metric-card emphasis">
      <div class="metric-header"><span>已管理资源</span><span class="metric-mark">01</span></div>
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
      <span>保存 SSH 或数据库连接后，它会显示在这面墙上。</span>
      <button type="button" on:click={() => onCreateConnection()}>新建连接</button>
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
            <button type="button" class="asset-card" on:click={() => onConnect(card.asset)} title={`连接 ${card.name}`}>
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
              {#if card.connected && (card.cpuPercent !== null || card.memPercent !== null)}
                <span class="asset-metrics">
                  {#if card.cpuPercent !== null}
                    <span class="metric-line"><span>CPU</span><i><b style="width: {Math.min(100, card.cpuPercent)}%"></b></i><em>{card.cpuPercent}%</em></span>
                  {/if}
                  {#if card.memPercent !== null}
                    <span class="metric-line"><span>内存</span><i><b class="violet" style="width: {Math.min(100, card.memPercent)}%"></b></i><em>{card.memPercent}%</em></span>
                  {/if}
                </span>
              {/if}
            </button>
          {/each}
        </div>
      </section>
    {/each}
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
  .pulse-dot, .status-dot { background: var(--ops-success); border-radius: 999px; height: 8px; width: 8px; flex-shrink: 0; }
  .pulse-dot { box-shadow: 0 0 0 0 color-mix(in srgb, var(--ops-success) 66%, transparent); margin: 6px 9px 0 0; animation: signal-pulse 2.2s infinite; }
  .status-dot.state-offline { background: var(--ops-alert); }
  .status-dot.state-unknown { background: var(--text-tertiary); }
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
  .asset-card { background: var(--panel); border: 1px solid var(--border-primary); border-radius: 12px; color: var(--text-primary); cursor: pointer; display: grid; gap: 7px; font: inherit; padding: 14px 15px; text-align: left; transition: border-color 180ms ease, background-color 180ms ease; }
  .asset-card:hover { background: var(--accent-subtle); border-color: color-mix(in srgb, var(--ops-signal) 45%, var(--border-primary)); }
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
  .metric-line em { color: var(--text-secondary); font-family: var(--terminal-font-family); font-size: 10px; font-style: normal; text-align: right; }
  .empty-panel { align-items: center; background: var(--panel); border: 1px solid var(--border-primary); border-radius: 14px; color: var(--text-secondary); display: flex; flex-direction: column; gap: 9px; justify-content: center; margin-top: 22px; min-height: 220px; text-align: center; }
  .empty-panel span { color: var(--text-tertiary); font-size: 12px; }
  .empty-panel button { background: transparent; border: 1px solid var(--border-primary); border-radius: 8px; color: var(--text-secondary); cursor: pointer; font: inherit; font-size: 12px; min-height: 36px; padding: 0 11px; }
  .empty-panel button:hover { border-color: var(--ops-signal); color: var(--ops-signal); }
  @keyframes signal-pulse { 70% { box-shadow: 0 0 0 7px transparent; } 100% { box-shadow: 0 0 0 0 transparent; } }
  @media (max-width: 900px) { .overview-grid { grid-template-columns: repeat(2, 1fr); } }
  @media (max-width: 620px) { .ops-dashboard { padding: 18px; } .dashboard-intro { align-items: flex-start; flex-direction: column; } .overview-grid { grid-template-columns: 1fr; } .live-status { width: 100%; } }
  @media (prefers-reduced-motion: reduce) { .pulse-dot { animation: none; } }
</style>
