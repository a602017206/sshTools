<script>
  import { createEventDispatcher } from 'svelte';

  export let visible = false;
  export let title = '输入密码';
  export let message = '请输入密码：';
  export let isPassword = true;
  export let defaultValue = '';
  export let showSaveOption = false;

  const dispatch = createEventDispatcher();

  let inputValue = defaultValue;
  let savePassword = false;

  function handleSubmit() {
    dispatch('submit', { value: inputValue, save: savePassword });
    close();
  }

  function handleCancel() {
    dispatch('cancel');
    close();
  }

  function close() {
    inputValue = '';
    savePassword = false;
    visible = false;
  }

  function handleKeydown(event) {
    if (event.key === 'Enter') {
      handleSubmit();
    } else if (event.key === 'Escape') {
      handleCancel();
    }
  }

  // Focus input when modal becomes visible
  $: if (visible) {
    setTimeout(() => {
      const input = document.getElementById('prompt-input');
      if (input) input.focus();
    }, 100);
  }
</script>

{#if visible}
  <div class="modal-overlay" on:click={handleCancel}>
    <div class="modal-content" on:click|stopPropagation>
      <h3>{title}</h3>
      <p class="message">{message}</p>

      <div class="input-group">
        {#if isPassword}
          <input
            id="prompt-input"
            type="password"
            bind:value={inputValue}
            on:keydown={handleKeydown}
            autocomplete="off"
          />
        {:else}
          <input
            id="prompt-input"
            type="text"
            bind:value={inputValue}
            on:keydown={handleKeydown}
            autocomplete="off"
          />
        {/if}
      </div>

      {#if showSaveOption}
        <div class="save-option">
          <label>
            <input type="checkbox" bind:checked={savePassword} />
            <span>保存密码（加密存储）</span>
          </label>
        </div>
      {/if}

      <div class="modal-actions">
        <button class="btn-secondary" on:click={handleCancel}>
          取消
        </button>
        <button class="btn-primary" on:click={handleSubmit}>
          确定
        </button>
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
    padding: 24px;
    min-width: 400px;
    max-width: 500px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
  }

  h3 {
    margin: 0 0 12px 0;
    font-size: 16px;
    font-weight: 500;
    color: var(--text-primary);
  }

  .message {
    margin: 0 0 16px 0;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .input-group {
    margin-bottom: 16px;
  }

  input[type="text"],
  input[type="password"] {
    width: 100%;
    padding: 10px 12px;
    background-color: var(--bg-input);
    border: 1px solid var(--border-primary);
    border-radius: 4px;
    color: var(--text-primary);
    font-size: 14px;
    box-sizing: border-box;
    transition: border-color 0.2s;
  }

  input[type="text"]:focus,
  input[type="password"]:focus {
    outline: none;
    border-color: var(--border-active);
  }

  .save-option {
    margin-bottom: 20px;
  }

  .save-option label {
    display: flex;
    align-items: center;
    cursor: pointer;
    font-size: 13px;
    color: var(--text-primary);
  }

  .save-option input[type="checkbox"] {
    margin-right: 8px;
    cursor: pointer;
  }

  .save-option span {
    user-select: none;
  }

  .modal-actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }

  .btn-primary,
  .btn-secondary {
    padding: 8px 20px;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
    transition: background-color 0.2s;
  }

  .btn-primary {
    background-color: var(--accent-primary);
    color: white;
  }

  .btn-primary:hover {
    background-color: var(--accent-hover);
  }

  .btn-secondary {
    background-color: var(--bg-input);
    color: var(--text-primary);
  }

  .btn-secondary:hover {
    background-color: var(--bg-hover);
  }
</style>
