<script>
  import { getWorkspaceMeta } from '../lib/workspaceTabs.js';

  export let workspace;
  export let onCreateConnection = () => {};

  $: meta = getWorkspaceMeta(workspace);
</script>

<main class="unavailable-workspace">
  <section>
    <p>WORKSPACE / {workspace.toUpperCase()}</p>
    <h1>{meta.title}</h1>
    <span class="line"></span>
    <h2>{meta.description}</h2>
    <div class="steps">
      <div><b>01</b><span>添加受管主机</span><small>从连接中心保存 SSH 或数据库凭据。</small></div>
      <div><b>02</b><span>建立会话</span><small>工作区会自动关联当前连接和权限。</small></div>
      <div><b>03</b><span>接入能力</span><small>后端接入后显示真实资源和可执行操作。</small></div>
    </div>
    <button type="button" on:click={() => onCreateConnection()}>添加连接</button>
  </section>
</main>

<style>
  .unavailable-workspace { align-items: center; color: var(--text-primary); display: flex; height: 100%; justify-content: center; overflow: auto; padding: 28px; }
  section { background: color-mix(in srgb, var(--ops-panel-bg) 88%, #14243b); border: 1px solid var(--border-primary); border-radius: 16px; max-width: 760px; padding: clamp(30px, 6vw, 70px); width: 100%; }
  p { color: var(--ops-signal); font-family: var(--terminal-font-family); font-size: 10px; font-weight: 700; letter-spacing: .14em; margin: 0; } h1 { font-size: clamp(30px, 5vw, 52px); letter-spacing: -.06em; margin: 13px 0 9px; } h2 { color: var(--text-secondary); font-size: 16px; font-weight: 500; line-height: 1.65; margin: 22px 0 30px; max-width: 470px; }.line { background: linear-gradient(90deg, var(--ops-signal), transparent); display: block; height: 2px; max-width: 230px; }.steps { display: grid; gap: 1px; grid-template-columns: repeat(3, 1fr); margin: 0 0 30px; }.steps div { background: color-mix(in srgb, var(--bg-tertiary) 76%, transparent); min-height: 116px; padding: 16px; }.steps b { color: var(--ops-signal); display: block; font-family: var(--terminal-font-family); font-size: 10px; }.steps span { display: block; font-size: 13px; font-weight: 700; margin-top: 14px; }.steps small { color: var(--text-tertiary); display: block; font-size: 11px; line-height: 1.5; margin-top: 4px; }button { background: var(--ops-signal); border: 0; border-radius: 8px; color: #02131a; cursor: pointer; font: inherit; font-size: 13px; font-weight: 800; min-height: 42px; padding: 0 16px; } @media (max-width: 620px) { .steps { grid-template-columns: 1fr; } }
</style>
