<script>
  import { onMount } from 'svelte';
  import ConfirmDialog from '../ui/ConfirmDialog.svelte';
  import RedisKeyEditor from '../RedisKeyEditor.svelte';
  import { nativeDatabaseWorkspace } from '../../lib/nativeDatabaseWorkspace.js';
  import {
    NATIVE_DB_OPERATIONS,
    buildRedisSavePayload,
    createRedisEditorState,
    formatMutationMessage,
    redisDatabaseOptions
  } from '../../lib/nativeDatabaseOperations.js';

  export let sessionId = null;
  export let dbConfig = null;

  let redisDatabases = [];
  let redisKeys = [];
  let selectedRedisDb = '';
  let loadingRedisKeys = false;
  let loading = false;
  let loadingDetails = false;
  let saving = false;
  let errorMessage = '';
  let actionMessage = '';
  let selectedResource = null;
  let selectedParent = '';
  let details = null;
  let redisEditor = null;
  let redisEditTTL = '';
  let showDeleteConfirm = false;

  $: databaseType = dbConfig?.metadata?.db_type || 'redis';
  $: workspace = nativeDatabaseWorkspace(databaseType || 'redis');
  $: redisDbOptions = redisDatabaseOptions(redisDatabases);

  onMount(loadResources);

  async function loadResources() {
    if (!window.wailsBindings || !sessionId) return;
    loading = true;
    errorMessage = '';
    try {
      redisDatabases = await window.wailsBindings.ListNativeDatabaseResources(sessionId) || [];
      const configuredDb = String(dbConfig?.metadata?.database ?? '').trim();
      selectedRedisDb = redisDatabases.some((item) => item.name === configuredDb)
        ? configuredDb
        : (redisDatabases[0]?.name || '0');
      await loadRedisKeys();
    } catch (error) {
      errorMessage = `加载${workspace.resourceLabel}失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loading = false;
    }
  }

  async function loadRedisKeys() {
    if (!selectedRedisDb || !window.wailsBindings) return;
    loadingRedisKeys = true;
    try {
      redisKeys = await window.wailsBindings.ListNativeDatabaseChildResources(sessionId, selectedRedisDb) || [];
    } catch (error) {
      errorMessage = `加载${workspace.childLabel}失败: ${error?.message || error || '未知错误'}`;
      redisKeys = [];
    } finally {
      loadingRedisKeys = false;
    }
  }

  async function handleRedisDbChange(event) {
    selectedRedisDb = event.target.value;
    selectedResource = null;
    selectedParent = '';
    details = null;
    redisEditor = null;
    actionMessage = '';
    await loadRedisKeys();
  }

  async function selectResource(resource) {
    selectedParent = selectedRedisDb;
    if (!resource?.name || !window.wailsBindings || !sessionId) return;
    selectedResource = resource.name;
    loadingDetails = true;
    errorMessage = '';
    actionMessage = '';
    try {
      details = await window.wailsBindings.DescribeNativeDatabaseResource(sessionId, selectedParent, resource.name);
      redisEditor = createRedisEditorState(details.content);
      redisEditTTL = redisEditor.ttlSeconds > 0 ? String(redisEditor.ttlSeconds) : '';
    } catch (error) {
      details = null;
      redisEditor = null;
      errorMessage = `加载资源详情失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loadingDetails = false;
    }
  }

  async function mutateResource(operation, payload) {
    if (!selectedResource || !window.wailsBindings?.MutateNativeDatabaseResource) return;
    saving = true;
    errorMessage = '';
    actionMessage = '';
    try {
      const result = await window.wailsBindings.MutateNativeDatabaseResource(
        sessionId,
        selectedParent,
        selectedResource,
        operation,
        payload
      );
      actionMessage = formatMutationMessage(result);
      await refreshSelectedResource();
      await loadRedisKeys();
    } catch (error) {
      errorMessage = `操作失败: ${error?.message || error || '未知错误'}`;
    } finally {
      saving = false;
    }
  }

  async function refreshSelectedResource() {
    if (!selectedResource) return;
    details = await window.wailsBindings.DescribeNativeDatabaseResource(sessionId, selectedParent, selectedResource);
    redisEditor = createRedisEditorState(details.content);
    redisEditTTL = redisEditor.ttlSeconds > 0 ? String(redisEditor.ttlSeconds) : '';
  }

  function saveRedisKey() {
    if (!redisEditor) return;
    mutateResource(
      NATIVE_DB_OPERATIONS.REDIS_SAVE,
      buildRedisSavePayload(redisEditor, redisEditTTL)
    );
  }

  function requestDeleteCurrentResource() {
    showDeleteConfirm = true;
  }

  async function confirmDelete() {
    showDeleteConfirm = false;
    await mutateResource(NATIVE_DB_OPERATIONS.REDIS_DELETE, '{}');
    selectedResource = null;
    details = null;
    redisEditor = null;
    await loadRedisKeys();
  }
</script>

<section class="native-database-panel" aria-label={workspace.title}>
  <header class="native-database-panel__header native-database-panel__context">
    <div>
      <div class="native-database-panel__eyebrow">{(databaseType || 'redis').toUpperCase()}</div>
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
      {#if actionMessage}
        <div class="native-database-panel__state native-database-panel__success" role="status">{actionMessage}</div>
      {/if}

      {#if loading}
        <div class="native-database-panel__state">正在加载{workspace.resourceLabel}…</div>
      {:else}
        <div class="native-database-panel__toolbar">
          <label class="native-database-panel__db-select">
            <span>逻辑库</span>
            <select value={selectedRedisDb} on:change={handleRedisDbChange}>
              {#each redisDbOptions as option}
                <option value={option.value}>{option.label}</option>
              {/each}
            </select>
          </label>
          <div class="native-database-panel__summary native-database-panel__resource-count">
            <span>{workspace.childLabel}</span>
            <strong>{redisKeys.length}</strong>
          </div>
        </div>
        {#if loadingRedisKeys}
          <div class="native-database-panel__state">正在加载键…</div>
        {:else if redisKeys.length === 0}
          <div class="native-database-panel__state">当前逻辑库中没有可显示的键。</div>
        {:else}
          <ul class="native-database-panel__tree native-database-panel__tree--flat">
            {#each redisKeys as key}
              <li>
                <button
                  class="native-database-panel__resource native-database-panel__resource--leaf"
                  type="button"
                  class:selected={selectedResource === key.name}
                  on:click={() => selectResource(key)}
                >
                  <span class="native-database-panel__leaf-mark">•</span>
                  <span>{key.name}</span>
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      {/if}
    </main>

    <aside class="native-database-panel__inspector">
      <h4>对象信息</h4>
      <dl>
        <div><dt>类型</dt><dd>{(databaseType || 'redis').toUpperCase()}</dd></div>
        <div><dt>资源</dt><dd>{workspace.childLabel}</dd></div>
        <div><dt>数量</dt><dd>{redisKeys.length}</dd></div>
        <div><dt>逻辑库</dt><dd>DB {selectedRedisDb}</dd></div>
        {#if selectedResource}<div><dt>当前选中</dt><dd>{selectedResource}</dd></div>{/if}
      </dl>
      <p>{workspace.description}</p>

      {#if loadingDetails}
        <p>正在读取对象详情…</p>
      {:else if details && redisEditor}
        <div class="native-database-panel__details">
          <strong>{details.name || selectedResource}</strong>
          <span>{details.summary}</span>
          <RedisKeyEditor
            bind:state={redisEditor}
            bind:ttlInput={redisEditTTL}
            {saving}
            onSave={saveRedisKey}
            onDelete={requestDeleteCurrentResource}
          />
        </div>
      {:else}
        <p>选择一个{workspace.childLabel}以查看详情或执行操作。</p>
      {/if}
    </aside>
  </div>
</section>

<ConfirmDialog
  isOpen={showDeleteConfirm}
  title="确认删除"
  message={`确定删除键 ${selectedResource} 吗？`}
  confirmText="删除"
  type="danger"
  onConfirm={confirmDelete}
  onCancel={() => { showDeleteConfirm = false; }}
/>

<style>
  .native-database-panel { height: 100%; overflow: auto; padding: 20px; background: var(--bg-primary); color: var(--text-primary); }
  .native-database-panel__header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; padding-bottom: 16px; border-bottom: 1px solid var(--border-primary); }
  h3 { margin: 3px 0 4px; font-size: 18px; line-height: 1.3; }
  p { max-width: 680px; margin: 0; color: var(--text-secondary); font-size: 12px; line-height: 1.6; }
  .native-database-panel__eyebrow { color: var(--accent-primary); font-size: 11px; font-weight: 750; letter-spacing: 0.08em; }
  .native-database-panel__refresh { flex: 0 0 auto; min-height: 32px; padding: 0 12px; border: 1px solid var(--border-primary); border-radius: 5px; background: var(--bg-secondary); color: var(--text-primary); cursor: pointer; }
  .native-database-panel__state { margin-top: 18px; padding: 18px; border: 1px dashed var(--border-primary); border-radius: 6px; background: var(--bg-secondary); color: var(--text-secondary); font-size: 13px; }
  .native-database-panel__error { border-color: #dc2626; color: #dc2626; }
  .native-database-panel__success { border-color: #059669; color: #047857; }
  .native-database-panel__toolbar { display: grid; gap: 10px; margin-bottom: 12px; }
  .native-database-panel__db-select { display: grid; gap: 6px; font-size: 11px; color: #6d7783; }
  .native-database-panel__db-select select {
    width: 100%; min-height: 34px; border: 1px solid #d9e0e4; border-radius: 4px; padding: 0 10px;
    background: #fff url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='%236d7783' stroke-width='2'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E") no-repeat right 0.65rem center / 0.9rem;
    appearance: none; color: #31414d; font-size: 12px;
  }
  .native-database-panel__tree--flat { margin-top: 0; border-radius: 4px; }
  .native-database-panel__summary { display: flex; align-items: center; justify-content: space-between; padding: 9px 12px; border: 1px solid var(--border-primary); border-radius: 6px 6px 0 0; background: var(--bg-tertiary); color: var(--text-secondary); font-size: 12px; }
  .native-database-panel__resource-count { margin-top: 0; border-radius: 4px; }
  .native-database-panel__tree { margin: 0; padding: 0; list-style: none; border: 1px solid var(--border-primary); border-top: 0; border-radius: 0 0 6px 6px; overflow: hidden; }
  .native-database-panel__resource { display: flex; align-items: center; width: 100%; min-height: 40px; gap: 9px; padding: 8px 12px; border: 0; background: var(--bg-secondary); color: var(--text-primary); font: inherit; font-size: 13px; text-align: left; cursor: pointer; }
  .native-database-panel__resource.selected, .native-database-panel__resource--leaf.selected { background: #eff6f5; color: #0e6674; }
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
  .native-database-panel__details { display: grid; gap: 8px; margin-top: 12px; }
  @media (max-width: 760px) {
    .native-database-panel__body { grid-template-columns: 1fr !important; }
    .native-database-panel__inspector { display: none; }
  }
</style>
