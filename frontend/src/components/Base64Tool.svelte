<script>
  export let themeStore;

  let inputText = '';
  let outputText = '';
  let mode = 'encode';
  let isLoading = false;

  async function handleAction() {
    if (!inputText.trim() || isLoading || !window.wailsBindings) return;

    isLoading = true;

    try {
      if (mode === 'encode') {
        outputText = await window.wailsBindings.EncodeBase64(inputText);
      } else {
        outputText = await window.wailsBindings.DecodeBase64(inputText);
      }
    } catch (error) {
      outputText = '';
      console.error('处理失败:', error);
      alert('处理失败: ' + (error.message || error));
    } finally {
      isLoading = false;
    }
  }

  async function copyOutput() {
    if (!outputText) return;

    try {
      await navigator.clipboard.writeText(outputText);
    } catch (error) {
      console.error('复制失败:', error);
    }
  }

  function clearAll() {
    inputText = '';
    outputText = '';
  }

  function swapMode() {
    mode = mode === 'encode' ? 'decode' : 'encode';
    inputText = '';
    outputText = '';
  }

  function handleKeyDown(event) {
    const ctrlOrCmd = event.ctrlKey || event.metaKey;

    if (ctrlOrCmd && event.key.toLowerCase() === 'enter') {
      if (!isLoading && inputText.trim()) {
        event.preventDefault();
        handleAction();
      }
    }
  }
</script>

<div class="space-y-6">
  <div class="flex gap-2 p-1 bg-gray-100 dark:bg-gray-800 rounded-xl">
    <button
      on:click={() => mode = 'encode'}
      class="flex-1 py-2.5 px-4 rounded-lg text-sm font-semibold transition-all {
        mode === 'encode'
          ? 'bg-white dark:bg-gray-700 text-purple-600 dark:text-purple-400 shadow-sm'
          : 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
      }"
    >
      编码
    </button>
    <button
      on:click={() => mode = 'decode'}
      class="flex-1 py-2.5 px-4 rounded-lg text-sm font-semibold transition-all {
        mode === 'decode'
          ? 'bg-white dark:bg-gray-700 text-blue-600 dark:text-blue-400 shadow-sm'
          : 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
      }"
    >
      解码
    </button>
  </div>

  <div>
    <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
      {mode === 'encode' ? '输入文本' : '输入 Base64'}
    </label>
    <textarea
      bind:value={inputText}
      placeholder={mode === 'encode' ? '输入要编码的文本...' : '输入 Base64 字符串...'}
      on:keydown={handleKeyDown}
      class="w-full h-32 px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm resize-none focus:outline-none focus:border-purple-500 dark:focus:border-purple-400 focus:ring-2 transition-all dark:text-white"
    />
  </div>

  <div class="flex items-center justify-center">
    <button
      on:click={handleAction}
      disabled={isLoading || !inputText.trim()}
      class="flex items-center justify-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-700 hover:to-cyan-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-xl"
    >
      {#if isLoading}
        <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        处理中...
      {:else}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
        </svg>
        {mode === 'encode' ? '编码' : '解码'}
      {/if}
    </button>
  </div>

  <div>
    <div class="flex items-center justify-between mb-2">
      <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300">
        {mode === 'encode' ? 'Base64 输出' : '解码结果'}
      </label>
      {#if outputText}
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
      bind:value={outputText}
      readonly
      placeholder={mode === 'encode' ? 'Base64 编码结果将显示在这里...' : '解码后的文本将显示在这里...'}
      class="w-full h-32 px-4 py-3 bg-gray-100 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm font-mono resize-none focus:outline-none dark:text-white"
    />
  </div>

  <div class="flex justify-between items-center">
    <button
      on:click={clearAll}
      class="text-sm px-4 py-2 text-gray-600 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
    >
      清空全部
    </button>
    <button
      on:click={swapMode}
      class="flex items-center gap-2 text-sm px-4 py-2 text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
      </svg>
      切换模式
    </button>
  </div>
</div>
