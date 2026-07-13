<script>
  import { onMount } from 'svelte';
  import { databaseTypeConfig } from '../lib/nativeDatabaseTypes.js';

  export let sessionId = null;
  export let dbConfig = null;

  let resources = [];
  let childrenByParent = {};
  let expanded = new Set();
  let loading = false;
  let errorMessage = '';

  $: databaseType = dbConfig?.metadata?.db_type || '';
  $: childLabel = databaseTypeConfig(databaseType)?.resourceLabel || '资源';

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
      errorMessage = `加载资源失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loading = false;
    }
  }

  async function toggleResource(resource) {
    const name = resource?.name;
    if (!name || resource.kind === 'index') return;
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
      errorMessage = `加载${childLabel}失败: ${error?.message || error || '未知错误'}`;
    }
  }
</script>

<div class="native-database-panel">
  <header>
    <div>
      <div class="native-database-panel__eyebrow">{databaseType.toUpperCase()}</div>
      <h3>{dbConfig?.name || '原生数据库'}</h3>
    </div>
    <button type="button" on:click={loadResources} disabled={loading}>刷新</button>
  </header>

  {#if errorMessage}
    <div class="native-database-panel__error">{errorMessage}</div>
  {:else if loading}
    <div class="native-database-panel__empty">正在加载资源...</div>
  {:else if resources.length === 0}
    <div class="native-database-panel__empty">没有可显示的资源。</div>
  {:else}
    <ul class="native-database-panel__tree">
      {#each resources as resource}
        <li>
          <button type="button" on:click={() => toggleResource(resource)}>{resource.name}</button>
          {#if expanded.has(resource.name)}
            <ul>
              {#each childrenByParent[resource.name] || [] as child}
                <li>{child.name}</li>
              {/each}
            </ul>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .native-database-panel { height: 100%; padding: 16px; overflow: auto; color: var(--text-primary); }
  header { display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid var(--border-color); padding-bottom: 12px; }
  h3 { margin: 2px 0 0; font-size: 16px; }
  button { border: 1px solid var(--border-color); border-radius: 6px; background: var(--bg-secondary); color: inherit; padding: 6px 10px; }
  .native-database-panel__eyebrow { color: var(--text-secondary); font-size: 11px; font-weight: 700; }
  .native-database-panel__tree { margin: 14px 0 0; padding: 0; list-style: none; }
  .native-database-panel__tree ul { padding-left: 20px; list-style: none; }
  .native-database-panel__tree li { margin: 6px 0; }
  .native-database-panel__empty, .native-database-panel__error { margin-top: 16px; color: var(--text-secondary); }
  .native-database-panel__error { color: #dc2626; }
</style>
