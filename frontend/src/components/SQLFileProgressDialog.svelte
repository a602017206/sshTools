<script>
  export let progress = null;
  export let onCancel = () => {};
  export let onDismiss = () => {};

  $: fileName = progress?.fileName || '';
  $: fileSize = Number(progress?.fileSize || 0);
  $: bytesRead = Number(progress?.bytesRead || 0);
  $: statements = Number(progress?.statements || 0);
  $: affected = Number(progress?.affected || 0);
  $: done = Boolean(progress?.done);
  $: canceled = Boolean(progress?.canceled);
  $: errorMessage = progress?.error || '';
  $: failedSql = progress?.failedSql || '';
  $: percent = Number(progress?.percent || 0);

  function formatSize(value) {
    const bytes = Number(value) || 0;
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  }
</script>

{#if progress}
  <div class="sql-file-progress" role="status">
    <div class="sql-file-progress__card">
      <h3>{done ? (errorMessage && !canceled ? '执行失败' : (canceled ? '已取消' : '执行完成')) : '正在运行 SQL 文件'}</h3>
      <p class="sql-file-progress__name" title={fileName}>{fileName}</p>
      <div class="sql-file-progress__bar" aria-valuemin="0" aria-valuemax="100" aria-valuenow={percent}>
        <span style={`width:${percent}%`}></span>
      </div>
      <p class="sql-file-progress__meta">{formatSize(bytesRead)} / {formatSize(fileSize)} · {statements} 条 · 影响 {affected} 行</p>
      {#if errorMessage}
        <p class="sql-file-progress__error">{errorMessage}</p>
      {/if}
      {#if failedSql}
        <pre class="sql-file-progress__sql">{failedSql}</pre>
      {/if}
      <div class="sql-file-progress__actions">
        {#if !done}
          <button type="button" on:click={onCancel}>取消</button>
        {:else}
          <button type="button" on:click={onDismiss}>关闭</button>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  .sql-file-progress {
    position: fixed;
    inset: 0;
    z-index: 140;
    display: grid;
    place-items: center;
    background: color-mix(in srgb, #020617 36%, transparent);
  }
  .sql-file-progress__card {
    width: min(420px, calc(100vw - 32px));
    padding: 16px 18px;
    border-radius: 14px;
    border: 1px solid var(--glass-border);
    background: var(--bg-primary);
    color: var(--text-primary);
    box-shadow: 0 18px 40px rgba(0,0,0,.22);
  }
  .sql-file-progress__card h3 { margin: 0; font-size: 14px; }
  .sql-file-progress__name, .sql-file-progress__meta { margin: 8px 0 0; font-size: 12px; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .sql-file-progress__bar { margin-top: 12px; height: 8px; border-radius: 999px; background: var(--bg-secondary); overflow: hidden; }
  .sql-file-progress__bar span { display: block; height: 100%; background: var(--accent-primary, #0f766e); }
  .sql-file-progress__error { margin: 10px 0 0; color: #dc2626; font-size: 12px; }
  .sql-file-progress__sql { margin: 8px 0 0; max-height: 120px; overflow: auto; padding: 8px; border-radius: 8px; background: var(--bg-secondary); font: 11px ui-monospace, SFMono-Regular, Menlo, monospace; }
  .sql-file-progress__actions { display: flex; justify-content: flex-end; margin-top: 14px; }
  .sql-file-progress__actions button { padding: 6px 12px; border-radius: 8px; border: 1px solid var(--glass-border); background: transparent; color: inherit; cursor: pointer; }
</style>
