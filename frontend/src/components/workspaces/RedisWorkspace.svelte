<script>
  import { onMount } from 'svelte';
  import ConfirmDialog from '../ui/ConfirmDialog.svelte';
  import RedisKeyEditor from '../RedisKeyEditor.svelte';
  import { nativeDatabaseWorkspace } from '../../lib/nativeDatabaseWorkspace.js';
  import {
    NATIVE_DB_OPERATIONS,
    buildRedisCLIQuery,
    buildRedisDeleteKeysPayload,
    buildRedisSavePayload,
    canSaveRedisEditor,
    createRedisEditorState,
    formatMutationMessage,
    parseNativeMutationArtifact,
    parseNativeResourceContent,
    parseNativeResourcePage,
    redisDatabaseOptions
  } from '../../lib/nativeDatabaseOperations.js';
  import { copilotStore } from '../../stores/copilot.js';
  import { COPILOT_APPLY_NATIVE, COPILOT_EXECUTE_NATIVE } from '../../lib/copilotApply.js';
  import { applyNativeArtifactToRedis, normalizeRedisCLICommand } from '../../lib/nativeCopilotApply.js';

  export let sessionId = null;
  export let dbConfig = null;

  let redisDatabases = [];
  let redisKeys = [];
  let selectedRedisDb = '';
  let keyPattern = '*';
  let scanCursor = '0';
  let hasMoreKeys = false;
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
  let showBatchDeleteConfirm = false;
  let activeTab = 'key';
  let cliCommand = 'PING';
  let cliResult = '';
  let selectedKeySet = new Set();
  let showCreateDialog = false;
  let createKeyName = '';
  let createKeyType = 'string';
  let createKeyValue = '';
  let createKeyTTL = '';

  $: databaseType = dbConfig?.metadata?.db_type || 'redis';
  $: workspace = nativeDatabaseWorkspace(databaseType || 'redis');
  $: redisDbOptions = redisDatabaseOptions(redisDatabases);
  $: selectedCount = selectedKeySet.size;
  $: if (sessionId) {
    copilotStore.setWorkspaceFocus(sessionId, {
      database: selectedRedisDb || '',
      objectKind: selectedResource ? 'key' : (selectedRedisDb ? 'database' : ''),
      objectName: selectedResource || selectedRedisDb || '',
      objectParent: selectedRedisDb || '',
      editorContent: activeTab === 'cli' ? cliCommand : (redisEditor ? JSON.stringify(redisEditor) : ''),
      pattern: keyPattern
    });
  }

  onMount(() => {
    loadResources();
    const handleApplyNative = (event) => {
      if (!event?.detail || event.detail.sessionId !== sessionId) return;
      const next = applyNativeArtifactToRedis(event.detail.artifact, {
        selectedRedisDb,
        selectedResource,
        activeTab,
        cliCommand,
        keyPattern
      });
      selectedRedisDb = next.selectedRedisDb ?? selectedRedisDb;
      selectedResource = next.selectedResource ?? selectedResource;
      activeTab = next.activeTab || activeTab;
      cliCommand = next.cliCommand ?? cliCommand;
      if (next.keyPattern != null) {
        keyPattern = next.keyPattern;
      }
      actionMessage = next.actionMessage || '';
      errorMessage = next.errorMessage || '';
      if (next.shouldReloadKeys) {
        reloadRedisKeys();
      }
      return next;
    };

    const handleExecuteNative = async (event) => {
      if (!event?.detail || event.detail.sessionId !== sessionId) return;
      const next = handleApplyNative(event);
      const artifact = event.detail.artifact;
      if (!artifact) return;
      if (String(artifact.type || '').toLowerCase() === 'native_mutation') {
        const mutation = parseNativeMutationArtifact(artifact.content);
        await mutateResource(
          mutation.operation,
          mutation.name || selectedResource,
          mutation.payload || '{}',
          mutation.parent || selectedRedisDb
        );
        return;
      }
      if (next?.applyTarget === 'match' || next?.shouldReloadKeys) {
        // MATCH 填入时 handleApplyNative 已触发扫描
        return;
      }
      activeTab = 'cli';
      await runCLI({ readOnly: true });
    };

    window.addEventListener(COPILOT_APPLY_NATIVE, handleApplyNative);
    window.addEventListener(COPILOT_EXECUTE_NATIVE, handleExecuteNative);
    return () => {
      window.removeEventListener(COPILOT_APPLY_NATIVE, handleApplyNative);
      window.removeEventListener(COPILOT_EXECUTE_NATIVE, handleExecuteNative);
    };
  });

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
      await reloadRedisKeys();
    } catch (error) {
      errorMessage = `加载失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loading = false;
    }
  }

  async function reloadRedisKeys() {
    scanCursor = '0';
    redisKeys = [];
    selectedKeySet = new Set();
    hasMoreKeys = false;
    await loadRedisKeysPage(false);
  }

  async function loadRedisKeysPage(append) {
    if (!selectedRedisDb || !window.wailsBindings) return;
    loadingRedisKeys = true;
    try {
      const pattern = String(keyPattern || '*').trim() || '*';
      const cursor = append ? scanCursor : '0';
      let page;
      if (typeof window.wailsBindings.ListNativeDatabaseChildResourcesPage === 'function') {
        page = parseNativeResourcePage(
          await window.wailsBindings.ListNativeDatabaseChildResourcesPage(
            sessionId, selectedRedisDb, pattern, cursor, 200
          )
        );
      } else {
        const items = await window.wailsBindings.ListNativeDatabaseChildResources(sessionId, selectedRedisDb) || [];
        page = { items, nextCursor: '0', hasMore: false };
      }
      redisKeys = append ? [...redisKeys, ...page.items] : page.items;
      scanCursor = page.nextCursor || '0';
      hasMoreKeys = Boolean(page.hasMore);
    } catch (error) {
      errorMessage = `加载键失败: ${error?.message || error || '未知错误'}`;
      if (!append) redisKeys = [];
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
    await reloadRedisKeys();
  }

  async function selectResource(resource) {
    selectedParent = selectedRedisDb;
    if (!resource?.name || !window.wailsBindings || !sessionId) return;
    selectedResource = resource.name;
    activeTab = 'key';
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

  function toggleKeySelection(name, checked) {
    const next = new Set(selectedKeySet);
    if (checked) next.add(name);
    else next.delete(name);
    selectedKeySet = next;
  }

  async function mutateResource(operation, name, payload, parent = selectedRedisDb) {
    if (!window.wailsBindings?.MutateNativeDatabaseResource) return;
    saving = true;
    errorMessage = '';
    actionMessage = '';
    try {
      const result = await window.wailsBindings.MutateNativeDatabaseResource(
        sessionId,
        parent,
        name || '',
        operation,
        payload
      );
      actionMessage = formatMutationMessage(result);
      await reloadRedisKeys();
      if (selectedResource) await refreshSelectedResource();
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
    if (!redisEditor || !canSaveRedisEditor(redisEditor)) return;
    mutateResource(
      NATIVE_DB_OPERATIONS.REDIS_SAVE,
      selectedResource,
      buildRedisSavePayload(redisEditor, redisEditTTL),
      selectedParent
    );
  }

  function requestDeleteCurrentResource() {
    showDeleteConfirm = true;
  }

  async function confirmDelete() {
    showDeleteConfirm = false;
    await mutateResource(NATIVE_DB_OPERATIONS.REDIS_DELETE, selectedResource, '{}', selectedParent);
    selectedResource = null;
    details = null;
    redisEditor = null;
  }

  async function confirmBatchDelete() {
    showBatchDeleteConfirm = false;
    const keys = [...selectedKeySet];
    await mutateResource(
      NATIVE_DB_OPERATIONS.REDIS_DELETE_KEYS,
      '',
      buildRedisDeleteKeysPayload(keys),
      selectedRedisDb
    );
    selectedKeySet = new Set();
  }

  async function createKey() {
    const name = String(createKeyName || '').trim();
    if (!name) {
      errorMessage = '请输入键名';
      return;
    }
    const state = {
      type: createKeyType,
      value: createKeyValue,
      fields: createKeyType === 'hash' ? [{ field: 'field', value: createKeyValue || '' }] : [],
      items: createKeyType === 'list' ? [createKeyValue || ''] : [],
      members: createKeyType === 'set' ? [createKeyValue || 'item'] : [],
      entries: createKeyType === 'zset' ? [{ member: createKeyValue || 'member', score: 0 }] : []
    };
    showCreateDialog = false;
    await mutateResource(
      NATIVE_DB_OPERATIONS.REDIS_CREATE_KEY,
      name,
      buildRedisSavePayload(state, createKeyTTL),
      selectedRedisDb
    );
    createKeyName = '';
    createKeyValue = '';
    createKeyTTL = '';
  }

  async function runCLI({ readOnly = false } = {}) {
    if (!window.wailsBindings?.ExecuteNativeDatabaseQuery) return;
    saving = true;
    errorMessage = '';
    try {
      cliCommand = normalizeRedisCLICommand(cliCommand);
      const result = await window.wailsBindings.ExecuteNativeDatabaseQuery(
        sessionId,
        selectedRedisDb,
        '',
        buildRedisCLIQuery(cliCommand, { readOnly })
      );
      actionMessage = formatMutationMessage(result);
      const parsed = parseNativeResourceContent(result.content);
      cliResult = JSON.stringify(parsed, null, 2);
    } catch (error) {
      errorMessage = `CLI 失败: ${error?.message || error || '未知错误'}`;
    } finally {
      saving = false;
    }
  }
</script>

<section class="native-database-panel" aria-label={workspace.title}>
  <header class="native-database-panel__header native-database-panel__context">
    <div>
      <div class="native-database-panel__eyebrow">缓存 · {(databaseType || 'redis').toUpperCase()}</div>
      <h3>{workspace.title}</h3>
      <p>{dbConfig?.name || 'Redis 连接'} · {dbConfig?.host}:{dbConfig?.port}</p>
    </div>
    <div class="native-database-panel__header-actions">
      <button class="native-database-panel__refresh" type="button" on:click={() => { showCreateDialog = true; }}>新建键</button>
      <button class="native-database-panel__refresh" type="button" on:click={loadResources} disabled={loading}>
        {loading ? '加载中…' : '刷新'}
      </button>
    </div>
  </header>

  <div class="native-database-panel__toolbar native-database-panel__toolbar--top">
    <label>
      <span>逻辑库</span>
      <select value={selectedRedisDb} on:change={handleRedisDbChange}>
        {#each redisDbOptions as option}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>
    </label>
    <label class="native-database-panel__grow">
      <span>MATCH</span>
      <input
        bind:value={keyPattern}
        placeholder="user:*"
        autocomplete="off"
        autocapitalize="off"
        autocorrect="off"
        spellcheck="false"
        on:keydown={(e) => e.key === 'Enter' && reloadRedisKeys()}
      />
    </label>
    <button type="button" on:click={reloadRedisKeys} disabled={loadingRedisKeys}>扫描</button>
    {#if selectedCount}
      <button type="button" class="danger" on:click={() => { showBatchDeleteConfirm = true; }}>删除选中 ({selectedCount})</button>
    {/if}
  </div>

  <div class="native-database-panel__tabs">
    <button type="button" class:active={activeTab === 'key'} on:click={() => activeTab = 'key'}>键详情</button>
    <button type="button" class:active={activeTab === 'cli'} on:click={() => activeTab = 'cli'}>CLI</button>
  </div>

  <div class="native-database-panel__body">
    <aside class="native-database-panel__list">
      {#if errorMessage}
        <div class="native-database-panel__state native-database-panel__error" role="alert">{errorMessage}</div>
      {/if}
      {#if actionMessage}
        <div class="native-database-panel__state native-database-panel__success" role="status">{actionMessage}</div>
      {/if}
      {#if loadingRedisKeys && redisKeys.length === 0}
        <div class="native-database-panel__state">正在扫描键…</div>
      {:else if redisKeys.length === 0}
        <div class="native-database-panel__state">没有匹配的键。</div>
      {:else}
        <ul class="native-database-panel__tree native-database-panel__tree--flat">
          {#each redisKeys as key}
            <li>
              <label class="native-database-panel__key-row">
                <input
                  type="checkbox"
                  checked={selectedKeySet.has(key.name)}
                  on:change={(e) => toggleKeySelection(key.name, e.currentTarget.checked)}
                />
                <button
                  class="native-database-panel__resource native-database-panel__resource--leaf"
                  type="button"
                  class:selected={selectedResource === key.name}
                  on:click={() => selectResource(key)}
                >
                  <span>{key.name}</span>
                </button>
              </label>
            </li>
          {/each}
        </ul>
        {#if hasMoreKeys}
          <button type="button" class="native-database-panel__more" on:click={() => loadRedisKeysPage(true)} disabled={loadingRedisKeys}>
            {loadingRedisKeys ? '加载中…' : '加载更多'}
          </button>
        {/if}
      {/if}
    </aside>

    <main class="native-database-panel__content">
      {#if activeTab === 'cli'}
        <div class="redis-cli">
          <p class="redis-cli__hint">命令（危险命令会被拒绝；写命令请确认后执行）。示例：<code>GET key</code>、<code>SCAN 0 MATCH mini* COUNT 100</code></p>
          <div class="redis-cli__composer">
            <textarea
              class="redis-cli__input"
              bind:value={cliCommand}
              rows="2"
              autocomplete="off"
              autocapitalize="off"
              autocorrect="off"
              spellcheck="false"
              placeholder="在此输入 Redis 命令…"
              on:keydown={(e) => {
                if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
                  e.preventDefault();
                  runCLI({ readOnly: false });
                }
              }}
            ></textarea>
            <div class="native-database-panel__editor-actions">
              <button type="button" on:click={() => runCLI({ readOnly: true })} disabled={saving}>只读执行</button>
              <button type="button" on:click={() => runCLI({ readOnly: false })} disabled={saving}>执行</button>
            </div>
          </div>
          {#if cliResult}
            <pre class="redis-cli__result">{cliResult}</pre>
          {:else}
            <div class="redis-cli__result redis-cli__result--empty">执行结果会显示在这里</div>
          {/if}
        </div>
      {:else if loadingDetails}
        <div class="native-database-panel__state">正在读取键…</div>
      {:else if details && redisEditor}
        <div class="native-database-panel__details">
          <strong>{details.name || selectedResource}</strong>
          <span>{details.summary}</span>
          {#if redisEditor.truncated}
            <p class="native-database-panel__warn">预览已截断，已禁用保存，避免用预览覆盖完整值。</p>
          {/if}
          <RedisKeyEditor
            bind:state={redisEditor}
            bind:ttlInput={redisEditTTL}
            {saving}
            saveDisabled={!canSaveRedisEditor(redisEditor)}
            onSave={saveRedisKey}
            onDelete={requestDeleteCurrentResource}
          />
        </div>
      {:else}
        <div class="native-database-panel__state">选择左侧键查看详情，或打开 CLI。</div>
      {/if}
    </main>
  </div>
</section>

{#if showCreateDialog}
  <div class="native-database-panel__modal" role="dialog">
    <div class="native-database-panel__modal-card">
      <h4>新建键</h4>
      <label><span>键名</span><input bind:value={createKeyName} /></label>
      <label>
        <span>类型</span>
        <select bind:value={createKeyType}>
          <option value="string">string</option>
          <option value="hash">hash</option>
          <option value="list">list</option>
          <option value="set">set</option>
          <option value="zset">zset</option>
        </select>
      </label>
      <label><span>初始值</span><input bind:value={createKeyValue} /></label>
      <label><span>TTL 秒（可选）</span><input bind:value={createKeyTTL} /></label>
      <div class="native-database-panel__editor-actions">
        <button type="button" on:click={() => { showCreateDialog = false; }}>取消</button>
        <button type="button" on:click={createKey} disabled={saving}>创建</button>
      </div>
    </div>
  </div>
{/if}

<ConfirmDialog
  isOpen={showDeleteConfirm}
  title="确认删除"
  message={`确定删除键 ${selectedResource} 吗？`}
  confirmText="删除"
  type="danger"
  onConfirm={confirmDelete}
  onCancel={() => { showDeleteConfirm = false; }}
/>

<ConfirmDialog
  isOpen={showBatchDeleteConfirm}
  title="批量删除"
  message={`确定删除选中的 ${selectedCount} 个键吗？`}
  confirmText="删除"
  type="danger"
  onConfirm={confirmBatchDelete}
  onCancel={() => { showBatchDeleteConfirm = false; }}
/>

<style>
  .native-database-panel { height: 100%; display: flex; flex-direction: column; overflow: hidden; background: #f7f8f5; color: #1d2935; }
  .native-database-panel__header { display: flex; align-items: center; justify-content: space-between; gap: 16px; min-height: 64px; padding: 0 20px; background: #fff; border-bottom: 1px solid #d9e0e4; }
  .native-database-panel__header-actions { display: flex; gap: 8px; }
  h3 { margin: 3px 0 4px; font-size: 18px; }
  p { margin: 0; color: #6d7783; font-size: 12px; }
  .native-database-panel__eyebrow { color: var(--accent-primary); font-size: 11px; font-weight: 750; letter-spacing: 0.08em; }
  .native-database-panel__refresh, .native-database-panel__toolbar button, .native-database-panel__editor-actions button, .native-database-panel__more {
    min-height: 32px; padding: 0 12px; border: 1px solid #d9e0e4; border-radius: 5px; background: #fff; cursor: pointer;
  }
  .danger { color: #b91c1c; border-color: #fecaca; }
  .native-database-panel__toolbar--top { display: flex; gap: 10px; align-items: end; padding: 10px 16px; background: #fff; border-bottom: 1px solid #d9e0e4; }
  .native-database-panel__toolbar label { display: grid; gap: 4px; font-size: 11px; color: #6d7783; }
  .native-database-panel__grow { flex: 1; }
  .native-database-panel__toolbar select, .native-database-panel__toolbar input, .native-database-panel__modal input, .native-database-panel__modal select {
    min-height: 34px; border: 1px solid #d9e0e4; border-radius: 4px; padding: 0 10px; background: #fff;
  }
  .native-database-panel__tabs { display: flex; gap: 4px; padding: 8px 16px 0; background: #fff; }
  .native-database-panel__tabs button { border: 0; background: transparent; padding: 8px 12px; cursor: pointer; color: #6d7783; }
  .native-database-panel__tabs button.active { color: #0e6674; border-bottom: 2px solid #0e6674; }
  .native-database-panel__body { min-height: 0; flex: 1; display: grid; grid-template-columns: minmax(220px, 320px) minmax(0, 1fr); overflow: hidden; }
  .native-database-panel__list, .native-database-panel__content { overflow: auto; padding: 12px 16px; background: #fff; min-width: 0; }
  .native-database-panel__list { border-right: 1px solid #d9e0e4; }
  .native-database-panel__tree { margin: 0; padding: 0; list-style: none; }
  .native-database-panel__key-row { display: flex; align-items: center; gap: 6px; }
  .native-database-panel__resource { flex: 1; display: flex; width: 100%; min-height: 36px; border: 0; background: transparent; text-align: left; cursor: pointer; padding: 6px 8px; border-radius: 4px; }
  .native-database-panel__resource.selected { background: #eff6f5; color: #0e6674; }
  .native-database-panel__state { margin: 8px 0; padding: 12px; border: 1px dashed #d9e0e4; border-radius: 6px; color: #6d7783; font-size: 13px; }
  .native-database-panel__error { border-color: #dc2626; color: #dc2626; }
  .native-database-panel__success { border-color: #059669; color: #047857; }
  .native-database-panel__warn { color: #b45309; font-size: 12px; }
  .native-database-panel__details { display: grid; gap: 10px; }
  .native-database-panel__editor-actions { display: flex; gap: 8px; flex-wrap: wrap; }
  .native-database-panel__more { width: 100%; margin-top: 8px; }
  .redis-cli {
    display: flex;
    flex-direction: column;
    gap: 10px;
    width: 100%;
    min-width: 0;
    max-width: 960px;
    height: 100%;
    min-height: 0;
    box-sizing: border-box;
  }
  .redis-cli__hint {
    margin: 0;
    color: #6d7783;
    font-size: 12px;
    line-height: 1.5;
  }
  .redis-cli__hint code {
    font-size: 11px;
    padding: 1px 4px;
    border-radius: 3px;
    background: #f0f3f4;
  }
  .redis-cli__composer {
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex: 0 0 auto;
  }
  .redis-cli__input {
    display: block;
    width: 100%;
    min-width: 0;
    min-height: 56px;
    max-height: 140px;
    box-sizing: border-box;
    resize: vertical;
    padding: 10px 12px;
    border: 1px solid #d9e0e4;
    border-radius: 6px;
    background: #fff;
    color: #1d2935;
    font: 13px/1.5 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }
  .redis-cli__result {
    margin: 0;
    flex: 1 1 auto;
    min-height: 180px;
    padding: 12px;
    background: #f7f8f5;
    border: 1px solid #e6ebef;
    border-radius: 6px;
    overflow: auto;
    font-size: 12px;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .redis-cli__result--empty {
    color: #6d7783;
    display: grid;
    place-items: center;
  }
  pre { margin: 0; padding: 12px; background: #f7f8f5; border-radius: 6px; overflow: auto; font-size: 12px; }
  .native-database-panel__modal { position: fixed; inset: 0; background: rgba(0,0,0,.35); display: grid; place-items: center; z-index: 80; }
  .native-database-panel__modal-card { width: min(420px, 92vw); background: #fff; border-radius: 10px; padding: 16px; display: grid; gap: 10px; }
  .native-database-panel__modal-card label { display: grid; gap: 4px; font-size: 12px; }
</style>
