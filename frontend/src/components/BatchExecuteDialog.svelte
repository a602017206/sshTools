<script>
  import Dialog from './ui/Dialog.svelte';

  export let isOpen = false;
  /** targets: [{ sessionId, title, subtitle, color }] */
  export let targets = [];
  export let onClose = () => {};
  /** onExecute({ command, sessionIds, appendEnter }) => Promise<{ sent, failed }> */
  export let onExecute = async () => ({ sent: 0, failed: 0 });

  let command = '';
  let selected = new Set();
  let appendEnter = true;
  let sending = false;
  let resultText = '';
  let wasOpen = false;

  $: if (isOpen && !wasOpen) {
    // 每次打开默认全选当前可执行页签
    selected = new Set((targets || []).map((target) => target.sessionId));
    resultText = '';
    wasOpen = true;
  }
  $: if (!isOpen) wasOpen = false;

  function toggleTarget(sessionId) {
    const next = new Set(selected);
    if (next.has(sessionId)) next.delete(sessionId);
    else next.add(sessionId);
    selected = next;
  }

  function selectAll() {
    selected = new Set((targets || []).map((target) => target.sessionId));
  }

  function selectNone() {
    selected = new Set();
  }

  async function handleExecute() {
    if (sending || !command.trim() || selected.size === 0) return;
    sending = true;
    resultText = '';
    try {
      const result = await onExecute({ command, sessionIds: [...selected], appendEnter });
      resultText = result.failed
        ? `已发送到 ${result.sent} 个页签，${result.failed} 个失败，请检查对应页签连接状态。`
        : `已发送到 ${result.sent} 个页签。`;
    } catch (error) {
      resultText = `发送失败：${error?.message || error}`;
    } finally {
      sending = false;
    }
  }
</script>

<Dialog bind:isOpen onClose={onClose} title="批量执行命令" size="md">
  <div class="space-y-4">
    <div class="space-y-2">
      <label class="block text-xs font-medium" style="color: var(--text-secondary);" for="batch-command-input">命令（可多行，逐行发送）</label>
      <textarea
        id="batch-command-input"
        bind:value={command}
        rows="4"
        placeholder={'例如：\nsystemctl restart nginx\njournalctl -u nginx -n 20 --no-pager'}
        class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm font-mono"
      ></textarea>
      <label class="flex items-center gap-2 text-xs" style="color: var(--text-secondary);">
        <input type="checkbox" bind:checked={appendEnter} class="w-4 h-4" style="accent-color: var(--accent-primary);" />
        发送后自动回车执行（不勾选则只把命令打到各页签输入行，不执行）
      </label>
    </div>

    <div class="space-y-2">
      <div class="flex items-center justify-between">
        <span class="text-xs font-medium" style="color: var(--text-secondary);">发送到这些页签（已选 {selected.size} / {targets.length}）</span>
        <span class="flex gap-2">
          <button type="button" class="text-xs hover:underline" style="color: var(--accent-primary);" on:click={selectAll}>全选</button>
          <button type="button" class="text-xs hover:underline" style="color: var(--accent-primary);" on:click={selectNone}>清空</button>
        </span>
      </div>
      {#if targets.length}
        <div class="max-h-48 overflow-y-auto rounded-lg border border-slate-200 dark:border-slate-600 divide-y divide-slate-100 dark:divide-slate-700">
          {#each targets as target}
            <label class="flex items-center gap-2.5 px-3 py-2 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-700/60">
              <input
                type="checkbox"
                checked={selected.has(target.sessionId)}
                on:change={() => toggleTarget(target.sessionId)}
                class="w-4 h-4 shrink-0"
                style="accent-color: var(--accent-primary);"
              />
              <span class="w-2.5 h-2.5 rounded-sm shrink-0" style="background: {target.color};"></span>
              <span class="flex-1 min-w-0 leading-tight">
                <span class="block text-xs font-medium truncate">{target.title}</span>
                <span class="block text-[10px] truncate" style="color: var(--text-tertiary);">{target.subtitle}</span>
              </span>
            </label>
          {/each}
        </div>
      {:else}
        <p class="text-xs" style="color: var(--text-tertiary);">当前没有已连接的 SSH 页签。</p>
      {/if}
    </div>

    {#if resultText}
      <p class="text-xs" style="color: var(--text-secondary);">{resultText}</p>
    {/if}

    <div class="flex justify-end gap-2 pt-1">
      <button type="button" on:click={onClose} class="ops-btn-glass px-3 py-1.5 rounded-lg text-xs font-medium">关闭</button>
      <button
        type="button"
        on:click={handleExecute}
        disabled={sending || !command.trim() || selected.size === 0}
        class="px-3 py-1.5 rounded-lg text-xs font-medium text-white accent-bg accent-bg-hover disabled:opacity-50"
      >{sending ? '发送中…' : `发送到 ${selected.size} 个页签`}</button>
    </div>
  </div>
</Dialog>
