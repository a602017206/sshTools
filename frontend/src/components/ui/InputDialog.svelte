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
      <p class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
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
        class="w-full px-4 py-2.5 bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-lg text-sm text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus-visible:ring-2 focus:border-transparent transition-all"
      />
    {:else}
      <input
        type="text"
        bind:this={inputElement}
        bind:value={inputValue}
        {placeholder}
        on:keydown={handleKeydown}
        class="w-full px-4 py-2.5 bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 rounded-lg text-sm text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus-visible:ring-2 focus:border-transparent transition-all"
      />
    {/if}

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
        on:click={handleConfirm}
        disabled={!allowEmpty && !inputValue.trim()}
        class="px-3 py-1.5 accent-bg accent-bg-hover text-white rounded-md text-xs font-medium transition-all shadow-sm focus-visible:outline-none focus-visible:ring-2 disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {confirmText}
      </button>
    </div>
  </div>
</Dialog>
