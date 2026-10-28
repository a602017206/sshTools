<script>
  import { onMount } from 'svelte';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  import DockerContainerDetail from './DockerContainerDetail.svelte';
  import DockerResourceDetail from './DockerResourceDetail.svelte';
  import {
    DOCKER_RESOURCE_TABS,
    formatDockerSize,
    imageDisplayTags,
    removeResourceConfirmCopy,
    resourceRowKey
  } from '../lib/dockerResources.js';
  import { portalToBody, resolveContextMenuPoint } from '../lib/contextMenu.js';
  import {
    actionSuccessLabel,
    containerActionMenu,
    isStillRunningError,
    removeConfirmCopy
  } from '../lib/dockerContainerActions.js';
  import {
    containerField,
    dockerTransportLabel,
    formatContainerCreated,
    isContainerRunning,
    usedCliFallback,
    visibleContainers,
    clampDockerDetailWidth,
    DEFAULT_DOCKER_DETAIL_WIDTH
  } from '../lib/dockerWorkspace.js';

  export let sessionId = '';
  export let connection = null;
  /** (connected: boolean, error: string) => void，供 TerminalPanel 同步页签状态 */
  export let onStatus = () => {};
  /** 打开/关闭详情时回调 (containerId, detail|null) */
  export let onSelect = () => {};
  /** 在容器内打开 Exec 终端页签：({ containerId, containerName }) => void */
  export let onExec = (payload) => {};

  let status = 'idle'; // idle | connecting | connected | error
  let errorMessage = '';
  let loading = false;
  let onlyRunning = false;
  let containers = [];
  let destroyed = false;

  // 行操作：菜单 / 忙碌 / 删除确认 / 选中容器（Task 7 侧栏使用）
  let menu = null; // { x, y, container }
  let busyId = '';
  let actionNotice = '';
  let noticeTimer = null;
  let removeTarget = null;
  let removeForce = false;
  let showRemoveConfirm = false;
  export let selectedContainerId = '';
  export let selectedDetail = null;
  let detailError = '';

  // 资源 Tab：containers | images | networks | volumes
  let resourceKind = 'containers';
  let resourceRows = [];
  let resourceToken = 0; // 使过期的列表结果失效
  let inspectToken = 0; // 使过期的 Inspect 结果失效
  let selectedResourceKey = '';
  let selectedResourceRow = null;
  let selectedResourceDetail = null;
  let resourceDetailError = '';
  let resRemoveTarget = null; // { kind, row }
  let resRemoveForce = false;
  let showResRemoveConfirm = false;
  let resBusyKey = '';

  // 详情侧栏宽度（仅会话内记忆，不落盘）
  let detailWidth = DEFAULT_DOCKER_DETAIL_WIDTH;
  let detailResizeStartX = 0;
  let detailResizeStartWidth = 0;
  let dockerMainEl = null;

  function startDetailResize(event) {
    event.preventDefault();
    detailResizeStartX = event.clientX;
    detailResizeStartWidth = clampDockerDetailWidth(detailWidth, dockerMainEl?.clientWidth || 0);
    window.addEventListener('pointermove', resizeDetail);
    window.addEventListener('pointerup', stopDetailResize, { once: true });
  }

  function resizeDetail(event) {
    const mainWidth = dockerMainEl?.clientWidth || 0;
    const proposed = detailResizeStartWidth - (event.clientX - detailResizeStartX);
    detailWidth = clampDockerDetailWidth(proposed, mainWidth);
  }

  function stopDetailResize() {
    window.removeEventListener('pointermove', resizeDetail);
    window.removeEventListener('pointerup', stopDetailResize);
  }

  $: transportLabel = dockerTransportLabel(connection);
  $: rows = visibleContainers(containers, onlyRunning);
  $: cliFallback = usedCliFallback(containers);
  $: connected = status === 'connected';

  function bindings() {
    return window.wailsBindings || {};
  }

  function errText(error) {
    return String(error?.message || error || '未知错误');
  }

  async function connect() {
    const { ConnectDocker } = bindings();
    if (typeof ConnectDocker !== 'function') {
      status = 'error';
      errorMessage = 'Docker 连接功能不可用';
      onStatus(false, errorMessage);
      return;
    }
    status = 'connecting';
    errorMessage = '';
    try {
      await ConnectDocker(sessionId, connection);
      if (destroyed) return;
      status = 'connected';
      onStatus(true, '');
      await refresh();
    } catch (error) {
      if (destroyed) return;
      status = 'error';
      errorMessage = errText(error);
      containers = [];
      onStatus(false, errorMessage);
    }
  }

  const LIST_METHODS = {
    images: 'ListDockerImages',
    networks: 'ListDockerNetworks',
    volumes: 'ListDockerVolumes'
  };
  const INSPECT_METHODS = {
    images: 'InspectDockerImage',
    networks: 'InspectDockerNetwork',
    volumes: 'InspectDockerVolume'
  };
  const REMOVE_METHODS = {
    images: 'RemoveDockerImage',
    networks: 'RemoveDockerNetwork',
    volumes: 'RemoveDockerVolume'
  };

  /** 刷新当前 Tab 的列表。 */
  async function refresh() {
    if (resourceKind === 'containers') return refreshContainers();
    return refreshResources();
  }

  async function refreshResources() {
    const kind = resourceKind;
    const list = bindings()[LIST_METHODS[kind]];
    if (!connected || typeof list !== 'function') return;
    const token = ++resourceToken;
    loading = true;
    try {
      const result = await list(sessionId);
      if (destroyed || token !== resourceToken) return;
      resourceRows = Array.isArray(result) ? result : [];
      errorMessage = '';
    } catch (error) {
      if (destroyed || token !== resourceToken) return;
      errorMessage = errText(error);
    } finally {
      if (token === resourceToken) loading = false;
    }
  }

  function clearResourceSelection() {
    inspectToken += 1; // 丢弃进行中的 Inspect
    selectedResourceKey = '';
    selectedResourceRow = null;
    selectedResourceDetail = null;
    resourceDetailError = '';
  }

  async function switchTab(kind) {
    if (kind === resourceKind) return;
    closeMenu();
    closeDetail();
    clearResourceSelection();
    resourceToken += 1; // 丢弃上一个 Tab 进行中的列表请求
    resourceKind = kind;
    resourceRows = [];
    errorMessage = '';
    actionNotice = '';
    loading = false;
    if (connected) await refresh();
  }

  function resourceId(kind, row) {
    return kind === 'volumes' ? String(row?.Name ?? '') : String(row?.ID ?? '');
  }

  function resourceLabel(kind, row) {
    if (!row) return '';
    if (kind === 'images') {
      const tags = imageDisplayTags(row);
      return tags !== '<none>' ? tags : String(row.ID ?? '').slice(0, 12);
    }
    if (kind === 'networks') return row.Name || String(row.ID ?? '').slice(0, 12);
    return row.Name || '';
  }

  async function openResourceDetail(row) {
    const kind = resourceKind;
    const key = resourceRowKey(kind, row);
    const id = resourceId(kind, row);
    if (!key || !id) return;
    const token = ++inspectToken;
    selectedResourceKey = key;
    selectedResourceRow = row;
    selectedResourceDetail = null;
    resourceDetailError = '';
    const inspect = bindings()[INSPECT_METHODS[kind]];
    if (typeof inspect !== 'function') {
      selectedResourceDetail = row;
      return;
    }
    try {
      const detail = await inspect(sessionId, id);
      if (destroyed || token !== inspectToken || selectedResourceKey !== key) return;
      selectedResourceDetail = detail || row;
    } catch (error) {
      if (destroyed || token !== inspectToken || selectedResourceKey !== key) return;
      // Inspect 失败：退回列表行字段，并提示原因
      selectedResourceDetail = row;
      resourceDetailError = errText(error);
    }
  }

  function closeResourceDetail() {
    clearResourceSelection();
  }

  function askRemoveResource(row) {
    resRemoveTarget = { kind: resourceKind, row };
    resRemoveForce = false;
    showResRemoveConfirm = true;
  }

  function cancelResourceRemove() {
    showResRemoveConfirm = false;
    resRemoveTarget = null;
    resRemoveForce = false;
  }

  async function confirmResourceRemove() {
    const target = resRemoveTarget;
    const force = resRemoveForce;
    showResRemoveConfirm = false;
    resRemoveTarget = null;
    resRemoveForce = false;
    if (!target) return;
    const { kind, row } = target;
    const remove = bindings()[REMOVE_METHODS[kind]];
    const id = resourceId(kind, row);
    if (!id || typeof remove !== 'function') {
      errorMessage = 'Docker 操作功能不可用';
      return;
    }
    const key = resourceRowKey(kind, row);
    resBusyKey = key;
    errorMessage = '';
    try {
      if (kind === 'images') await remove(sessionId, id, force);
      else await remove(sessionId, id);
      if (destroyed) return;
      if (kind === resourceKind) {
        if (selectedResourceKey === key) clearResourceSelection();
        flashNotice(`${resourceLabel(kind, row)} 已删除`);
        await refreshResources();
      }
    } catch (error) {
      if (destroyed) return;
      errorMessage = errText(error);
    } finally {
      resBusyKey = '';
    }
  }

  async function refreshContainers() {
    const { ListDockerContainers } = bindings();
    if (!connected || typeof ListDockerContainers !== 'function') return;
    loading = true;
    try {
      // all=true 取全部容器，「仅运行中」在前端筛选，切换无需再请求
      const list = await ListDockerContainers(sessionId, true);
      if (destroyed) return;
      containers = Array.isArray(list) ? list : [];
      errorMessage = '';
    } catch (error) {
      if (destroyed) return;
      errorMessage = errText(error);
    } finally {
      loading = false;
    }
  }

  function containerKey(container) {
    return String(containerField(container, 'ID') || containerField(container, 'Name'));
  }

  function containerLabel(container) {
    return containerField(container, 'Name') || String(containerField(container, 'ID')).slice(0, 12) || '';
  }

  function flashNotice(text) {
    actionNotice = text;
    if (noticeTimer) clearTimeout(noticeTimer);
    noticeTimer = setTimeout(() => {
      actionNotice = '';
    }, 2500);
  }

  function openMenu(event, container) {
    const point = resolveContextMenuPoint(event, { menuWidth: 140, menuHeight: 190 });
    menu = { x: point.x, y: point.y, container };
  }

  function openMenuFromButton(event, container) {
    const rect = event.currentTarget.getBoundingClientRect();
    openMenu(
      {
        clientX: rect.right - 140,
        clientY: rect.bottom + 2,
        preventDefault: () => event.preventDefault(),
        stopPropagation: () => event.stopPropagation()
      },
      container
    );
  }

  function closeMenu() {
    menu = null;
  }

  async function openDetail(container) {
    const id = containerKey(container);
    if (!id) return;
    selectedContainerId = id;
    selectedDetail = null;
    detailError = '';
    onSelect(id, null);
    const { InspectDockerContainer } = bindings();
    if (typeof InspectDockerContainer !== 'function') return;
    try {
      const detail = await InspectDockerContainer(sessionId, id);
      if (destroyed || selectedContainerId !== id) return;
      selectedDetail = detail || null;
      onSelect(id, selectedDetail);
    } catch (error) {
      if (destroyed || selectedContainerId !== id) return;
      detailError = errText(error);
    }
  }

  export function closeDetail() {
    selectedContainerId = '';
    selectedDetail = null;
    detailError = '';
    onSelect('', null);
  }

  async function runAction(actionId, container, extra) {
    const id = containerKey(container);
    const methodName = {
      start: 'StartDockerContainer',
      stop: 'StopDockerContainer',
      restart: 'RestartDockerContainer',
      remove: 'RemoveDockerContainer'
    }[actionId];
    const method = bindings()[methodName];
    if (!id || typeof method !== 'function') {
      errorMessage = 'Docker 操作功能不可用';
      return false;
    }
    busyId = id;
    errorMessage = '';
    try {
      if (actionId === 'remove') await method(sessionId, id, Boolean(extra));
      else await method(sessionId, id);
      if (destroyed) return true;
      if (actionId === 'remove' && selectedContainerId === id) closeDetail();
      flashNotice(`${containerLabel(container)} ${actionSuccessLabel(actionId)}`);
      await refresh();
      return true;
    } catch (error) {
      if (destroyed) return false;
      errorMessage = errText(error);
      if (actionId === 'remove' && !extra && isStillRunningError(error)) {
        // 删除失败且容器仍在运行：二次确认强制删除
        removeTarget = container;
        removeForce = true;
        showRemoveConfirm = true;
      }
      return false;
    } finally {
      busyId = '';
    }
  }

  function handleMenuAction(actionId) {
    const container = menu?.container;
    closeMenu();
    if (!container) return;
    if (actionId === 'detail') {
      openDetail(container);
    } else if (actionId === 'remove') {
      removeTarget = container;
      removeForce = false;
      showRemoveConfirm = true;
    } else if (actionId === 'exec') {
      onExec({ containerId: containerKey(container), containerName: containerLabel(container) });
    } else {
      runAction(actionId, container);
    }
  }

  function cancelRemove() {
    showRemoveConfirm = false;
    removeTarget = null;
    removeForce = false;
  }

  async function confirmRemove() {
    const target = removeTarget;
    const force = removeForce;
    showRemoveConfirm = false;
    removeTarget = null;
    removeForce = false;
    if (target) await runAction('remove', target, force);
  }

  $: resourceTabLabel = DOCKER_RESOURCE_TABS.find((tab) => tab.id === resourceKind)?.label || '';
  $: resRemoveCopy = removeResourceConfirmCopy(
    resRemoveTarget?.kind || 'images',
    resRemoveTarget ? resourceLabel(resRemoveTarget.kind, resRemoveTarget.row) : '',
    resRemoveForce
  );
  $: menuItems = menu ? containerActionMenu(menu.container) : [];
  $: removeCopy = removeConfirmCopy(removeTarget ? containerLabel(removeTarget) : '', removeForce);

  onMount(() => {
    connect();
    return () => {
      destroyed = true;
      stopDetailResize();
      if (noticeTimer) clearTimeout(noticeTimer);
    };
  });
</script>

<svelte:window
  on:click={closeMenu}
  on:keydown={(event) => {
    if (event.key === 'Escape') closeMenu();
  }}
/>

<div class="docker-workspace">
  <div class="docker-toolbar">
    <div class="docker-title">
      <span class="docker-name">{connection?.name || 'Docker'}</span>
      <span class="docker-badge">{transportLabel}</span>
      {#if cliFallback}
        <span class="docker-hint">CLI 回退</span>
      {/if}
    </div>
    <div class="docker-actions">
      {#if resourceKind === 'containers'}
        <label class="docker-filter">
          <input type="checkbox" bind:checked={onlyRunning} />
          仅运行中
        </label>
      {/if}
      <button
        type="button"
        class="docker-btn"
        disabled={status === 'connecting' || loading}
        on:click={() => (connected ? refresh() : connect())}
      >
        {connected ? (loading ? '刷新中…' : '刷新') : '重新连接'}
      </button>
    </div>
  </div>

  <div class="docker-resource-tabs" role="tablist" aria-label="Docker 资源">
    {#each DOCKER_RESOURCE_TABS as tab (tab.id)}
      <button
        type="button"
        role="tab"
        class="docker-resource-tab"
        class:is-active={resourceKind === tab.id}
        aria-selected={resourceKind === tab.id}
        on:click={() => switchTab(tab.id)}
      >{tab.label}</button>
    {/each}
  </div>

  <div class="docker-main" bind:this={dockerMainEl}>
  <div class="docker-body">
    {#if status === 'connecting' || status === 'idle'}
      <div class="docker-empty" role="status">
        <p class="kicker">容器</p>
        <h2>正在连接 Docker…</h2>
      </div>
    {:else if status === 'error'}
      <div class="docker-empty" role="alert">
        <p class="kicker">容器</p>
        <h2>未连接到 Docker</h2>
        <p class="detail">{errorMessage}</p>
      </div>
    {:else if resourceKind !== 'containers'}
      {#if errorMessage}<div class="docker-banner docker-error">{errorMessage}</div>{/if}
      {#if actionNotice}<div class="docker-banner docker-notice" role="status">{actionNotice}</div>{/if}
      {#if resourceRows.length === 0}
        <div class="docker-empty" role="status">
          <p class="kicker">{resourceTabLabel}</p>
          <h2>{loading ? '加载中…' : `没有${resourceTabLabel}`}</h2>
          {#if !loading}<p class="detail">该 Docker 引擎上还没有{resourceTabLabel}。</p>{/if}
        </div>
      {:else}
        <div class="docker-table-wrap">
          <table class="docker-table">
            <thead>
              <tr>
                {#if resourceKind === 'images'}
                  <th>标签</th><th>ID</th><th>大小</th><th>创建时间</th>
                {:else if resourceKind === 'networks'}
                  <th>名称</th><th>ID</th><th>驱动</th><th>范围</th>
                {:else}
                  <th>名称</th><th>驱动</th><th>挂载点</th><th>创建时间</th>
                {/if}
                <th class="docker-col-actions">操作</th>
              </tr>
            </thead>
            <tbody>
              {#each resourceRows as row (resourceRowKey(resourceKind, row))}
                <tr
                  class:docker-row--selected={selectedResourceKey !== '' && selectedResourceKey === resourceRowKey(resourceKind, row)}
                  class:docker-row--busy={resBusyKey !== '' && resBusyKey === resourceRowKey(resourceKind, row)}
                  on:click={() => openResourceDetail(row)}
                >
                  {#if resourceKind === 'images'}
                    <td class="docker-cell-name" title={imageDisplayTags(row)}>{imageDisplayTags(row)}</td>
                    <td class="docker-mono">{String(row.ID ?? '').replace(/^sha256:/, '').slice(0, 12) || '-'}</td>
                    <td>{formatDockerSize(row.Size)}</td>
                    <td>{formatContainerCreated(row.Created)}</td>
                  {:else if resourceKind === 'networks'}
                    <td class="docker-cell-name">{row.Name || '-'}</td>
                    <td class="docker-mono">{String(row.ID ?? '').slice(0, 12) || '-'}</td>
                    <td>{row.Driver || '-'}</td>
                    <td>{row.Scope || '-'}</td>
                  {:else}
                    <td class="docker-cell-name">{row.Name || '-'}</td>
                    <td>{row.Driver || '-'}</td>
                    <td class="docker-mono" title={row.Mountpoint || ''}>{row.Mountpoint || '-'}</td>
                    <td>{formatContainerCreated(row.Created)}</td>
                  {/if}
                  <td class="docker-col-actions">
                    <button
                      type="button"
                      class="docker-res-delete"
                      title="删除"
                      aria-label="删除"
                      disabled={resBusyKey !== '' && resBusyKey === resourceRowKey(resourceKind, row)}
                      on:click|stopPropagation={() => askRemoveResource(row)}
                    >删除</button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    {:else if rows.length === 0}
      <div class="docker-empty" role="status">
        <p class="kicker">容器</p>
        <h2>{containers.length === 0 ? '没有容器' : '没有运行中的容器'}</h2>
        <p class="detail">
          {containers.length === 0 ? '该 Docker 引擎上还没有容器。' : '取消「仅运行中」可查看已停止的容器。'}
        </p>
        {#if errorMessage}<p class="detail docker-error">{errorMessage}</p>{/if}
      </div>
    {:else}
      {#if errorMessage}<div class="docker-banner docker-error">{errorMessage}</div>{/if}
      {#if actionNotice}<div class="docker-banner docker-notice" role="status">{actionNotice}</div>{/if}
      {#if detailError}<div class="docker-banner docker-error">详情加载失败：{detailError}</div>{/if}
      <div class="docker-table-wrap">
        <table class="docker-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>镜像</th>
              <th>状态</th>
              <th>端口</th>
              <th>ID</th>
              <th>创建时间</th>
              <th class="docker-col-actions">操作</th>
            </tr>
          </thead>
          <tbody>
            {#each rows as container (containerField(container, 'ID') || containerField(container, 'Name'))}
              <tr
                class:docker-row--selected={selectedContainerId !== '' && selectedContainerId === containerKey(container)}
                class:docker-row--busy={busyId !== '' && busyId === containerKey(container)}
                on:click={() => openDetail(container)}
                on:contextmenu={(event) => openMenu(event, container)}
              >
                <td class="docker-cell-name">
                  <span class="ops-status-dot docker-dot" class:docker-dot--up={isContainerRunning(container)}></span>
                  {containerField(container, 'Name') || '-'}
                </td>
                <td>{containerField(container, 'Image') || '-'}</td>
                <td>{containerField(container, 'Status') || '-'}</td>
                <td class="docker-mono">{containerField(container, 'Ports') || '-'}</td>
                <td class="docker-mono">{String(containerField(container, 'ID')).slice(0, 12) || '-'}</td>
                <td>{formatContainerCreated(containerField(container, 'Created'))}</td>
                <td class="docker-col-actions">
                  <button
                    type="button"
                    class="docker-more"
                    title="容器操作"
                    aria-label="容器操作"
                    aria-haspopup="menu"
                    disabled={busyId !== '' && busyId === containerKey(container)}
                    on:click|stopPropagation={(event) => openMenuFromButton(event, container)}
                  >⋯</button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
  {#if connected && resourceKind !== 'containers' && selectedResourceKey}
    <div
      class="docker-detail-shell"
      style={`width: ${detailWidth}px;`}
    >
      <div
        class="docker-detail-resizer"
        role="separator"
        aria-orientation="vertical"
        aria-label="调整详情栏宽度"
        on:pointerdown={startDetailResize}
      ></div>
      {#key selectedResourceKey}
        <DockerResourceDetail
          kind={resourceKind}
          title={resourceLabel(resourceKind, selectedResourceRow)}
          detail={selectedResourceDetail}
          detailError={resourceDetailError}
          onClose={closeResourceDetail}
        />
      {/key}
    </div>
  {:else if connected && resourceKind === 'containers' && selectedContainerId}
    <!-- key：切换容器时销毁旧侧栏（onDestroy 内 StopLogs），再挂载新的 -->
    <div
      class="docker-detail-shell"
      style={`width: ${detailWidth}px;`}
    >
      <div
        class="docker-detail-resizer"
        role="separator"
        aria-orientation="vertical"
        aria-label="调整详情栏宽度"
        on:pointerdown={startDetailResize}
      ></div>
      {#key selectedContainerId}
        <DockerContainerDetail
          {sessionId}
          containerId={selectedContainerId}
          detail={selectedDetail}
          {detailError}
          onClose={closeDetail}
          onExec={() => onExec({
            containerId: selectedContainerId,
            containerName: containerField(selectedDetail, 'Name') || String(selectedContainerId).slice(0, 12)
          })}
        />
      {/key}
    </div>
  {/if}
  </div>
</div>

{#if menu}
  <div
    class="docker-menu ops-flyout fixed z-[80] rounded-xl text-xs py-1"
    style={`left: ${menu.x}px; top: ${menu.y}px; width: 140px;`}
    use:portalToBody
    on:mousedown|stopPropagation
  >
    {#each menuItems as item (item.id)}
      <button
        type="button"
        class="docker-menu-item"
        class:is-danger={item.danger}
        disabled={item.disabled}
        on:click|stopPropagation={() => handleMenuAction(item.id)}
      >{item.label}</button>
    {/each}
  </div>
{/if}

<ConfirmDialog
  bind:isOpen={showRemoveConfirm}
  title={removeCopy.title}
  message={removeCopy.message}
  type="danger"
  confirmText={removeCopy.confirmText}
  cancelText="取消"
  onConfirm={confirmRemove}
  onCancel={cancelRemove}
>
  <label class="docker-force">
    <input type="checkbox" bind:checked={removeForce} />
    强制删除（容器运行中也会被终止）
  </label>
</ConfirmDialog>

<ConfirmDialog
  bind:isOpen={showResRemoveConfirm}
  title={resRemoveCopy.title}
  message={resRemoveCopy.message}
  type="danger"
  confirmText={resRemoveCopy.confirmLabel}
  cancelText={resRemoveCopy.cancelLabel}
  onConfirm={confirmResourceRemove}
  onCancel={cancelResourceRemove}
>
  {#if resRemoveTarget?.kind === 'images'}
    <label class="docker-force">
      <input type="checkbox" bind:checked={resRemoveForce} />
      强制删除（镜像被引用时也尝试删除）
    </label>
  {/if}
</ConfirmDialog>

<style>
  .docker-workspace {
    height: 100%;
    display: flex;
    flex-direction: column;
    min-height: 0;
    color: var(--text-primary);
  }

  .docker-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 14px;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-title {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .docker-name {
    font-size: 14px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .docker-badge {
    flex-shrink: 0;
    padding: 1px 8px;
    border-radius: 999px;
    border: 1px solid var(--glass-border);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.06em;
    color: var(--ops-signal);
    font-family: var(--terminal-font-family);
  }

  .docker-hint {
    font-size: 10px;
    color: var(--text-secondary);
  }

  .docker-actions {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
  }

  .docker-filter {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--text-secondary);
    cursor: pointer;
  }

  .docker-btn {
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-size: 12px;
    font-weight: 600;
    min-height: 30px;
    padding: 0 14px;
    border-radius: 10px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    color: var(--text-primary);
    box-shadow: var(--shadow-glass);
    transition: border-color var(--trans-fast), color var(--trans-fast);
  }

  .docker-btn:hover:not(:disabled) {
    border-color: var(--ops-signal);
    color: var(--ops-signal);
  }

  .docker-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .docker-btn:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 2px;
  }

  .docker-resource-tabs {
    display: flex;
    gap: 4px;
    padding: 6px 14px 0;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-resource-tab {
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-size: 12px;
    font-weight: 600;
    padding: 6px 14px;
    border: 0;
    border-bottom: 2px solid transparent;
    background: transparent;
    color: var(--text-secondary);
  }

  .docker-resource-tab:hover {
    color: var(--text-primary);
  }

  .docker-resource-tab.is-active {
    color: var(--ops-signal);
    border-bottom-color: var(--ops-signal);
  }

  .docker-resource-tab:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 1px;
  }

  .docker-res-delete {
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-size: 11px;
    padding: 2px 8px;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--ops-alert);
  }

  .docker-res-delete:hover:not(:disabled) {
    border-color: var(--glass-border);
    background: var(--glass-bg);
  }

  .docker-res-delete:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 1px;
  }

  .docker-main {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: row;
  }

  .docker-detail-shell {
    flex-shrink: 0;
    max-width: max(320px, 55%);
    display: flex;
    flex-direction: row;
    min-width: 0;
    min-height: 0;
    border-left: 1px solid var(--glass-border);
    background: var(--glass-bg-strong);
  }

  .docker-detail-resizer {
    width: 6px;
    flex: 0 0 6px;
    cursor: col-resize;
    touch-action: none;
    background: transparent;
  }

  .docker-detail-resizer:hover,
  .docker-detail-resizer:focus-visible {
    background: color-mix(in srgb, var(--ops-signal, #1687d4) 28%, transparent);
  }

  .docker-body {
    flex: 1;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .docker-empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 32px;
    text-align: center;
  }

  .kicker {
    margin: 0;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: var(--ops-signal);
    font-family: var(--terminal-font-family);
  }

  h2 {
    margin: 0;
    font-size: 20px;
    letter-spacing: -0.03em;
  }

  .detail {
    margin: 0;
    max-width: 480px;
    font-size: 13px;
    color: var(--text-secondary);
    word-break: break-word;
  }

  .docker-error {
    color: var(--ops-danger, #ef4444);
  }

  .docker-banner {
    padding: 6px 14px;
    font-size: 12px;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-table-wrap {
    flex: 1;
    min-height: 0;
    overflow: auto;
  }

  .docker-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 12px;
  }

  .docker-table th {
    position: sticky;
    top: 0;
    z-index: 1;
    text-align: left;
    padding: 8px 14px;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
    color: var(--text-secondary);
    background: var(--glass-bg-strong);
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-table td {
    padding: 7px 14px;
    border-bottom: 1px solid var(--glass-border);
    max-width: 280px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .docker-table tbody tr:hover {
    background: var(--glass-bg);
  }

  .docker-cell-name {
    font-weight: 600;
  }

  .docker-dot {
    display: inline-block;
    margin-right: 8px;
    background: var(--text-secondary);
    opacity: 0.5;
  }

  .docker-dot--up {
    background: #22c55e;
    opacity: 1;
  }

  .docker-mono {
    font-family: var(--terminal-font-family);
  }

  .docker-table tbody tr {
    cursor: pointer;
  }

  .docker-row--selected {
    background: color-mix(in srgb, var(--ops-signal) 12%, transparent);
  }

  .docker-row--busy {
    opacity: 0.55;
    pointer-events: none;
  }

  .docker-notice {
    color: var(--ops-signal);
  }

  .docker-col-actions {
    width: 56px;
    text-align: center !important;
  }

  .docker-more {
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-size: 16px;
    line-height: 1;
    width: 28px;
    height: 24px;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--text-secondary);
  }

  .docker-more:hover:not(:disabled) {
    border-color: var(--glass-border);
    color: var(--text-primary);
    background: var(--glass-bg);
  }

  .docker-more:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 1px;
  }

  .docker-menu {
    pointer-events: auto;
    color: var(--text-primary);
  }

  .docker-menu-item {
    display: block;
    width: 100%;
    min-height: 30px;
    text-align: left;
    padding: 5px 12px;
    color: var(--text-primary);
    background: transparent;
    border: 0;
    cursor: pointer;
  }

  .docker-menu-item:hover:not(:disabled) {
    background: color-mix(in srgb, var(--ops-signal) 16%, transparent);
  }

  .docker-menu-item:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .docker-menu-item.is-danger {
    color: var(--ops-alert);
  }

  .docker-force {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--text-secondary);
    cursor: pointer;
  }
</style>
