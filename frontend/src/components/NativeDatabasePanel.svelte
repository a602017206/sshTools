<script>
  import { onMount } from 'svelte';
  import { nativeDatabaseWorkspace } from '../lib/nativeDatabaseWorkspace.js';

  export let sessionId = null;
  export let dbConfig = null;

  let resources = [];
  let childrenByParent = {};
  let expanded = new Set();
  let loading = false;
  let errorMessage = '';

  $: databaseType = dbConfig?.metadata?.db_type || '';
  $: workspace = nativeDatabaseWorkspace(databaseType);

  onMount(loadResources);

  async function loadResources() {
    if (!window.wailsBindings || !sessionId) return;
    loading = true;
    errorMessage = '';
    try {
      resources = await window.wailsBindings.ListNativeDatabaseResources(sessionId) || [];
      childrenByParent = {};
      expanded = new Set();
    } catch (error) {
      errorMessage = `加载${workspace.resourceLabel}失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loading = false;
    }
  }

  async function toggleResource(resource) {
    const name = resource?.name;
    if (!name || !workspace.canExpand) return;

    const next = new Set(expanded);
    if (next.has(name)) {
      next.delete(name);
      expanded = next;
      return;
    }

    try {
      if (!childrenByParent[name]) {
        const children = await window.wailsBindings.ListNativeDatabaseChildResources(sessionId, name);
        childrenByParent = { ...childrenByParent, [name]: children || [] };
      }
      next.add(name);
      expanded = next;
    } catch (error) {
      errorMessage = `加载${workspace.childLabel || '子资源'}失败: ${error?.message || error || '未知错误'}`;
    }
  }
</script>

<section class="native-database-panel" aria-label={workspace.title}>
  <header class="native-database-panel__header native-database-panel__context">
    <div>
      <div class="native-database-panel__eyebrow">{databaseType.toUpperCase() || 'NATIVE DATABASE'}</div>
      <h3>{workspace.title}</h3>
      <p>{dbConfig?.name || '原生数据库连接'} · {workspace.description}</p>
    </div>
    <button class="native-database-panel__refresh" type="button" on:click={loadResources} disabled={loading}>
      {loading ? '加载中…' : '刷新'}
    </button>
  </header>

  <div class="native-database-panel__body">
    <main class="native-database-panel__content">
    {#if errorMessage}
      <div class="native-database-panel__state native-database-panel__error" role="alert">{errorMessage}</div>
    {:else if loading}
      <div class="native-database-panel__state">正在加载{workspace.resourceLabel}…</div>
    {:else if resources.length === 0}
      <div class="native-database-panel__state">未发现可显示的{workspace.resourceLabel}。</div>
    {:else}
    <div class="native-database-panel__summary native-database-panel__resource-count">
      <span>{workspace.resourceLabel}</span>
      <strong>{resources.length}</strong>
    </div>
    <ul class="native-database-panel__tree">
      {#each resources as resource}
        <li class:expanded={expanded.has(resource.name)}>
          {#if workspace.canExpand}
            <button class="native-database-panel__resource" type="button" on:click={() => toggleResource(resource)} aria-expanded={expanded.has(resource.name)}>
              <span class="native-database-panel__chevron">{expanded.has(resource.name) ? '⌄' : '›'}</span>
              <span>{resource.name}</span>
            </button>
          {:else}
            <div class="native-database-panel__resource native-database-panel__resource--leaf">
              <span class="native-database-panel__leaf-mark">•</span>
              <span>{resource.name}</span>
            </div>
          {/if}

          {#if expanded.has(resource.name)}
            <ul class="native-database-panel__children">
              {#if (childrenByParent[resource.name] || []).length === 0}
                <li class="native-database-panel__children-empty">该{workspace.resourceLabel}中没有可显示的{workspace.childLabel}。</li>
              {:else}
                {#each childrenByParent[resource.name] || [] as child}
                  <li class="native-database-panel__child"><span>↳</span>{child.name}</li>
                {/each}
              {/if}
            </ul>
          {/if}
        </li>
      {/each}
    </ul>
    {/if}
    </main>
    <aside class="native-database-panel__inspector">
      <h4>对象信息</h4>
      <dl><div><dt>类型</dt><dd>{databaseType.toUpperCase() || 'NATIVE'}</dd></div><div><dt>资源</dt><dd>{workspace.resourceLabel}</dd></div><div><dt>数量</dt><dd>{resources.length}</dd></div></dl>
      <p>{workspace.description}</p>
    </aside>
  </div>
</section>

<style>
  .native-database-panel {
    height: 100%;
    overflow: auto;
    padding: 20px;
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .native-database-panel__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid var(--border-primary);
  }

  h3 { margin: 3px 0 4px; font-size: 18px; line-height: 1.3; }
  p { max-width: 680px; margin: 0; color: var(--text-secondary); font-size: 12px; line-height: 1.6; }

  .native-database-panel__eyebrow {
    color: var(--accent-primary);
    font-size: 11px;
    font-weight: 750;
    letter-spacing: 0.08em;
  }

  .native-database-panel__refresh {
    flex: 0 0 auto;
    min-height: 32px;
    padding: 0 12px;
    border: 1px solid var(--border-primary);
    border-radius: 5px;
    background: var(--bg-secondary);
    color: var(--text-primary);
    cursor: pointer;
  }

  .native-database-panel__refresh:hover:not(:disabled) { border-color: var(--border-active); color: var(--accent-primary); }
  .native-database-panel__refresh:disabled { cursor: wait; opacity: 0.65; }

  .native-database-panel__state {
    margin-top: 18px;
    padding: 18px;
    border: 1px dashed var(--border-primary);
    border-radius: 6px;
    background: var(--bg-secondary);
    color: var(--text-secondary);
    font-size: 13px;
  }

  .native-database-panel__error { border-color: #dc2626; color: #dc2626; }

  .native-database-panel__summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 18px;
    padding: 9px 12px;
    border: 1px solid var(--border-primary);
    border-radius: 6px 6px 0 0;
    background: var(--bg-tertiary);
    color: var(--text-secondary);
    font-size: 12px;
  }

  .native-database-panel__summary strong { color: var(--text-primary); font-size: 13px; }
  .native-database-panel__tree, .native-database-panel__children { margin: 0; padding: 0; list-style: none; }
  .native-database-panel__tree { border: 1px solid var(--border-primary); border-top: 0; border-radius: 0 0 6px 6px; overflow: hidden; }
  .native-database-panel__tree > li + li { border-top: 1px solid var(--border-primary); }

  .native-database-panel__resource {
    display: flex;
    align-items: center;
    width: 100%;
    min-height: 40px;
    gap: 9px;
    padding: 8px 12px;
    border: 0;
    background: var(--bg-secondary);
    color: var(--text-primary);
    font: inherit;
    font-size: 13px;
    text-align: left;
  }

  button.native-database-panel__resource { cursor: pointer; }
  button.native-database-panel__resource:hover { background: var(--bg-hover); }
  .native-database-panel__resource--leaf { cursor: default; }
  .native-database-panel__chevron { width: 12px; color: var(--text-secondary); font-size: 18px; line-height: 1; text-align: center; }
  .native-database-panel__leaf-mark { color: var(--accent-primary); font-size: 18px; line-height: 1; }

  .native-database-panel__children { padding: 2px 0 7px 33px; background: var(--bg-primary); }
  .native-database-panel__child, .native-database-panel__children-empty { min-height: 29px; padding: 5px 12px; color: var(--text-secondary); font-size: 12px; }
  .native-database-panel__child { display: flex; gap: 8px; align-items: center; color: var(--text-primary); }
  .native-database-panel__child span { color: var(--text-tertiary); }

  @media (max-width: 600px) {
    .native-database-panel { padding: 14px; }
    .native-database-panel__header { gap: 10px; }
    p { font-size: 11px; }
  }

  .native-database-panel { padding: 0; display: flex; flex-direction: column; overflow: hidden; background: #f7f8f5; color: #1d2935; font-family: "PingFang SC", "Hiragino Sans GB", -apple-system, sans-serif; }
  .native-database-panel__context { min-height: 64px; padding: 0 20px; align-items: center; background: #fff; border-bottom-color: #d9e0e4; }
  .native-database-panel__header h3 { margin: 2px 0; font-size: 15px; }
  .native-database-panel__header p { color: #6d7783; }
  .native-database-panel__eyebrow { color: #0e6674; letter-spacing: .1em; }
  .native-database-panel__refresh { min-height: 30px; border-radius: 4px; border-color: #bdd1d4; background: #eff6f5; color: #0e6674; }
  .native-database-panel__body { min-height: 0; flex: 1; display: grid; grid-template-columns: minmax(0, 1fr) 250px; overflow: hidden; }
  .native-database-panel__content { min-width: 0; overflow: auto; padding: 16px; background: #fff; }
  .native-database-panel__resource-count { margin-top: 0; border-radius: 4px 4px 0 0; border-color: #d9e0e4; background: #f4f6f5; }
  .native-database-panel__tree { border-color: #d9e0e4; border-radius: 0 0 4px 4px; }
  .native-database-panel__resource { min-height: 38px; background: #fff; }
  button.native-database-panel__resource:hover { background: #eff6f5; color: #0e6674; }
  .native-database-panel__leaf-mark { color: #0e6674; }
  .native-database-panel__inspector { padding: 18px 16px; border-left: 1px solid #d9e0e4; background: #f7f8f5; overflow: auto; }
  .native-database-panel__inspector h4 { margin: 0 0 14px; color: #31414d; font-size: 11px; letter-spacing: .04em; }
  .native-database-panel__inspector dl { margin: 0; display: grid; gap: 12px; }
  .native-database-panel__inspector dl div { display: grid; gap: 3px; }
  .native-database-panel__inspector dt { color: #7b8791; font-size: 11px; }
  .native-database-panel__inspector dd { margin: 0; color: #31414d; font-size: 12px; }
  .native-database-panel__inspector p { margin-top: 26px; color: #6d7783; font-size: 12px; line-height: 1.6; }
  @media (max-width: 760px) { .native-database-panel__body { grid-template-columns: 1fr; } .native-database-panel__inspector { display: none; } }
</style>
