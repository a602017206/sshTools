<script>
  export let isOpen = false;
  export let onClose = () => {};
  export let title = '';
  export let size = 'md'; // sm, md, lg, xl

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
    <div class="absolute inset-0 bg-black/20 backdrop-blur-sm transition-opacity pointer-events-auto" />
    
    <!-- 对话框 -->
    <div class="relative w-full max-w-xs bg-white dark:bg-gray-800 rounded-xl shadow-2xl m-4 max-h-[90vh] overflow-hidden flex flex-col transform transition-all">
      <!-- 头部 -->
      <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700">
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">{title}</h2>
        <button
          on:click={onClose}
          class="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
        >
          <svg class="w-5 h-5 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
