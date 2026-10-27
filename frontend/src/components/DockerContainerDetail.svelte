<script>
  import { onDestroy, tick } from 'svelte';
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import { containerField, formatContainerCreated, isContainerRunning } from '../lib/dockerWorkspace.js';
  import {
    DEFAULT_LOG_TAIL,
    LOG_TAIL_OPTIONS,
    appendLogText,
    asStringList,
    createThrottledAppender,
    dockerLogsEventName,
    isNearBottom,
    logChunkText,
    normalizeLogTail
  } from '../lib/dockerLogs.js';

  export let sessionId = '';
  export let containerId = '';
  /** Inspect 结果（可能仍为 null：加载中） */
  export let detail = null;
  export let detailError = '';
  export let onClose = () => {};
  /** 打开 Exec 终端页签 */
  export let onExec = () => {};

  let activeTab = 'overview'; // overview | logs
  let tail = DEFAULT_LOG_TAIL;
  let follow = false;
  let logText = '';
  let logError = '';
  let logLoading = false;
  let logEl = null;
  let autoScroll = true;
  let copied = false;
  let copiedTimer = null;
  let unsubscribe = null;
  let followActive = false; // Start 已发起（含进行中），destroy 时据此 Stop
  let destroyed = false;
  let logToken = 0; // 使过期的异步结果失效

  const appender = createThrottledAppender((text) => {
    logText = appendLogText(logText, text);
    scrollToBottomIfNeeded();
  });

  $: shortId = String(containerId).slice(0, 12);
  $: title = containerField(detail, 'Name') || shortId;
  $: running = isContainerRunning(detail);
  $: mounts = asStringList(detail?.Mounts ?? detail?.mounts);
  $: envList = asStringList(detail?.Env ?? detail?.env);
  $: cmdText = asStringList(detail?.Cmd ?? detail?.cmd).join(' ');
  $: entrypointText = asStringList(detail?.Entrypoint ?? detail?.entrypoint).join(' ');

  function bindings() {
    return window.wailsBindings || {};
  }

  function errText(error) {
    return String(error?.message || error || '未知错误');
  }

  async function scrollToBottomIfNeeded() {
    if (!autoScroll) return;
    await tick();
    if (logEl && autoScroll) logEl.scrollTop = logEl.scrollHeight;
  }

  function onLogScroll() {
    if (!logEl) return;
    autoScroll = isNearBottom({
      scrollTop: logEl.scrollTop,
      scrollHeight: logEl.scrollHeight,
      clientHeight: logEl.clientHeight
    });
  }

  function resumeAutoScroll() {
    autoScroll = true;
    scrollToBottomIfNeeded();
  }

  function detachListener() {
    if (unsubscribe) {
      unsubscribe();
      unsubscribe = null;
    }
  }

  /** 停止跟流（幂等）：先摘除监听并丢弃节流缓存，再通知后端。 */
  async function stopFollow() {
    logToken += 1;
    detachListener();
    appender.flush();
    const wasActive = followActive;
    followActive = false;
    follow = false;
    if (!wasActive) return;
    const { StopDockerContainerLogs } = bindings();
    if (typeof StopDockerContainerLogs !== 'function') return;
    try {
      await StopDockerContainerLogs(sessionId, containerId);
    } catch (error) {
      if (!destroyed) logError = errText(error);
    }
  }

  async function loadLogs() {
    const { GetDockerContainerLogs } = bindings();
    const token = ++logToken;
    appender.cancel();
    logError = '';
    if (typeof GetDockerContainerLogs !== 'function') {
      logError = '日志功能不可用';
      return;
    }
    logLoading = true;
    try {
      const text = await GetDockerContainerLogs(sessionId, containerId, tail);
      if (destroyed || token !== logToken) return;
      logText = appendLogText('', text || '');
      autoScroll = true;
      scrollToBottomIfNeeded();
    } catch (error) {
      if (destroyed || token !== logToken) return;
      logError = errText(error);
    } finally {
      if (token === logToken) logLoading = false;
    }
  }

  async function startFollow() {
    const { StartDockerContainerLogs } = bindings();
    if (typeof StartDockerContainerLogs !== 'function') {
      follow = false;
      logError = '日志跟流功能不可用';
      return;
    }
    const token = ++logToken;
    logError = '';
    // 跟流自带尾部 N 行，先清空避免与一次性拉取的内容重复
    appender.cancel();
    logText = '';
    autoScroll = true;
    detachListener();
    unsubscribe = EventsOn(dockerLogsEventName(sessionId, containerId), (payload) => {
      if (destroyed) return;
      appender.push(logChunkText(payload));
    });
    followActive = true;
    try {
      await StartDockerContainerLogs(sessionId, containerId, tail);
      if (destroyed || token !== logToken) return;
    } catch (error) {
      if (destroyed || token !== logToken) return;
      detachListener();
      followActive = false;
      follow = false;
      logError = errText(error);
    }
  }

  async function toggleFollow() {
    if (follow) await stopFollow();
    else {
      follow = true;
      await startFollow();
    }
  }

  async function selectTab(tab) {
    if (tab === activeTab) return;
    if (activeTab === 'logs') await stopFollow();
    activeTab = tab;
    if (tab === 'logs') {
      logText = '';
      await loadLogs();
    }
  }

  async function onTailChange() {
    tail = normalizeLogTail(tail);
    if (follow) {
      await stopFollow();
      follow = true;
      await startFollow();
    } else {
      await loadLogs();
    }
  }

  function clearLogs() {
    appender.cancel();
    logText = '';
  }

  async function copyLogs() {
    appender.flush();
    try {
      await navigator.clipboard.writeText(logText);
      copied = true;
      if (copiedTimer) clearTimeout(copiedTimer);
      copiedTimer = setTimeout(() => {
        copied = false;
      }, 1500);
    } catch (error) {
      logError = `复制失败：${errText(error)}`;
    }
  }

  onDestroy(() => {
    destroyed = true;
    if (copiedTimer) clearTimeout(copiedTimer);
    appender.cancel();
    detachListener();
    if (followActive) {
      followActive = false;
      const { StopDockerContainerLogs } = bindings();
      if (typeof StopDockerContainerLogs === 'function') {
        Promise.resolve(StopDockerContainerLogs(sessionId, containerId)).catch(() => {});
      }
    }
  });
</script>

<aside class="docker-detail" aria-label="容器详情">
  <header class="docker-detail-head">
    <span class="ops-status-dot docker-detail-dot" class:docker-detail-dot--up={running}></span>
    <div class="docker-detail-title">
      <span class="docker-detail-name" title={title}>{title}</span>
      <span class="docker-detail-id">{shortId}</span>
    </div>
    <button
      type="button"
      class="docker-detail-exec"
      title={running ? '在容器内打开终端' : '仅运行中的容器可 Exec'}
      disabled={!running}
      on:click={onExec}
    >Exec</button>
    <button type="button" class="docker-detail-close" title="关闭详情" aria-label="关闭详情" on:click={onClose}>✕</button>
  </header>

  <div class="docker-detail-tabs" role="tablist">
    <button
      type="button"
      role="tab"
      class="docker-detail-tab"
      class:is-active={activeTab === 'overview'}
      aria-selected={activeTab === 'overview'}
      on:click={() => selectTab('overview')}
    >概览</button>
    <button
      type="button"
      role="tab"
      class="docker-detail-tab"
      class:is-active={activeTab === 'logs'}
      aria-selected={activeTab === 'logs'}
      on:click={() => selectTab('logs')}
    >日志</button>
  </div>

  {#if activeTab === 'overview'}
    <div class="docker-detail-body">
      {#if detailError}
        <p class="docker-detail-error">详情加载失败：{detailError}</p>
      {:else if !detail}
        <p class="docker-detail-muted" role="status">加载中…</p>
      {:else}
        <dl class="docker-kv">
          <dt>状态</dt><dd>{containerField(detail, 'Status') || containerField(detail, 'State') || '-'}</dd>
          <dt>镜像</dt><dd class="docker-detail-mono">{containerField(detail, 'Image') || '-'}</dd>
          <dt>ID</dt><dd class="docker-detail-mono">{containerField(detail, 'ID') || containerId}</dd>
          <dt>创建时间</dt><dd>{formatContainerCreated(containerField(detail, 'Created'))}</dd>
          <dt>端口</dt><dd class="docker-detail-mono">{containerField(detail, 'Ports') || '-'}</dd>
          <dt>网络</dt><dd class="docker-detail-mono">{detail.NetworkMode || detail.network_mode || '-'}</dd>
        </dl>

        <h3 class="docker-detail-section">挂载</h3>
        {#if mounts.length === 0}
          <p class="docker-detail-muted">无挂载</p>
        {:else}
          <ul class="docker-detail-list">
            {#each mounts as mount}
              <li class="docker-detail-mono">{mount}</li>
            {/each}
          </ul>
        {/if}

        <details class="docker-fold">
          <summary>命令 / Entrypoint</summary>
          <dl class="docker-kv">
            <dt>Cmd</dt><dd class="docker-detail-mono">{cmdText || '-'}</dd>
            <dt>Entrypoint</dt><dd class="docker-detail-mono">{entrypointText || '-'}</dd>
          </dl>
        </details>

        <details class="docker-fold">
          <summary>环境变量（{envList.length}）</summary>
          {#if envList.length === 0}
            <p class="docker-detail-muted">无环境变量</p>
          {:else}
            <ul class="docker-detail-list">
              {#each envList as item}
                <li class="docker-detail-mono">{item}</li>
              {/each}
            </ul>
          {/if}
        </details>
      {/if}
    </div>
  {:else}
    <div class="docker-logs">
      <div class="docker-logs-bar">
        <label class="docker-logs-field">
          尾部
          <select class="docker-logs-select" bind:value={tail} on:change={onTailChange}>
            {#each LOG_TAIL_OPTIONS as option}
              <option value={option}>{option} 行</option>
            {/each}
          </select>
        </label>
        <label class="docker-logs-field docker-logs-follow">
          <input type="checkbox" checked={follow} on:change={toggleFollow} />
          跟流
        </label>
        <button type="button" class="docker-logs-btn" on:click={clearLogs}>清空</button>
        <button type="button" class="docker-logs-btn" disabled={!logText} on:click={copyLogs}>
          {copied ? '已复制' : '复制'}
        </button>
      </div>
      {#if logError}<div class="docker-detail-error docker-logs-error">{logError}</div>{/if}
      <div class="docker-logs-wrap">
        <pre class="docker-logs-view" bind:this={logEl} on:scroll={onLogScroll}>{logText || (logLoading ? '加载中…' : follow ? '等待日志输出…' : '暂无日志')}</pre>
        {#if !autoScroll}
          <button type="button" class="docker-logs-resume" on:click={resumeAutoScroll}>已暂停滚动 · 回到底部</button>
        {/if}
      </div>
    </div>
  {/if}
</aside>

<style>
  .docker-detail {
    flex: 1 1 auto;
    width: 100%;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    border-left: none; /* 左边框由外层 docker-detail-shell 承担，避免双线 */
    background: var(--glass-bg-strong);
    color: var(--text-primary);
    font-size: 12px;
  }

  .docker-detail-head {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-detail-dot {
    flex-shrink: 0;
    background: var(--text-secondary);
    opacity: 0.5;
  }

  .docker-detail-dot--up {
    background: #22c55e;
    opacity: 1;
  }

  .docker-detail-title {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: baseline;
    gap: 8px;
  }

  .docker-detail-name {
    font-size: 14px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .docker-detail-id {
    flex-shrink: 0;
    font-size: 10px;
    color: var(--text-secondary);
    font-family: var(--terminal-font-family);
  }

  .docker-detail-exec {
    appearance: none;
    cursor: pointer;
    flex-shrink: 0;
    height: 26px;
    padding: 0 10px;
    border-radius: 8px;
    border: 1px solid var(--glass-border);
    background: transparent;
    color: var(--text-primary);
    font: inherit;
    font-size: 12px;
  }

  .docker-detail-exec:hover:not(:disabled) {
    background: var(--glass-bg-hover, rgba(127, 127, 127, 0.12));
  }

  .docker-detail-exec:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .docker-detail-close {
    appearance: none;
    cursor: pointer;
    flex-shrink: 0;
    width: 26px;
    height: 26px;
    border-radius: 8px;
    border: 1px solid transparent;
    background: transparent;
    color: var(--text-secondary);
    font: inherit;
  }

  .docker-detail-close:hover {
    border-color: var(--glass-border);
    color: var(--text-primary);
    background: var(--glass-bg);
  }

  .docker-detail-close:focus-visible,
  .docker-detail-tab:focus-visible,
  .docker-logs-btn:focus-visible,
  .docker-logs-resume:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 1px;
  }

  .docker-detail-tabs {
    display: flex;
    gap: 4px;
    padding: 6px 12px 0;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-detail-tab {
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-weight: 600;
    padding: 6px 14px;
    border: 0;
    border-bottom: 2px solid transparent;
    background: transparent;
    color: var(--text-secondary);
  }

  .docker-detail-tab:hover {
    color: var(--text-primary);
  }

  .docker-detail-tab.is-active {
    color: var(--ops-signal);
    border-bottom-color: var(--ops-signal);
  }

  .docker-detail-body {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 12px;
  }

  .docker-kv {
    display: grid;
    grid-template-columns: 72px 1fr;
    gap: 6px 10px;
    margin: 0 0 12px;
  }

  .docker-kv dt {
    color: var(--text-secondary);
  }

  .docker-kv dd {
    margin: 0;
    min-width: 0;
    word-break: break-all;
  }

  .docker-detail-mono {
    font-family: var(--terminal-font-family);
  }

  .docker-detail-section {
    margin: 0 0 6px;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.06em;
    color: var(--text-secondary);
  }

  .docker-detail-list {
    list-style: none;
    margin: 0 0 12px;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .docker-detail-list li {
    padding: 4px 8px;
    border-radius: 8px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    word-break: break-all;
  }

  .docker-fold {
    margin-bottom: 10px;
  }

  .docker-fold summary {
    cursor: pointer;
    padding: 4px 0;
    font-weight: 600;
    color: var(--text-secondary);
  }

  .docker-fold summary:hover {
    color: var(--text-primary);
  }

  .docker-detail-muted {
    margin: 0 0 12px;
    color: var(--text-secondary);
  }

  .docker-detail-error {
    margin: 0;
    color: var(--ops-danger, #ef4444);
    word-break: break-word;
  }

  .docker-logs {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .docker-logs-bar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-logs-field {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: var(--text-secondary);
  }

  .docker-logs-follow {
    cursor: pointer;
  }

  .docker-logs-select {
    font: inherit;
    padding: 2px 6px;
    border-radius: 8px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    color: var(--text-primary);
  }

  .docker-logs-btn {
    appearance: none;
    cursor: pointer;
    font: inherit;
    font-weight: 600;
    min-height: 26px;
    padding: 0 10px;
    border-radius: 8px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    color: var(--text-primary);
  }

  .docker-logs-btn:hover:not(:disabled) {
    border-color: var(--ops-signal);
    color: var(--ops-signal);
  }

  .docker-logs-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .docker-logs-error {
    padding: 6px 12px;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-logs-wrap {
    position: relative;
    flex: 1;
    min-height: 0;
    display: flex;
  }

  .docker-logs-view {
    flex: 1;
    min-width: 0;
    margin: 0;
    padding: 8px 12px;
    overflow: auto;
    font-family: var(--terminal-font-family);
    font-size: 11px;
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-all;
    user-select: text;
  }

  .docker-logs-resume {
    appearance: none;
    cursor: pointer;
    position: absolute;
    right: 14px;
    bottom: 10px;
    font: inherit;
    font-size: 11px;
    font-weight: 600;
    padding: 4px 10px;
    border-radius: 999px;
    border: 1px solid var(--ops-signal);
    background: var(--glass-bg-strong);
    color: var(--ops-signal);
    box-shadow: var(--shadow-glass);
  }
</style>
