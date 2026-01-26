<script>
  export let isOpen = false;
  
  let activeToolId = null;
  let jsonInput = '';
  let jsonOutput = '';
  
  const tools = [
    { id: 'json', name: 'JSON 格式化', icon: '📄', color: 'text-purple-600' },
    { id: 'base64', name: 'Base64 编解码', icon: '🔐', color: 'text-blue-600' },
    { id: 'hash', name: 'Hash 计算', icon: '#️⃣', color: 'text-green-600' },
    { id: 'timestamp', name: '时间戳转换', icon: '🕐', color: 'text-amber-600' },
    { id: 'uuid', name: 'UUID 生成', icon: '🆔', color: 'text-indigo-600' },
  ];

  function formatJSON() {
    try {
      const parsed = JSON.parse(jsonInput);
      jsonOutput = JSON.stringify(parsed, null, 2);
    } catch (error) {
      jsonOutput = 'JSON 格式错误: ' + error.message;
    }
  }

  function handleBackdropClick(event) {
    if (event.target === event.currentTarget) {
      isOpen = false;
    }
  }
</script>

{#if isOpen}
  <div
    class="fixed inset-0 z-50 flex items-start justify-center pt-16"
    on:click={handleBackdropClick}
    role="dialog"
    aria-modal="true"
  >
    <div class="absolute inset-0 bg-black/20 backdrop-blur-sm" />
    
    <div class="relative w-full max-w-4xl bg-white dark:bg-gray-800 rounded-2xl shadow-2xl m-4 max-h-[80vh] overflow-hidden flex">
      <div class="w-64 bg-gray-50 dark:bg-gray-900 border-r border-gray-200 dark:border-gray-700 p-4">
        <div class="flex items-center justify-between mb-4">
          <h3 class="text-sm font-semibold text-gray-900 dark:text-white">开发工具</h3>
          <button
            on:click={() => isOpen = false}
            class="p-1.5 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition-colors"
          >
            <svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="space-y-1">
          {#each tools as tool (tool.id)}
            <button
              on:click={() => activeToolId = tool.id}
              class="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all {
                activeToolId === tool.id
                  ? 'bg-white dark:bg-gray-800 shadow-sm border border-gray-200 dark:border-gray-700'
                  : 'hover:bg-gray-100 dark:hover:bg-gray-800'
              }"
            >
              <span class="text-xl">{tool.icon}</span>
              <span class={`text-sm font-medium ${
                activeToolId === tool.id ? 'text-gray-900 dark:text-white' : 'text-gray-700 dark:text-gray-300'
              }`}>
                {tool.name}
              </span>
            </button>
          {/each}
        </div>
      </div>

      <div class="flex-1 p-6 overflow-y-auto">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-6">
          {tools.find(t => t.id === activeToolId)?.name || '选择工具'}
        </h2>
        
        {#if activeToolId === 'json'}
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">输入 JSON</label>
              <textarea
                bind:value={jsonInput}
                placeholder={"{&quot;name&quot;: &quot;test&quot;}"}
                class="w-full h-32 px-3 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-mono resize-none focus:outline-none focus:ring-2 focus:ring-purple-500 dark:text-white"
              />
            </div>
            <button
              on:click={formatJSON}
              class="w-full px-4 py-2.5 bg-purple-600 hover:bg-purple-700 text-white rounded-lg font-medium transition-colors"
            >
              格式化
            </button>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">输出</label>
              <textarea
                bind:value={jsonOutput}
                readonly
                class="w-full h-32 px-3 py-2 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-mono resize-none focus:outline-none dark:text-white"
              />
            </div>
          </div>
        {:else}
          <div class="text-center py-12 text-gray-500 dark:text-gray-400">
            {tools.find(t => t.id === activeToolId)?.name + ' 功能待实现'}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
