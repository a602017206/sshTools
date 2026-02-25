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
    <p class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
      {message}
    </p>

    <div class="flex gap-2 pt-2">
      <button
        type="button"
        on:click={onCancel}
        class="px-3 py-1.5 bg-slate-100 dark:bg-slate-700 hover:bg-slate-200 dark:hover:bg-slate-600 text-slate-700 dark:text-slate-200 rounded-md text-xs font-medium transition-colors focus-visible:outline-none focus-visible:ring-2"
      >
        {cancelText}
      </button>
      <button
        type="button"
        on:click={onConfirm}
        class="px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm {
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
