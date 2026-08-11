<script>
  import { createEventDispatcher } from 'svelte';
  import { APP_MODES } from '../lib/workspaceTabs.js';

  export let activeMode = 'ssh';

  const dispatch = createEventDispatcher();

  function selectMode(id) {
    dispatch('select', id);
  }
</script>

<nav class="mode-switch" aria-label="主模式">
  <div class="mode-pill" role="tablist">
    {#each APP_MODES as mode}
      <button
        type="button"
        role="tab"
        class:active={activeMode === mode.id}
        aria-selected={activeMode === mode.id}
        on:click={() => selectMode(mode.id)}
      >
        {mode.label}
      </button>
    {/each}
  </div>
</nav>

<style>
  .mode-switch {
    display: flex;
    justify-content: center;
  }

  .mode-pill {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 3px;
    border-radius: 999px;
    border: 1px solid var(--glass-border);
    background: var(--glass-bg);
    box-shadow: var(--shadow-glass);
    backdrop-filter: blur(var(--glass-blur)) saturate(var(--glass-saturate));
    -webkit-backdrop-filter: blur(var(--glass-blur)) saturate(var(--glass-saturate));
  }

  button {
    appearance: none;
    border: 0;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    font: inherit;
    font-size: 13px;
    font-weight: 600;
    min-height: 36px;
    min-width: 96px;
    padding: 0 16px;
    border-radius: 999px;
    transition: background-color var(--trans-fast), color var(--trans-fast), box-shadow var(--trans-fast), transform var(--trans-fast);
  }

  button:hover {
    color: var(--text-primary);
  }

  button.active {
    color: var(--text-primary);
    background: #ffffff;
    box-shadow: var(--shadow-sm), inset 0 0 0 1px var(--glass-border);
  }

  :global(.dark) button.active {
    color: var(--text-primary);
    background: color-mix(in srgb, var(--glass-bg-strong) 92%, transparent);
  }

  button:active {
    transform: scale(0.98);
  }

  button:focus-visible {
    outline: 2px solid var(--focus-ring);
    outline-offset: 2px;
  }

  @media (prefers-reduced-motion: reduce) {
    button {
      transition: none;
    }
  }
</style>
