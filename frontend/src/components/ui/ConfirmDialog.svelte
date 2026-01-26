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
    <p class="text-sm text-gray-700 dark:text-gray-300 leading-relaxed">
      {message}
    </p>

    <div class="flex gap-2 pt-2">
      <button
        type="button"
        on:click={onCancel}
        class="px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-md text-xs font-medium transition-colors"
      >
        {cancelText}
      </button>
      <button
        type="button"
        on:click={onConfirm}
        class="flex-1 px-3 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm {
          type === 'danger'
            ? 'bg-red-600 hover:bg-red-700 text-white'
            : type === 'warning'
            ? 'bg-yellow-500 hover:bg-yellow-600 text-white'
            : 'bg-purple-600 hover:bg-purple-700 text-white'
        }"
      >
        {confirmText}
      </button>
    </div>
  </div>
</Dialog>
