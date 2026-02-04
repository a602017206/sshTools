<script>
  export let themeStore;

  let inputText = '';
  let outputHash = '';
  let selectedAlgorithm = 'md5';
  let isLoading = false;

  const algorithms = [
    { id: 'md5', name: 'MD5', color: 'from-red-500 to-pink-500', icon: '🔒' },
    { id: 'sha256', name: 'SHA-256', color: 'from-blue-500 to-cyan-500', icon: '🔐' },
    { id: 'sha512', name: 'SHA-512', color: 'from-purple-500 to-indigo-500', icon: '🔑' },
  ];

  async function calculateHash() {
    if (!inputText.trim() || isLoading || !window.wailsBindings) return;

    isLoading = true;
    outputHash = '';

    try {
      outputHash = await window.wailsBindings.CalculateHash(inputText, selectedAlgorithm);
    } catch (error) {
      outputHash = '';
      console.error('哈希计算失败:', error);
      alert('哈希计算失败: ' + (error.message || error));
    } finally {
      isLoading = false;
    }
  }

  async function copyHash() {
    if (!outputHash) return;

    try {
      await navigator.clipboard.writeText(outputHash);
    } catch (error) {
      console.error('复制失败:', error);
    }
  }

  function clearAll() {
    inputText = '';
    outputHash = '';
  }

  function handleKeyDown(event) {
    const ctrlOrCmd = event.ctrlKey || event.metaKey;

    if (ctrlOrCmd && event.key.toLowerCase() === 'enter') {
      if (!isLoading && inputText.trim()) {
        event.preventDefault();
        calculateHash();
      }
    }
  }
</script>

<div class="space-y-6">
  <div>
    <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">选择哈希算法</label>
    <div class="grid grid-cols-3 gap-3">
      {#each algorithms as algo}
        <button
          on:click={() => selectedAlgorithm = algo.id}
          class="relative p-4 rounded-xl transition-all duration-200 group {
            selectedAlgorithm === algo.id
              ? 'bg-gradient-to-br ' + algo.color + ' text-white shadow-lg transform scale-105'
              : 'bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-700 border-2 border-transparent hover:border-gray-300 dark:hover:border-gray-600'
          }"
        >
          <div class="flex flex-col items-center gap-2">
            <span class="text-2xl">{algo.icon}</span>
            <span class="text-sm font-semibold">{algo.name}</span>
            {#if selectedAlgorithm === algo.id}
              <svg class="absolute top-2 right-2 w-4 h-4 text-white/80" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
              </svg>
            {/if}
          </div>
        </button>
      {/each}
    </div>
  </div>

  <div>
    <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">输入文本</label>
    <textarea
      bind:value={inputText}
      placeholder="输入要计算哈希的文本..."
      on:keydown={handleKeyDown}
      class="w-full h-32 px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm resize-none focus:outline-none focus:border-green-500 dark:focus:border-green-400 focus:ring-2 transition-all dark:text-white"
    />
    <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
      {inputText.length} 字符
    </div>
  </div>

  <div class="flex justify-center">
    <button
      on:click={calculateHash}
      disabled={isLoading || !inputText.trim()}
      class="flex items-center justify-center gap-2 px-8 py-3 bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-700 hover:to-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-xl"
    >
      {#if isLoading}
        <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        计算中...
      {:else}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
        </svg>
        计算哈希
      {/if}
    </button>
  </div>

  {#if outputHash}
    <div>
      <div class="flex items-center justify-between mb-2">
        <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
          <span class="bg-gradient-to-r {algorithms.find(a => a.id === selectedAlgorithm)?.color} bg-clip-text text-transparent">
            {algorithms.find(a => a.id === selectedAlgorithm)?.name}
          </span>
          <span>哈希值</span>
        </label>
        <button
          on:click={copyHash}
          class="text-xs flex items-center gap-1 px-3 py-1.5 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 rounded-lg transition-colors text-gray-700 dark:text-gray-300"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
          复制
        </button>
      </div>
      <div class="p-4 bg-gray-100 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl">
        <code class="text-sm font-mono break-all text-gray-800 dark:text-gray-200">
          {outputHash}
        </code>
      </div>
      <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
        哈希长度: {outputHash.length} 字符
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
