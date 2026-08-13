<script>
  import Dialog from './ui/Dialog.svelte';
  import { uploadStore, activeTransfers, completedTransfers, formatFileSize, formatSpeed, getTransferPercentage } from '../stores/uploadStore.js';
  import { CancelTransfer } from '../../wailsjs/go/main/App.js';
</script>

<Dialog
  isOpen={$uploadStore.isPanelOpen}
  onClose={() => uploadStore.setPanelOpen(false)}
  title="上传任务"
  size="md"
>
  <div class="w-full h-[min(620px,calc(100vh-10rem))] flex flex-col" style="color: var(--text-primary);">
    <div class="p-4 border-b" style="border-color: var(--glass-border);">
      <div class="flex gap-2 p-1 rounded-full" style="border: 1px solid var(--glass-border); background: var(--glass-bg);">
        <button on:click={() => uploadStore.setActiveTab('active')} class="flex-1 py-1.5 px-3 text-xs font-medium rounded-full transition-colors {$uploadStore.activeTab === 'active' ? 'text-white' : 'ops-muted'}" style={$uploadStore.activeTab === 'active' ? 'background: var(--ops-signal);' : ''}>
          进行中 ({$activeTransfers.length})
        </button>
        <button on:click={() => uploadStore.setActiveTab('history')} class="flex-1 py-1.5 px-3 text-xs font-medium rounded-full transition-colors {$uploadStore.activeTab === 'history' ? 'text-white' : 'ops-muted'}" style={$uploadStore.activeTab === 'history' ? 'background: var(--ops-signal);' : ''}>
          历史 ({$completedTransfers.length})
        </button>
      </div>
    </div>

    {#if $uploadStore.activeTab === 'active' && $activeTransfers.length === 0}
      <div class="flex-1 flex flex-col items-center justify-center ops-muted gap-3"><span class="text-sm">暂无上传任务</span></div>
    {:else if $uploadStore.activeTab === 'active'}
      <div class="flex-1 overflow-y-auto p-4 space-y-3">
        {#each $activeTransfers as transfer (transfer.id)}
          <div class="rounded-xl border p-3" style="background: var(--glass-bg); border-color: var(--glass-border);">
            <div class="flex items-center justify-between mb-2"><div class="flex-1 min-w-0"><div class="text-sm font-medium truncate" title={transfer.filename}>{transfer.filename}</div><div class="text-xs ops-muted mt-0.5">{formatFileSize(transfer.bytesSent)} / {formatFileSize(transfer.totalBytes)} {#if transfer.speed}• {formatSpeed(transfer.speed)}{/if}</div></div><button on:click={async () => { await CancelTransfer(transfer.id); uploadStore.cancelTransfer(transfer.id); }} class="p-1.5 rounded text-red-500" title="取消上传">×</button></div>
            <div class="h-2 rounded-full overflow-hidden" style="background: color-mix(in srgb, var(--glass-bg) 60%, #94a3b8);"><div class="h-full transition-all duration-300" style={`width: ${Math.min(100, Math.max(0, getTransferPercentage(transfer)))}%; background: var(--ops-signal);`}></div></div>
          </div>
        {/each}
      </div>
    {:else if $completedTransfers.length === 0}
      <div class="flex-1 flex flex-col items-center justify-center ops-muted gap-3"><span class="text-sm">暂无历史记录</span></div>
    {:else}
      <div class="flex-1 flex flex-col"><div class="p-3 border-b" style="border-color: var(--glass-border);"><button on:click={() => uploadStore.clearCompleted()} class="w-full py-2 px-3 text-xs font-medium rounded-lg text-red-600 dark:text-red-400" style="background: color-mix(in srgb, var(--ops-alert) 8%, transparent);">清空历史记录</button></div><div class="flex-1 overflow-y-auto p-4 space-y-3">{#each $completedTransfers as transfer (transfer.id)}<div class="rounded-xl border p-3" style="background: var(--glass-bg); border-color: var(--glass-border);"><div class="flex items-center justify-between"><div class="flex-1 min-w-0"><div class="text-sm font-medium truncate" title={transfer.filename}>{transfer.filename}</div><div class="text-xs ops-muted mt-0.5 flex items-center gap-2"><span>{transfer.status === 'completed' ? '完成' : transfer.status === 'failed' ? '失败' : '已取消'}</span><span>•</span><span>{formatFileSize(transfer.totalBytes)}</span></div>{#if transfer.status === 'failed' && transfer.error}<div class="text-xs text-red-500 dark:text-red-400 mt-1 truncate" title={transfer.error}>{transfer.error}</div>{/if}</div><button on:click={() => uploadStore.removeTransfer(transfer.id)} class="ops-icon-button p-1.5 rounded transition-colors ml-3" title="删除记录">×</button></div></div>{/each}</div></div>
    {/if}
  </div>
</Dialog>
