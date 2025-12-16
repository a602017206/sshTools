<script>
  import { createEventDispatcher, onMount } from 'svelte';

  export let visible = false;
  export let title = '';
  export let label = '';
  export let value = '';
  export let placeholder = '';

  const dispatch = createEventDispatcher();
  let inputElement;

  onMount(() => {
    if (visible && inputElement) {
      inputElement.focus();
      inputElement.select();
    }
  });

  $: if (visible && inputElement) {
    setTimeout(() => {
      inputElement.focus();
      inputElement.select();
    }, 100);
  }

  function handleSubmit() {
    if (value.trim()) {
      dispatch('confirm', value.trim());
      close();
    }
  }

  function handleCancel() {
    dispatch('cancel');
    close();
  }

  function close() {
    visible = false;
    value = '';
  }

  function handleKeyDown(event) {
    if (event.key === 'Enter') {
      event.preventDefault();
      handleSubmit();
    } else if (event.key === 'Escape') {
      event.preventDefault();
      handleCancel();
    }
  }

  function handleOverlayClick(event) {
    if (event.target === event.currentTarget) {
      handleCancel();
    }
  }
</script>

{#if visible}
  <div class="modal-overlay" on:click={handleOverlayClick} role="dialog" aria-modal="true">
    <div class="modal-content">
      <div class="modal-header">
        <h3>{title}</h3>
        <button class="btn-close" on:click={handleCancel}>✕</button>
      </div>

      <div class="modal-body">
        {#if label}
          <label for="input-field">{label}</label>
        {/if}
        <input
          id="input-field"
          type="text"
          bind:value
          bind:this={inputElement}
          on:keydown={handleKeyDown}
          placeholder={placeholder}
        />
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" on:click={handleCancel}>取消</button>
        <button class="btn-primary" on:click={handleSubmit} disabled={!value.trim()}>
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
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    -webkit-app-region: no-drag;
  }

  .modal-content {
    background: var(--bg-secondary);
    border: 1px solid var(--border-primary);
    border-radius: 8px;
    width: 90%;
    max-width: 400px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-primary);
  }

  .modal-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .btn-close {
    padding: 4px 8px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    font-size: 18px;
    line-height: 1;
  }

  .btn-close:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
    border-radius: 4px;
  }

  .modal-body {
    padding: 20px;
  }

  .modal-body label {
    display: block;
    margin-bottom: 8px;
    font-size: 13px;
    color: var(--text-primary);
  }

  .modal-body input {
    width: 100%;
    padding: 8px 12px;
    background: var(--bg-input);
    border: 1px solid var(--border-primary);
    border-radius: 4px;
    color: var(--text-primary);
    font-size: 13px;
    font-family: inherit;
    outline: none;
    transition: border-color 0.2s;
  }

  .modal-body input:focus {
    border-color: var(--accent-primary);
  }

  .modal-body input::placeholder {
    color: var(--text-secondary);
    opacity: 0.6;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding: 16px 20px;
    border-top: 1px solid var(--border-primary);
  }

  .btn-primary,
  .btn-secondary {
    padding: 8px 16px;
    border: none;
    border-radius: 4px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: var(--accent-primary);
    color: white;
  }

  .btn-primary:hover:not(:disabled) {
    background: #0d5a99;
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: var(--bg-tertiary);
    color: var(--text-primary);
  }

  .btn-secondary:hover {
    background: var(--bg-hover);
  }
</style>
