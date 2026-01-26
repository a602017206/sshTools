<script>
  import { createEventDispatcher } from 'svelte';

  export let transfer;

  const dispatch = createEventDispatcher();

  // Format bytes
  function formatBytes(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
  }

  // Format speed
  function formatSpeed(bytesPerSec) {
    return formatBytes(bytesPerSec) + '/s';
  }

  // Get filename from path
  function getFilename(path) {
    return path.split('/').pop() || path;
  }

  // Get status color
  function getStatusColor(status) {
    switch (status) {
      case 'completed':
        return 'var(--accent-success)';
      case 'failed':
        return 'var(--accent-error)';
      case 'cancelled':
        return 'var(--text-secondary)';
      default:
        return 'var(--accent-primary)';
    }
  }

  // Get status icon
  function getStatusIcon(status) {
    switch (status) {
      case 'completed':
        return '✓';
      case 'failed':
        return '✗';
      case 'cancelled':
        return '⊘';
      default:
        return '';
    }
  }

  function handleCancel() {
    dispatch('cancel', transfer.transfer_id);
  }
</script>

<div class="transfer-item" class:completed={transfer.status === 'completed'}>
  <div class="transfer-header">
    <div class="transfer-info">
      <span class="transfer-filename">{getFilename(transfer.filename)}</span>
      {#if transfer.status === 'running'}
        <span class="transfer-speed">{formatSpeed(transfer.speed)}</span>
      {/if}
    </div>
    <div class="transfer-actions">
      {#if transfer.status === 'running'}
        <button class="btn-cancel" on:click={handleCancel} title="取消传输">
          ✕
        </button>
      {:else}
        <span class="status-icon" style="color: {getStatusColor(transfer.status)}">
          {getStatusIcon(transfer.status)}
        </span>
      {/if}
    </div>
  </div>

  <div class="progress-bar-container">
    <div
      class="progress-bar-fill"
      style="width: {transfer.percentage}%; background: {getStatusColor(transfer.status)};"
    ></div>
  </div>

  <div class="transfer-footer">
    <span class="transfer-progress">
      {formatBytes(transfer.bytes_sent)} / {formatBytes(transfer.total_bytes)}
    </span>
    <span class="transfer-percentage">{transfer.percentage.toFixed(1)}%</span>
  </div>

  {#if transfer.status === 'failed' && transfer.error}
    <div class="transfer-error">
      {transfer.error}
    </div>
  {/if}
</div>

<style>
  .transfer-item {
    padding: 12px;
    background: var(--bg-tertiary);
    border-radius: 4px;
    margin-bottom: 8px;
    border: 1px solid var(--border-primary);
  }

  .transfer-item.completed {
    opacity: 0.7;
  }

  .transfer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .transfer-info {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .transfer-filename {
    font-size: 12px;
    color: var(--text-primary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .transfer-speed {
    font-size: 11px;
    color: var(--text-secondary);
    flex-shrink: 0;
  }

  .transfer-actions {
    display: flex;
    align-items: center;
    margin-left: 8px;
  }

  .btn-cancel {
    padding: 2px 6px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 14px;
    line-height: 1;
  }

  .btn-cancel:hover {
    color: var(--accent-error);
    background: var(--bg-hover);
    border-radius: 2px;
  }

  .status-icon {
    font-size: 14px;
    font-weight: bold;
  }

  .progress-bar-container {
    height: 4px;
    background: var(--bg-secondary);
    border-radius: 2px;
    overflow: hidden;
    margin-bottom: 6px;
  }

  .progress-bar-fill {
    height: 100%;
    transition: width 0.3s ease;
    border-radius: 2px;
  }

  .transfer-footer {
    display: flex;
    justify-content: space-between;
    font-size: 11px;
    color: var(--text-secondary);
  }

  .transfer-error {
    margin-top: 8px;
    padding: 6px 8px;
    background: rgba(244, 67, 54, 0.1);
    border-left: 2px solid var(--accent-error);
    font-size: 11px;
    color: var(--accent-error);
    border-radius: 2px;
  }
</style>
