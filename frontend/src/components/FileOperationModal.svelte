<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { fade, scale } from 'svelte/transition';
  import Icon from './Icon.svelte';

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
  <div class="modal-overlay" on:click={handleOverlayClick} role="dialog" aria-modal="true" transition:fade={{ duration: 150 }}>
    <div class="modal-content" transition:scale={{ start: 0.95, duration: 150 }}>
      <div class="modal-header">
        <div class="title-container">
          <Icon name="edit" size={18} color="var(--accent-primary)" />
          <h3>{title}</h3>
        </div>
        <button class="btn-close" on:click={handleCancel}>
          <Icon name="close" size={16} />
        </button>
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
          autocomplete="off"
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
    backdrop-filter: blur(2px);
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
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-primary);
  }

  .title-container {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .modal-header h3 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .btn-close {
    padding: 4px;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: all 0.2s;
  }

  .btn-close:hover {
    color: var(--text-primary);
    background: var(--bg-hover);
  }

  .modal-body {
    padding: 24px 20px;
  }

  .modal-body label {
    display: block;
    margin-bottom: 8px;
    font-size: 13px;
    color: var(--text-secondary);
    font-weight: 500;
  }

  .modal-body input {
    width: 100%;
    padding: 10px 12px;
    background: var(--bg-input);
    border: 1px solid var(--border-primary);
    border-radius: 6px;
    color: var(--text-primary);
    font-size: 14px;
    font-family: inherit;
    outline: none;
    transition: all 0.2s;
  }

  .modal-body input:focus {
    border-color: var(--accent-primary);
    box-shadow: 0 0 0 2px rgba(14, 99, 156, 0.2);
  }

  .modal-body input::placeholder {
    color: var(--text-secondary);
    opacity: 0.6;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 10px;
    padding: 16px 20px;
    border-top: 1px solid var(--border-primary);
    background: var(--bg-tertiary);
    border-bottom-left-radius: 8px;
    border-bottom-right-radius: 8px;
  }

  .btn-primary,
  .btn-secondary {
    padding: 8px 20px;
    border: 1px solid transparent;
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
    background: var(--accent-hover);
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: transparent;
    border-color: var(--border-primary);
    color: var(--text-primary);
  }

  .btn-secondary:hover {
    background: var(--bg-hover);
    border-color: var(--text-secondary);
  }
</style>
