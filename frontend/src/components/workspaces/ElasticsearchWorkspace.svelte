<script>
  import { onMount } from 'svelte';
  import ConfirmDialog from '../ui/ConfirmDialog.svelte';
  import { nativeDatabaseWorkspace } from '../../lib/nativeDatabaseWorkspace.js';
  import { copilotStore } from '../../stores/copilot.js';
  import {
    NATIVE_DB_OPERATIONS,
    buildElasticsearchDevToolsQuery,
    buildElasticsearchDocumentPayload,
    buildElasticsearchPagedQuery,
    defaultElasticsearchQuery,
    filterNativeResources,
    formatMutationMessage,
    parseElasticsearchClusterOverview,
    parseElasticsearchIndexMetadata,
    parseElasticsearchQueryHits,
    parseNativeMutationArtifact,
    parseNativeResourceContent
  } from '../../lib/nativeDatabaseOperations.js';
  import { COPILOT_APPLY_NATIVE, COPILOT_EXECUTE_NATIVE } from '../../lib/copilotApply.js';
  import { applyNativeArtifactToElasticsearch } from '../../lib/nativeCopilotApply.js';

  export let sessionId = null;
  export let dbConfig = null;

  let resources = [];
  let loading = false;
  let loadingDetails = false;
  let saving = false;
  let errorMessage = '';
  let actionMessage = '';
  let selectedResource = null;
  let selectedParent = '';
  let details = null;
  let activeTab = 'discover';
  let queryText = defaultElasticsearchQuery();
  let queryFrom = 0;
  let querySize = 20;
  let queryResult = null;
  let esDocumentId = '';
  let esDocumentBody = '{\n  \n}';
  let showDeleteConfirm = false;
  let pendingDeleteTarget = null;
  let resourceSearch = '';
  let sessionOverview = null;
  let inspectorWidth = 320;
  let resizingInspector = false;
  let devMethod = 'GET';
  let devPath = '/_cluster/health';
  let devBody = '';
  let devResult = '';
  let showCreateIndex = false;
  let newIndexName = '';
  let newIndexBody = '{\n  "settings": { "number_of_shards": 1, "number_of_replicas": 0 }\n}';

  $: databaseType = dbConfig?.metadata?.db_type || 'elasticsearch';
  $: workspace = nativeDatabaseWorkspace(databaseType || 'elasticsearch');
  $: queryHits = queryResult ? parseElasticsearchQueryHits(queryResult.content) : [];
  $: filteredResources = filterNativeResources(resources, resourceSearch);
  $: clusterOverview = sessionOverview ? parseElasticsearchClusterOverview(sessionOverview.content) : null;
  $: indexMetadata = details ? parseElasticsearchIndexMetadata(details.content) : null;
  $: bodyStyle = `grid-template-columns: minmax(220px, 280px) minmax(0, 1fr) 6px ${inspectorWidth}px;`;
  $: if (sessionId) {
    copilotStore.setWorkspaceFocus(sessionId, {
      objectKind: selectedResource ? 'index' : '',
      objectName: selectedResource || '',
      objectParent: selectedParent || '',
      editorContent: activeTab === 'devtools' ? buildElasticsearchDevToolsQuery(devMethod, devPath, devBody) : queryText
    });
  }

  onMount(() => {
    loadResources();
    const handleApplyNative = (event) => {
      if (!event?.detail || event.detail.sessionId !== sessionId) return;
      const next = applyNativeArtifactToElasticsearch(event.detail.artifact, {
        selectedResource,
        activeTab,
        queryText,
        esDocumentId,
        esDocumentBody,
        showCreateIndex,
        newIndexName,
        newIndexBody,
        devMethod,
        devPath,
        devBody
      });
      selectedResource = next.selectedResource ?? selectedResource;
      activeTab = next.activeTab || activeTab;
      queryText = next.queryText ?? queryText;
      esDocumentId = next.esDocumentId ?? esDocumentId;
      esDocumentBody = next.esDocumentBody ?? esDocumentBody;
      showCreateIndex = Boolean(next.showCreateIndex);
      newIndexName = next.newIndexName ?? newIndexName;
      newIndexBody = next.newIndexBody ?? newIndexBody;
      devMethod = next.devMethod ?? devMethod;
      devPath = next.devPath ?? devPath;
      devBody = next.devBody ?? devBody;
      actionMessage = next.actionMessage || '';
      errorMessage = next.errorMessage || '';
    };

    const handleExecuteNative = async (event) => {
      if (!event?.detail || event.detail.sessionId !== sessionId) return;
      handleApplyNative(event);
      const artifact = event.detail.artifact;
      if (!artifact) return;
      if (String(artifact.type || '').toLowerCase() === 'native_mutation') {
        const mutation = parseNativeMutationArtifact(artifact.content);
        await mutateResource(mutation.operation, mutation.name || selectedResource, mutation.payload || '{}');
        return;
      }
      if (activeTab === 'devtools') {
        await runDevTools();
      } else {
        await runElasticsearchQuery();
      }
    };

    window.addEventListener(COPILOT_APPLY_NATIVE, handleApplyNative);
    window.addEventListener(COPILOT_EXECUTE_NATIVE, handleExecuteNative);
    return () => {
      stopInspectorResize();
      window.removeEventListener(COPILOT_APPLY_NATIVE, handleApplyNative);
      window.removeEventListener(COPILOT_EXECUTE_NATIVE, handleExecuteNative);
    };
  });

  async function loadResources() {
    if (!window.wailsBindings || !sessionId) return;
    loading = true;
    errorMessage = '';
    try {
      resources = await window.wailsBindings.ListNativeDatabaseResources(sessionId) || [];
      if (typeof window.wailsBindings.DescribeNativeDatabaseSession === 'function') {
        try {
          sessionOverview = await window.wailsBindings.DescribeNativeDatabaseSession(sessionId);
        } catch (_) {
          sessionOverview = null;
        }
      }
    } catch (error) {
      errorMessage = `加载索引失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loading = false;
    }
  }

  function startInspectorResize(event) {
    event.preventDefault();
    resizingInspector = true;
    window.addEventListener('mousemove', handleInspectorResize);
    window.addEventListener('mouseup', stopInspectorResize);
  }

  function handleInspectorResize(event) {
    if (!resizingInspector) return;
    const body = document.querySelector('.native-database-panel__body');
    const bounds = body?.getBoundingClientRect?.();
    if (!bounds) return;
    inspectorWidth = Math.min(560, Math.max(240, Math.round(bounds.right - event.clientX)));
  }

  function stopInspectorResize() {
    resizingInspector = false;
    window.removeEventListener('mousemove', handleInspectorResize);
    window.removeEventListener('mouseup', stopInspectorResize);
  }

  async function selectResource(resource) {
    selectedParent = '';
    if (!resource?.name || !window.wailsBindings || !sessionId) return;
    selectedResource = resource.name;
    loadingDetails = true;
    errorMessage = '';
    try {
      details = await window.wailsBindings.DescribeNativeDatabaseResource(sessionId, '', resource.name);
      queryText = defaultElasticsearchQuery(querySize, 0);
      queryFrom = 0;
      queryResult = null;
      esDocumentId = '';
      esDocumentBody = '{\n  \n}';
      devPath = `/${resource.name}/_search`;
      activeTab = 'discover';
    } catch (error) {
      details = null;
      errorMessage = `加载索引详情失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loadingDetails = false;
    }
  }

  async function runElasticsearchQuery() {
    if (!selectedResource || !window.wailsBindings?.ExecuteNativeDatabaseQuery) return;
    saving = true;
    errorMessage = '';
    try {
      const body = buildElasticsearchPagedQuery(queryText, queryFrom, querySize);
      queryText = body;
      queryResult = await window.wailsBindings.ExecuteNativeDatabaseQuery(sessionId, '', selectedResource, body);
      activeTab = 'discover';
    } catch (error) {
      errorMessage = `查询失败: ${error?.message || error || '未知错误'}`;
    } finally {
      saving = false;
    }
  }

  async function runDevTools() {
    if (!window.wailsBindings?.ExecuteNativeDatabaseQuery) return;
    saving = true;
    errorMessage = '';
    try {
      const query = buildElasticsearchDevToolsQuery(devMethod, devPath, devBody);
      const result = await window.wailsBindings.ExecuteNativeDatabaseQuery(sessionId, '', selectedResource || '', query);
      actionMessage = formatMutationMessage(result);
      devResult = JSON.stringify(parseNativeResourceContent(result.content), null, 2);
      activeTab = 'devtools';
    } catch (error) {
      errorMessage = `Dev Tools 失败: ${error?.message || error || '未知错误'}`;
    } finally {
      saving = false;
    }
  }

  async function mutateResource(operation, name, payload) {
    if (!window.wailsBindings?.MutateNativeDatabaseResource) return;
    saving = true;
    errorMessage = '';
    actionMessage = '';
    try {
      const result = await window.wailsBindings.MutateNativeDatabaseResource(
        sessionId, '', name || selectedResource || '', operation, payload
      );
      actionMessage = formatMutationMessage(result);
      await loadResources();
      if (selectedResource && operation !== NATIVE_DB_OPERATIONS.ES_DELETE_INDEX) {
        await refreshSelectedResource();
        if (activeTab === 'discover') await runElasticsearchQuery();
      }
      if (operation === NATIVE_DB_OPERATIONS.ES_DELETE_INDEX) {
        selectedResource = null;
        details = null;
      }
    } catch (error) {
      errorMessage = `操作失败: ${error?.message || error || '未知错误'}`;
    } finally {
      saving = false;
    }
  }

  async function refreshSelectedResource() {
    if (!selectedResource) return;
    details = await window.wailsBindings.DescribeNativeDatabaseResource(sessionId, '', selectedResource);
  }

  function requestDeleteDocument(id) {
    pendingDeleteTarget = { kind: 'document', id };
    showDeleteConfirm = true;
  }

  function requestDeleteIndex() {
    pendingDeleteTarget = { kind: 'index', id: selectedResource };
    showDeleteConfirm = true;
  }

  async function confirmDelete() {
    showDeleteConfirm = false;
    if (!pendingDeleteTarget) return;
    if (pendingDeleteTarget.kind === 'index') {
      await mutateResource(NATIVE_DB_OPERATIONS.ES_DELETE_INDEX, pendingDeleteTarget.id, '{}');
    } else {
      await mutateResource(
        NATIVE_DB_OPERATIONS.ES_DELETE,
        selectedResource,
        buildElasticsearchDocumentPayload(pendingDeleteTarget.id, {}, NATIVE_DB_OPERATIONS.ES_DELETE)
      );
    }
    pendingDeleteTarget = null;
  }

  function saveElasticsearchDocument(mode) {
    const operation = mode === 'update' ? NATIVE_DB_OPERATIONS.ES_UPDATE : NATIVE_DB_OPERATIONS.ES_INDEX;
    mutateResource(operation, selectedResource, buildElasticsearchDocumentPayload(esDocumentId, esDocumentBody, operation));
  }

  function loadHitIntoEditor(hit) {
    esDocumentId = hit.id || '';
    esDocumentBody = JSON.stringify(hit.document ?? {}, null, 2);
    activeTab = 'discover';
  }

  async function createIndex() {
    const name = String(newIndexName || '').trim();
    if (!name) {
      errorMessage = '请输入索引名';
      return;
    }
    showCreateIndex = false;
    await mutateResource(NATIVE_DB_OPERATIONS.ES_CREATE_INDEX, name, newIndexBody || '{}');
    newIndexName = '';
  }

  function formatDetails(content) {
    try { return JSON.stringify(JSON.parse(content || '{}'), null, 2); } catch (_) { return content || '{}'; }
  }

  function nextPage() {
    queryFrom += querySize;
    runElasticsearchQuery();
  }

  function prevPage() {
    queryFrom = Math.max(0, queryFrom - querySize);
    runElasticsearchQuery();
  }
</script>

<section class="native-database-panel" aria-label={workspace.title} class:is-resizing={resizingInspector}>
  <header class="native-database-panel__header native-database-panel__context">
    <div>
      <div class="native-database-panel__eyebrow">搜索 · {(databaseType || 'elasticsearch').toUpperCase()}</div>
      <h3>{workspace.title}</h3>
      <p>{dbConfig?.name || 'Elasticsearch 连接'} · {dbConfig?.host}:{dbConfig?.port}</p>
      {#if clusterOverview}
        <div class="native-database-panel__cluster">
          <span class={`health-${clusterOverview.health}`}>健康 {clusterOverview.health}</span>
          <span>集群 {clusterOverview.clusterName || '-'}</span>
          <span>节点 {clusterOverview.nodeCount}</span>
          <span>未分配 {clusterOverview.unassignedShards}</span>
          <span>版本 {clusterOverview.version || '-'}</span>
        </div>
      {/if}
    </div>
    <div class="native-database-panel__header-actions">
      <button type="button" class="native-database-panel__refresh" on:click={() => { showCreateIndex = true; }}>新建索引</button>
      <button class="native-database-panel__refresh" type="button" on:click={loadResources} disabled={loading}>{loading ? '加载中…' : '刷新'}</button>
    </div>
  </header>

  <div class="native-database-panel__tabs">
    <button type="button" class:active={activeTab === 'discover'} on:click={() => activeTab = 'discover'}>Discover</button>
    <button type="button" class:active={activeTab === 'devtools'} on:click={() => activeTab = 'devtools'}>Dev Tools</button>
    <button type="button" class:active={activeTab === 'index'} on:click={() => activeTab = 'index'} disabled={!selectedResource}>索引详情</button>
  </div>

  <div class="native-database-panel__body" style={bodyStyle}>
    <aside class="native-database-panel__list">
      <label class="native-database-panel__search">
        <span>过滤索引</span>
        <input bind:value={resourceSearch} placeholder="按名称过滤" />
      </label>
      {#if loading}
        <div class="native-database-panel__state">加载中…</div>
      {:else}
        <ul class="native-database-panel__tree">
          {#each filteredResources as resource}
            <li>
              <button type="button" class="native-database-panel__resource" class:selected={selectedResource === resource.name} on:click={() => selectResource(resource)}>
                {resource.name}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </aside>

    <main class="native-database-panel__content">
      {#if errorMessage}<div class="native-database-panel__state native-database-panel__error" role="alert">{errorMessage}</div>{/if}
      {#if actionMessage}<div class="native-database-panel__state native-database-panel__success" role="status">{actionMessage}</div>{/if}

      {#if activeTab === 'devtools'}
        <div class="native-database-panel__query">
          <div class="native-database-panel__devtools-bar">
            <select bind:value={devMethod}><option>GET</option><option>POST</option><option>PUT</option><option>HEAD</option></select>
            <input bind:value={devPath} placeholder="/index/_search" />
            <button type="button" on:click={runDevTools} disabled={saving}>{saving ? '执行中…' : '发送'}</button>
          </div>
          <textarea bind:value={devBody} rows="8" placeholder="可选 JSON body" spellcheck="false"></textarea>
          {#if devResult}<pre>{devResult}</pre>{/if}
        </div>
      {:else if activeTab === 'index'}
        {#if loadingDetails}
          <div class="native-database-panel__state">读取索引…</div>
        {:else if details && indexMetadata}
          <div class="native-database-panel__details">
            <strong>{selectedResource}</strong>
            <span>{details.summary}</span>
            <div class="native-database-panel__editor-actions">
              <button type="button" on:click={() => mutateResource(NATIVE_DB_OPERATIONS.ES_REFRESH_INDEX, selectedResource, '{}')}>刷新索引</button>
              <button type="button" class="danger" on:click={requestDeleteIndex}>删除索引</button>
            </div>
            <pre>{formatDetails(JSON.stringify(indexMetadata.mapping ?? {}))}</pre>
          </div>
        {:else}
          <div class="native-database-panel__state">请先选择索引</div>
        {/if}
      {:else}
        <div class="native-database-panel__query">
          <div class="native-database-panel__query-header">
            <strong>{selectedResource || '请选择索引'}</strong>
            <div class="native-database-panel__editor-actions">
              <label>from <input type="number" min="0" bind:value={queryFrom} style="width:72px" /></label>
              <label>size <input type="number" min="1" max="100" bind:value={querySize} style="width:72px" /></label>
              <button type="button" on:click={runElasticsearchQuery} disabled={!selectedResource || saving}>{saving ? '执行中…' : '执行查询'}</button>
            </div>
          </div>
          <textarea bind:value={queryText} spellcheck="false" rows="8"></textarea>
          {#if queryResult}
            <div class="native-database-panel__query-summary">{queryResult.summary}</div>
            <div class="native-database-panel__editor-actions">
              <button type="button" on:click={prevPage} disabled={queryFrom <= 0 || saving}>上一页</button>
              <button type="button" on:click={nextPage} disabled={saving || queryHits.length < querySize}>下一页</button>
            </div>
            <ul class="native-database-panel__hits">
              {#each queryHits as hit}
                <li>
                  <div class="native-database-panel__hit-head">
                    <code>{hit.id || '(auto id)'}</code>
                    <div class="native-database-panel__hit-actions">
                      <button type="button" on:click={() => loadHitIntoEditor(hit)}>编辑</button>
                      {#if hit.id}<button type="button" class="danger" on:click={() => requestDeleteDocument(hit.id)}>删除</button>{/if}
                    </div>
                  </div>
                  <pre>{formatDetails(JSON.stringify(hit.document ?? {}))}</pre>
                </li>
              {/each}
            </ul>
          {/if}
          {#if selectedResource}
            <div class="native-database-panel__editor">
              <h4>写入 / 更新文档</h4>
              <label><span>文档 ID（可留空自动生成）</span><input bind:value={esDocumentId} placeholder="_id" /></label>
              <textarea bind:value={esDocumentBody} rows="8" spellcheck="false"></textarea>
              <div class="native-database-panel__editor-actions">
                <button type="button" on:click={() => saveElasticsearchDocument('index')} disabled={saving}>写入文档</button>
                <button type="button" on:click={() => saveElasticsearchDocument('update')} disabled={saving || !esDocumentId}>更新文档</button>
              </div>
            </div>
          {/if}
        </div>
      {/if}
    </main>

    <div class="native-database-panel__splitter" role="separator" on:mousedown={startInspectorResize}></div>
    <aside class="native-database-panel__inspector">
      <h4>对象信息</h4>
      <dl>
        <div><dt>索引</dt><dd>{selectedResource || '-'}</dd></div>
        {#if indexMetadata}
          <div><dt>健康</dt><dd>{indexMetadata.health}</dd></div>
          <div><dt>文档</dt><dd>{indexMetadata.docsCount}</dd></div>
          <div><dt>大小</dt><dd>{indexMetadata.storeSize}</dd></div>
        {/if}
      </dl>
    </aside>
  </div>
</section>

{#if showCreateIndex}
  <div class="native-database-panel__modal" role="dialog">
    <div class="native-database-panel__modal-card">
      <h4>新建索引</h4>
      <label><span>索引名</span><input bind:value={newIndexName} /></label>
      <label><span>settings / mappings JSON</span><textarea bind:value={newIndexBody} rows="8"></textarea></label>
      <div class="native-database-panel__editor-actions">
        <button type="button" on:click={() => { showCreateIndex = false; }}>取消</button>
        <button type="button" on:click={createIndex} disabled={saving}>创建</button>
      </div>
    </div>
  </div>
{/if}

<ConfirmDialog
  isOpen={showDeleteConfirm}
  title="确认删除"
  message={pendingDeleteTarget?.kind === 'index'
    ? `确定删除索引 ${pendingDeleteTarget?.id || ''} 吗？此操作不可恢复。`
    : `确定删除文档 ${pendingDeleteTarget?.id || ''} 吗？`}
  confirmText="删除"
  type="danger"
  onConfirm={confirmDelete}
  onCancel={() => { showDeleteConfirm = false; pendingDeleteTarget = null; }}
/>

<style>
  .native-database-panel { height: 100%; display: flex; flex-direction: column; overflow: hidden; background: #f7f8f5; color: #1d2935; }
  .native-database-panel__header { display: flex; justify-content: space-between; gap: 16px; padding: 12px 20px; background: #fff; border-bottom: 1px solid #d9e0e4; }
  .native-database-panel__header-actions { display: flex; gap: 8px; align-items: start; }
  .native-database-panel__eyebrow { color: var(--accent-primary); font-size: 11px; font-weight: 750; letter-spacing: .08em; }
  h3 { margin: 3px 0 4px; font-size: 18px; }
  p { margin: 0; color: #6d7783; font-size: 12px; }
  .native-database-panel__cluster { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 8px; }
  .native-database-panel__cluster span { padding: 2px 8px; border-radius: 999px; border: 1px solid #d9e0e4; background: #eff6f5; font-size: 11px; color: #0e6674; }
  .health-red { background: #fef2f2 !important; color: #b91c1c !important; }
  .health-yellow { background: #fffbeb !important; color: #b45309 !important; }
  .native-database-panel__refresh, .native-database-panel__editor-actions button, .native-database-panel__devtools-bar button, .native-database-panel__hit-actions button {
    min-height: 32px; padding: 0 12px; border: 1px solid #d9e0e4; border-radius: 5px; background: #fff; cursor: pointer;
  }
  .danger { color: #b91c1c; }
  .native-database-panel__tabs { display: flex; gap: 4px; padding: 8px 16px 0; background: #fff; border-bottom: 1px solid #d9e0e4; }
  .native-database-panel__tabs button { border: 0; background: transparent; padding: 8px 12px; cursor: pointer; color: #6d7783; }
  .native-database-panel__tabs button.active { color: #0e6674; border-bottom: 2px solid #0e6674; }
  .native-database-panel__body { min-height: 0; flex: 1; display: grid; overflow: hidden; }
  .native-database-panel__list, .native-database-panel__content, .native-database-panel__inspector { overflow: auto; padding: 12px; background: #fff; }
  .native-database-panel__list { border-right: 1px solid #d9e0e4; }
  .native-database-panel__search { display: grid; gap: 4px; font-size: 11px; color: #6d7783; margin-bottom: 8px; }
  .native-database-panel__search input, .native-database-panel__query textarea, .native-database-panel__editor input, .native-database-panel__editor textarea, .native-database-panel__devtools-bar input, .native-database-panel__modal input, .native-database-panel__modal textarea, .native-database-panel__query-header input {
    border: 1px solid #d9e0e4; border-radius: 4px; padding: 8px; width: 100%; box-sizing: border-box;
  }
  .native-database-panel__tree { list-style: none; margin: 0; padding: 0; }
  .native-database-panel__resource { width: 100%; text-align: left; border: 0; background: transparent; padding: 8px; border-radius: 4px; cursor: pointer; }
  .native-database-panel__resource.selected { background: #eff6f5; color: #0e6674; }
  .native-database-panel__splitter { cursor: col-resize; background: #e7ecee; }
  .native-database-panel__state { padding: 12px; border: 1px dashed #d9e0e4; border-radius: 6px; color: #6d7783; margin-bottom: 8px; }
  .native-database-panel__error { color: #dc2626; border-color: #dc2626; }
  .native-database-panel__success { color: #047857; border-color: #059669; }
  .native-database-panel__query, .native-database-panel__editor, .native-database-panel__details { display: grid; gap: 10px; }
  .native-database-panel__query-header, .native-database-panel__editor-actions, .native-database-panel__devtools-bar, .native-database-panel__hit-head, .native-database-panel__hit-actions { display: flex; gap: 8px; align-items: center; flex-wrap: wrap; }
  .native-database-panel__devtools-bar { display: grid; grid-template-columns: 90px minmax(0, 1fr) auto; }
  .native-database-panel__hits { list-style: none; margin: 0; padding: 0; display: grid; gap: 10px; }
  pre { margin: 0; padding: 10px; background: #f7f8f5; border-radius: 6px; overflow: auto; font-size: 11px; }
  .native-database-panel__inspector h4 { margin: 0 0 12px; font-size: 11px; color: #31414d; }
  .native-database-panel__inspector dl { display: grid; gap: 10px; margin: 0; }
  .native-database-panel__inspector dt { color: #7b8791; font-size: 11px; }
  .native-database-panel__inspector dd { margin: 0; font-size: 12px; }
  .native-database-panel__modal { position: fixed; inset: 0; background: rgba(0,0,0,.35); display: grid; place-items: center; z-index: 80; }
  .native-database-panel__modal-card { width: min(480px, 92vw); background: #fff; border-radius: 10px; padding: 16px; display: grid; gap: 10px; }
</style>
