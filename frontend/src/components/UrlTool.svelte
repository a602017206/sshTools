<script>
  export let themeStore;

  let inputText = '';
  let outputText = '';
  let mode = 'encode';
  let encodeMode = 'query';
  let isLoading = false;
  let parsedParams = null;

  const encodeModes = [
    { value: 'query', label: 'Query 参数', desc: '空格转为 +' },
    { value: 'path', label: '路径', desc: '保留斜杠' },
    { value: 'full', label: '完整 URL', desc: '解析并编码' },
  ];

  async function handleAction() {
    if (!inputText.trim() || isLoading || !window.wailsBindings) return;

    isLoading = true;
    parsedParams = null;

    try {
      if (mode === 'encode') {
        const result = await window.wailsBindings.URLEncode(inputText, encodeMode);
        outputText = result.encoded;
      } else {
        const result = await window.wailsBindings.URLDecode(inputText, encodeMode);
        outputText = result.decoded;
        if (result.params) {
          parsedParams = result.params;
        }
      }
    } catch (error) {
      outputText = '';
      console.error('处理失败:', error);
      alert('处理失败: ' + (error.message || error));
    } finally {
      isLoading = false;
    }
  }

  async function parseURL() {
    if (!inputText.trim() || !window.wailsBindings) return;

    try {
      const result = await window.wailsBindings.ParseURL(inputText);
      outputText = JSON.stringify(result, null, 2);
      if (result.queryParams) {
        parsedParams = {};
        for (const [key, values] of Object.entries(result.queryParams)) {
          parsedParams[key] = Array.isArray(values) ? values[0] : values;
        }
      }
    } catch (error) {
      console.error('解析失败:', error);
      alert('URL解析失败: ' + (error.message || error));
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
    parsedParams = null;
  }

  function swapMode() {
    mode = mode === 'encode' ? 'decode' : 'encode';
    inputText = '';
    outputText = '';
    parsedParams = null;
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
  <!-- 模式选择 -->
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

  <!-- 编码/解码模式选择 -->
  <div>
    <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
      编码/解码模式
    </label>
    <div class="grid grid-cols-3 gap-2">
      {#each encodeModes as em}
        <button
          on:click={() => encodeMode = em.value}
          class="p-3 rounded-xl border-2 text-left transition-all {
            encodeMode === em.value
              ? 'border-purple-500 dark:border-purple-400 bg-purple-50 dark:bg-purple-900/20'
              : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
          }"
        >
          <div class="text-sm font-semibold {encodeMode === em.value ? 'text-purple-700 dark:text-purple-300' : 'text-gray-700 dark:text-gray-300'}">
            {em.label}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            {em.desc}
          </div>
        </button>
      {/each}
    </div>
  </div>

  <!-- 输入区域 -->
  <div>
    <div class="flex items-center justify-between mb-2">
      <label class="text-sm font-semibold text-gray-700 dark:text-gray-300">
        {mode === 'encode' ? '输入文本' : '输入 URL 编码'}
      </label>
      <button
        on:click={parseURL}
        class="text-xs flex items-center gap-1 px-3 py-1.5 bg-indigo-100 dark:bg-indigo-900/30 hover:bg-indigo-200 dark:hover:bg-indigo-900/50 rounded-lg transition-colors text-indigo-700 dark:text-indigo-300"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
        解析 URL
      </button>
    </div>
    <textarea
      bind:value={inputText}
      placeholder={mode === 'encode' ? '输入要编码的文本...' : '输入 URL 编码字符串...'}
      on:keydown={handleKeyDown}
      class="w-full h-28 px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm resize-none focus:outline-none focus:border-purple-500 dark:focus:border-purple-400 focus:ring-2 transition-all dark:text-white"
    />
  </div>

  <!-- 操作按钮 -->
  <div class="flex items-center justify-center">
    <button
      on:click={handleAction}
      disabled={isLoading || !inputText.trim()}
      class="flex items-center justify-center gap-2 px-6 py-3 accent-gradient hover:brightness-95 disabled:opacity-50 disabled:cursor-not-allowed rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-xl"
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

  <!-- 输出区域 -->
  <div>
    <div class="flex items-center justify-between mb-2">
      <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300">
        {mode === 'encode' ? 'URL 编码结果' : '解码结果'}
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
      placeholder={mode === 'encode' ? 'URL 编码结果将显示在这里...' : '解码后的文本将显示在这里...'}
      class="w-full h-28 px-4 py-3 bg-gray-100 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm font-mono resize-none focus:outline-none dark:text-white"
    />
  </div>

  <!-- 解析的查询参数 -->
  {#if parsedParams && Object.keys(parsedParams).length > 0}
    <div class="border border-gray-200 dark:border-gray-700 rounded-xl p-4 bg-gray-50 dark:bg-gray-800/50">
      <h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">查询参数</h4>
      <div class="space-y-2 max-h-40 overflow-y-auto">
        {#each Object.entries(parsedParams) as [key, value]}
          <div class="flex items-center gap-2 text-sm">
            <span class="px-2 py-1 bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 rounded-lg font-medium">
              {key}
            </span>
            <span class="text-gray-400">=</span>
            <span class="text-gray-700 dark:text-gray-300 font-mono break-all">{value}</span>
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- 底部操作 -->
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

  <!-- 使用提示 -->
  <div class="text-xs text-gray-500 dark:text-gray-400 bg-gray-100 dark:bg-gray-800 p-3 rounded-lg">
    <p class="font-semibold mb-1">💡 提示:</p>
    <ul class="space-y-1 ml-4 list-disc">
      <li><b>Query 参数</b>：编码空格为 +，适合表单数据</li>
      <li><b>路径</b>：保留斜杠，适合 URL 路径部分</li>
      <li><b>完整 URL</b>：解析并显示 URL 各组成部分</li>
      <li>按 <kbd class="px-1 py-0.5 bg-gray-200 dark:bg-gray-700 rounded">Ctrl+Enter</kbd> 快速执行</li>
    </ul>
  </div>
</div>
