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
    background-color: #2a2a2a;
    border: 1px solid #3c3c3c;
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
    color: #cccccc;
  }

  .message {
    margin: 0 0 16px 0;
    font-size: 13px;
    color: #999999;
  }

  .input-group {
    margin-bottom: 16px;
  }

  input[type="text"],
  input[type="password"] {
    width: 100%;
    padding: 10px 12px;
    background-color: #1e1e1e;
    border: 1px solid #555555;
    border-radius: 4px;
    color: #cccccc;
    font-size: 14px;
    box-sizing: border-box;
    transition: border-color 0.2s;
  }

  input[type="text"]:focus,
  input[type="password"]:focus {
    outline: none;
    border-color: #0e639c;
  }

  .save-option {
    margin-bottom: 20px;
  }

  .save-option label {
    display: flex;
    align-items: center;
    cursor: pointer;
    font-size: 13px;
    color: #cccccc;
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
    background-color: #0e639c;
    color: white;
  }

  .btn-primary:hover {
    background-color: #1177bb;
  }

  .btn-secondary {
    background-color: #3c3c3c;
    color: #cccccc;
  }

  .btn-secondary:hover {
    background-color: #505050;
  }
</style>
