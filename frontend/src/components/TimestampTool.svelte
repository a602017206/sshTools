<script>
  export let themeStore;

  let mode = 'toDateTime';
  let timestampInput = '';
  let datetimeInput = '';
  let output = '';
  let isLoading = false;
  let currentTimestamp = 0;

  $: updateCurrentTimestamp();

  function updateCurrentTimestamp() {
    if (window.wailsBindings) {
      currentTimestamp = Math.floor(Date.now() / 1000);
    }
  }

  async function convert() {
    if (!window.wailsBindings) return;

    isLoading = true;
    output = '';

    try {
      if (mode === 'toDateTime') {
        const ts = parseInt(timestampInput.trim());
        if (isNaN(ts)) {
          throw new Error('请输入有效的时间戳');
        }
        output = await window.wailsBindings.TimestampToDateTime(ts, '2006-01-02 15:04:05');
      } else {
        if (!datetimeInput.trim()) {
          throw new Error('请输入日期时间');
        }
        const ts = await window.wailsBindings.DateTimeToTimestamp(datetimeInput, '2006-01-02 15:04:05');
        output = ts.toString();
      }
    } catch (error) {
      output = '';
      console.error('转换失败:', error);
      alert('转换失败: ' + (error.message || error));
    } finally {
      isLoading = false;
    }
  }

  async function useCurrentTimestamp() {
    if (!window.wailsBindings) return;
    try {
      const ts = await window.wailsBindings.GetCurrentTimestamp();
      timestampInput = ts.toString();
    } catch (error) {
      console.error('获取当前时间戳失败:', error);
    }
  }

  async function useCurrentDateTime() {
    if (!window.wailsBindings) return;
    try {
      const ts = await window.wailsBindings.GetCurrentTimestamp();
      const dt = await window.wailsBindings.TimestampToDateTime(ts, '2006-01-02 15:04:05');
      datetimeInput = dt;
    } catch (error) {
      console.error('获取当前时间失败:', error);
    }
  }

  async function copyOutput() {
    if (!output) return;

    try {
      await navigator.clipboard.writeText(output);
    } catch (error) {
      console.error('复制失败:', error);
    }
  }

  function clearAll() {
    timestampInput = '';
    datetimeInput = '';
    output = '';
  }

  function handleKeyDown(event) {
    const ctrlOrCmd = event.ctrlKey || event.metaKey;

    if (ctrlOrCmd && event.key.toLowerCase() === 'enter') {
      if (!isLoading && ((mode === 'toDateTime' && timestampInput.trim()) || (mode === 'toTimestamp' && datetimeInput.trim()))) {
        event.preventDefault();
        convert();
      }
    }
  }
</script>

<div class="space-y-6">
  <div class="flex gap-2 p-1 bg-gray-100 dark:bg-gray-800 rounded-xl">
    <button
      on:click={() => { mode = 'toDateTime'; clearAll(); }}
      class="flex-1 py-2.5 px-4 rounded-lg text-sm font-semibold transition-all {
        mode === 'toDateTime'
          ? 'bg-white dark:bg-gray-700 text-purple-600 dark:text-purple-400 shadow-sm'
          : 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
      }"
    >
      时间戳 → 日期时间
    </button>
    <button
      on:click={() => { mode = 'toTimestamp'; clearAll(); }}
      class="flex-1 py-2.5 px-4 rounded-lg text-sm font-semibold transition-all {
        mode === 'toTimestamp'
          ? 'bg-white dark:bg-gray-700 text-blue-600 dark:text-blue-400 shadow-sm'
          : 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
      }"
    >
      日期时间 → 时间戳
    </button>
  </div>

  <div class="bg-gradient-to-br from-amber-50 to-orange-50 dark:from-amber-950/30 dark:to-orange-950/30 rounded-xl p-4 border border-amber-200 dark:border-amber-800/50">
    <div class="flex items-center gap-2 text-sm text-amber-800 dark:text-amber-300">
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <span>当前时间戳: <strong class="font-mono">{currentTimestamp}</strong></span>
    </div>
  </div>

  {#if mode === 'toDateTime'}
    <div>
      <div class="flex items-center justify-between mb-2">
        <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300">输入 Unix 时间戳（秒）</label>
        <button
          on:click={useCurrentTimestamp}
          class="text-xs flex items-center gap-1 px-3 py-1.5 bg-purple-100 dark:bg-purple-900/50 hover:bg-purple-200 dark:hover:bg-purple-800/50 rounded-lg transition-colors text-purple-700 dark:text-purple-300"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          使用当前
        </button>
      </div>
      <input
        type="number"
        bind:value={timestampInput}
        placeholder="例如: 1234567890"
        on:keydown={handleKeyDown}
        class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm font-mono focus:outline-none focus:border-purple-500 dark:focus:border-purple-400 focus:ring-2 transition-all dark:text-white"
      />
    </div>
  {:else}
    <div>
      <div class="flex items-center justify-between mb-2">
        <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300">输入日期时间</label>
        <button
          on:click={useCurrentDateTime}
          class="text-xs flex items-center gap-1 px-3 py-1.5 bg-blue-100 dark:bg-blue-900/50 hover:bg-blue-200 dark:hover:bg-blue-800/50 rounded-lg transition-colors text-blue-700 dark:text-blue-300"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          使用当前
        </button>
      </div>
      <input
        type="datetime-local"
        bind:value={datetimeInput}
        on:keydown={handleKeyDown}
        class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm focus:outline-none focus:border-blue-500 dark:focus:border-blue-400 focus:ring-2 transition-all dark:text-white"
      />
      <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
        格式: YYYY-MM-DD HH:MM:SS
      </div>
    </div>
  {/if}

  <div class="flex justify-center">
    <button
      on:click={convert}
      disabled={isLoading || (mode === 'toDateTime' ? !timestampInput.trim() : !datetimeInput.trim())}
      class="flex items-center justify-center gap-2 px-8 py-3 bg-gradient-to-r from-amber-600 to-orange-600 hover:from-amber-700 hover:to-orange-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-xl"
    >
      {#if isLoading}
        <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        转换中...
      {:else}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
        </svg>
        转换
      {/if}
    </button>
  </div>

  {#if output}
    <div>
      <div class="flex items-center justify-between mb-2">
        <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300">
          {mode === 'toDateTime' ? '日期时间' : '时间戳'}
        </label>
        <button
          on:click={copyOutput}
          class="text-xs flex items-center gap-1 px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg transition-colors text-gray-700 dark:text-gray-300"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
          复制
        </button>
      </div>
      <div class="p-4 bg-gray-100 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl">
        <code class="text-sm font-mono text-gray-800 dark:text-gray-200">
          {output}
        </code>
      </div>
    </div>
  {/if}

  <div class="flex justify-end">
    <button
      on:click={clearAll}
      class="text-sm px-4 py-2 text-gray-600 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
    >
      清空全部
    </button>
  </div>
</div>
