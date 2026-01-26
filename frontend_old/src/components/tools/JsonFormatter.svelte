<script>
  import { FormatJSON, ValidateJSON } from '../../../wailsjs/go/main/App.js';

  let inputText = '';
  let outputText = '';
  let errorMessage = '';
  let isFormatting = false;
  let validationStatus = null; // 'valid' | 'invalid' | null

  // 实时验证（防抖）
  let validationTimer;
  $: {
    clearTimeout(validationTimer);
    if (inputText.trim()) {
      validationTimer = setTimeout(validateInput, 500);
    } else {
      validationStatus = null;
      errorMessage = '';
    }
  }

  async function validateInput() {
    if (!inputText.trim()) return;

    try {
      const result = await ValidateJSON(inputText);
      if (result.valid) {
        validationStatus = 'valid';
        errorMessage = '';
      } else {
        validationStatus = 'invalid';
        errorMessage = result.error || '无效的JSON格式';
      }
    } catch (err) {
      validationStatus = 'invalid';
      errorMessage = err.toString();
    }
  }

  async function formatJson() {
    if (!inputText.trim()) {
      errorMessage = '请输入JSON内容';
      return;
    }

    isFormatting = true;
    errorMessage = '';

    try {
      const result = await FormatJSON(inputText);
      outputText = result;
      validationStatus = 'valid';
    } catch (err) {
      errorMessage = `格式化失败: ${err}`;
      validationStatus = 'invalid';
      outputText = '';
    } finally {
      isFormatting = false;
    }
  }

  function clearAll() {
    inputText = '';
    outputText = '';
    errorMessage = '';
    validationStatus = null;
  }

  function copyOutput() {
    if (!outputText) return;

    navigator.clipboard.writeText(outputText)
      .then(() => {
        // 显示临时提示
        const btn = document.querySelector('.copy-btn');
        if (btn) {
          const originalText = btn.textContent;
          btn.textContent = '✓ 已复制';
          setTimeout(() => {
            btn.textContent = originalText;
          }, 1500);
        }
      })
      .catch(err => {
        errorMessage = `复制失败: ${err}`;
      });
  }

  function minifyJson() {
    if (!outputText) return;
    try {
      const minified = JSON.stringify(JSON.parse(outputText));
      outputText = minified;
    } catch (err) {
      errorMessage = `压缩失败: ${err}`;
    }
  }

  // 语法高亮处理（简单实现）
  function highlightJson(text) {
    if (!text) return '';

    try {
      return text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, (match) => {
          let cls = 'json-number';
          if (/^"/.test(match)) {
            if (/:$/.test(match)) {
              cls = 'json-key';
            } else {
              cls = 'json-string';
            }
          } else if (/true|false/.test(match)) {
            cls = 'json-boolean';
          } else if (/null/.test(match)) {
            cls = 'json-null';
          }
          return `<span class="${cls}">${match}</span>`;
        });
    } catch (e) {
      return text;
    }
  }

  $: highlightedOutput = highlightJson(outputText);
  $: charCount = inputText.length;
  $: lineCount = inputText ? inputText.split('\n').length : 0;
</script>

<div class="json-formatter">
  <div class="toolbar">
    <button
      class="btn btn-primary"
      on:click={formatJson}
      disabled={isFormatting || !inputText.trim()}
      title="格式化JSON（美化输出）"
    >
      {#if isFormatting}
        <span class="spinner"></span> 格式化中...
      {:else}
        ✨ 格式化
      {/if}
    </button>

    <button
      class="btn btn-secondary"
      on:click={minifyJson}
      disabled={!outputText}
      title="压缩JSON（移除空白）"
    >
      🗜️ 压缩
    </button>

    <button
      class="btn btn-secondary"
      on:click={clearAll}
      title="清空所有内容"
    >
      🗑️ 清空
    </button>

    <div class="spacer"></div>

    {#if validationStatus === 'valid'}
      <span class="status-badge valid" title="JSON格式有效">✓ 有效</span>
    {:else if validationStatus === 'invalid'}
      <span class="status-badge invalid" title={errorMessage}>✗ 无效</span>
    {/if}
  </div>

  {#if errorMessage}
    <div class="error-message">
      <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
        <path d="M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0zM7 11.5v1h2v-1H7zm0-7v5h2v-5H7z"/>
      </svg>
      <span>{errorMessage}</span>
      <button class="dismiss-btn" on:click={() => errorMessage = ''}>×</button>
    </div>
  {/if}

  <div class="editor-container">
    <div class="editor-panel">
      <div class="panel-header">
        <span class="panel-title">📝 输入</span>
        <div class="panel-stats">
          <span class="stat">{charCount} 字符</span>
          <span class="stat-divider">|</span>
          <span class="stat">{lineCount} 行</span>
        </div>
      </div>
      <textarea
        class="json-input"
        bind:value={inputText}
        placeholder="在此粘贴或输入JSON内容..."
        spellcheck="false"
      ></textarea>
    </div>

    <div class="divider"></div>

    <div class="editor-panel">
      <div class="panel-header">
        <span class="panel-title">📄 输出</span>
        {#if outputText}
          <button class="copy-btn" on:click={copyOutput} title="复制到剪贴板">
            📋 复制
          </button>
        {/if}
      </div>
      <div class="json-output" class:empty={!outputText}>
        {#if outputText}
          <pre><code>{@html highlightedOutput}</code></pre>
        {:else}
          <div class="placeholder">
            <div class="placeholder-icon">JSON</div>
            <div class="placeholder-text">格式化后的JSON将在这里显示</div>
            <div class="placeholder-hint">支持语法高亮</div>
          </div>
        {/if}
      </div>
    </div>
  </div>
</div>

<style>
  .json-formatter {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--bg-primary);
  }

  .toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px;
    border-bottom: 1px solid var(--border-primary);
    background: var(--bg-secondary);
    flex-wrap: wrap;
  }

  .btn {
    padding: 6px 12px;
    border: 1px solid var(--border-primary);
    border-radius: 4px;
    background: var(--bg-primary);
    color: var(--text-primary);
    cursor: pointer;
    font-size: 13px;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    gap: 4px;
    white-space: nowrap;
  }

  .btn:hover:not(:disabled) {
    background: var(--bg-hover);
    transform: translateY(-1px);
  }

  .btn:active:not(:disabled) {
    transform: translateY(0);
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    background: var(--accent-primary);
    color: white;
    border-color: var(--accent-primary);
  }

  .btn-primary:hover:not(:disabled) {
    opacity: 0.9;
    background: var(--accent-primary);
  }

  .spacer {
    flex: 1;
  }

  .spinner {
    display: inline-block;
    width: 12px;
    height: 12px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .status-badge {
    padding: 4px 10px;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 600;
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .status-badge.valid {
    background: rgba(34, 197, 94, 0.15);
    color: #22c55e;
  }

  .status-badge.invalid {
    background: rgba(239, 68, 68, 0.15);
    color: #ef4444;
  }

  .error-message {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    background: rgba(239, 68, 68, 0.1);
    color: #ef4444;
    border-bottom: 1px solid rgba(239, 68, 68, 0.3);
    font-size: 13px;
  }

  .error-message span {
    flex: 1;
  }

  .dismiss-btn {
    width: 20px;
    height: 20px;
    border: none;
    background: transparent;
    color: currentColor;
    cursor: pointer;
    font-size: 18px;
    line-height: 1;
    padding: 0;
    opacity: 0.7;
    transition: opacity 0.2s;
  }

  .dismiss-btn:hover {
    opacity: 1;
  }

  .editor-container {
    display: flex;
    flex: 1;
    overflow: hidden;
  }

  .editor-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-primary);
    font-size: 12px;
    gap: 8px;
  }

  .panel-title {
    font-weight: 600;
    color: var(--text-primary);
  }

  .panel-stats {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--text-tertiary);
    font-size: 11px;
  }

  .stat {
    font-variant-numeric: tabular-nums;
  }

  .stat-divider {
    opacity: 0.5;
  }

  .copy-btn {
    padding: 4px 8px;
    border: none;
    background: transparent;
    color: var(--accent-primary);
    cursor: pointer;
    font-size: 12px;
    border-radius: 3px;
    transition: all 0.2s;
    font-weight: 500;
  }

  .copy-btn:hover {
    background: var(--bg-hover);
  }

  .divider {
    width: 1px;
    background: var(--border-primary);
    flex-shrink: 0;
  }

  .json-input,
  .json-output {
    flex: 1;
    padding: 12px;
    font-family: 'SF Mono', 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
    font-size: 13px;
    line-height: 1.6;
    overflow: auto;
  }

  .json-input {
    border: none;
    background: var(--bg-primary);
    color: var(--text-primary);
    resize: none;
    outline: none;
  }

  .json-input::placeholder {
    color: var(--text-tertiary);
    opacity: 0.6;
  }

  .json-output {
    background: var(--bg-primary);
    color: var(--text-primary);
  }

  .json-output.empty {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .placeholder {
    text-align: center;
    color: var(--text-tertiary);
  }

  .placeholder-icon {
    font-size: 48px;
    font-family: 'SF Mono', 'Monaco', monospace;
    margin-bottom: 12px;
    opacity: 0.3;
  }

  .placeholder-text {
    font-size: 14px;
    margin-bottom: 4px;
  }

  .placeholder-hint {
    font-size: 12px;
    opacity: 0.7;
  }

  .json-output pre {
    margin: 0;
    white-space: pre-wrap;
    word-wrap: break-word;
  }

  .json-output code {
    font-family: inherit;
    font-size: inherit;
  }

  /* JSON语法高亮样式 */
  :global(.json-key) {
    color: #0066cc;
    font-weight: 600;
  }

  :global(.json-string) {
    color: #22863a;
  }

  :global(.json-number) {
    color: #005cc5;
  }

  :global(.json-boolean) {
    color: #d73a49;
    font-weight: 600;
  }

  :global(.json-null) {
    color: #6f42c1;
    font-style: italic;
  }

  /* 暗色主题适配 */
  :global([data-theme="dark"]) :global(.json-key) {
    color: #79b8ff;
  }

  :global([data-theme="dark"]) :global(.json-string) {
    color: #85e89d;
  }

  :global([data-theme="dark"]) :global(.json-number) {
    color: #79b8ff;
  }

  :global([data-theme="dark"]) :global(.json-boolean) {
    color: #f97583;
  }

  :global([data-theme="dark"]) :global(.json-null) {
    color: #b392f0;
  }

  /* 滚动条样式 */
  .json-input::-webkit-scrollbar,
  .json-output::-webkit-scrollbar {
    width: 10px;
    height: 10px;
  }

  .json-input::-webkit-scrollbar-track,
  .json-output::-webkit-scrollbar-track {
    background: var(--bg-secondary);
  }

  .json-input::-webkit-scrollbar-thumb,
  .json-output::-webkit-scrollbar-thumb {
    background: var(--border-primary);
    border-radius: 5px;
  }

  .json-input::-webkit-scrollbar-thumb:hover,
  .json-output::-webkit-scrollbar-thumb:hover {
    background: var(--text-tertiary);
  }
</style>
