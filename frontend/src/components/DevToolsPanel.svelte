<script>
  import { onMount, tick } from 'svelte';
  import JsonFormatter from './JsonFormatter.svelte';
  import Base64Tool from './Base64Tool.svelte';
  import HashTool from './HashTool.svelte';
  import TimestampTool from './TimestampTool.svelte';
  import UuidTool from './UuidTool.svelte';

  export let isOpen = false;
  export let themeStore;

  let activeToolId = null;
  let dialogElement;

  const tools = [
    { id: 'json', name: 'JSON 格式化', icon: '📄', color: 'text-purple-500' },
    { id: 'base64', name: 'Base64 编解码', icon: '🔐', color: 'text-blue-500' },
    { id: 'hash', name: 'Hash 计算', icon: '#️⃣', color: 'text-green-500' },
    { id: 'timestamp', name: '时间戳转换', icon: '🕐', color: 'text-amber-500' },
    { id: 'uuid', name: 'UUID 生成', icon: '🆔', color: 'text-indigo-500' },
  ];

  function handleBackdropClick(event) {
    if (event.target === event.currentTarget) {
      isOpen = false;
    }
  }

  function selectTool(toolId) {
    activeToolId = toolId;
  }

  function getActiveTool() {
    return tools.find(t => t.id === activeToolId);
  }

  function handleKeyDown(event) {
    if (event.key === 'Escape' && isOpen) {
      event.preventDefault();
      isOpen = false;
    }
  }

  // Focus the dialog when it opens
  $: if (isOpen) {
    tick().then(() => {
      if (dialogElement) {
        dialogElement.focus();
      }
    });
  }
</script>

{#if isOpen}
  <div
    bind:this={dialogElement}
    class="fixed inset-0 z-50 flex items-start justify-center pt-16"
    on:click={handleBackdropClick}
    on:keydown={handleKeyDown}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
  >
    <div class="absolute inset-0 bg-black/30 backdrop-blur-sm transition-opacity" />

    <div class="relative w-full max-w-5xl bg-white dark:bg-gray-900 rounded-2xl shadow-2xl m-4 max-h-[85vh] overflow-hidden flex border border-gray-200 dark:border-gray-700">
      <div class="w-72 bg-gradient-to-b from-gray-50 to-white dark:from-gray-800 dark:to-gray-900 border-r border-gray-200 dark:border-gray-700 p-4 flex flex-col">
        <div class="flex items-center justify-between mb-6">
          <h3 class="text-lg font-bold bg-gradient-to-r from-purple-600 to-blue-600 bg-clip-text text-transparent">开发工具</h3>
          <button
            on:click={() => isOpen = false}
            class="p-2 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-all duration-200 group"
            title="关闭"
          >
            <svg class="w-4 h-4 text-gray-500 dark:text-gray-400 group-hover:text-gray-700 dark:group-hover:text-gray-200 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="space-y-2 flex-1 overflow-y-auto">
          {#each tools as tool (tool.id)}
            <button
              on:click={() => selectTool(tool.id)}
              class="w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 group {
                activeToolId === tool.id
                  ? 'bg-white dark:bg-gray-800 shadow-lg border-2 border-purple-500 dark:border-purple-400 transform scale-[1.02]'
                  : 'hover:bg-white dark:hover:bg-gray-800 border-2 border-transparent hover:border-gray-200 dark:hover:border-gray-600'
              }"
            >
              <span class="text-2xl">{tool.icon}</span>
              <span class={`text-sm font-semibold transition-colors ${
                activeToolId === tool.id ? 'text-purple-600 dark:text-purple-400' : 'text-gray-700 dark:text-gray-300 group-hover:text-gray-900 dark:group-hover:text-white'
              }`}>
                {tool.name}
              </span>
            </button>
          {/each}
        </div>

        <div class="mt-4 pt-4 border-t border-gray-200 dark:border-gray-700">
          <div class="text-xs text-gray-500 dark:text-gray-400 text-center">
            快捷键: 按 ESC 关闭
          </div>
        </div>
      </div>

      <div class="flex-1 flex flex-col overflow-hidden">
        {#if activeToolId}
          <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gradient-to-r from-purple-50 to-blue-50 dark:from-purple-950/30 dark:to-blue-950/30">
            <h2 class="text-xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
              <span class="text-2xl">{getActiveTool()?.icon}</span>
              {getActiveTool()?.name}
            </h2>
          </div>

          <div class="flex-1 overflow-y-auto p-6">
            {#if activeToolId === 'json'}
              <svelte:component this={JsonFormatter} {themeStore} />
            {:else if activeToolId === 'base64'}
              <svelte:component this={Base64Tool} {themeStore} />
            {:else if activeToolId === 'hash'}
              <svelte:component this={HashTool} {themeStore} />
            {:else if activeToolId === 'timestamp'}
              <svelte:component this={TimestampTool} {themeStore} />
            {:else if activeToolId === 'uuid'}
              <svelte:component this={UuidTool} {themeStore} />
            {/if}
          </div>
        {:else}
          <div class="flex-1 flex items-center justify-center p-12">
            <div class="text-center">
              <div class="text-6xl mb-4">🛠️</div>
              <h3 class="text-xl font-semibold text-gray-700 dark:text-gray-300 mb-2">选择一个工具</h3>
              <p class="text-gray-500 dark:text-gray-400">从左侧选择一个开发工具开始使用</p>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

<style>
  :global(.scrollbar-thin)::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }

  :global(.scrollbar-thin)::-webkit-scrollbar-track {
    background: transparent;
  }

  :global(.scrollbar-thin)::-webkit-scrollbar-thumb {
    background: #d1d5db;
    border-radius: 3px;
  }

  :global(.dark .scrollbar-thin)::-webkit-scrollbar-thumb {
    background: #4b5563;
  }

  :global(.scrollbar-thin)::-webkit-scrollbar-thumb:hover {
    background: #9ca3af;
  }
</style>
