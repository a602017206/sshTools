<script>
  export let themeStore;

  let generatedUuid = '';
  let uuidHistory = [];
  let isGenerating = false;
  let copiedUuid = null;

  async function generateUuid() {
    if (!window.wailsBindings) return;

    isGenerating = true;

    try {
      const uuid = await window.wailsBindings.GenerateUUIDv4();
      generatedUuid = uuid;

      if (uuidHistory.length < 10) {
        uuidHistory = [uuid, ...uuidHistory];
      } else {
        uuidHistory = [uuid, ...uuidHistory.slice(0, 9)];
      }
    } catch (error) {
      console.error('UUID 生成失败:', error);
      alert('UUID 生成失败: ' + (error.message || error));
    } finally {
      isGenerating = false;
    }
  }

  async function copyUuid(uuid) {
    try {
      await navigator.clipboard.writeText(uuid);
      copiedUuid = uuid;
      setTimeout(() => copiedUuid = null, 2000);
    } catch (error) {
      console.error('复制失败:', error);
    }
  }

  function clearHistory() {
    uuidHistory = [];
  }

  function useFromHistory(uuid) {
    generatedUuid = uuid;
  }
</script>

<div class="space-y-6">
  <div class="text-center">
    <button
      on:click={generateUuid}
      disabled={isGenerating}
      class="group relative inline-flex items-center justify-center gap-3 px-8 py-4 bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-700 hover:to-purple-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-2xl font-bold text-lg transition-all duration-200 shadow-2xl hover:shadow-3xl"
    >
      {#if isGenerating}
        <svg class="animate-spin w-6 h-6" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        生成中...
      {:else}
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v1m6 11h2m-6 0h-2v4h2v-4zM5 22h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v14a2 2 0 002 2zM12 8a2 2 0 110-4 2 2 0 010 4z" />
        </svg>
        <span>生成 UUID v4</span>
      {/if}

      <div class="absolute inset-0 rounded-2xl bg-white/20 opacity-0 group-hover:opacity-100 transition-opacity" />
    </button>
  </div>

  {#if generatedUuid}
    <div>
      <div class="flex items-center justify-between mb-3">
        <label class="text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
          <span class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
          生成的 UUID
        </label>
        <button
          on:click={() => copyUuid(generatedUuid)}
          class="flex items-center gap-1.5 px-4 py-2 bg-indigo-100 dark:bg-indigo-900/50 hover:bg-indigo-200 dark:hover:bg-indigo-800/50 rounded-lg transition-colors text-indigo-700 dark:text-indigo-300 font-medium text-sm"
        >
          {#if copiedUuid === generatedUuid}
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
            已复制
          {:else}
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            复制
          {/if}
        </button>
      </div>
      <div class="p-4 bg-gradient-to-br from-indigo-50 to-purple-50 dark:from-indigo-950/50 dark:to-purple-950/50 border-2 border-indigo-300 dark:border-indigo-700 rounded-xl">
        <code class="text-lg font-mono text-indigo-900 dark:text-indigo-100 break-all tracking-wide">
          {generatedUuid}
        </code>
      </div>
    </div>
  {/if}

  {#if uuidHistory.length > 0}
    <div>
      <div class="flex items-center justify-between mb-3">
        <label class="text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          历史记录
          <span class="text-xs font-normal text-gray-500 dark:text-gray-400">
            ({uuidHistory.length}/10)
          </span>
        </label>
        <button
          on:click={clearHistory}
          class="text-xs flex items-center gap-1 px-3 py-1.5 bg-red-100 dark:bg-red-900/50 hover:bg-red-200 dark:hover:bg-red-800/50 rounded-lg transition-colors text-red-700 dark:text-red-300"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
          </svg>
          清空
        </button>
      </div>

      <div class="space-y-2 max-h-64 overflow-y-auto scrollbar-thin">
        {#each uuidHistory as uuid, index}
          <div
            class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-750 transition-colors group"
          >
            <code class="text-sm font-mono text-gray-700 dark:text-gray-300 truncate mr-4">
              {uuid}
            </code>
            <div class="flex items-center gap-2 shrink-0">
              <button
                on:click={() => useFromHistory(uuid)}
                class="p-1.5 text-gray-500 dark:text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
                title="使用此 UUID"
              >
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                </svg>
              </button>
              <button
                on:click={() => copyUuid(uuid)}
                class="p-1.5 text-gray-500 dark:text-gray-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
                title="复制"
              >
                {#if copiedUuid === uuid}
                  <svg class="w-4 h-4 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
                  </svg>
                {:else}
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                {/if}
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  {#if !generatedUuid}
    <div class="text-center py-8">
      <div class="text-6xl mb-4 opacity-50">🎲</div>
      <p class="text-gray-500 dark:text-gray-400">
        点击上方按钮生成 UUID v4
      </p>
    </div>
  {/if}
</div>
