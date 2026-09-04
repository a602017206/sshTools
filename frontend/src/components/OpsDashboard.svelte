<script>
  import { assetsStore, connectionsStore } from '../stores.js';

  export let onCreateConnection = () => {};
  export let onSelectWorkspace = () => {};

  const quickActions = [
    { title: '新建 SSH', detail: '添加一台受管主机', action: 'connect' },
    { title: '新建查询', detail: '进入数据库工作区', action: 'database' },
    { title: '上传文件', detail: '打开远程文件管理', action: 'files' },
    { title: '容器视图', detail: '查看 Docker 资源', action: 'docker' }
  ];

  $: assets = $assetsStore || [];
  $: sessions = Array.from($connectionsStore?.values?.() || []);
  $: onlineAssets = assets.filter((asset) => asset.status !== 'offline');
  $: databaseAssets = assets.filter((asset) => asset.type === 'database');
  $: sshAssets = assets.filter((asset) => asset.type !== 'database');
  $: activeSessions = sessions.filter((session) => session?.connected);
  $: uptimeText = activeSessions.length ? `${activeSessions.length} 个实时会话` : '等待连接';

  function connectionState(asset) {
    return asset.status === 'offline' ? '离线' : asset.dbConnected || asset.status === 'online' ? '在线' : '待连接';
  }

  function useAction(action) {
    if (action === 'connect') {
      onCreateConnection();
      return;
    }
    onSelectWorkspace(action);
  }
</script>

<main class="ops-dashboard">
  <section class="dashboard-intro" aria-labelledby="dashboard-title">
    <div>
      <p class="eyebrow"><span></span> OPERATOR OVERVIEW</p>
      <h1 id="dashboard-title">所有资源，保持在同一条命令线上。</h1>
      <p>从连接状态开始，然后进入终端、文件或数据库工作区完成工作。</p>
    </div>
    <div class="live-status" aria-label="实时会话状态">
      <span class="pulse-dot"></span>
      <strong>{uptimeText}</strong>
      <small>最后同步：刚刚</small>
    </div>
  </section>

  <section class="overview-grid" aria-label="系统概览">
    <article class="metric-card emphasis">
      <div class="metric-header"><span>已管理资源</span><span class="metric-mark">01</span></div>
      <strong>{assets.length}</strong>
      <p>{onlineAssets.length} 个资源可用</p>
      <div class="signal-line"><i style="width: 78%"></i></div>
    </article>
    <article class="metric-card">
      <div class="metric-header"><span>活跃连接</span><span class="metric-mark">02</span></div>
      <strong>{activeSessions.length}</strong>
      <p>SSH 与数据库会话</p>
      <div class="spark-bars"><i></i><i></i><i></i><i></i><i></i><i></i><i></i></div>
    </article>
    <article class="metric-card">
      <div class="metric-header"><span>数据库实例</span><span class="metric-mark">03</span></div>
      <strong>{databaseAssets.length}</strong>
      <p>统一进入查询工作区</p>
      <div class="ring-line"><span></span></div>
    </article>
  </section>

  <section class="dashboard-columns">
    <article class="dashboard-panel connections-panel">
      <header><div><p class="panel-kicker">LIVE CONNECTIONS</p><h2>资源状态</h2></div><button type="button" on:click={() => onCreateConnection()}>添加连接</button></header>
      {#if assets.length}
        <div class="resource-list">
          {#each assets.slice(0, 6) as asset}
            <div class="resource-row">
              <span class:offline={asset.status === 'offline'} class="status-dot"></span>
              <div><strong>{asset.name}</strong><small>{asset.host || '未配置主机'} · {asset.username || '默认用户'}</small></div>
              <span class="state-label">{connectionState(asset)}</span>
            </div>
          {/each}
        </div>
      {:else}
        <div class="empty-panel"><strong>从第一台服务器开始</strong><span>保存 SSH 或数据库连接后，它会显示在此处。</span><button type="button" on:click={() => onCreateConnection()}>新建 SSH 连接</button></div>
      {/if}
    </article>

    <article class="dashboard-panel resource-panel">
      <header><div><p class="panel-kicker">RESOURCE PULSE</p><h2>运行轮廓</h2></div><button type="button" on:click={() => onSelectWorkspace('performance')}>查看性能</button></header>
      <div class="resource-visuals">
        <div class="dial"><div><strong>{activeSessions.length ? '64' : '0'}%</strong><span>会话容量</span></div></div>
        <div class="telemetry"><div><span>CPU</span><strong>{activeSessions.length ? '18.6%' : '—'}</strong><i class="cyan"></i></div><div><span>内存</span><strong>{activeSessions.length ? '63.7%' : '—'}</strong><i class="violet"></i></div><div><span>网络</span><strong>{activeSessions.length ? '2.41 Mbps' : '—'}</strong><i class="mint"></i></div></div>
      </div>
      <p class="panel-note">连接任意 SSH 主机后，性能页将显示真实采样数据。</p>
    </article>
  </section>

  <section class="quick-actions" aria-label="快捷操作">
    {#each quickActions as action}
      <button type="button" on:click={() => useAction(action.action)}>
        <span class="action-arrow">↗</span><strong>{action.title}</strong><small>{action.detail}</small>
      </button>
    {/each}
  </section>

  <section class="resource-groups" aria-label="资源分类">
    <div><span>SSH CONNECTIONS</span><strong>{sshAssets.length}</strong></div>
    <div><span>DATABASES</span><strong>{databaseAssets.length}</strong></div>
    <div><span>DOCKER HOSTS</span><strong>—</strong></div>
    <div><span>CLOUD SERVERS</span><strong>—</strong></div>
  </section>
</main>

<style>
  .ops-dashboard { --panel: color-mix(in srgb, var(--ops-panel-bg) 84%, #14243b); color: var(--text-primary); height: 100%; overflow: auto; padding: clamp(20px, 3vw, 42px); }
  .dashboard-intro, .dashboard-panel header, .resource-row, .metric-header { display: flex; align-items: center; justify-content: space-between; gap: 18px; }
  .dashboard-intro { border-bottom: 1px solid var(--border-secondary); padding: 0 0 26px; }
  .eyebrow, .panel-kicker { color: var(--ops-signal); font-family: var(--terminal-font-family); font-size: 10px; font-weight: 700; letter-spacing: .12em; margin: 0 0 8px; }
  .eyebrow span { background: var(--ops-signal); box-shadow: 0 0 14px var(--ops-signal); display: inline-block; height: 7px; margin-right: 7px; width: 7px; }
  h1 { font-size: clamp(23px, 2.2vw, 34px); letter-spacing: -.04em; line-height: 1.15; margin: 0; max-width: 620px; }
  h2 { font-size: 16px; letter-spacing: -.02em; margin: 0; } .dashboard-intro p:not(.eyebrow) { color: var(--text-secondary); margin: 10px 0 0; }
  .live-status { background: var(--accent-subtle); border: 1px solid color-mix(in srgb, var(--ops-signal) 22%, var(--border-primary)); border-radius: 12px; display: grid; grid-template-columns: auto 1fr; padding: 12px 15px; min-width: 180px; }
  .live-status small { color: var(--text-tertiary); grid-column: 2; margin-top: 2px; }.pulse-dot, .status-dot { background: var(--ops-success); border-radius: 999px; height: 8px; width: 8px; }.pulse-dot { box-shadow: 0 0 0 0 color-mix(in srgb, var(--ops-success) 66%, transparent); margin: 6px 9px 0 0; animation: signal-pulse 2.2s infinite; }.overview-grid { display: grid; gap: 14px; grid-template-columns: repeat(3, minmax(0, 1fr)); margin: 22px 0 14px; }.metric-card, .dashboard-panel { background: var(--panel); border: 1px solid var(--border-primary); border-radius: 14px; box-shadow: var(--shadow-sm); }.metric-card { min-height: 174px; padding: 17px; position: relative; overflow: hidden; }.metric-card.emphasis { border-color: color-mix(in srgb, var(--ops-signal) 36%, var(--border-primary)); }.metric-card strong { display: block; font-family: var(--terminal-font-family); font-size: 34px; letter-spacing: -.06em; margin-top: 19px; }.metric-card p, .panel-note { color: var(--text-tertiary); font-size: 12px; margin: 4px 0; }.metric-header { color: var(--text-secondary); font-size: 12px; }.metric-mark { color: var(--text-tertiary); font-family: var(--terminal-font-family); font-size: 10px; }.signal-line { background: color-mix(in srgb, var(--border-primary) 58%, transparent); bottom: 19px; height: 3px; left: 17px; position: absolute; right: 17px; }.signal-line i { background: linear-gradient(90deg, var(--ops-signal), var(--ops-blue)); box-shadow: 0 0 15px color-mix(in srgb, var(--ops-signal) 60%, transparent); display: block; height: 100%; }.spark-bars { align-items: end; display: flex; gap: 4px; height: 22px; margin-top: 15px; }.spark-bars i { background: var(--ops-blue); display: block; height: 40%; width: 6px; }.spark-bars i:nth-child(2n) { height: 84%; }.spark-bars i:nth-child(3n) { height: 63%; }.ring-line { border: 3px solid var(--ops-violet); border-left-color: transparent; border-radius: 50%; height: 25px; margin: 14px 0 0; transform: rotate(-42deg); width: 25px; }.dashboard-columns { display: grid; gap: 14px; grid-template-columns: minmax(0, 1.15fr) minmax(360px, .85fr); }.dashboard-panel { min-height: 285px; padding: 20px; }.dashboard-panel header button, .empty-panel button { background: transparent; border: 1px solid var(--border-primary); border-radius: 8px; color: var(--text-secondary); cursor: pointer; font: inherit; font-size: 12px; min-height: 36px; padding: 0 11px; }.dashboard-panel header button:hover, .empty-panel button:hover { border-color: var(--ops-signal); color: var(--ops-signal); }.resource-list { margin-top: 12px; }.resource-row { border-top: 1px solid var(--border-secondary); min-height: 55px; }.resource-row > div { flex: 1; min-width: 0; }.resource-row strong, .resource-row small { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }.resource-row small { color: var(--text-tertiary); font-size: 11px; margin-top: 2px; }.status-dot { margin-right: 9px; }.status-dot.offline { background: var(--ops-alert); }.state-label { color: var(--text-secondary); font-size: 11px; }.resource-visuals { align-items: center; display: flex; gap: 27px; justify-content: center; padding: 18px 0 12px; }.dial { background: conic-gradient(var(--ops-signal) 0 64%, var(--border-primary) 64% 100%); border-radius: 50%; display: grid; height: 134px; place-items: center; position: relative; width: 134px; }.dial::after { background: var(--panel); border-radius: inherit; content: ''; inset: 10px; position: absolute; }.dial div { position: relative; text-align: center; z-index: 1; }.dial strong, .dial span { display: block; }.dial strong { font-family: var(--terminal-font-family); font-size: 25px; }.dial span { color: var(--text-tertiary); font-size: 10px; }.telemetry { display: grid; gap: 13px; flex: 1; }.telemetry div { display: grid; grid-template-columns: 45px 1fr; gap: 5px; }.telemetry span { color: var(--text-tertiary); font-size: 11px; }.telemetry strong { font-family: var(--terminal-font-family); font-size: 12px; text-align: right; }.telemetry i { background: var(--ops-signal); grid-column: 1 / -1; height: 3px; opacity: .86; }.telemetry i.violet { background: var(--ops-violet); width: 64%; }.telemetry i.mint { background: var(--ops-success); width: 44%; }.empty-panel { align-items: center; color: var(--text-secondary); display: flex; flex-direction: column; gap: 9px; justify-content: center; min-height: 185px; text-align: center; }.empty-panel span { color: var(--text-tertiary); font-size: 12px; }.quick-actions { display: grid; gap: 14px; grid-template-columns: repeat(4, minmax(0, 1fr)); margin-top: 14px; }.quick-actions button { background: color-mix(in srgb, var(--panel) 80%, transparent); border: 1px solid var(--border-primary); border-radius: 12px; color: var(--text-primary); cursor: pointer; display: grid; min-height: 118px; padding: 16px; text-align: left; transition: border-color 180ms ease, background-color 180ms ease; }.quick-actions button:hover { background: var(--accent-subtle); border-color: color-mix(in srgb, var(--ops-signal) 45%, var(--border-primary)); }.quick-actions strong, .quick-actions small { display: block; }.quick-actions strong { font-size: 13px; }.quick-actions small { color: var(--text-tertiary); font-size: 11px; margin-top: 4px; }.action-arrow { color: var(--ops-signal); font-family: var(--terminal-font-family); font-size: 18px; margin-bottom: 12px; }.resource-groups { display: grid; gap: 1px; grid-template-columns: repeat(4, 1fr); margin: 20px 0 6px; }.resource-groups div { border-left: 1px solid var(--border-primary); padding: 0 14px; }.resource-groups div:first-child { border-left: 0; padding-left: 0; }.resource-groups span { color: var(--text-tertiary); display: block; font-family: var(--terminal-font-family); font-size: 9px; letter-spacing: .09em; }.resource-groups strong { display: block; font-family: var(--terminal-font-family); font-size: 17px; margin-top: 5px; } @keyframes signal-pulse { 70% { box-shadow: 0 0 0 7px transparent; } 100% { box-shadow: 0 0 0 0 transparent; } } @media (max-width: 900px) { .dashboard-columns { grid-template-columns: 1fr; }.quick-actions { grid-template-columns: repeat(2, 1fr); }.resource-groups { grid-template-columns: repeat(2, 1fr); row-gap: 14px; }.resource-groups div:nth-child(3) { border-left: 0; padding-left: 0; } } @media (max-width: 620px) { .ops-dashboard { padding: 18px; }.dashboard-intro { align-items: flex-start; flex-direction: column; }.overview-grid { grid-template-columns: 1fr; }.resource-visuals { align-items: flex-start; flex-direction: column; }.live-status { width: 100%; } } @media (prefers-reduced-motion: reduce) { .pulse-dot { animation: none; } }
</style>
