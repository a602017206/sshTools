<script>
  import { onMount } from 'svelte';
  import { nativeDatabaseWorkspace } from '../../lib/nativeDatabaseWorkspace.js';
  import { copilotStore } from '../../stores/copilot.js';

  export let sessionId = null;
  export let dbConfig = null;

  let resources = [];
  let loading = false;
  let loadingDetails = false;
  let errorMessage = '';
  let selectedResource = null;
  let details = null;

  $: databaseType = dbConfig?.metadata?.db_type || 'kafka';
  $: workspace = nativeDatabaseWorkspace(databaseType);
  $: if (sessionId) {
    copilotStore.setWorkspaceFocus(sessionId, {
      objectKind: selectedResource ? 'topic' : '',
      objectName: selectedResource || '',
      objectParent: ''
    });
  }

  onMount(loadResources);

  async function loadResources() {
    if (!window.wailsBindings || !sessionId) return;
    loading = true;
    errorMessage = '';
    try {
      resources = await window.wailsBindings.ListNativeDatabaseResources(sessionId) || [];
    } catch (error) {
      errorMessage = `加载${workspace.resourceLabel}失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loading = false;
    }
  }

  async function selectResource(resource) {
    if (!resource?.name || !window.wailsBindings || !sessionId) return;
    selectedResource = resource.name;
    loadingDetails = true;
    errorMessage = '';
    try {
      details = await window.wailsBindings.DescribeNativeDatabaseResource(sessionId, '', resource.name);
    } catch (error) {
      details = null;
      errorMessage = `加载资源详情失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loadingDetails = false;
    }
  }

  function formatDetails(content) {
    try { return JSON.stringify(JSON.parse(content || '{}'), null, 2); } catch (_) { return content || '{}'; }
  }
</script>

<section class="native-database-panel" aria-label={workspace.title}>
  <header class="native-database-panel__header native-database-panel__context">
    <div>
      <div class="native-database-panel__eyebrow">{(databaseType || 'kafka').toUpperCase()}</div>
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
      {/if}

      {#if loading}
        <div class="native-database-panel__state">正在加载{workspace.resourceLabel}…</div>
      {:else if resources.length === 0}
        <div class="native-database-panel__state">未发现可显示的{workspace.resourceLabel}。</div>
      {:else}
        <div class="native-database-panel__toolbar native-database-panel__toolbar--resources">
          <div class="native-database-panel__summary native-database-panel__resource-count">
            <span>{workspace.resourceLabel}</span>
            <strong>{resources.length}</strong>
          </div>
        </div>
        <ul class="native-database-panel__tree">
          {#each resources as resource}
            <li>
              <button
                class="native-database-panel__resource native-database-panel__resource--leaf"
                type="button"
                class:selected={selectedResource === resource.name}
                on:click={() => selectResource(resource)}
              >
                <span class="native-database-panel__leaf-mark">•</span>
                <span>{resource.name}</span>
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </main>

    <aside class="native-database-panel__inspector">
      <h4>对象信息</h4>
      <dl>
        <div><dt>类型</dt><dd>{(databaseType || 'kafka').toUpperCase()}</dd></div>
        <div><dt>资源</dt><dd>{workspace.resourceLabel}</dd></div>
        <div><dt>数量</dt><dd>{resources.length}</dd></div>
        {#if selectedResource}<div><dt>当前选中</dt><dd>{selectedResource}</dd></div>{/if}
      </dl>
      <p>{workspace.description}</p>
      <p class="native-database-panel__hint">当前为只读元数据浏览</p>

      {#if loadingDetails}
        <p>正在读取对象详情…</p>
      {:else if details}
        <div class="native-database-panel__details">
          <strong>{details.name || selectedResource}</strong>
          <span>{details.summary}</span>
          <pre>{formatDetails(details.content)}</pre>
        </div>
      {:else}
        <p>选择一个{workspace.resourceLabel}以查看详情。</p>
      {/if}
    </aside>
  </div>
</section>

<style>
  .native-database-panel { height: 100%; overflow: auto; padding: 20px; background: var(--bg-primary); color: var(--text-primary); }
  .native-database-panel__header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; padding-bottom: 16px; border-bottom: 1px solid var(--border-primary); }
  h3 { margin: 3px 0 4px; font-size: 18px; line-height: 1.3; }
  p { max-width: 680px; margin: 0; color: var(--text-secondary); font-size: 12px; line-height: 1.6; }
  .native-database-panel__eyebrow { color: var(--accent-primary); font-size: 11px; font-weight: 750; letter-spacing: 0.08em; }
  .native-database-panel__refresh { flex: 0 0 auto; min-height: 32px; padding: 0 12px; border: 1px solid var(--border-primary); border-radius: 5px; background: var(--bg-secondary); color: var(--text-primary); cursor: pointer; }
  .native-database-panel__state { margin-top: 18px; padding: 18px; border: 1px dashed var(--border-primary); border-radius: 6px; background: var(--bg-secondary); color: var(--text-secondary); font-size: 13px; }
  .native-database-panel__error { border-color: #dc2626; color: #dc2626; }
  .native-database-panel__toolbar { display: grid; gap: 10px; margin-bottom: 12px; }
  .native-database-panel__toolbar--resources { margin-top: 0; }
  .native-database-panel__summary { display: flex; align-items: center; justify-content: space-between; padding: 9px 12px; border: 1px solid var(--border-primary); border-radius: 6px 6px 0 0; background: var(--bg-tertiary); color: var(--text-secondary); font-size: 12px; }
  .native-database-panel__resource-count { margin-top: 0; border-radius: 4px; }
  .native-database-panel__tree { margin: 0; padding: 0; list-style: none; border: 1px solid var(--border-primary); border-top: 0; border-radius: 0 0 6px 6px; overflow: hidden; }
  .native-database-panel__resource { display: flex; align-items: center; width: 100%; min-height: 40px; gap: 9px; padding: 8px 12px; border: 0; background: var(--bg-secondary); color: var(--text-primary); font: inherit; font-size: 13px; text-align: left; cursor: pointer; }
  .native-database-panel__resource.selected, .native-database-panel__resource--leaf.selected { background: #eff6f5; color: #0e6674; }
  .native-database-panel__hint { font-size: 11px; color: #6d7783; margin-top: 8px; }
  .native-database-panel__details pre { max-width: 100%; max-height: 280px; margin: 8px 0 0; overflow: auto; padding: 9px; border: 1px solid #d9e0e4; border-radius: 4px; background: #fff; font: 10px/1.5 ui-monospace, SFMono-Regular, Menlo, monospace; white-space: pre-wrap; overflow-wrap: anywhere; }
  .native-database-panel { padding: 0; display: flex; flex-direction: column; overflow: hidden; background: #f7f8f5; color: #1d2935; font-family: "PingFang SC", "Hiragino Sans GB", -apple-system, sans-serif; }
  .native-database-panel__context { min-height: 64px; padding: 0 20px; align-items: center; background: #fff; border-bottom-color: #d9e0e4; }
  .native-database-panel__body { min-height: 0; flex: 1; display: grid; grid-template-columns: minmax(0, 1fr) 280px; overflow: hidden; }
  .native-database-panel__content { min-width: 0; overflow: auto; padding: 16px; background: #fff; }
  .native-database-panel__inspector { min-width: 0; padding: 18px 16px; border-left: 1px solid #d9e0e4; background: #f7f8f5; overflow: auto; }
  .native-database-panel__inspector h4 { margin: 0 0 14px; color: #31414d; font-size: 11px; letter-spacing: .04em; }
  .native-database-panel__inspector dl { margin: 0; display: grid; gap: 12px; }
  .native-database-panel__inspector dl div { display: grid; gap: 3px; }
  .native-database-panel__inspector dt { color: #7b8791; font-size: 11px; }
  .native-database-panel__inspector dd { margin: 0; color: #31414d; font-size: 12px; overflow-wrap: anywhere; }
  @media (max-width: 760px) {
    .native-database-panel__body { grid-template-columns: 1fr !important; }
    .native-database-panel__inspector { display: none; }
  }
</style>
