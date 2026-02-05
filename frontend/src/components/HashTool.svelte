<script>
  export let themeStore;

  let inputText = '';
  let outputText = '';
  let selectedAlgorithm = 'md5';
  let mode = 'encrypt';
  let keyHex = '';
  let ivHex = '';
  let isLoading = false;

  const algorithms = [
    { id: 'md5', name: 'MD5', type: 'hash' },
    { id: 'sha256', name: 'SHA-256', type: 'hash' },
    { id: 'sha512', name: 'SHA-512', type: 'hash' },
    { id: 'aes-gcm', name: 'AES-GCM', type: 'crypto' },
    { id: 'aes-cbc', name: 'AES-CBC', type: 'crypto' },
    { id: 'sm4-cbc', name: 'SM4-CBC', type: 'crypto' },
  ];

  function isHashAlgorithm() {
    const algo = algorithms.find(a => a.id === selectedAlgorithm);
    return algo?.type === 'hash';
  }

  function getAlgorithmName() {
    return algorithms.find(a => a.id === selectedAlgorithm)?.name || '';
  }

  function getIvLabel() {
    return selectedAlgorithm === 'aes-gcm' ? 'Nonce' : 'IV';
  }

  function getIvLengthHint() {
    return selectedAlgorithm === 'aes-gcm' ? '12 字节（24位Hex）' : '16 字节（32位Hex）';
  }

  function getKeyLengthHint() {
    if (selectedAlgorithm === 'sm4-cbc') return '16 字节（32位Hex）';
    return '16/24/32 字节（32/48/64位Hex）';
  }

  function getInputPlaceholder() {
    if (isHashAlgorithm()) return '输入要计算哈希的文本...';
    if (mode === 'decrypt') return '输入 Base64 密文...';
    return '输入要加密的明文...';
  }

  async function runOperation() {
    if (!inputText.trim() || isLoading || !window.wailsBindings) return;

    if (!isHashAlgorithm()) {
      if (!keyHex.trim()) return;
      if (!ivHex.trim()) return;
    }

    isLoading = true;
    outputText = '';

    try {
      if (isHashAlgorithm()) {
        outputText = await window.wailsBindings.CalculateHash(inputText, selectedAlgorithm);
      } else if (mode === 'encrypt') {
        outputText = await window.wailsBindings.EncryptText(inputText, selectedAlgorithm, keyHex, ivHex);
      } else {
        outputText = await window.wailsBindings.DecryptText(inputText, selectedAlgorithm, keyHex, ivHex);
      }
    } catch (error) {
      outputText = '';
      console.error('操作失败:', error);
      alert('操作失败: ' + (error.message || error));
    } finally {
      isLoading = false;
    }
  }

  async function copyHash() {
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
    keyHex = '';
    ivHex = '';
  }

  function handleKeyDown(event) {
    const ctrlOrCmd = event.ctrlKey || event.metaKey;

    if (ctrlOrCmd && event.key.toLowerCase() === 'enter') {
      if (!isLoading && inputText.trim()) {
        event.preventDefault();
        runOperation();
      }
    }
  }
</script>

<div class="space-y-6">
  <div>
    <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">选择算法</label>
    <div class="flex flex-col md:flex-row gap-3">
      <div class="flex-1">
        <select
          bind:value={selectedAlgorithm}
          on:change={() => { clearAll(); }}
          class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm focus:outline-none focus:border-green-500 dark:focus:border-green-400 focus:ring-2 transition-all dark:text-white"
        >
          <optgroup label="Hash">
            {#each algorithms.filter(a => a.type === 'hash') as algo}
              <option value={algo.id}>{algo.name}</option>
            {/each}
          </optgroup>
          <optgroup label="加密解密">
            {#each algorithms.filter(a => a.type === 'crypto') as algo}
              <option value={algo.id}>{algo.name}</option>
            {/each}
          </optgroup>
        </select>
      </div>
      {#if !isHashAlgorithm()}
        <div class="flex gap-2 p-1 bg-gray-100 dark:bg-gray-800 rounded-xl">
          <button
            on:click={() => { mode = 'encrypt'; clearAll(); }}
            class="px-4 py-2 rounded-lg text-xs font-semibold transition-all {
              mode === 'encrypt'
                ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm'
                : 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
            }"
          >
            加密
          </button>
          <button
            on:click={() => { mode = 'decrypt'; clearAll(); }}
            class="px-4 py-2 rounded-lg text-xs font-semibold transition-all {
              mode === 'decrypt'
                ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm'
                : 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
            }"
          >
            解密
          </button>
        </div>
      {/if}
    </div>
  </div>

  {#if !isHashAlgorithm()}
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">密钥（Hex）</label>
        <input
          type="text"
          bind:value={keyHex}
          placeholder="例如: 00112233445566778899aabbccddeeff"
          class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm font-mono focus:outline-none focus:border-green-500 dark:focus:border-green-400 focus:ring-2 transition-all dark:text-white"
        />
        <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">密钥长度: {getKeyLengthHint()}</div>
      </div>
      <div>
        <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">{getIvLabel()}（Hex）</label>
        <input
          type="text"
          bind:value={ivHex}
          placeholder={getIvLabel() === 'Nonce' ? '例如: 0f0e0d0c0b0a090807060504' : '例如: 0102030405060708090a0b0c0d0e0f10'}
          class="w-full px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm font-mono focus:outline-none focus:border-green-500 dark:focus:border-green-400 focus:ring-2 transition-all dark:text-white"
        />
        <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">{getIvLabel()}长度: {getIvLengthHint()}</div>
      </div>
    </div>
  {/if}

  <div>
    <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
      {isHashAlgorithm() ? '输入文本' : (mode === 'decrypt' ? '输入 Base64 密文' : '输入明文')}
    </label>
    <textarea
      bind:value={inputText}
      placeholder={getInputPlaceholder()}
      on:keydown={handleKeyDown}
      class="w-full h-32 px-4 py-3 bg-gray-50 dark:bg-gray-800 border-2 border-gray-300 dark:border-gray-600 rounded-xl text-sm resize-none focus:outline-none focus:border-green-500 dark:focus:border-green-400 focus:ring-2 transition-all dark:text-white"
    />
    <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
      {inputText.length} 字符
    </div>
  </div>

  <div class="flex justify-center">
    <button
      on:click={runOperation}
      disabled={isLoading || !inputText.trim() || (!isHashAlgorithm() && (!keyHex.trim() || !ivHex.trim()))}
      class="flex items-center justify-center gap-2 px-8 py-3 bg-gradient-to-r from-green-600 to-emerald-600 hover:from-green-700 hover:to-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed text-white rounded-xl font-semibold transition-all duration-200 shadow-lg hover:shadow-xl"
    >
      {#if isLoading}
        <svg class="animate-spin w-4 h-4" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        处理中...
      {:else}
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
        </svg>
        {isHashAlgorithm() ? '计算哈希' : (mode === 'decrypt' ? '解密' : '加密')}
      {/if}
    </button>
  </div>

  {#if outputText}
    <div>
      <div class="flex items-center justify-between mb-2">
        <label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
          <span>{getAlgorithmName()}</span>
          <span>{isHashAlgorithm() ? '哈希值' : (mode === 'decrypt' ? '明文' : 'Base64 密文')}</span>
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
          {outputText}
        </code>
      </div>
      <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
        长度: {outputText.length} 字符
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
