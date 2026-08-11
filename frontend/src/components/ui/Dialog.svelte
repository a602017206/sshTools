<script>
  import { onDestroy } from 'svelte';

  export let isOpen = false;
  export let onClose = () => {};
  export let title = '';
  export let size = 'md'; // sm, md, lg, xl

  const sizeClasses = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-2xl',
    xl: 'max-w-5xl'
  };

  $: dialogWidthClass = sizeClasses[size] || sizeClasses.md;

  let portalTarget;
  let host;

  function handleEscapeKey(event) {
    if (event.key === 'Escape' && isOpen) {
      onClose();
    }
  }

  function mountPortal(node) {
    host = node;
    if (typeof document === 'undefined') return {};
    portalTarget = document.createElement('div');
    portalTarget.className = 'ops-dialog-portal';
    document.body.appendChild(portalTarget);
    portalTarget.appendChild(node);
    return {
      destroy() {
        if (portalTarget?.parentNode) {
          portalTarget.parentNode.removeChild(portalTarget);
        }
        portalTarget = null;
        host = null;
      }
    };
  }

  $: if (typeof document !== 'undefined') {
    if (isOpen) {
      document.addEventListener('keydown', handleEscapeKey);
      document.body.style.overflow = 'hidden';
    } else {
      document.removeEventListener('keydown', handleEscapeKey);
      document.body.style.overflow = '';
    }
  }

  onDestroy(() => {
    if (typeof document !== 'undefined') {
      document.removeEventListener('keydown', handleEscapeKey);
      document.body.style.overflow = '';
    }
  });
</script>

{#if isOpen}
  <div
    class="ops-dialog-root"
    use:mountPortal
    role="dialog"
    aria-modal="true"
    aria-label={title}
  >
    <button
      type="button"
      class="ops-overlay-scrim ops-dialog-scrim"
      aria-label="关闭对话框"
      on:click={onClose}
    ></button>

    <div class={`ops-modal-panel ops-dialog-panel ${dialogWidthClass}`}>
      <div class="ops-dialog-header">
        <h2>{title}</h2>
        <button
          type="button"
          on:click={onClose}
          class="ops-icon-button p-2 rounded-md transition-colors focus-visible:outline-none focus-visible:ring-2"
          aria-label="关闭"
        >
          <svg class="w-5 h-5" style="color: var(--text-secondary);" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div class="ops-dialog-body">
        <slot />
      </div>
    </div>
  </div>
{/if}

<style>
  :global(.ops-dialog-portal) {
    position: relative;
    z-index: 10000;
  }

  .ops-dialog-root {
    position: fixed;
    inset: 0;
    z-index: 10000;
    display: flex;
    align-items: flex-start;
    justify-content: center;
    padding: max(1.25rem, 5vh) 1rem;
    overflow-y: auto;
    overscroll-behavior: contain;
  }

  .ops-dialog-scrim {
    position: fixed;
    inset: 0;
    border: 0;
    padding: 0;
    cursor: default;
  }

  .ops-dialog-panel {
    position: relative;
    z-index: 1;
    width: 100%;
    max-height: min(92vh, 880px);
    overflow: hidden;
    display: flex;
    flex-direction: column;
    margin-top: auto;
    margin-bottom: auto;
    flex-shrink: 0;
  }

  .ops-dialog-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--glass-border);
    flex-shrink: 0;
  }

  .ops-dialog-header h2 {
    margin: 0;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .ops-dialog-body {
    flex: 1;
    overflow-y: auto;
    padding: 1.25rem;
  }
</style>
