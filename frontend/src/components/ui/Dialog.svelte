<script>
  export let isOpen = false;
  export let onClose = () => {};
  export let title = '';
  export let size = 'md'; // sm, md, lg, xl

  const sizeClasses = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-2xl',
    xl: 'max-w-5xl'
  };

  $: dialogWidthClass = sizeClasses[size] || sizeClasses.md;

  function handleBackdropClick(event) {
    if (event.target === event.currentTarget) {
      onClose();
    }
  }

  function handleEscapeKey(event) {
    if (event.key === 'Escape' && isOpen) {
      onClose();
    }
  }

  $: if (isOpen) {
    document.addEventListener('keydown', handleEscapeKey);
  } else {
    document.removeEventListener('keydown', handleEscapeKey);
  }
</script>

{#if isOpen}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center"
    on:click={handleBackdropClick}
    role="dialog"
    aria-modal="true"
  >
    <!-- 背景遮罩 -->
    <div class="absolute inset-0 bg-slate-900/35 dark:bg-black/45 backdrop-blur-sm transition-opacity pointer-events-auto" />
    
    <!-- 对话框 -->
    <div class={`relative w-full ${dialogWidthClass} bg-white dark:bg-slate-800 rounded-xl shadow-2xl border border-slate-200 dark:border-slate-700 m-4 max-h-[92vh] overflow-hidden flex flex-col transform transition-all`}>
      <!-- 头部 -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-slate-200 dark:border-slate-700">
        <h2 class="text-base font-semibold text-slate-900 dark:text-white">{title}</h2>
        <button
          on:click={onClose}
          class="p-2 hover:bg-slate-100 dark:hover:bg-slate-700 rounded-lg transition-colors focus-visible:outline-none focus-visible:ring-2"
        >
          <svg class="w-5 h-5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
      
      <!-- 内容 -->
      <div class="flex-1 overflow-y-auto p-6">
        <slot />
      </div>
    </div>
  </div>
{/if}
