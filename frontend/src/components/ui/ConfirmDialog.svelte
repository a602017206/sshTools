<script>
  import Dialog from './Dialog.svelte';

  export let isOpen = false;
  export let title = '';
  export let message = '';
  export let confirmText = '确定';
  export let cancelText = '取消';
  export let onConfirm = () => {};
  export let onCancel = () => {};
  export let type = 'default'; // default, warning, danger
</script>

<Dialog
  bind:isOpen={isOpen}
  onClose={onCancel}
  title={title}
  size="sm"
>
  <div class="space-y-4">
    <p class="text-sm leading-relaxed" style="color: var(--text-secondary);">
      {message}
    </p>

    <div class="flex gap-2 pt-2">
      <button
        type="button"
        on:click={onCancel}
        class="ops-btn-glass px-3 py-1.5 rounded-lg text-xs font-medium transition-all focus-visible:outline-none focus-visible:ring-2"
      >
        {cancelText}
      </button>
      <button
        type="button"
        on:click={onConfirm}
        class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm {
          type === 'danger'
            ? 'bg-rose-600 hover:bg-rose-700 text-white'
            : type === 'warning'
            ? 'bg-amber-500 hover:bg-amber-600 text-white'
            : 'accent-bg accent-bg-hover text-white'
        }"
      >
        {confirmText}
      </button>
    </div>
  </div>
</Dialog>
