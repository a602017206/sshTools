<script>
  import Dialog from './ui/Dialog.svelte';
  import { ACCENT_PRESETS, FONT_PRESETS, TERMINAL_FONT_PRESETS, getDefaultAppSettings } from '../settings/appearance.js';

  export let isOpen = false;
  export let value = getDefaultAppSettings();
  export let onSave = () => {};
  export let onCancel = () => {};
  export let onPreview = () => {};

  let draft = getDefaultAppSettings();
  let initializedForOpen = false;

  $: if (isOpen && !initializedForOpen) {
    draft = normalizeDraft({ ...getDefaultAppSettings(), ...value });
    initializedForOpen = true;
  }

  $: if (!isOpen && initializedForOpen) {
    initializedForOpen = false;
  }

  function normalizeDraft(settings) {
    return {
      ...settings,
      font_size: Number(settings.font_size) || 14,
      terminal_font_size: Number(settings.terminal_font_size) || 14
    };
  }

  function handleSave() {
    onSave(normalizeDraft(draft));
  }

  function getEffectiveMode() {
    if (draft.theme_mode === 'dark' || draft.theme_mode === 'light') {
      return draft.theme_mode;
    }
    if (typeof document !== 'undefined' && document.documentElement.classList.contains('dark')) {
      return 'dark';
    }
    return 'light';
  }

  function getPresetMainColor(id) {
    const preset = ACCENT_PRESETS[id] || ACCENT_PRESETS.teal;
    const darkTheme = getEffectiveMode() === 'dark';
    return darkTheme ? preset.dark.accentPrimary : preset.light.accentPrimary;
  }

  function getPresetHoverColor(id) {
    const preset = ACCENT_PRESETS[id] || ACCENT_PRESETS.teal;
    const darkTheme = getEffectiveMode() === 'dark';
    return darkTheme ? preset.dark.accentHover : preset.light.accentHover;
  }

  function triggerPreview() {
    queueMicrotask(() => {
      onPreview(normalizeDraft(draft));
    });
  }

  function handleReset() {
    draft = getDefaultAppSettings();
    triggerPreview();
  }
</script>

<Dialog bind:isOpen={isOpen} onClose={onCancel} title="全局设置" size="xl">
  <div class="space-y-6">
    <div class="rounded-xl border border-slate-200 dark:border-slate-700 p-4 bg-slate-50/70 dark:bg-slate-900/50">
      <div class="text-sm font-semibold text-slate-900 dark:text-slate-100 mb-3">实时预览</div>
      <div class="rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 p-4">
        <div class="flex items-center justify-between">
          <div>
            <div class="text-base font-semibold text-slate-900 dark:text-slate-100">AHa SSH Manager</div>
            <div class="text-xs text-slate-500 dark:text-slate-400">主题色、字体和字号实时生效</div>
          </div>
          <button
            type="button"
            class="px-3 py-2 rounded-lg text-white text-xs font-medium shadow-sm"
            style={`background: linear-gradient(135deg, ${getPresetMainColor(draft.accent_color)}, ${getPresetHoverColor(draft.accent_color)});`}
          >
            预览按钮
          </button>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-2 gap-5">
      <div class="space-y-2">
        <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">主题模式</div>
        <div class="grid grid-cols-3 gap-2">
          <button type="button" class="px-3 py-2 rounded-lg text-xs font-medium transition-colors {draft.theme_mode === 'light' ? 'text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200'}" style={draft.theme_mode === 'light' ? 'background: linear-gradient(90deg, var(--accent-primary), var(--accent-hover));' : ''} on:click={() => { draft.theme_mode = 'light'; triggerPreview(); }}>浅色</button>
          <button type="button" class="px-3 py-2 rounded-lg text-xs font-medium transition-colors {draft.theme_mode === 'dark' ? 'text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200'}" style={draft.theme_mode === 'dark' ? 'background: linear-gradient(90deg, var(--accent-primary), var(--accent-hover));' : ''} on:click={() => { draft.theme_mode = 'dark'; triggerPreview(); }}>深色</button>
          <button type="button" class="px-3 py-2 rounded-lg text-xs font-medium transition-colors {draft.theme_mode === 'system' ? 'text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200'}" style={draft.theme_mode === 'system' ? 'background: linear-gradient(90deg, var(--accent-primary), var(--accent-hover));' : ''} on:click={() => { draft.theme_mode = 'system'; triggerPreview(); }}>跟随系统</button>
        </div>
      </div>

      <div class="space-y-2">
        <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">主题色</div>
        <div class="grid grid-cols-4 gap-2">
          {#each Object.entries(ACCENT_PRESETS) as [id, preset]}
            <button
              type="button"
              class="px-2 py-2 rounded-lg border text-xs transition-colors {draft.accent_color === id ? 'text-white shadow-md' : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300'}"
              style={draft.accent_color === id ? `background: linear-gradient(90deg, ${getPresetMainColor(id)}, ${getPresetHoverColor(id)}); border-color: ${getPresetMainColor(id)};` : `border-color: ${getPresetMainColor(id)}66;`}
              on:click={() => { draft.accent_color = id; triggerPreview(); }}
            >
              <span class="inline-flex items-center gap-1.5">
                <span class="w-2.5 h-2.5 rounded-full border border-white/30" style={`background: ${getPresetMainColor(id)};`}></span>
                {preset.label}
              </span>
            </button>
          {/each}
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <label class="space-y-2 block">
        <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">界面字体</div>
        <select bind:value={draft.font_family} on:change={triggerPreview} class="w-full px-3 py-2 rounded-lg bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm">
          {#each FONT_PRESETS as font}
            <option value={font.value}>{font.label}</option>
          {/each}
        </select>
      </label>

      <label class="space-y-2 block">
        <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">终端字体</div>
        <select bind:value={draft.terminal_font_family} on:change={triggerPreview} class="w-full px-3 py-2 rounded-lg bg-slate-50 dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm">
          {#each TERMINAL_FONT_PRESETS as font}
            <option value={font.value}>{font.label}</option>
          {/each}
        </select>
      </label>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-2">
        <div class="flex items-center justify-between text-sm font-semibold text-slate-900 dark:text-slate-100">
          <span>界面字号</span>
          <span style="color: var(--accent-primary);">{draft.font_size}px</span>
        </div>
        <input type="range" min="12" max="18" step="1" bind:value={draft.font_size} on:input={triggerPreview} class="w-full" style="accent-color: var(--accent-primary);" />
      </div>

      <div class="space-y-2">
        <div class="flex items-center justify-between text-sm font-semibold text-slate-900 dark:text-slate-100">
          <span>终端字号</span>
          <span style="color: var(--accent-primary);">{draft.terminal_font_size}px</span>
        </div>
        <input type="range" min="12" max="20" step="1" bind:value={draft.terminal_font_size} on:input={triggerPreview} class="w-full" style="accent-color: var(--accent-primary);" />
      </div>
    </div>

    <div class="rounded-lg border border-slate-200 dark:border-slate-700 p-3 space-y-3">
      <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">扩展设置</div>
      <label class="flex items-center justify-between text-sm text-slate-700 dark:text-slate-300">
        <span>紧凑模式（预留）</span>
        <input type="checkbox" bind:checked={draft.compact_mode} on:change={triggerPreview} class="w-4 h-4" style="accent-color: var(--accent-primary);" />
      </label>
      <label class="flex items-center justify-between text-sm text-slate-700 dark:text-slate-300">
        <span>减少动画（预留）</span>
        <input type="checkbox" bind:checked={draft.reduced_motion} on:change={triggerPreview} class="w-4 h-4" style="accent-color: var(--accent-primary);" />
      </label>
    </div>

    <div class="flex items-center justify-between pt-2">
      <button type="button" on:click={handleReset} class="px-3 py-2 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200">恢复默认</button>
      <div class="flex gap-2">
        <button type="button" on:click={onCancel} class="px-3 py-2 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200">取消</button>
        <button type="button" on:click={handleSave} class="px-3 py-2 text-xs rounded-lg text-white" style="background: linear-gradient(90deg, var(--accent-primary), var(--accent-hover));">保存设置</button>
      </div>
    </div>
  </div>
</Dialog>
