<script>
  import Dialog from './ui/Dialog.svelte';

  export let isOpen = false;
  export let mode = 'file';
  export let name = '';
  export let kind = 'file';
  export let isDir = false;
  export let conflictCount = 0;
  export let remaining = 1;
  export let suggestedName = '';
  export let onChoose = () => {};
  export let onCancel = () => {};

  $: title = mode === 'folder'
    ? '文件夹已存在'
    : kind === 'type'
      ? '名称冲突'
      : '文件已存在';

  $: message = mode === 'folder'
    ? `远程已存在文件夹 “${name}”，其中有 ${conflictCount} 个同名文件。`
    : kind === 'type'
      ? (isDir
        ? `远程已存在同名文件 “${name}”，无法直接创建文件夹。`
        : `远程已存在同名文件夹 “${name}”，无法直接上传文件。`)
      : `远程已存在文件 “${name}”。可覆盖原文件，或重命名为 “${suggestedName || `${name} copy`}”。`;

  $: showApplyAll = mode === 'file' && remaining > 1;
</script>

<Dialog
  bind:isOpen={isOpen}
  onClose={onCancel}
  title={title}
  size="sm"
>
  <div class="space-y-4">
    <p class="text-sm leading-relaxed" style="color: var(--text-secondary);">
      {message}
    </p>

    <div class="flex flex-wrap gap-2 pt-1">
      <button
        type="button"
        class="ops-btn-glass px-3 py-1.5 rounded-lg text-xs font-medium transition-all"
        on:click={() => onCancel()}
      >
        取消
      </button>
      <button
        type="button"
        class="ops-btn-glass px-3 py-1.5 rounded-lg text-xs font-medium transition-all"
        on:click={() => onChoose('rename')}
      >
        重命名
      </button>
      {#if mode === 'folder'}
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm accent-bg accent-bg-hover text-white"
          on:click={() => onChoose('oneByOne')}
        >
          逐个选择
        </button>
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm bg-amber-500 hover:bg-amber-600 text-white"
          on:click={() => onChoose('overwriteAll')}
        >
          全部覆盖
        </button>
      {:else}
        <button
          type="button"
          class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm bg-amber-500 hover:bg-amber-600 text-white"
          on:click={() => onChoose('overwrite')}
        >
          覆盖
        </button>
        {#if showApplyAll}
          <button
            type="button"
            class="ops-btn-glass px-3 py-1.5 rounded-lg text-xs font-medium transition-all"
            on:click={() => onChoose('renameAll')}
          >
            全部重命名
          </button>
          <button
            type="button"
            class="px-3 py-1.5 rounded-lg text-xs font-medium transition-all shadow-sm bg-amber-500 hover:bg-amber-600 text-white"
            on:click={() => onChoose('overwriteAll')}
          >
            全部覆盖
          </button>
        {/if}
      {/if}
    </div>
  </div>
</Dialog>
