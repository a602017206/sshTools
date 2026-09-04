<script>
  import Dialog from './ui/Dialog.svelte';
  import JDBCDriverManager from './JDBCDriverManager.svelte';
  import { ACCENT_PRESETS, FONT_PRESETS, TERMINAL_FONT_PRESETS, getDefaultAppSettings } from '../settings/appearance.js';
  import { TERMINAL_THEME_PRESETS } from '../lib/terminalTheme.js';

  export let isOpen = false;
  export let value = getDefaultAppSettings();
  export let onSave = () => {};
  export let onCancel = () => {};
  export let onPreview = () => {};

  let draft = getDefaultAppSettings();
  let initializedForOpen = false;
  let activeSection = 'appearance';
  let copilotApiKey = '';
  let hasCopilotAPIKey = false;
  let copilotKeyBusy = false;
  let copilotKeyError = '';

  $: if (isOpen && !initializedForOpen) {
    draft = normalizeDraft({ ...getDefaultAppSettings(), ...value });
    copilotApiKey = '';
    copilotKeyError = '';
    initializedForOpen = true;
    refreshHasCopilotAPIKey();
  }

  $: if (!isOpen && initializedForOpen) {
    initializedForOpen = false;
    copilotApiKey = '';
    copilotKeyError = '';
  }

  function normalizeDraft(settings) {
    const rest = { ...(settings || {}) };
    delete rest.copilot_api_key;
    return {
      ...rest,
      font_size: Number(rest.font_size) || 14,
      terminal_theme: rest.terminal_theme === 'light' || rest.terminal_theme === 'follow' ? rest.terminal_theme : 'dark',
      terminal_font_size: Number(rest.terminal_font_size) || 14,
      copilot_provider: rest.copilot_provider || 'openai_compatible',
      copilot_base_url: rest.copilot_base_url || '',
      copilot_model: rest.copilot_model || ''
      ,copilot_max_tool_rounds: rest.copilot_max_tool_rounds || 4, copilot_max_tool_result_chars: rest.copilot_max_tool_result_chars || 8000
    };
  }

  function handleSave() {
    const next = normalizeDraft(draft);
    const key = String(copilotApiKey || '').trim();
    if (key) {
      next.copilot_api_key = key;
    }
    onSave(next);
  }

  async function refreshHasCopilotAPIKey() {
    const api = window.wailsBindings || {};
    if (typeof api.HasCopilotAPIKey !== 'function') {
      hasCopilotAPIKey = false;
      return;
    }
    try {
      hasCopilotAPIKey = Boolean(await api.HasCopilotAPIKey());
    } catch (error) {
      console.error('Failed to check copilot API key:', error);
      hasCopilotAPIKey = false;
    }
  }

  async function handleClearCopilotAPIKey() {
    const api = window.wailsBindings || {};
    if (typeof api.ClearCopilotAPIKey !== 'function') {
      return;
    }
    copilotKeyBusy = true;
    copilotKeyError = '';
    try {
      await api.ClearCopilotAPIKey();
      copilotApiKey = '';
      await refreshHasCopilotAPIKey();
    } catch (error) {
      console.error('Failed to clear copilot API key:', error);
      copilotKeyError = '清除密钥失败';
    } finally {
      copilotKeyBusy = false;
    }
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
    const preset = ACCENT_PRESETS[id] || ACCENT_PRESETS.blue;
    const darkTheme = getEffectiveMode() === 'dark';
    return darkTheme ? preset.dark.accentPrimary : preset.light.accentPrimary;
  }

  function getPresetHoverColor(id) {
    const preset = ACCENT_PRESETS[id] || ACCENT_PRESETS.blue;
    const darkTheme = getEffectiveMode() === 'dark';
    return darkTheme ? preset.dark.accentHover : preset.light.accentHover;
  }

  function getPresetSecondaryColor(id) {
    const preset = ACCENT_PRESETS[id] || ACCENT_PRESETS.blue;
    const darkTheme = getEffectiveMode() === 'dark';
    const colors = darkTheme ? preset.dark : preset.light;
    return colors.accentSecondary || colors.accentHover;
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

  async function handleSelectBackground() {
    try {
      const api = window.wailsBindings || {};
      if (typeof api.SelectBackgroundImage !== 'function') {
        return;
      }
      const result = await api.SelectBackgroundImage();
      if (!result) {
        return;
      }
      draft = {
        ...draft,
        background_image_enabled: true,
        background_image_path: result.path || '',
        background_image_data_url: result.data_url || '',
        background_image_fit: result.fit || draft.background_image_fit || 'cover',
        background_image_opacity: result.opacity ?? draft.background_image_opacity ?? 35
      };
      triggerPreview();
    } catch (error) {
      console.error('Failed to select background image:', error);
    }
  }

  async function handleClearBackground() {
    // Draft-only clear; file + persisted settings are removed when the user clicks Save.
    draft = {
      ...draft,
      background_image_enabled: false,
      background_image_path: '',
      background_image_data_url: ''
    };
    triggerPreview();
  }
</script>

<Dialog bind:isOpen={isOpen} onClose={onCancel} title="全局设置" size="xl">
  <div class="settings-shell">
    <nav class="settings-nav" aria-label="全局设置分类">
      <button
        type="button"
        class:active={activeSection === 'appearance'}
        on:click={() => (activeSection = 'appearance')}
      >
        <span>外观</span>
        <small>主题、字体、字号</small>
      </button>
      <button
        type="button"
        class:active={activeSection === 'jdbc'}
        on:click={() => (activeSection = 'jdbc')}
      >
        <span>数据库驱动</span>
        <small>JRE、驱动、agent</small>
      </button>
      <button
        type="button"
        class:active={activeSection === 'copilot'}
        on:click={() => (activeSection = 'copilot')}
      >
        <span>AI Copilot</span>
        <small>接口、模型、密钥</small>
      </button>
    </nav>

    <div class="settings-content">
      {#if activeSection === 'appearance'}
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
            style={`background: linear-gradient(135deg, ${getPresetMainColor(draft.accent_color)}, ${getPresetSecondaryColor(draft.accent_color)});`}
          >
            预览按钮
          </button>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 xl:grid-cols-2 gap-5">
      <div class="space-y-2">
        <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">整体外观</div>
        <div class="grid grid-cols-3 gap-2">
          <button type="button" class="px-3 py-2 rounded-lg text-xs font-medium transition-colors {draft.theme_mode === 'light' ? 'text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200'}" style={draft.theme_mode === 'light' ? 'background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));' : ''} on:click={() => { draft.theme_mode = 'light'; triggerPreview(); }}>浅色</button>
          <button type="button" class="px-3 py-2 rounded-lg text-xs font-medium transition-colors {draft.theme_mode === 'dark' ? 'text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200'}" style={draft.theme_mode === 'dark' ? 'background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));' : ''} on:click={() => { draft.theme_mode = 'dark'; triggerPreview(); }}>深色</button>
          <button type="button" class="px-3 py-2 rounded-lg text-xs font-medium transition-colors {draft.theme_mode === 'system' ? 'text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200'}" style={draft.theme_mode === 'system' ? 'background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));' : ''} on:click={() => { draft.theme_mode = 'system'; triggerPreview(); }}>跟随系统</button>
        </div>
      </div>

      <div class="space-y-2">
        <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">终端主题</div>
        <div class="grid grid-cols-3 gap-2">
          {#each TERMINAL_THEME_PRESETS as option}
            <button
              type="button"
              class="px-3 py-2 rounded-lg text-xs font-medium transition-colors {draft.terminal_theme === option.id ? 'text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200'}"
              style={draft.terminal_theme === option.id ? 'background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));' : ''}
              on:click={() => { draft.terminal_theme = option.id; triggerPreview(); }}
            >
              {option.label}
            </button>
          {/each}
        </div>
        <p class="text-[11px] leading-5 text-slate-500 dark:text-slate-400">与整体外观分开保存。跟随界面时，终端颜色随浅色/深色外观一起切换。</p>
      </div>
    </div>

    <div class="space-y-2">
        <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">主题色</div>
        <div class="grid grid-cols-5 gap-2">
          {#each Object.entries(ACCENT_PRESETS) as [id, preset]}
            <button
              type="button"
              class="px-2 py-2 rounded-lg border text-xs transition-colors {draft.accent_color === id ? 'text-white shadow-md' : 'border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300'}"
              style={draft.accent_color === id ? `background: linear-gradient(90deg, ${getPresetMainColor(id)}, ${getPresetSecondaryColor(id)}); border-color: ${getPresetMainColor(id)};` : `border-color: ${getPresetMainColor(id)}66;`}
              on:click={() => { draft.accent_color = id; triggerPreview(); }}
            >
              <span class="inline-flex items-center gap-1.5">
                <span class="w-2.5 h-2.5 rounded-full border border-white/30" style={`background: linear-gradient(135deg, ${getPresetMainColor(id)}, ${getPresetSecondaryColor(id)});`}></span>
                {preset.label}
              </span>
            </button>
          {/each}
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
      <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">背景图片</div>
      <p class="text-xs text-slate-500 dark:text-slate-400">自定义工作台背景；面板仍保持浅色可读，图片仅作为底层氛围。</p>
      <div class="flex flex-wrap items-center gap-2">
        <button
          type="button"
          class="px-3 py-2 text-xs rounded-lg text-white"
          style="background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));"
          on:click={handleSelectBackground}
        >
          选择图片
        </button>
        <button
          type="button"
          class="px-3 py-2 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200"
          disabled={!draft.background_image_enabled && !draft.background_image_data_url}
          on:click={handleClearBackground}
        >
          清除背景
        </button>
        <label class="flex items-center gap-2 text-xs text-slate-600 dark:text-slate-300 ml-auto">
          <input
            type="checkbox"
            bind:checked={draft.background_image_enabled}
            disabled={!draft.background_image_path && !draft.background_image_data_url}
            on:change={triggerPreview}
            class="w-4 h-4"
            style="accent-color: var(--accent-primary);"
          />
          启用背景图
        </label>
      </div>
      {#if draft.background_image_data_url}
        <div
          class="h-24 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-100 dark:bg-slate-800 overflow-hidden"
          style={`background-image: url("${draft.background_image_data_url}"); background-size: ${draft.background_image_fit === 'contain' ? 'contain' : 'cover'}; background-position: center; background-repeat: no-repeat;`}
        ></div>
      {/if}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
        <div class="space-y-2">
          <div class="text-xs font-medium text-slate-700 dark:text-slate-300">填充方式</div>
          <div class="grid grid-cols-2 gap-2">
            <button type="button" class="px-2 py-2 rounded-lg text-xs {draft.background_image_fit !== 'contain' ? 'text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200'}" style={draft.background_image_fit !== 'contain' ? 'background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));' : ''} on:click={() => { draft.background_image_fit = 'cover'; triggerPreview(); }}>铺满</button>
            <button type="button" class="px-2 py-2 rounded-lg text-xs {draft.background_image_fit === 'contain' ? 'text-white' : 'bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200'}" style={draft.background_image_fit === 'contain' ? 'background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));' : ''} on:click={() => { draft.background_image_fit = 'contain'; triggerPreview(); }}>完整显示</button>
          </div>
        </div>
        <div class="space-y-2">
          <div class="flex items-center justify-between text-xs font-medium text-slate-700 dark:text-slate-300">
            <span>背景强度</span>
            <span style="color: var(--accent-primary);">{draft.background_image_opacity}%</span>
          </div>
          <input type="range" min="5" max="80" step="1" bind:value={draft.background_image_opacity} on:input={triggerPreview} class="w-full" style="accent-color: var(--accent-primary);" />
        </div>
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
        <button type="button" on:click={handleSave} class="px-3 py-2 text-xs rounded-lg text-white" style="background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));">保存设置</button>
      </div>
    </div>
  </div>
      {:else if activeSection === 'copilot'}
        <div class="space-y-6">
          <div class="rounded-xl border border-slate-200 dark:border-slate-700 p-4 bg-slate-50/70 dark:bg-slate-900/50 space-y-4">
            <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">AI Copilot</div>

            <label class="space-y-2 block">
              <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">Base URL</div>
              <input
                type="text"
                bind:value={draft.copilot_base_url}
                placeholder="https://api.deepseek.com/v1"
                autocomplete="off"
                class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm"
              />
            </label>
            <div class="grid grid-cols-2 gap-3"><label class="space-y-2 block"><div class="text-sm font-semibold">最大工具轮次</div><input type="number" min="1" max="8" bind:value={draft.copilot_max_tool_rounds} class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border" /></label><label class="space-y-2 block"><div class="text-sm font-semibold">单次工具结果上限</div><input type="number" min="1000" max="20000" step="1000" bind:value={draft.copilot_max_tool_result_chars} class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border" /></label></div>

            <label class="space-y-2 block">
              <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">模型名称</div>
              <input
                type="text"
                bind:value={draft.copilot_model}
                placeholder="deepseek-chat"
                autocomplete="off"
                class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm"
              />
              <p class="text-xs text-slate-500 dark:text-slate-400">模型名称请按服务商官方文档填写，例如 deepseek-chat</p>
            </label>

            <label class="space-y-2 block">
              <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">API Key</div>
              <input
                type="password"
                bind:value={copilotApiKey}
                placeholder="留空则保留已保存的密钥"
                autocomplete="off"
                class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm"
              />
            </label>

            <div class="flex items-center justify-between gap-2">
              <span class="text-xs text-slate-500 dark:text-slate-400">
                {hasCopilotAPIKey ? '已保存密钥' : '尚未保存密钥'}
              </span>
              <button
                type="button"
                class="px-3 py-2 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200 disabled:opacity-50"
                disabled={!hasCopilotAPIKey || copilotKeyBusy}
                on:click={handleClearCopilotAPIKey}
              >
                清除密钥
              </button>
            </div>
            {#if copilotKeyError}
              <p class="text-xs text-red-500">{copilotKeyError}</p>
            {/if}
          </div>

          <div class="flex items-center justify-end pt-2">
            <div class="flex gap-2">
              <button type="button" on:click={onCancel} class="px-3 py-2 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200">取消</button>
              <button type="button" on:click={handleSave} class="px-3 py-2 text-xs rounded-lg text-white" style="background: linear-gradient(90deg, var(--accent-primary), var(--accent-secondary));">保存设置</button>
            </div>
          </div>
        </div>
      {:else}
        <JDBCDriverManager />
      {/if}
    </div>
  </div>
</Dialog>

<style>
  .settings-shell {
    display: grid;
    grid-template-columns: 180px minmax(0, 1fr);
    gap: 16px;
    min-height: 420px;
  }

  .settings-nav {
    border: 1px solid var(--glass-border);
    border-radius: 12px;
    background: var(--glass-bg);
    padding: 8px;
  }

  .settings-nav button {
    width: 100%;
    display: block;
    border: 1px solid transparent;
    border-radius: 7px;
    background: transparent;
    color: var(--text-secondary);
    padding: 10px;
    text-align: left;
  }

  .settings-nav button.active {
    border-color: var(--accent-primary);
    background: var(--accent-subtle);
    color: var(--text-primary);
  }

  .settings-nav span,
  .settings-nav small {
    display: block;
  }

  .settings-nav span {
    font-size: 13px;
    font-weight: 700;
  }

  .settings-nav small {
    color: var(--text-tertiary);
    font-size: 11px;
  }

  .settings-content {
    min-width: 0;
  }

  @media (max-width: 900px) {
    .settings-shell {
      grid-template-columns: 1fr;
    }

    .settings-nav {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 8px;
    }
  }
</style>
