<script>
  import { resolveFileIconKind } from '../../lib/fileType.js';

  export let file = null;
  export let kind = '';
  export let size = 18;

  const uid = `ft-${Math.random().toString(36).slice(2, 8)}`;
  $: icon = kind || resolveFileIconKind(file);
</script>

<span class="file-type-icon" style={`width:${size}px;height:${size}px;`}>
  {#if icon === 'parent'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="4" y1="6" x2="20" y2="20" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#fde68a" /><stop offset="1" stop-color="#d97706" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-g)`} d="M3.5 8.2c0-1.2 1-2.2 2.2-2.2h4.1l1.6 1.7h7c1.2 0 2.1 1 2.1 2.2v7.9c0 1.2-1 2.2-2.1 2.2H5.7c-1.2 0-2.2-1-2.2-2.2V8.2z" />
      <path fill="#fef3c7" opacity=".5" d="M5 9.2h14v1.4H5z" />
      <path fill="none" stroke="#78350f" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" d="M11.2 16 8.4 13.2 11.2 10.4" />
    </svg>
  {:else if icon === 'folder'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-tab" x1="4" y1="5" x2="12" y2="10" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#fcd34d" /><stop offset="1" stop-color="#f59e0b" />
        </linearGradient>
        <linearGradient id="{uid}-body" x1="3" y1="8" x2="21" y2="20" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#fbbf24" /><stop offset="1" stop-color="#d97706" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-tab)`} d="M3.4 8.4V7.1c0-.9.7-1.6 1.6-1.6h4.2l1.7 1.7H3.4z" />
      <path fill={`url(#${uid}-body)`} d="M3.4 8.2h17.2c.9 0 1.6.7 1.6 1.6v8.4c0 .9-.7 1.6-1.6 1.6H5c-.9 0-1.6-.7-1.6-1.6V8.2z" />
      <path fill="#fde68a" opacity=".38" d="M5.2 9.4h13.6v1.6H5.2z" />
    </svg>
  {:else if icon === 'archive'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="6" y1="3" x2="18" y2="21" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#fde68a" /><stop offset="1" stop-color="#b45309" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-g)`} d="M7 3.6h10c.8 0 1.4.6 1.4 1.4v14c0 .8-.6 1.4-1.4 1.4H7c-.8 0-1.4-.6-1.4-1.4v-14c0-.8.6-1.4 1.4-1.4z" />
      <path fill="#fff7ed" d="M10.2 3.6h3.6v17H10.2z" />
      <rect x="10.6" y="10.2" width="2.8" height="2.4" rx=".5" fill="#b45309" />
    </svg>
  {:else if icon === 'java'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="6" y1="3" x2="18" y2="21" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#fecaca" /><stop offset="1" stop-color="#b91c1c" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-g)`} d="M7.2 3.5h7.1L18.8 8v11.4c0 .8-.7 1.5-1.5 1.5H7.2c-.8 0-1.5-.7-1.5-1.5V5c0-.8.7-1.5 1.5-1.5z" />
      <path fill="#fee2e2" d="M14.3 3.5 18.8 8h-3.3c-.7 0-1.2-.5-1.2-1.2V3.5z" />
      <circle cx="12" cy="14.6" r="3.1" fill="#7f1d1d" />
      <path fill="#fecaca" d="M10.7 13.4h2.6v.7h-2.6zm.4 1.2h1.8v.7h-1.8z" />
    </svg>
  {:else if icon === 'json'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="6" y1="3" x2="18" y2="21" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#fde68a" /><stop offset="1" stop-color="#ca8a04" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-g)`} d="M7.2 3.5h7.1L18.8 8v11.4c0 .8-.7 1.5-1.5 1.5H7.2c-.8 0-1.5-.7-1.5-1.5V5c0-.8.7-1.5 1.5-1.5z" />
      <path fill="#fef9c3" d="M14.3 3.5 18.8 8h-3.3c-.7 0-1.2-.5-1.2-1.2V3.5z" />
      <path fill="none" stroke="#713f12" stroke-width="1.5" stroke-linecap="round" d="M9.4 12.4c-.8.2-1.3.8-1.3 1.6s.5 1.4 1.3 1.6m5.2-3.2c.8.2 1.3.8 1.3 1.6s-.5 1.4-1.3 1.6" />
    </svg>
  {:else if icon === 'image'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="4" y1="4" x2="20" y2="20" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#6ee7b7" /><stop offset="1" stop-color="#0f766e" />
        </linearGradient>
      </defs>
      <rect x="4.2" y="5.2" width="15.6" height="13.6" rx="2" fill={`url(#${uid}-g)`} />
      <circle cx="9" cy="10" r="1.6" fill="#ecfdf5" />
      <path fill="#d1fae5" d="M6.2 16.4 10 12.4l2.4 2.3 2.6-3.2 4.6 4.9H6.2z" />
    </svg>
  {:else if icon === 'log'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="6" y1="3" x2="18" y2="21" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#cbd5e1" /><stop offset="1" stop-color="#475569" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-g)`} d="M7.2 3.5h7.1L18.8 8v11.4c0 .8-.7 1.5-1.5 1.5H7.2c-.8 0-1.5-.7-1.5-1.5V5c0-.8.7-1.5 1.5-1.5z" />
      <path fill="#e2e8f0" d="M8.4 11h7.2v1.1H8.4zm0 2.4h7.2v1.1H8.4zm0 2.4h5.1v1.1H8.4z" />
    </svg>
  {:else if icon === 'config'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="6" y1="3" x2="18" y2="21" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#c4b5fd" /><stop offset="1" stop-color="#6d28d9" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-g)`} d="M7.2 3.5h7.1L18.8 8v11.4c0 .8-.7 1.5-1.5 1.5H7.2c-.8 0-1.5-.7-1.5-1.5V5c0-.8.7-1.5 1.5-1.5z" />
      <circle cx="12" cy="14.4" r="2.3" fill="none" stroke="#ede9fe" stroke-width="1.4" />
    </svg>
  {:else if icon === 'markdown'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="6" y1="3" x2="18" y2="21" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#93c5fd" /><stop offset="1" stop-color="#1d4ed8" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-g)`} d="M7.2 3.5h7.1L18.8 8v11.4c0 .8-.7 1.5-1.5 1.5H7.2c-.8 0-1.5-.7-1.5-1.5V5c0-.8.7-1.5 1.5-1.5z" />
      <path fill="#dbeafe" d="M8.4 12.2h1.5l1.1 2.3 1.1-2.3h1.5v4.2h-1.2v-2.6l-1.4 2.6h-.8l-1.4-2.6v2.6H8.4z" />
    </svg>
  {:else if icon === 'code'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="6" y1="3" x2="18" y2="21" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#5eead4" /><stop offset="1" stop-color="#0f766e" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-g)`} d="M7.2 3.5h7.1L18.8 8v11.4c0 .8-.7 1.5-1.5 1.5H7.2c-.8 0-1.5-.7-1.5-1.5V5c0-.8.7-1.5 1.5-1.5z" />
      <path fill="#ccfbf1" d="M14.3 3.5 18.8 8h-3.3c-.7 0-1.2-.5-1.2-1.2V3.5z" />
      <path fill="none" stroke="#ecfdf5" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round" d="M9.2 13.2 7.8 14.8 9.2 16.4m5.6 0 1.4-1.6-1.4-1.6" />
    </svg>
  {:else}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <defs>
        <linearGradient id="{uid}-g" x1="6" y1="3" x2="18" y2="21" gradientUnits="userSpaceOnUse">
          <stop offset="0" stop-color="#e2e8f0" /><stop offset="1" stop-color="#64748b" />
        </linearGradient>
      </defs>
      <path fill={`url(#${uid}-g)`} d="M7.2 3.5h7.1L18.8 8v11.4c0 .8-.7 1.5-1.5 1.5H7.2c-.8 0-1.5-.7-1.5-1.5V5c0-.8.7-1.5 1.5-1.5z" />
      <path fill="#f8fafc" d="M14.3 3.5 18.8 8h-3.3c-.7 0-1.2-.5-1.2-1.2V3.5z" />
      <path fill="#cbd5e1" d="M8.5 11.2h7v1.05h-7zm0 2.3h7v1.05h-7zm0 2.3h4.8v1.05H8.5z" />
    </svg>
  {/if}
</span>

<style>
  .file-type-icon {
    display: inline-flex;
    flex-shrink: 0;
    align-items: center;
    justify-content: center;
    filter: drop-shadow(0 1px 1px rgba(15, 23, 42, 0.18));
  }
  .file-type-icon svg {
    width: 100%;
    height: 100%;
    display: block;
  }
</style>
