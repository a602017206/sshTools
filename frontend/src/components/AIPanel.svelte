<script>
  import { onMount, tick } from 'svelte';
  import { connectionsStore } from '../stores.js';
  import { copilotStore } from '../stores/copilot.js';
  import {
    applySqlEvent,
    executeSqlEvent,
    shellExecutePayload
  } from '../lib/copilotApply.js';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';

  export let sessionId = null;
  export let mode = 'ssh';
  export let hasSession = false;
  export let onOpenSettings = () => {};
  export let onInsertShell = null;

  let draft = '';
  let generating = false;
  let hasApiKey = false;
  let checkingKey = true;
  let errorMessage = '';
  let showDangerConfirm = false;
  let dangerTitle = '确认执行危险操作';
  let dangerMessage = '';
  let resolveDangerConfirm = null;

  $: messages = ($copilotStore.messagesBySession?.[sessionId] || []);
  $: backendSessionId = resolveBackendSessionId(sessionId, $connectionsStore);
  $: copilotMode = mode === 'database' ? 'database' : 'ssh';

  function resolveBackendSessionId(id, conns) {
    if (!id || !conns || typeof conns.get !== 'function') return id;
    const session = conns.get(id);
    return session?.dbSessionId || session?.sessionId || id;
  }

  function getBindings() {
    return window.wailsBindings || window.go?.main?.App || {};
  }

  async function refreshHasApiKey() {
    const api = getBindings();
    checkingKey = true;
    try {
      if (typeof api.HasCopilotAPIKey !== 'function') {
        hasApiKey = false;
        return;
      }
      hasApiKey = Boolean(await api.HasCopilotAPIKey());
    } catch (error) {
      console.error('Failed to check copilot API key:', error);
      hasApiKey = false;
    } finally {
      checkingKey = false;
    }
  }

  let keyCheckedForOpen = false;

  onMount(() => {
    refreshHasApiKey();
  });

  $: if ($copilotStore.open && !keyCheckedForOpen) {
    keyCheckedForOpen = true;
    refreshHasApiKey();
  }
  $: if (!$copilotStore.open) {
    keyCheckedForOpen = false;
  }

  function historyForRequest() {
    return messages
      .filter((item) => item && (item.role === 'user' || item.role === 'assistant') && item.content)
      .map((item) => ({
        Role: item.role,
        Content: String(item.content)
      }));
  }

  function normalizeChatResponse(raw) {
    if (!raw || typeof raw !== 'object') {
      return { reply: '', artifact: null, toolNotes: [] };
    }
    const artifact = raw.artifact || raw.Artifact || null;
    return {
      reply: raw.reply || raw.Reply || '',
      artifact: artifact && (artifact.content || artifact.Content)
        ? {
            type: artifact.type || artifact.Type || '',
            content: artifact.content || artifact.Content || '',
            summary: artifact.summary || artifact.Summary || '',
            destructive: Boolean(artifact.destructive ?? artifact.Destructive)
          }
        : null,
      toolNotes: raw.tool_notes || raw.ToolNotes || []
    };
  }

  function formatError(error) {
    if (!error) return '请求失败';
    if (typeof error === 'string') return error;
    return error.message || String(error);
  }

  async function sendMessage() {
    const text = String(draft || '').trim();
    if (!text || generating || !hasSession || !hasApiKey) return;
    const api = getBindings();
    if (typeof api.CopilotChat !== 'function') {
      errorMessage = 'Copilot 绑定不可用';
      return;
    }

    generating = true;
    errorMessage = '';
    draft = '';
    const history = historyForRequest();
    copilotStore.appendMessage(sessionId, { role: 'user', content: text });

    try {
      const response = await api.CopilotChat({
        SessionID: backendSessionId,
        Mode: copilotMode,
        Message: text,
        History: history,
        EditorContent: '',
        TerminalTail: ''
      });
      const normalized = normalizeChatResponse(response);
      copilotStore.appendMessage(sessionId, {
        role: 'assistant',
        content: normalized.reply || (normalized.artifact ? normalized.artifact.summary : ''),
        artifact: normalized.artifact,
        toolNotes: normalized.toolNotes
      });
    } catch (error) {
      errorMessage = formatError(error);
      copilotStore.appendMessage(sessionId, {
        role: 'assistant',
        content: `生成失败：${formatError(error)}`
      });
    } finally {
      generating = false;
    }
  }

  function handleComposerKeydown(event) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      sendMessage();
    }
  }

  function applyArtifact(artifact) {
    if (!artifact?.content || !backendSessionId) return;
    if ((artifact.type || copilotMode) === 'sql' || copilotMode === 'database') {
      window.dispatchEvent(applySqlEvent(backendSessionId, artifact.content));
      return;
    }
    if (typeof onInsertShell === 'function') {
      onInsertShell(backendSessionId, artifact.content);
      return;
    }
    const api = getBindings();
    if (typeof api.SendSSHData === 'function') {
      api.SendSSHData(backendSessionId, artifact.content);
    }
  }

  function confirmDanger(reason) {
    dangerMessage = reason || '该操作可能造成破坏，确认后才会执行。';
    showDangerConfirm = true;
    return new Promise((resolve) => {
      resolveDangerConfirm = resolve;
    });
  }

  function handleDangerConfirm() {
    showDangerConfirm = false;
    if (resolveDangerConfirm) {
      resolveDangerConfirm(true);
      resolveDangerConfirm = null;
    }
  }

  function handleDangerCancel() {
    showDangerConfirm = false;
    if (resolveDangerConfirm) {
      resolveDangerConfirm(false);
      resolveDangerConfirm = null;
    }
  }

  async function classifyAndConfirm(kind, content) {
    const api = getBindings();
    if (typeof api.CopilotClassify !== 'function') {
      return true;
    }
    const result = await api.CopilotClassify(kind, content);
    const destructive = Boolean(result?.Destructive ?? result?.destructive);
    if (!destructive) return true;
    return confirmDanger(result?.Reason || result?.reason || '');
  }

  async function executeArtifact(artifact) {
    if (!artifact?.content || !backendSessionId || generating) return;
    const kind = copilotMode === 'database' || artifact.type === 'sql' ? 'sql' : 'shell';
    const confirmed = await classifyAndConfirm(kind, artifact.content);
    if (!confirmed) return;

    try {
      if (kind === 'sql') {
        window.dispatchEvent(applySqlEvent(backendSessionId, artifact.content));
        await tick();
        const handled = { value: false };
        window.dispatchEvent(executeSqlEvent(backendSessionId, handled));
        await tick();
        if (!handled.value) {
          const api = getBindings();
          if (typeof api.ExecuteDatabaseQuery !== 'function') {
            throw new Error('数据库执行不可用');
          }
          const sql = String(artifact.content).trim().replace(/;+\s*$/g, '');
          const raw = await api.ExecuteDatabaseQuery(backendSessionId, sql);
          let summary = '已执行';
          try {
            const data = typeof raw === 'string' ? JSON.parse(raw) : raw;
            const rows = data?.rows?.length ?? data?.rowCount;
            if (typeof rows === 'number') {
              summary = `已执行，返回 ${rows} 行`;
            }
          } catch {
            summary = '已执行';
          }
          copilotStore.appendMessage(sessionId, { role: 'assistant', content: summary });
          return;
        }
        copilotStore.appendMessage(sessionId, { role: 'assistant', content: '已执行，结果见数据库面板。' });
        return;
      }

      const payload = shellExecutePayload(artifact.content);
      if (typeof onInsertShell === 'function') {
        onInsertShell(backendSessionId, payload);
      } else {
        const api = getBindings();
        if (typeof api.SendSSHData !== 'function') {
          throw new Error('终端执行不可用');
        }
        await api.SendSSHData(backendSessionId, payload);
      }
      copilotStore.appendMessage(sessionId, { role: 'assistant', content: '已发送到终端执行。' });
    } catch (error) {
      copilotStore.appendMessage(sessionId, {
        role: 'assistant',
        content: `执行失败：${formatError(error)}`
      });
    }
  }
</script>

<aside class="ai-panel" aria-label="AI Copilot">
  <header class="ai-panel__header">
    <div>
      <p class="ai-panel__kicker">AI Copilot</p>
      <h2>{copilotMode === 'database' ? 'SQL 助手' : 'Shell 助手'}</h2>
    </div>
    <button
      type="button"
      class="ops-icon-button flex items-center justify-center w-7 h-7 rounded-md"
      title="关闭"
      on:click={() => copilotStore.setOpen(false)}
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
      </svg>
    </button>
  </header>

  {#if !hasSession}
    <div class="ai-panel__empty" role="status">
      <strong>先连接主机或数据库</strong>
      <span>侧栏会跟随当前标签生成 SQL 或 Shell。</span>
    </div>
  {:else if checkingKey}
    <div class="ai-panel__empty" role="status">
      <span>正在检查 API Key…</span>
    </div>
  {:else if !hasApiKey}
    <div class="ai-panel__empty" role="status">
      <strong>尚未配置 API Key</strong>
      <span>请先在设置中填写 Base URL、模型名称和密钥。</span>
      <button type="button" class="ai-panel__cta" on:click={onOpenSettings}>去设置</button>
    </div>
  {:else}
    <div class="ai-panel__thread">
      {#if messages.length === 0}
        <div class="ai-panel__hint">用自然语言描述你想查的数据或要跑的命令。</div>
      {/if}
      {#each messages as item, index (index)}
        <div class="ai-panel__bubble" class:ai-panel__bubble--user={item.role === 'user'}>
          <p>{item.content}</p>
          {#if item.artifact}
            <div class="ai-panel__artifact">
              {#if item.artifact.summary}
                <p class="ai-panel__summary">{item.artifact.summary}</p>
              {/if}
              <pre><code>{item.artifact.content}</code></pre>
              <div class="ai-panel__actions">
                <button type="button" on:click={() => applyArtifact(item.artifact)}>填入</button>
                <button type="button" class="ai-panel__run" on:click={() => executeArtifact(item.artifact)}>执行</button>
              </div>
            </div>
          {/if}
        </div>
      {/each}
      {#if generating}
        <div class="ai-panel__hint">正在生成…</div>
      {/if}
      {#if errorMessage}
        <div class="ai-panel__error">{errorMessage}</div>
      {/if}
    </div>

    <form class="ai-panel__composer" on:submit|preventDefault={sendMessage}>
      <textarea
        bind:value={draft}
        placeholder={copilotMode === 'database' ? '描述要生成的 SQL…' : '描述要生成的命令…'}
        disabled={generating}
        on:keydown={handleComposerKeydown}
      ></textarea>
      <button type="submit" disabled={generating || !draft.trim()}>发送</button>
    </form>
  {/if}
</aside>

<ConfirmDialog
  bind:isOpen={showDangerConfirm}
  title={dangerTitle}
  message={dangerMessage}
  type="danger"
  confirmText="执行"
  cancelText="取消"
  onConfirm={handleDangerConfirm}
  onCancel={handleDangerCancel}
/>

<style>
  .ai-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
    background: transparent;
    color: var(--text-primary);
  }

  .ai-panel__header {
    flex-shrink: 0;
    padding: 12px 14px 10px;
    border-bottom: 1px solid var(--glass-border);
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 10px;
    background: color-mix(in srgb, var(--glass-bg) 70%, transparent);
  }

  .ai-panel__kicker {
    margin: 0;
    font-size: 10px;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--text-secondary);
  }

  .ai-panel__header h2 {
    margin: 2px 0 0;
    font-size: 14px;
    font-weight: 650;
  }

  .ai-panel__empty,
  .ai-panel__thread {
    flex: 1;
    min-height: 0;
    overflow-y: auto;
  }

  .ai-panel__empty {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    justify-content: center;
    gap: 8px;
    padding: 24px 16px;
    color: var(--text-secondary);
  }

  .ai-panel__empty strong {
    color: var(--text-primary);
    font-size: 14px;
  }

  .ai-panel__empty span {
    font-size: 12px;
    line-height: 1.5;
  }

  .ai-panel__cta,
  .ai-panel__composer button,
  .ai-panel__actions button {
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: var(--glass-bg);
    color: var(--text-primary);
    cursor: pointer;
    font-size: 12px;
  }

  .ai-panel__cta {
    margin-top: 6px;
    padding: 6px 12px;
    background: var(--accent-primary);
    border-color: var(--accent-primary);
    color: #fff;
  }

  .ai-panel__thread {
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .ai-panel__hint,
  .ai-panel__error {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .ai-panel__error {
    color: #c43832;
  }

  .ai-panel__bubble {
    align-self: flex-start;
    max-width: 100%;
    padding: 8px 10px;
    border: 1px solid var(--glass-border);
    border-radius: 12px;
    background: color-mix(in srgb, var(--glass-bg) 80%, transparent);
    font-size: 13px;
    line-height: 1.5;
    white-space: pre-wrap;
    word-break: break-word;
  }

  .ai-panel__bubble--user {
    align-self: flex-end;
    background: color-mix(in srgb, var(--accent-primary) 14%, transparent);
  }

  .ai-panel__bubble p {
    margin: 0;
  }

  .ai-panel__artifact {
    margin-top: 8px;
    display: grid;
    gap: 8px;
  }

  .ai-panel__summary {
    color: var(--text-secondary);
    font-size: 12px;
  }

  .ai-panel__artifact pre {
    margin: 0;
    max-height: 180px;
    overflow: auto;
    padding: 8px;
    border-radius: 8px;
    background: var(--bg-secondary);
    font: 11px ui-monospace, SFMono-Regular, Menlo, monospace;
  }

  .ai-panel__actions {
    display: flex;
    gap: 8px;
  }

  .ai-panel__actions button {
    padding: 4px 10px;
  }

  .ai-panel__run {
    background: var(--accent-primary) !important;
    border-color: var(--accent-primary) !important;
    color: #fff !important;
  }

  .ai-panel__composer {
    flex-shrink: 0;
    padding: 10px 12px 12px;
    border-top: 1px solid var(--glass-border);
    display: grid;
    gap: 8px;
  }

  .ai-panel__composer textarea {
    width: 100%;
    min-height: 72px;
    resize: vertical;
    padding: 8px 10px;
    border: 1px solid var(--glass-border);
    border-radius: 10px;
    background: var(--bg-primary);
    color: var(--text-primary);
    font-size: 13px;
  }

  .ai-panel__composer textarea:disabled,
  .ai-panel__composer button:disabled {
    opacity: 0.55;
    cursor: not-allowed;
  }

  .ai-panel__composer button {
    justify-self: end;
    padding: 6px 14px;
    background: var(--accent-primary);
    border-color: var(--accent-primary);
    color: #fff;
  }
</style>
