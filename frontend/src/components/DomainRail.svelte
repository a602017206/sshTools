<script>
  import { ASSET_DOMAINS } from '../lib/assetDomain.js';

  export let activeDomain = 'all';
  export let onSelect = () => {};
  export let disabled = false;
</script>

<nav class="domain-rail" aria-label="资产域切换">
  {#each ASSET_DOMAINS as domain}
    <button
      type="button"
      class="domain-rail__item"
      class:active={activeDomain === domain.id}
      class:is-docker={domain.id === 'docker'}
      title={domain.label}
      aria-label={domain.label}
      aria-pressed={activeDomain === domain.id}
      {disabled}
      on:click={() => onSelect(domain.id)}
    >
      {#if domain.id === 'all'}
        <svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M4 5h6v6H4V5zm10 0h6v6h-6V5zM4 13h6v6H4v-6zm10 0h6v6h-6v-6z"/></svg>
      {:else if domain.id === 'ssh'}
        <svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M4 6h16a2 2 0 012 2v7a2 2 0 01-2 2h-6l-2 3-2-3H4a2 2 0 01-2-2V8a2 2 0 012-2zm2 3v2h2V9H6zm4 0v2h8V9h-8z"/></svg>
      {:else if domain.id === 'database'}
        <svg viewBox="0 0 24 24" aria-hidden="true"><ellipse cx="12" cy="6" rx="7" ry="2.4" fill="currentColor"/><path fill="currentColor" d="M5 6v9.2c0 1.3 3.1 2.4 7 2.4s7-1.1 7-2.4V6c0 1.3-3.1 2.4-7 2.4S5 7.3 5 6z"/></svg>
      {:else if domain.id === 'cache'}
        <svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M12 3l8 4.2v4.1c0 4.7-3.2 9-8 10.2C7.2 20.3 4 16 4 11.3V7.2L12 3zm0 2.3L6.5 8v3.3c0 3.5 2.3 6.6 5.5 7.5 3.2-.9 5.5-4 5.5-7.5V8L12 5.3z"/></svg>
      {:else if domain.id === 'search'}
        <svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M10.5 4a6.5 6.5 0 014.96 10.7l3.92 3.92-1.41 1.41-3.92-3.92A6.5 6.5 0 1110.5 4zm0 2a4.5 4.5 0 100 9 4.5 4.5 0 000-9z"/></svg>
      {:else if domain.id === 'mq'}
        <svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M4 6h4v12H4V6zm6 3h4v9h-4V9zm6-5h4v14h-4V4z"/></svg>
      {:else}
        <svg viewBox="0 0 24 24" aria-hidden="true"><path fill="currentColor" d="M4 8h4v3H4V8zm5 0h4v3H9V8zm5 0h4v3h-4V8zM6.5 12h3v3h-3v-3zm4 0h3v3h-3v-3zM4 18h16v2H4v-2z"/></svg>
      {/if}
      <span class="domain-rail__label">{domain.shortLabel}</span>
    </button>
  {/each}
</nav>

<style>
  .domain-rail {
    width: 52px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 8px 6px;
    border-right: 1px solid var(--border-primary, #d9e0e4);
    background: color-mix(in srgb, var(--bg-secondary, #f4f6f5) 88%, transparent);
    overflow-y: auto;
  }

  .domain-rail__item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 2px;
    width: 100%;
    min-height: 44px;
    border: 0;
    border-radius: 8px;
    background: transparent;
    color: var(--text-secondary, #6d7783);
    cursor: pointer;
    padding: 6px 2px;
  }

  .domain-rail__item svg {
    width: 18px;
    height: 18px;
    display: block;
  }

  .domain-rail__label {
    font-size: 10px;
    line-height: 1.1;
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  .domain-rail__item:hover:not(:disabled) {
    color: var(--text-primary, #1d2935);
    background: var(--bg-hover, #eff6f5);
  }

  .domain-rail__item.active {
    color: #0e6674;
    background: #eff6f5;
    box-shadow: inset 0 0 0 1px #bdd1d4;
  }

  .domain-rail__item.is-docker:not(.active) {
    opacity: 0.72;
  }

  .domain-rail__item:disabled {
    cursor: not-allowed;
    opacity: 0.45;
  }
</style>
