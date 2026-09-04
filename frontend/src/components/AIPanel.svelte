<script>
  import { onMount, tick } from 'svelte';
  import { connectionsStore, databaseNavigationStore } from '../stores.js';
  import { copilotStore } from '../stores/copilot.js';
  import {
    applySqlEvent,
    executeSqlEvent,
    peekSqlEvent,
    shouldUsePanelPath,
    shellExecutePayload,
    applyNativeEvent,
    executeNativeEvent,
    isNativeArtifact
  } from '../lib/copilotApply.js';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  import {
    buildChatHistory,
    buildCopilotWorkspaceContext,
    copilotAssistantTitle,
    copilotChatPayload,
    formatCopilotWorkspaceLabel,
    resolveWorkspaceFocus
  } from '../lib/copilotContext.js';
  import { isCopilotCancelError, shouldSubmitComposerOnEnter } from '../lib/composerKeys.js';

  export let sessionId = null;
  export let mode = 'ssh';
  export let hasSession = false;
  export let onOpenSettings = () => {};
  export let onInsertShell = null;

  let draft = '';
  let generating = false;
  let generationToken = 0;
  let hasApiKey = false;
  let checkingKey = true;
  let errorMessage = '';
  let showDangerConfirm = false;
  let dangerTitle = '确认执行危险操作';
  let dangerMessage = '';
  let resolveDangerConfirm = null;

  $: messages = ($copilotStore.messagesBySession?.[sessionId] || []);
  $: terminalTail = $copilotStore.terminalTailsBySession?.[sessionId] || '';
  $: backendSessionId = resolveBackendSessionId(sessionId, $connectionsStore);
  $: copilotMode = mode === 'database' ? 'database' : 'ssh';
  $: copilotSession = ($connectionsStore && typeof $connectionsStore.get === 'function')
    ? $connectionsStore.get(sessionId)
    : null;
  $: copilotNavigation = $databaseNavigationStore?.[backendSessionId] || $databaseNavigationStore?.[sessionId] || null;
  $: copilotFocus = resolveWorkspaceFocus($copilotStore.workspaceFocusBySession, sessionId, backendSessionId);
  $: workspaceContext = buildCopilotWorkspaceContext({
    session: copilotSession,
    navigation: copilotNavigation,
    focus: copilotFocus,
    mode: copilotMode
  });
  $: workspaceLabel = formatCopilotWorkspaceLabel(workspaceContext);
  $: assistantTitle = copilotAssistantTitle(workspaceContext, copilotMode);
  $: composerPlaceholder = workspaceContext?.workspaceKind === 'native'
    ? (assistantTitle.includes('搜索')
      ? '描述要查的索引、DSL 或文档变更…'
      : assistantTitle.includes('缓存')
        ? '描述要 SCAN 的键、查看内容或起草删键…'
        : '描述要查询或变更的资源…')
    : (copilotMode === 'database' ? '描述要生成的 SQL…' : '描述要生成的命令…');
  $: emptyHint = workspaceContext?.workspaceKind === 'native'
    ? '可让我列出资源、解释 mapping、生成 DSL，或起草变更（需你确认后执行）。'
    : '用自然语言描述你想查的数据或要跑的命令。';

  function resolveBackendSessionId(id, conns) {
    if (!id || !conns || typeof conns.get !== 'function') return id;
    const session = conns.get(id);
    return session?.dbSessionId || session?.sessionId || id;
  }

  function getBindings() {
    // 优先 Wails 运行时注入对象（含全部 Go 导出方法，包括尚未生成的 Copilot 系列），
    // 再回退到生成的绑定模块，最后空对象，避免运行时回退不可达。
    return window.go?.main?.App || window.wailsBindings || {};
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
    return buildChatHistory(messages);
  }

  function startNewChat() {
    if (generating || !sessionId) return;
    copilotStore.clearSession(sessionId);
    errorMessage = '';
  }

  function normalizeChatResponse(raw) {
    if (!raw || typeof raw !== 'object') {
      return { reply: '', artifact: null, toolNotes: [] };
    }
    const artifact = raw.artifact || raw.Artifact || null;
    let normalized = artifact && (artifact.content || artifact.Content)
      ? {
          type: artifact.type || artifact.Type || '',
          content: artifact.content || artifact.Content || '',
          summary: artifact.summary || artifact.Summary || '',
          destructive: Boolean(artifact.destructive ?? artifact.Destructive)
        }
      : null;
    // 原生会话若模型仍回 sql/shell，前端也改成 native_query，保证出现填入/执行。
    if (normalized && workspaceContext?.workspaceKind === 'native') {
      const type = String(normalized.type || '').toLowerCase();
      if (type === 'sql' || type === 'shell') {
        normalized = { ...normalized, type: 'native_query', destructive: false };
      }
    }
    return {
      reply: raw.reply || raw.Reply || '',
      artifact: normalized,
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

    const token = ++generationToken;
    generating = true;
    errorMessage = '';
    draft = '';
    const history = historyForRequest();
    const requestSessionId = backendSessionId;
    copilotStore.appendMessage(sessionId, { role: 'user', content: text });

    try {
      const response = await api.CopilotChat(copilotChatPayload(workspaceContext, {
        sessionID: requestSessionId,
        mode: copilotMode,
        message: text,
        history,
        terminalTail
      }));
      if (token !== generationToken) return;
      const normalized = normalizeChatResponse(response);
      copilotStore.appendMessage(sessionId, {
        role: 'assistant',
        content: normalized.reply || (normalized.artifact ? normalized.artifact.summary : ''),
        artifact: normalized.artifact,
        toolNotes: normalized.toolNotes
      });
    } catch (error) {
      if (token !== generationToken || isCopilotCancelError(error)) {
        return;
      }
      errorMessage = formatError(error);
      copilotStore.appendMessage(sessionId, {
        role: 'assistant',
        content: `生成失败：${formatError(error)}`
      });
    } finally {
      if (token === generationToken) {
        generating = false;
      }
    }
  }

  function stopGeneration() {
    if (!generating) return;
    generationToken += 1;
    generating = false;
    const api = getBindings();
    const id = backendSessionId || sessionId;
    if (typeof api.CopilotCancel === 'function' && id) {
      try {
        api.CopilotCancel(id);
      } catch (error) {
        console.warn('CopilotCancel failed', error);
      }
    }
    if (sessionId) {
      copilotStore.appendMessage(sessionId, {
        role: 'assistant',
        content: '已停止生成。'
      });
    }
  }

  function handleComposerKeydown(event) {
    if (!shouldSubmitComposerOnEnter(event)) return;
    event.preventDefault();
    sendMessage();
  }

  function applyArtifact(artifact) {
    if (!artifact?.content || !backendSessionId) return;
    if (isNativeArtifact(artifact)) {
      window.dispatchEvent(applyNativeEvent(backendSessionId, artifact));
      return;
    }
    if ((artifact.type || copilotMode) === 'sql' || (copilotMode === 'database' && artifact.type !== 'shell')) {
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
      // 绑定缺失：写操作一律二次确认，fail-closed，不静默放行。
      return confirmDanger('无法完成操作分类，请手动确认后再执行。');
    }
    try {
      const result = await api.CopilotClassify(kind, content);
      const destructive = Boolean(result?.Destructive ?? result?.destructive);
      if (!destructive) return true;
      return confirmDanger(result?.Reason || result?.reason || '');
    } catch (error) {
      // 分类抛错：fail-closed，弹危险确认，用户取消则不执行。
      return confirmDanger('操作分类失败，请手动确认后再执行。');
    }
  }

  async function executeArtifact(artifact) {
    if (!artifact?.content || !backendSessionId || generating) return;

    try {
      if (isNativeArtifact(artifact)) {
        const needsConfirm = artifact.type === 'native_mutation';
        const confirmed = needsConfirm
          ? await confirmDanger(artifact.summary || '确认执行该原生变更？')
          : true;
        if (!confirmed) return;
        window.dispatchEvent(executeNativeEvent(backendSessionId, artifact));
        copilotStore.appendMessage(sessionId, {
          role: 'assistant',
          content: needsConfirm ? '已确认并交给工作区执行。' : '已交给工作区执行（见 Redis / ES 面板）。'
        });
        return;
      }

      const kind = copilotMode === 'database' || artifact.type === 'sql' ? 'sql' : 'shell';

      if (kind === 'sql') {
        // 先 peek：有打开的查询/表面板时，由面板同步回填当前编辑器 query。
        const peek = { found: false, query: '' };
        window.dispatchEvent(peekSqlEvent(backendSessionId, peek));
        await tick();

        if (shouldUsePanelPath(peek)) {
          // 有打开的查询/表面板且编辑器非空：分类将要执行的 SQL（当前编辑器 query，可能已被用户改过）。
          // 不先覆写编辑器，由面板 executeQuery() 跑当前 query。
          const targetSql = peek.query;
          const confirmed = await classifyAndConfirm('sql', targetSql);
          if (!confirmed) return;
          const handled = { value: false };
          window.dispatchEvent(executeSqlEvent(backendSessionId, handled));
          await tick();
          if (!handled.value) {
            // 面板未认领执行（理论上 peek.found 即会认领），回退直接执行同一内容。
            const api = getBindings();
            if (typeof api.ExecuteDatabaseQuery !== 'function') {
              throw new Error('数据库执行不可用');
            }
            const sql = String(targetSql).trim().replace(/;+\s*$/g, '');
            await api.ExecuteDatabaseQuery(backendSessionId, sql);
          }
          copilotStore.appendMessage(sessionId, { role: 'assistant', content: '已执行，结果见数据库面板。' });
          return;
        }

        // 无打开的查询面板（或编辑器为空）：分类 artifact.content，确认后先填入再执行同一内容。
        const confirmed = await classifyAndConfirm('sql', artifact.content);
        if (!confirmed) return;
        window.dispatchEvent(applySqlEvent(backendSessionId, artifact.content));
        await tick();
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

      const confirmed = await classifyAndConfirm('shell', artifact.content);
      if (!confirmed) return;
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
      <h2>{assistantTitle}</h2>
      {#if workspaceLabel}
        <p class="ai-panel__context" title={workspaceLabel}>当前：{workspaceLabel}</p>
      {/if}
    </div>
    <div class="ai-panel__header-actions">
    <button type="button" class="ai-panel__new-chat" disabled={generating || !sessionId} on:click={startNewChat}>新对话</button>
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
    </div>
  </header>

  {#if !hasSession}
    <div class="ai-panel__empty" role="status">
      <strong>先连接主机或数据源</strong>
      <span>侧栏会跟随当前标签生成 SQL、查询 DSL 或 Shell。</span>
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
        <div class="ai-panel__hint">{emptyHint}</div>
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
        <div class="ai-panel__hint">正在生成… <button type="button" class="ai-panel__stop-inline" on:click={stopGeneration}>停止</button></div>
      {/if}
      {#if errorMessage}
        <div class="ai-panel__error">{errorMessage}</div>
      {/if}
    </div>

    <form class="ai-panel__composer" on:submit|preventDefault={sendMessage}>
      <textarea
        bind:value={draft}
        placeholder={composerPlaceholder}
        on:keydown={handleComposerKeydown}
      ></textarea>
      {#if generating}
        <button type="button" class="ai-panel__stop" on:click={stopGeneration}>停止</button>
      {:else}
        <button type="submit" disabled={!draft.trim()}>发送</button>
      {/if}
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

  .ai-panel__context {
    margin: 4px 0 0;
    max-width: 180px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 11px;
    color: var(--text-secondary);
  }
  .ai-panel__header-actions { display: flex; align-items: center; gap: 6px; }
  .ai-panel__new-chat { border: 1px solid var(--glass-border); border-radius: 6px; background: transparent; color: var(--text-secondary); padding: 4px 7px; font-size: 11px; cursor: pointer; }
  .ai-panel__new-chat:hover:not(:disabled) { color: var(--text-primary); background: var(--bg-secondary); }
  .ai-panel__new-chat:disabled { opacity: .5; cursor: not-allowed; }

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

  .ai-panel__stop {
    background: transparent !important;
    border-color: #fecaca !important;
    color: #b91c1c !important;
  }

  .ai-panel__stop-inline {
    margin-left: 8px;
    border: 1px solid #fecaca;
    border-radius: 6px;
    background: transparent;
    color: #b91c1c;
    padding: 2px 8px;
    font-size: 11px;
    cursor: pointer;
  }
</style>
