<script>
  import { resourceDetailFields } from '../lib/dockerResources.js';

  /** images | networks | volumes */
  export let kind = 'images';
  export let title = '';
  /** Inspect 结果或列表行（可能为 null：加载中） */
  export let detail = null;
  export let detailError = '';
  export let onClose = () => {};

  const KIND_LABELS = { images: '镜像', networks: '网络', volumes: '卷' };

  $: kindLabel = KIND_LABELS[kind] || '资源';
  $: fields = resourceDetailFields(kind, detail);
</script>

<aside class="docker-res-detail" aria-label={`${kindLabel}详情`}>
  <header class="docker-res-head">
    <div class="docker-res-title">
      <span class="docker-res-name" title={title}>{title || '-'}</span>
      <span class="docker-res-kind">{kindLabel}</span>
    </div>
    <button type="button" class="docker-res-close" title="关闭详情" aria-label="关闭详情" on:click={onClose}>✕</button>
  </header>

  <div class="docker-res-body">
    {#if detailError}
      <p class="docker-res-error">详情加载失败：{detailError}</p>
    {/if}
    {#if !detail}
      {#if !detailError}<p class="docker-res-muted" role="status">加载中…</p>{/if}
    {:else if fields.length === 0}
      <p class="docker-res-muted">暂无详情</p>
    {:else}
      <dl class="docker-res-kv">
        {#each fields as field (field.label)}
          <dt>{field.label}</dt><dd class="docker-res-mono">{field.value}</dd>
        {/each}
      </dl>
    {/if}
  </div>
</aside>

<style>
  .docker-res-detail {
    flex: 1 1 auto;
    width: 100%;
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
    background: var(--glass-bg-strong);
    color: var(--text-primary);
    font-size: 12px;
  }

  .docker-res-head {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    border-bottom: 1px solid var(--glass-border);
  }

  .docker-res-title {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: baseline;
    gap: 8px;
  }

  .docker-res-name {
    font-size: 14px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .docker-res-kind {
    flex-shrink: 0;
    font-size: 10px;
    color: var(--text-secondary);
  }

  .docker-res-close {
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

  .docker-res-close:hover {
    border-color: var(--glass-border);
    color: var(--text-primary);
    background: var(--glass-bg);
  }

  .docker-res-close:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 1px;
  }

  .docker-res-body {
    flex: 1;
    min-height: 0;
    overflow: auto;
    padding: 12px;
  }

  .docker-res-kv {
    display: grid;
    grid-template-columns: 72px 1fr;
    gap: 6px 10px;
    margin: 0;
  }

  .docker-res-kv dt {
    color: var(--text-secondary);
  }

  .docker-res-kv dd {
    margin: 0;
    min-width: 0;
    word-break: break-all;
  }

  .docker-res-mono {
    font-family: var(--terminal-font-family);
  }

  .docker-res-muted {
    margin: 0 0 12px;
    color: var(--text-secondary);
  }

  .docker-res-error {
    margin: 0 0 12px;
    color: var(--ops-danger, #ef4444);
    word-break: break-word;
  }
</style>
