<script>
  export let themeStore;

  let jsonInput = '';
  let jsonOutput = '';
  let isLoading = false;
  let errorMessage = '';
  let isValid = false;
  let stats = { chars: 0, lines: 0 };

  $: validateJSON();
  $: updateStats();

  async function validateJSON() {
    if (!jsonInput.trim()) {
      isValid = false;
      errorMessage = '';
      return;
    }

    if (!window.wailsBindings) {
      isValid = false;
      errorMessage = 'Wails 绑定未加载';
      return;
    }

    try {
      const result = await window.wailsBindings.ValidateJSON(jsonInput);
      isValid = result.valid;
      errorMessage = result.error || '';
    } catch (error) {
      isValid = false;
      errorMessage = error.message;
    }
  }

  function updateStats() {
    stats.chars = jsonInput.length;
    stats.lines = jsonInput.split('\n').length;
  }

  async function formatJSON() {
    if (!jsonInput.trim() || isLoading || !window.wailsBindings) return;

    isLoading = true;
    errorMessage = '';

    try {
      jsonOutput = await window.wailsBindings.FormatJSON(jsonInput);
      isValid = true;
      errorMessage = '';
    } catch (error) {
      errorMessage = error.message || '格式化失败';
    } finally {
      isLoading = false;
    }
  }

  async function minifyJSON() {
    if (!jsonInput.trim() || isLoading || !window.wailsBindings) return;

    isLoading = true;
    errorMessage = '';

    try {
      jsonOutput = await window.wailsBindings.MinifyJSON(jsonInput);
      isValid = true;
      errorMessage = '';
    } catch (error) {
      errorMessage = error.message || '压缩失败';
    } finally {
      isLoading = false;
    }
  }

  async function escapeJSON() {
    if (!jsonInput.trim() || isLoading || !window.wailsBindings) return;

    isLoading = true;
    errorMessage = '';

    try {
      jsonOutput = await window.wailsBindings.EscapeJSON(jsonInput);
      isValid = true;
      errorMessage = '';
    } catch (error) {
      errorMessage = error.message || '转义失败';
    } finally {
      isLoading = false;
    }
  }

  async function copyOutput() {
    if (!jsonOutput) return;

    try {
      await navigator.clipboard.writeText(jsonOutput);
    } catch (error) {
      console.error('复制失败:', error);
    }
  }

  function clearAll() {
    jsonInput = '';
    jsonOutput = '';
    isValid = false;
  }

  function handleKeyDown(event) {
    const ctrlOrCmd = event.ctrlKey || event.metaKey;

    if (ctrlOrCmd && event.key.toLowerCase() === 'enter') {
      if (!isLoading && jsonInput.trim()) {
        event.preventDefault();
        formatJSON();
      }
    }
  }
</script>

<div class="space-y-6">
  <div class="relative">
    <div class="flex items-center justify-between mb-2">
      <label class="text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
        <span>输入 JSON</span>
        {#if isValid}
          <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300">
            ✓ 有效
          </span>
        {:else if errorMessage}
          <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300">
            ✗ 无效
          </span>
        {/if}
      </label>
      <div class="text-xs text-gray-500 dark:text-gray-400">
        {stats.chars} 字符 · {stats.lines} 行
      </div>
    </div>

    <textarea
      bind:value={jsonInput}
      placeholder="输入 JSON..."
      on:keydown={handleKeyDown}
      class="w-full h-48 px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 {errorMessage ? 'border-red-500 focus:ring-red-500' : 'border-gray-300 dark:border-gray-600 focus:border-purple-500 dark:focus:border-purple-400'} rounded-xl text-sm font-mono resize-none focus:outline-none focus:ring-2 transition-all dark:text-white"
    />

    {#if errorMessage}
      <div class="mt-2 text-sm text-red-600 dark:text-red-400 flex items-center gap-1">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
        {errorMessage}
      </div>
    {/if}
  </div>

  <div class="grid grid-cols-3 gap-3">
    <button
      on:click={formatJSON}
      disabled={isLoading || !jsonInput.trim()}
      class="flex items-center justify-center gap-2 px-4 py-3 bg-gradient-to-r from-purple-600 to-blue-600 hover:from-purple-700 hover:to-blue-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-xl"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16m-7 6h7" />
      </svg>
      <span class="text-sm">格式化</span>
    </button>

    <button
      on:click={minifyJSON}
      disabled={isLoading || !jsonInput.trim()}
      class="flex items-center justify-center gap-2 px-4 py-3 bg-gradient-to-r from-amber-500 to-orange-500 hover:from-amber-600 hover:to-orange-600 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-xl"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 5l7 7-7 7M5 5l7 7-7 7" />
      </svg>
      <span class="text-sm">压缩</span>
    </button>

    <button
      on:click={escapeJSON}
      disabled={isLoading || !jsonInput.trim()}
      class="flex items-center justify-center gap-2 px-4 py-3 accent-gradient hover:brightness-95 disabled:opacity-50 disabled:cursor-not-allowed rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-xl"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21a4 4 0 01-4-4V5a2 2 0 012-2h4a2 2 0 012 2v12a4 4 0 01-4 4zm0 0h12a2 2 0 002-2v-4a2 2 0 00-2-2h-2.343M11 7.343l1.657-1.657a2 2 0 012.828 0l2.829 2.829a2 2 0 010 2.828l-8.486 8.485M7 17h.01" />
      </svg>
      <span class="text-sm">转义</span>
    </button>
  </div>

  <div>
    <div class="flex items-center justify-between mb-2">
      <label class="text-sm font-semibold text-gray-700 dark:text-gray-300">输出</label>
      {#if jsonOutput}
        <button
          on:click={copyOutput}
          class="text-xs flex items-center gap-1 px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg transition-colors text-gray-700 dark:text-gray-300"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
          复制
        </button>
      {/if}
    </div>
    <textarea
      bind:value={jsonOutput}
      readonly
      placeholder="格式化后的 JSON 将显示在这里..."
      class="w-full h-48 px-4 py-3 bg-gray-100 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm font-mono resize-none focus:outline-none dark:text-white"
    />
  </div>

  <div class="flex justify-end">
    <button
      on:click={clearAll}
      class="text-sm px-4 py-2 text-gray-600 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
    >
      清空全部
    </button>
  </div>
</div>
