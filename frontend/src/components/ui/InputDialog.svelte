<script>
  import Dialog from './Dialog.svelte';

  export let isOpen = false;
  export let title = '';
  export let message = '';
  export let defaultValue = '';
  export let placeholder = '';
  export let inputType = 'text';
  export let allowEmpty = false;
  export let trimValue = true;
  export let confirmText = '确定';
  export let cancelText = '取消';
  export let onConfirm = () => {};
  export let onCancel = () => {};

  let inputValue = '';
  let inputElement;

  // Reset input and focus when dialog opens
  $: if (isOpen) {
    inputValue = defaultValue;
  }

  // Focus input after dialog opens
  $: if (isOpen && inputElement) {
    setTimeout(() => {
      inputElement?.focus();
      inputElement?.select();
    }, 0);
  }

  function handleConfirm() {
    if (allowEmpty) {
      onConfirm(trimValue ? inputValue.trim() : inputValue);
      return;
    }

    const value = trimValue ? inputValue.trim() : inputValue;
    if (value) {
      onConfirm(value);
    }
  }

  function handleKeydown(event) {
    if (event.key === 'Enter') {
      handleConfirm();
    }
  }
</script>

<Dialog
  bind:isOpen={isOpen}
  onClose={onCancel}
  title={title}
  size="sm"
>
  <div class="space-y-4">
    {#if message}
      <p class="text-sm text-gray-700 dark:text-gray-300 leading-relaxed">
        {message}
      </p>
    {/if}

    {#if inputType === 'password'}
      <input
        type="password"
        bind:this={inputElement}
        bind:value={inputValue}
        {placeholder}
        on:keydown={handleKeydown}
        class="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
      />
    {:else}
      <input
        type="text"
        bind:this={inputElement}
        bind:value={inputValue}
        {placeholder}
        on:keydown={handleKeydown}
        class="w-full px-4 py-2.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm text-gray-900 dark:text-white placeholder-gray-400 dark:placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent transition-all"
      />
    {/if}

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
        on:click={handleConfirm}
        disabled={!allowEmpty && !inputValue.trim()}
        class="px-3 py-1.5 bg-purple-600 hover:bg-purple-700 text-white rounded-md text-xs font-medium transition-all shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {confirmText}
      </button>
    </div>
  </div>
</Dialog>
