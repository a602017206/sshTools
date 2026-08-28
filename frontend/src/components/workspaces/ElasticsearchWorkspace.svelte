<script>
  import { onMount } from 'svelte';
  import ConfirmDialog from '../ui/ConfirmDialog.svelte';
  import { nativeDatabaseWorkspace } from '../../lib/nativeDatabaseWorkspace.js';
  import {
    NATIVE_DB_OPERATIONS,
    buildElasticsearchDocumentPayload,
    defaultElasticsearchQuery,
    filterNativeResources,
    formatMutationMessage,
    parseElasticsearchClusterOverview,
    parseElasticsearchIndexMetadata,
    parseElasticsearchQueryHits,
    parseNativeResourceContent
  } from '../../lib/nativeDatabaseOperations.js';

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
  let activeTab = 'resources';
  let queryText = defaultElasticsearchQuery();
  let queryResult = null;
  let esDocumentId = '';
  let esDocumentBody = '{\n  \n}';
  let showDeleteConfirm = false;
  let pendingDeleteTarget = null;
  let resourceSearch = '';
  let sessionOverview = null;
  let inspectorWidth = 320;
  let resizingInspector = false;

  $: databaseType = dbConfig?.metadata?.db_type || 'elasticsearch';
  $: workspace = nativeDatabaseWorkspace(databaseType || 'elasticsearch');
  $: queryHits = queryResult ? parseElasticsearchQueryHits(queryResult.content) : [];
  $: filteredResources = filterNativeResources(resources, resourceSearch);
  $: clusterOverview = sessionOverview ? parseElasticsearchClusterOverview(sessionOverview.content) : null;
  $: indexMetadata = details ? parseElasticsearchIndexMetadata(details.content) : null;
  $: bodyStyle = `grid-template-columns: minmax(0, 1fr) 6px ${inspectorWidth}px;`;

  onMount(() => {
    loadResources();
    return () => stopInspectorResize();
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
        } catch (error) {
          sessionOverview = null;
          console.warn('Failed to load native session overview:', error);
        }
      } else {
        sessionOverview = null;
      }
    } catch (error) {
      errorMessage = `加载${workspace.resourceLabel}失败: ${error?.message || error || '未知错误'}`;
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
    const nextWidth = Math.round(bounds.right - event.clientX);
    inspectorWidth = Math.min(560, Math.max(240, nextWidth));
  }

  function stopInspectorResize() {
    resizingInspector = false;
    window.removeEventListener('mousemove', handleInspectorResize);
    window.removeEventListener('mouseup', stopInspectorResize);
  }

  async function selectResource(resource, parent = '') {
    selectedParent = parent;
    if (!resource?.name || !window.wailsBindings || !sessionId) return;
    selectedResource = resource.name;
    loadingDetails = true;
    errorMessage = '';
    actionMessage = '';
    try {
      details = await window.wailsBindings.DescribeNativeDatabaseResource(sessionId, selectedParent, resource.name);
      queryText = defaultElasticsearchQuery();
      queryResult = null;
      esDocumentId = '';
      esDocumentBody = '{\n  \n}';
    } catch (error) {
      details = null;
      errorMessage = `加载资源详情失败: ${error?.message || error || '未知错误'}`;
    } finally {
      loadingDetails = false;
    }
  }

  async function runElasticsearchQuery() {
    if (!selectedResource || !window.wailsBindings?.ExecuteNativeDatabaseQuery) return;
    saving = true;
    errorMessage = '';
    actionMessage = '';
    try {
      queryResult = await window.wailsBindings.ExecuteNativeDatabaseQuery(sessionId, selectedParent, selectedResource, queryText);
      activeTab = 'query';
    } catch (error) {
      errorMessage = `查询失败: ${error?.message || error || '未知错误'}`;
    } finally {
      saving = false;
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
      await runElasticsearchQuery();
      await refreshSelectedResource();
    } catch (error) {
      errorMessage = `操作失败: ${error?.message || error || '未知错误'}`;
    } finally {
      saving = false;
    }
  }

  async function refreshSelectedResource() {
    if (!selectedResource) return;
    details = await window.wailsBindings.DescribeNativeDatabaseResource(sessionId, selectedParent, selectedResource);
  }

  function requestDeleteDocument(id) {
    pendingDeleteTarget = { kind: 'document', id };
    showDeleteConfirm = true;
  }

  async function confirmDelete() {
    showDeleteConfirm = false;
    if (!pendingDeleteTarget || pendingDeleteTarget.kind !== 'document') return;
    await mutateResource(
      NATIVE_DB_OPERATIONS.ES_DELETE,
      buildElasticsearchDocumentPayload(pendingDeleteTarget.id, {}, NATIVE_DB_OPERATIONS.ES_DELETE)
    );
    pendingDeleteTarget = null;
  }

  function saveElasticsearchDocument(mode) {
    const operation = mode === 'update'
      ? NATIVE_DB_OPERATIONS.ES_UPDATE
      : NATIVE_DB_OPERATIONS.ES_INDEX;
    mutateResource(
      operation,
      buildElasticsearchDocumentPayload(esDocumentId, esDocumentBody, operation)
    );
  }

  function loadHitIntoEditor(hit) {
    esDocumentId = hit.id || '';
    esDocumentBody = JSON.stringify(hit.document ?? {}, null, 2);
    activeTab = 'query';
  }

  function formatDetails(content) {
    try { return JSON.stringify(JSON.parse(content || '{}'), null, 2); } catch (_) { return content || '{}'; }
  }

  function formatQueryResult(content) {
    const parsed = parseNativeResourceContent(content);
    if (parsed.raw) {
      try { return JSON.stringify(parsed.raw, null, 2); } catch (_) { return content || '{}'; }
    }
    return formatDetails(content);
  }
</script>

<section class="native-database-panel" aria-label={workspace.title} class:is-resizing={resizingInspector}>
  <header class="native-database-panel__header native-database-panel__context">
    <div>
      <div class="native-database-panel__eyebrow">{(databaseType || 'elasticsearch').toUpperCase()}</div>
      <h3>{workspace.title}</h3>
      <p>{dbConfig?.name || '原生数据库连接'} · {workspace.description}</p>
      {#if clusterOverview}
        <div class="native-database-panel__cluster">
          <span>集群 {clusterOverview.clusterName || '-'}</span>
          <span>健康 {clusterOverview.health}</span>
          <span>节点 {clusterOverview.nodeCount}</span>
          <span>数据节点 {clusterOverview.dataNodeCount}</span>
          <span>版本 {clusterOverview.version || '-'}</span>
          <span>分片 {clusterOverview.activeShards}</span>
        </div>
      {/if}
    </div>
    <button class="native-database-panel__refresh" type="button" on:click={loadResources} disabled={loading}>
      {loading ? '加载中…' : '刷新'}
    </button>
  </header>

  <div class="native-database-panel__tabs">
    <button type="button" class:active={activeTab === 'resources'} on:click={() => activeTab = 'resources'}>资源</button>
    <button type="button" class:active={activeTab === 'query'} on:click={() => activeTab = 'query'} disabled={!selectedResource}>查询</button>
  </div>

  <div class="native-database-panel__body" style={bodyStyle}>
    <main class="native-database-panel__content">
      {#if errorMessage}
        <div class="native-database-panel__state native-database-panel__error" role="alert">{errorMessage}</div>
      {/if}
      {#if actionMessage}
        <div class="native-database-panel__state native-database-panel__success" role="status">{actionMessage}</div>
      {/if}

      {#if activeTab === 'query'}
        <div class="native-database-panel__query">
          <div class="native-database-panel__query-header">
            <strong>{selectedResource || '请选择索引'}</strong>
            <button type="button" on:click={runElasticsearchQuery} disabled={!selectedResource || saving}>
              {saving ? '执行中…' : '执行查询'}
            </button>
          </div>
          <textarea bind:value={queryText} spellcheck="false" placeholder="输入 Elasticsearch DSL JSON"></textarea>
          {#if queryResult}
            <div class="native-database-panel__query-result">
              <div class="native-database-panel__query-summary">{queryResult.summary}</div>
              {#if queryHits.length}
                <ul class="native-database-panel__hits">
                  {#each queryHits as hit}
                    <li>
                      <div class="native-database-panel__hit-head">
                        <code>{hit.id || '(auto id)'}</code>
                        <div class="native-database-panel__hit-actions">
                          <button type="button" on:click={() => loadHitIntoEditor(hit)}>编辑</button>
                          {#if hit.id}
                            <button type="button" class="danger" on:click={() => requestDeleteDocument(hit.id)}>删除</button>
                          {/if}
                        </div>
                      </div>
                      <pre>{formatDetails(JSON.stringify(hit.document ?? {}))}</pre>
                    </li>
                  {/each}
                </ul>
              {:else}
                <pre>{formatQueryResult(queryResult.content)}</pre>
              {/if}
            </div>
          {/if}

          {#if selectedResource}
            <div class="native-database-panel__editor">
              <h4>写入 / 更新文档</h4>
              <label>
                <span>文档 ID（可选）</span>
                <input bind:value={esDocumentId} placeholder="_id，留空则自动生成" />
              </label>
              <textarea bind:value={esDocumentBody} spellcheck="false" rows="8"></textarea>
              <div class="native-database-panel__editor-actions">
                <button type="button" on:click={() => saveElasticsearchDocument('index')} disabled={saving}>写入文档</button>
                <button type="button" on:click={() => saveElasticsearchDocument('update')} disabled={saving || !esDocumentId}>更新文档</button>
              </div>
            </div>
          {/if}
        </div>
      {:else if loading}
        <div class="native-database-panel__state">正在加载{workspace.resourceLabel}…</div>
      {:else if resources.length === 0}
        <div class="native-database-panel__state">未发现可显示的{workspace.resourceLabel}。</div>
      {:else}
        <div class="native-database-panel__toolbar native-database-panel__toolbar--resources">
          {#if workspace.canSearchResources}
            <label class="native-database-panel__search">
              <span>搜索{workspace.resourceLabel}</span>
              <input bind:value={resourceSearch} placeholder={`按${workspace.resourceLabel}名称过滤`} />
            </label>
          {/if}
          <div class="native-database-panel__summary native-database-panel__resource-count">
            <span>{workspace.resourceLabel}</span>
            <strong>{filteredResources.length}{#if resourceSearch}/ {resources.length}{/if}</strong>
          </div>
        </div>
        {#if filteredResources.length === 0}
          <div class="native-database-panel__state">没有匹配“{resourceSearch}”的{workspace.resourceLabel}。</div>
        {:else}
          <ul class="native-database-panel__tree">
            {#each filteredResources as resource}
              <li class:selected={selectedResource === resource.name}>
                <button
                  class="native-database-panel__resource native-database-panel__resource--leaf"
                  type="button"
                  on:click={() => selectResource(resource)}
                  class:selected={selectedResource === resource.name}
                >
                  <span class="native-database-panel__leaf-mark">•</span>
                  <span>{resource.name}</span>
                </button>
              </li>
            {/each}
          </ul>
        {/if}
      {/if}
    </main>

    <div
      class="native-database-panel__splitter"
      role="separator"
      aria-orientation="vertical"
      aria-label="调整对象信息宽度"
      on:mousedown={startInspectorResize}
    ></div>

    <aside class="native-database-panel__inspector">
      <h4>对象信息</h4>
      <dl>
        <div><dt>类型</dt><dd>{(databaseType || 'elasticsearch').toUpperCase()}</dd></div>
        <div><dt>资源</dt><dd>{workspace.resourceLabel}</dd></div>
        <div><dt>数量</dt><dd>{resources.length}</dd></div>
        {#if selectedResource}<div><dt>当前选中</dt><dd>{selectedResource}</dd></div>{/if}
      </dl>
      <p>{workspace.description}</p>

      {#if loadingDetails}
        <p>正在读取对象详情…</p>
      {:else if details}
        <div class="native-database-panel__details">
          <strong>{details.name || selectedResource}</strong>
          <span>{details.summary}</span>

          {#if indexMetadata}
            <dl class="native-database-panel__meta">
              <div><dt>健康</dt><dd>{indexMetadata.health}</dd></div>
              <div><dt>状态</dt><dd>{indexMetadata.status || '-'}</dd></div>
              <div><dt>文档数</dt><dd>{indexMetadata.docsCount}</dd></div>
              <div><dt>已删文档</dt><dd>{indexMetadata.docsDeleted}</dd></div>
              <div><dt>存储大小</dt><dd>{indexMetadata.storeSize}</dd></div>
              <div><dt>主分片存储</dt><dd>{indexMetadata.priStoreSize}</dd></div>
              <div><dt>主分片</dt><dd>{indexMetadata.primaries}</dd></div>
              <div><dt>副本</dt><dd>{indexMetadata.replicas}</dd></div>
            </dl>
            <div class="native-database-panel__mapping">
              <h5>Mapping</h5>
              <pre>{formatDetails(JSON.stringify(indexMetadata.mapping ?? {}))}</pre>
            </div>
            <button type="button" class="native-database-panel__link-action" on:click={() => activeTab = 'query'}>打开查询编辑器</button>
          {/if}
        </div>
      {:else}
        <p>选择一个{workspace.resourceLabel}以查看详情或执行操作。</p>
      {/if}
    </aside>
  </div>
</section>

<ConfirmDialog
  isOpen={showDeleteConfirm}
  title="确认删除"
  message={`确定删除文档 ${pendingDeleteTarget?.id || ''} 吗？`}
  confirmText="删除"
  type="danger"
  onConfirm={confirmDelete}
  onCancel={() => { showDeleteConfirm = false; pendingDeleteTarget = null; }}
/>

<style>
  .native-database-panel { height: 100%; overflow: auto; padding: 20px; background: var(--bg-primary); color: var(--text-primary); }
  .native-database-panel__header { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; padding-bottom: 16px; border-bottom: 1px solid var(--border-primary); }
  h3 { margin: 3px 0 4px; font-size: 18px; line-height: 1.3; }
  p { max-width: 680px; margin: 0; color: var(--text-secondary); font-size: 12px; line-height: 1.6; }
  .native-database-panel__eyebrow { color: var(--accent-primary); font-size: 11px; font-weight: 750; letter-spacing: 0.08em; }
  .native-database-panel__refresh { flex: 0 0 auto; min-height: 32px; padding: 0 12px; border: 1px solid var(--border-primary); border-radius: 5px; background: var(--bg-secondary); color: var(--text-primary); cursor: pointer; }
  .native-database-panel__tabs { display: flex; gap: 8px; padding: 10px 16px 0; background: #fff; border-bottom: 1px solid #d9e0e4; }
  .native-database-panel__tabs button { border: 0; background: transparent; color: #6d7783; padding: 8px 12px; cursor: pointer; border-bottom: 2px solid transparent; }
  .native-database-panel__tabs button.active { color: #0e6674; border-bottom-color: #0e6674; font-weight: 600; }
  .native-database-panel__state { margin-top: 18px; padding: 18px; border: 1px dashed var(--border-primary); border-radius: 6px; background: var(--bg-secondary); color: var(--text-secondary); font-size: 13px; }
  .native-database-panel__error { border-color: #dc2626; color: #dc2626; }
  .native-database-panel__success { border-color: #059669; color: #047857; }
  .native-database-panel__toolbar { display: grid; gap: 10px; margin-bottom: 12px; }
  .native-database-panel__toolbar--resources { margin-top: 0; }
  .native-database-panel__search { display: grid; gap: 6px; font-size: 11px; color: #6d7783; }
  .native-database-panel__search input {
    width: 100%; min-height: 34px; border: 1px solid #d9e0e4; border-radius: 4px; padding: 0 10px;
    box-sizing: border-box; color: #31414d; font-size: 12px; background: #fff;
  }
  .native-database-panel__cluster {
    display: flex; flex-wrap: wrap; gap: 8px; margin-top: 8px;
  }
  .native-database-panel__cluster span {
    display: inline-flex; align-items: center; min-height: 24px; padding: 0 8px;
    border: 1px solid #d9e0e4; border-radius: 999px; background: #eff6f5; color: #0e6674; font-size: 11px;
  }
  .native-database-panel__summary { display: flex; align-items: center; justify-content: space-between; padding: 9px 12px; border: 1px solid var(--border-primary); border-radius: 6px 6px 0 0; background: var(--bg-tertiary); color: var(--text-secondary); font-size: 12px; }
  .native-database-panel__resource-count { margin-top: 0; border-radius: 4px; }
  .native-database-panel__tree { margin: 0; padding: 0; list-style: none; border: 1px solid var(--border-primary); border-top: 0; border-radius: 0 0 6px 6px; overflow: hidden; }
  .native-database-panel__resource { display: flex; align-items: center; width: 100%; min-height: 40px; gap: 9px; padding: 8px 12px; border: 0; background: var(--bg-secondary); color: var(--text-primary); font: inherit; font-size: 13px; text-align: left; cursor: pointer; }
  .native-database-panel__resource.selected, .native-database-panel__resource--leaf.selected { background: #eff6f5; color: #0e6674; }
  .native-database-panel__query { display: grid; gap: 12px; }
  .native-database-panel__query-header { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
  .native-database-panel__query textarea, .native-database-panel__editor textarea, .native-database-panel__editor input { width: 100%; border: 1px solid #d9e0e4; border-radius: 4px; padding: 10px; font: 12px/1.5 ui-monospace, SFMono-Regular, Menlo, monospace; box-sizing: border-box; }
  .native-database-panel__query-result pre, .native-database-panel__details pre { max-width: 100%; max-height: 280px; margin: 0; overflow: auto; padding: 9px; border: 1px solid #d9e0e4; border-radius: 4px; background: #fff; font: 10px/1.5 ui-monospace, SFMono-Regular, Menlo, monospace; white-space: pre-wrap; overflow-wrap: anywhere; }
  .native-database-panel__hits { list-style: none; padding: 0; margin: 0; display: grid; gap: 10px; }
  .native-database-panel__hit-head { display: flex; align-items: center; justify-content: space-between; gap: 8px; margin-bottom: 6px; }
  .native-database-panel__hit-actions { display: flex; gap: 6px; }
  .native-database-panel__editor { display: grid; gap: 10px; margin-top: 14px; }
  .native-database-panel__editor label { display: grid; gap: 4px; font-size: 11px; color: #6d7783; }
  .native-database-panel__editor-actions { display: flex; gap: 8px; flex-wrap: wrap; }
  button.danger { color: #b91c1c; }
  .native-database-panel__link-action { margin-top: 10px; border: 0; background: transparent; color: #0e6674; cursor: pointer; padding: 0; }
  .native-database-panel { padding: 0; display: flex; flex-direction: column; overflow: hidden; background: #f7f8f5; color: #1d2935; font-family: "PingFang SC", "Hiragino Sans GB", -apple-system, sans-serif; }
  .native-database-panel__context { min-height: 64px; padding: 0 20px; align-items: center; background: #fff; border-bottom-color: #d9e0e4; }
  .native-database-panel__body { min-height: 0; flex: 1; display: grid; grid-template-columns: minmax(0, 1fr) 280px; overflow: hidden; }
  .native-database-panel__splitter {
    width: 6px; cursor: col-resize; background: #e7ecee; border-left: 1px solid #d9e0e4; border-right: 1px solid #d9e0e4;
  }
  .native-database-panel__splitter:hover, .native-database-panel.is-resizing .native-database-panel__splitter {
    background: #cfe3e5;
  }
  .native-database-panel__content { min-width: 0; overflow: auto; padding: 16px; background: #fff; }
  .native-database-panel__inspector { min-width: 0; padding: 18px 16px; border-left: 1px solid #d9e0e4; background: #f7f8f5; overflow: auto; }
  .native-database-panel__meta { margin: 12px 0 0; display: grid; gap: 8px; }
  .native-database-panel__meta div { display: grid; grid-template-columns: 84px minmax(0, 1fr); gap: 8px; align-items: start; }
  .native-database-panel__meta dt { color: #7b8791; font-size: 11px; }
  .native-database-panel__meta dd { margin: 0; color: #31414d; font-size: 12px; overflow-wrap: anywhere; }
  .native-database-panel__mapping { display: grid; gap: 6px; margin-top: 12px; }
  .native-database-panel__mapping h5 { margin: 0; color: #31414d; font-size: 11px; letter-spacing: .04em; }
  .native-database-panel__inspector h4 { margin: 0 0 14px; color: #31414d; font-size: 11px; letter-spacing: .04em; }
  .native-database-panel__inspector dl { margin: 0; display: grid; gap: 12px; }
  .native-database-panel__inspector dl div { display: grid; gap: 3px; }
  .native-database-panel__inspector dt { color: #7b8791; font-size: 11px; }
  .native-database-panel__inspector dd { margin: 0; color: #31414d; font-size: 12px; overflow-wrap: anywhere; }
  @media (max-width: 760px) {
    .native-database-panel__body { grid-template-columns: 1fr !important; }
    .native-database-panel__splitter, .native-database-panel__inspector { display: none; }
  }
</style>
