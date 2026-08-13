<script>
  import Dialog from './ui/Dialog.svelte';
  import JsonFormatter from './JsonFormatter.svelte';
  import Base64Tool from './Base64Tool.svelte';
  import HashTool from './HashTool.svelte';
  import TimestampTool from './TimestampTool.svelte';
  import UuidTool from './UuidTool.svelte';
  import UrlTool from './UrlTool.svelte';

  export let isOpen = false;
  export let themeStore;

  let activeToolId = null;

  const tools = [
    { id: 'json', name: 'JSON 格式化', icon: '📄', color: 'text-purple-500' },
    { id: 'base64', name: 'Base64 编解码', icon: '🔐', color: 'text-blue-500' },
    { id: 'url', name: 'URL 编解码', icon: '🔗', color: 'text-orange-500' },
    { id: 'hash', name: '加密解密', icon: '#️⃣', color: 'text-green-500' },
    { id: 'timestamp', name: '时间戳转换', icon: '🕐', color: 'text-amber-500' },
    { id: 'uuid', name: 'UUID 生成', icon: '🆔', color: 'text-indigo-500' },
  ];

  function selectTool(toolId) {
    activeToolId = toolId;
  }

  function getActiveTool() {
    return tools.find(t => t.id === activeToolId);
  }

</script>

<Dialog bind:isOpen={isOpen} onClose={() => isOpen = false} title="开发工具" size="xl">
    <div class="w-full h-[min(720px,calc(100vh-10rem))] bg-white/92 dark:bg-slate-900/85 rounded-xl overflow-hidden flex border border-slate-200/70 dark:border-slate-700/60">
      <div class="w-72 bg-gradient-to-b from-slate-50 via-white to-slate-100 dark:from-slate-900 dark:via-slate-900/80 dark:to-slate-950 border-r border-slate-200/70 dark:border-slate-800/80 p-5 flex flex-col">
        <div class="flex items-center justify-between mb-6">
          <div>
            <h3 class="text-lg font-semibold tracking-tight bg-clip-text text-transparent" style="background-image: linear-gradient(90deg, var(--text-primary), var(--accent-primary));">开发工具</h3>
            <div class="text-[11px] uppercase tracking-[0.2em] text-slate-400 dark:text-slate-500 mt-1">Toolkit</div>
          </div>
          <button
            on:click={() => isOpen = false}
            class="p-2 rounded-lg transition-all duration-200 group hover:bg-slate-200/70 dark:hover:bg-slate-800/70"
            title="关闭"
          >
            <svg class="w-4 h-4 text-slate-500 dark:text-slate-400 group-hover:text-slate-700 dark:group-hover:text-slate-200 transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="space-y-2 flex-1 overflow-y-auto">
          {#each tools as tool (tool.id)}
            <button
              on:click={() => selectTool(tool.id)}
              class="relative w-full flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200 group border {
                activeToolId === tool.id
                  ? 'bg-white/90 dark:bg-slate-900/80 border-slate-200 dark:border-slate-700 shadow-sm'
                  : 'bg-transparent border-transparent hover:bg-white/70 dark:hover:bg-slate-900/60 hover:border-slate-200/70 dark:hover:border-slate-700/70'
              }"
            >
              {#if activeToolId === tool.id}
                <span class="absolute left-2 top-2 bottom-2 w-1 rounded-full" style="background: linear-gradient(180deg, var(--accent-primary), var(--accent-hover));" />
              {/if}
              <span class={`text-2xl ${tool.color}`}>{tool.icon}</span>
              <span class={`text-sm font-semibold transition-colors ${
                activeToolId === tool.id ? 'text-slate-800 dark:text-slate-100' : 'text-slate-600 dark:text-slate-300 group-hover:text-slate-900 dark:group-hover:text-white'
              }`}>
                {tool.name}
              </span>
            </button>
          {/each}
        </div>

        <div class="mt-4 pt-4 border-t border-slate-200/70 dark:border-slate-800/80">
          <div class="text-[11px] uppercase tracking-[0.2em] text-slate-400 dark:text-slate-500 text-center">
            快捷键: 按 ESC 关闭
          </div>
        </div>
      </div>

        <div class="flex-1 flex flex-col overflow-hidden">
          {#if activeToolId}
            <div class="px-6 py-4 border-b border-slate-200/70 dark:border-slate-800/80 bg-gradient-to-r from-slate-50 via-white to-slate-50 dark:from-slate-900 dark:via-slate-900/80 dark:to-slate-950">
              <h2 class="text-xl font-semibold text-slate-900 dark:text-slate-100 flex items-center gap-2">
                <span class={`text-2xl ${getActiveTool()?.color}`}>{getActiveTool()?.icon}</span>
                {getActiveTool()?.name}
              </h2>
            </div>

            <div class="flex-1 overflow-y-auto p-6 bg-slate-50/60 dark:bg-slate-950/30">
              <div class="rounded-2xl border border-slate-200/70 dark:border-slate-800/70 bg-white/80 dark:bg-slate-900/70 p-5 shadow-sm">
                {#if activeToolId === 'json'}
                  <svelte:component this={JsonFormatter} {themeStore} />
                {:else if activeToolId === 'base64'}
                  <svelte:component this={Base64Tool} {themeStore} />
                {:else if activeToolId === 'url'}
                  <svelte:component this={UrlTool} {themeStore} />
                {:else if activeToolId === 'hash'}
                  <svelte:component this={HashTool} {themeStore} />
                {:else if activeToolId === 'timestamp'}
                  <svelte:component this={TimestampTool} {themeStore} />
                {:else if activeToolId === 'uuid'}
                  <svelte:component this={UuidTool} {themeStore} />
                {/if}
              </div>
            </div>
          {:else}
            <div class="flex-1 flex items-center justify-center p-12 bg-slate-50/60 dark:bg-slate-950/30">
              <div class="text-center max-w-sm">
                <div class="text-6xl mb-4">🛠️</div>
                <h3 class="text-xl font-semibold text-slate-700 dark:text-slate-200 mb-2">选择一个工具</h3>
                <p class="text-sm text-slate-500 dark:text-slate-400">从左侧选择一个开发工具开始使用</p>
              </div>
            </div>
          {/if}
        </div>
      </div>
</Dialog>

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
