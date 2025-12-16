<script>
  import { createEventDispatcher } from 'svelte';
  import { themeStore } from '../stores/theme.js';

  export let visible = false;

  const dispatch = createEventDispatcher();
  let currentTheme = 'dark';

  themeStore.subscribe(state => {
    currentTheme = state.theme;
  });

  async function handleThemeChange(newTheme) {
    try {
      await themeStore.setTheme(newTheme);
    } catch (error) {
      console.error('Failed to change theme:', error);
    }
  }

  function handleClose() {
    visible = false;
    dispatch('close');
  }

  function handleKeydown(event) {
    if (event.key === 'Escape') {
      handleClose();
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if visible}
  <div class="modal-overlay" on:click={handleClose}>
    <div class="modal-content" on:click|stopPropagation>
      <div class="modal-header">
        <h3>设置</h3>
        <button class="close-btn" on:click={handleClose}>✕</button>
      </div>

      <div class="settings-body">
        <section>
          <h4>外观</h4>
          <div class="setting-item">
            <label>主题</label>
            <div class="theme-selector">
              <label class="radio-label">
                <input
                  type="radio"
                  name="theme"
                  value="light"
                  checked={currentTheme === 'light'}
                  on:change={() => handleThemeChange('light')}
                />
                <span>浅色</span>
              </label>
              <label class="radio-label">
                <input
                  type="radio"
                  name="theme"
                  value="dark"
                  checked={currentTheme === 'dark'}
                  on:change={() => handleThemeChange('dark')}
                />
                <span>深色</span>
              </label>
            </div>
          </div>
        </section>

        <section>
          <h4>终端</h4>
          <div class="setting-item">
            <label>字体大小</label>
            <select disabled>
              <option>14</option>
            </select>
            <p class="hint">即将推出</p>
          </div>
        </section>
      </div>

      <div class="modal-footer">
        <button class="btn-primary" on:click={handleClose}>关闭</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    -webkit-app-region: no-drag;
  }

  .modal-content {
    background-color: var(--bg-tertiary);
    border: 1px solid var(--border-primary);
    border-radius: 8px;
    width: 500px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-primary);
  }

  h3 {
    margin: 0;
    font-size: 16px;
    color: var(--text-primary);
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    font-size: 20px;
    cursor: pointer;
    padding: 4px 8px;
    line-height: 1;
  }

  .close-btn:hover {
    color: var(--text-primary);
  }

  .settings-body {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  section {
    margin-bottom: 24px;
  }

  section:last-child {
    margin-bottom: 0;
  }

  h4 {
    margin: 0 0 12px 0;
    font-size: 14px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .setting-item {
    margin-bottom: 16px;
  }

  .setting-item:last-child {
    margin-bottom: 0;
  }

  .setting-item > label {
    display: block;
    margin-bottom: 8px;
    font-size: 13px;
    color: var(--text-primary);
  }

  .theme-selector {
    display: flex;
    gap: 16px;
  }

  .radio-label {
    display: flex;
    align-items: center;
    cursor: pointer;
    font-size: 13px;
    color: var(--text-primary);
  }

  .radio-label input {
    margin-right: 6px;
    cursor: pointer;
  }

  select {
    padding: 6px 8px;
    background-color: var(--bg-input);
    border: 1px solid var(--border-primary);
    color: var(--text-primary);
    border-radius: 4px;
    font-size: 13px;
  }

  .hint {
    margin: 4px 0 0 0;
    font-size: 11px;
    color: var(--text-secondary);
    font-style: italic;
  }

  .modal-footer {
    padding: 16px 24px;
    border-top: 1px solid var(--border-primary);
    display: flex;
    justify-content: flex-end;
  }

  .btn-primary {
    padding: 8px 20px;
    background-color: var(--accent-primary);
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
  }

  .btn-primary:hover {
    background-color: var(--accent-hover);
  }
</style>
