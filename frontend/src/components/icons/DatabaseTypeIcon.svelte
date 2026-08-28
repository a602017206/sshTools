<script>
  import { databaseTypeLabel, resolveDatabaseTypeIconKind } from '../../lib/databaseTypeIcon.js';

  export let databaseType = '';
  export let size = 16;
  export let kind = '';

  $: icon = kind || resolveDatabaseTypeIconKind(databaseType);
  $: title = kind === 'ssh' ? 'SSH' : kind === 'docker' ? 'Docker' : databaseTypeLabel(databaseType);
</script>

<span class="database-type-icon" style={`width:${size}px;height:${size}px;`} {title}>
  {#if icon === 'ssh'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="3" y="5" width="18" height="12" rx="2" fill="#0f766e"/>
      <rect x="5" y="7" width="8" height="2" rx="1" fill="#99f6e4"/>
      <rect x="5" y="11" width="5" height="2" rx="1" fill="#5eead4"/>
      <circle cx="17" cy="11" r="1.5" fill="#ccfbf1"/>
      <path d="M9 17h6l1.5 2h-9L9 17z" fill="#0f766e"/>
    </svg>
  {:else if icon === 'docker'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#2496ed" d="M4 11h2.2v2.2H4V11zm2.8 0H9v2.2H6.8V11zm2.8 0h2.2v2.2H9.6V11zm2.8 0H15v2.2h-2.2V11zM6.8 8.2H9V10.4H6.8V8.2zm2.8 0h2.2V10.4H9.6V8.2zM4 14h12.5c2.8 0 4.7-1.2 5.5-3.4.5.1 1.6.1 2.2-1-.1.8-.5 2.1-1.5 3.1C21.2 14.2 19.4 15 17 15H4v-1z"/>
    </svg>
  {:else if icon === 'mysql'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#00758f" d="M4.8 15.2c1.4 1.5 3.4 2.4 7.2 2.4 3.9 0 5.8-1 7.2-2.5-.7 2.6-3.5 4.4-7.2 4.4-3.8 0-6.5-1.8-7.2-4.3z"/>
      <path fill="#f29111" d="M7.2 7.1c.4-.2 2.4-1 4.8-1 2.6 0 4.5.8 4.9 1-.3 1.6-2.4 2.7-4.9 2.7S7.5 8.7 7.2 7.1z"/>
      <path fill="#00758f" d="M6.2 9.8c.5 1.7 2.7 2.9 5.8 2.9s5.3-1.2 5.8-2.9c-.6 2.2-2.9 3.7-5.8 3.7s-5.2-1.5-5.8-3.7z"/>
      <path fill="#f29111" d="M16.8 5.4c.7.2 1.3.8 1.4 1.5.4-.8.3-1.8-.5-2.4-.7-.5-1.6-.4-2.2.1.5.1 1 .3 1.3.8z"/>
    </svg>
  {:else if icon === 'postgresql'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#336791" d="M12.2 3.5c-2.4 0-4.3 1.5-4.6 3.7-.8.3-1.4 1-1.5 1.9-.1.8.2 1.5.8 2-.2.6-.2 1.2 0 1.8.3.8 1 1.3 1.8 1.5.4 1.4 1.4 2.5 2.7 3v2.3h1.6v-2.2c1.5-.4 2.6-1.6 3.1-3.1.7-.2 1.3-.7 1.5-1.4.3-.7.1-1.4-.3-2 .5-.5.7-1.2.5-1.9-.2-.8-.8-1.4-1.6-1.6C15.8 4.7 14.2 3.5 12.2 3.5z"/>
      <circle cx="9.4" cy="9.2" r="1" fill="#fff"/>
      <circle cx="14.8" cy="9.2" r="1" fill="#fff"/>
      <path fill="#fff" d="M10.2 12.2c.5.5 1.2.8 1.9.8s1.4-.3 1.9-.8" stroke="#fff" stroke-width="1" fill="none"/>
    </svg>
  {:else if icon === 'sqlite'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#003b57" d="M6 3.8h9.4L18.8 7v13.2c0 .8-.6 1.4-1.4 1.4H6c-.8 0-1.4-.6-1.4-1.4V5.2c0-.8.6-1.4 1.4-1.4z"/>
      <path fill="#044a64" d="M15.4 3.8V6.6c0 .7.6 1.3 1.3 1.3h2.1"/>
      <path fill="#7cc4e0" d="M7.4 10.2h9.2v1.3H7.4zm0 2.6h9.2v1.3H7.4zm0 2.6h6.4v1.3H7.4z"/>
    </svg>
  {:else if icon === 'oracle'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="3.5" y="8" width="17" height="8" rx="4" fill="none" stroke="#f80000" stroke-width="2.4"/>
      <path fill="#f80000" d="M8.2 10.2h7.6c.7 0 1.2.5 1.2 1.2v1.2c0 .7-.5 1.2-1.2 1.2H8.2c-.7 0-1.2-.5-1.2-1.2v-1.2c0-.7.5-1.2 1.2-1.2z"/>
    </svg>
  {:else if icon === 'sqlserver'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#cc2927" d="M6 4h12l-1.2 3.2H7.2L6 4zm1.5 4.5h9l-.9 16.2c0 .7-.6 1.3-1.3 1.3H9.7c-.7 0-1.3-.6-1.3-1.3L7.5 8.5z"/>
      <path fill="#fff" d="M10 11h4v1.4h-4V11zm0 2.8h4v1.4h-4v-1.4z"/>
    </svg>
  {:else if icon === 'dm'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="8.2" fill="#007a33"/>
      <path fill="#fff" d="M8.2 8.5h2.2l1.6 4.2 1.6-4.2H16v7h-1.8v-4.1L12.7 15.5h-1.4L9.9 11.4V15.5H8.2v-7z"/>
    </svg>
  {:else if icon === 'kingbase'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="3.5" y="4" width="17" height="16" rx="3" fill="#005bac"/>
      <path fill="#ffd200" d="M8.2 7.5h2l1.8 3.2 1.8-3.2h2L13.2 13v3.5h-2.2V13L8.2 7.5z"/>
    </svg>
  {:else if icon === 'opengauss'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="8.2" fill="#e60012"/>
      <circle cx="12" cy="12" r="4.8" fill="#fff"/>
      <circle cx="12" cy="12" r="2.2" fill="#e60012"/>
    </svg>
  {:else if icon === 'redis'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#a41e11" d="M4.5 14.8 12 18.8l7.5-4V10L12 14l-7.5-4v4.8z"/>
      <path fill="#dc382d" d="M4.5 10.2 12 14.2l7.5-4V5.6L12 9.6l-7.5-4v4.6z"/>
      <path fill="#91261d" d="M12 9.6 7.2 7.1 12 4.6l4.8 2.5L12 9.6z"/>
      <path fill="#fff" opacity=".9" d="M10.2 7.8h3.6v1.1h-3.6z"/>
    </svg>
  {:else if icon === 'mongodb'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#10aa50" d="M12.1 2.8s-.2 1.4-.2 2.8c0 4.6-1.5 7.5-2.7 9.4-1 1.5-1.5 2.5-1.5 3.7 0 1.7 1.1 2.6 2.4 2.8l.4.9h1.1l.3-.8c1.5-.2 2.7-1.3 2.7-3 0-1.2-.5-2.2-1.5-3.7-1.3-1.9-2.8-4.7-2.8-9.3 0-1.4.1-2.5.1-2.8h-.3z"/>
      <path fill="#b8c4c2" d="M11.7 18.8h.8v2.2h-.8z"/>
    </svg>
  {:else if icon === 'elasticsearch'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="3.5" y="4" width="17" height="16" rx="3" fill="#343741"/>
      <rect x="6" y="7" width="12" height="2.2" rx="1.1" fill="#00bfb3"/>
      <rect x="6" y="11" width="9" height="2.2" rx="1.1" fill="#f04e98"/>
      <rect x="6" y="15" width="10.5" height="2.2" rx="1.1" fill="#fec514"/>
    </svg>
  {:else if icon === 'memcached'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="4" y="5" width="16" height="14" rx="2" fill="#008080"/>
      <path fill="#fff" d="M7 8h10v2H7V8zm0 3.5h10v2H7v-2zm0 3.5h7v2H7v-2z"/>
    </svg>
  {:else if icon === 'cassandra'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="8" fill="none" stroke="#1287b1" stroke-width="1.8"/>
      <circle cx="12" cy="12" r="4.6" fill="none" stroke="#1287b1" stroke-width="1.5"/>
      <circle cx="12" cy="12" r="1.6" fill="#1287b1"/>
      <circle cx="12" cy="5.2" r="1.1" fill="#1287b1"/>
      <circle cx="17.8" cy="15.4" r="1.1" fill="#1287b1"/>
      <circle cx="6.2" cy="15.4" r="1.1" fill="#1287b1"/>
    </svg>
  {:else if icon === 'couchbase'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="3.5" y="4.5" width="17" height="15" rx="3" fill="#ea2328"/>
      <path fill="#fff" d="M7.5 8.2c2.1-1.4 5.1-1.4 7.2 0-2.1 1.4-5.1 1.4-7.2 0zm0 3.4c2.1-1.4 5.1-1.4 7.2 0-2.1 1.4-5.1 1.4-7.2 0zm0 3.4c2.1-1.4 5.1-1.4 7.2 0-2.1 1.4-5.1 1.4-7.2 0z"/>
    </svg>
  {:else if icon === 'influxdb'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#22adf6" d="M12 3.5 19.2 7.5v9L12 20.5 4.8 16.5v-9L12 3.5z"/>
      <path fill="#fff" d="M8.8 12.2c1-2 3.1-3.2 5.5-3.1.9 1.8.4 4-1.2 5.3-1.8.4-3.6-.5-4.3-2.2z"/>
    </svg>
  {:else if icon === 'neo4j'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="7.5" cy="8.5" r="2.4" fill="#018bff"/>
      <circle cx="16.5" cy="8.5" r="2.4" fill="#018bff"/>
      <circle cx="12" cy="16.2" r="2.4" fill="#00cb5b"/>
      <path stroke="#018bff" stroke-width="1.6" d="M9.4 9.4 10.8 14.2M14.6 9.4 13.2 14.2M9.6 8.5h4.8"/>
    </svg>
  {:else if icon === 'kafka'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="3.5" y="4" width="17" height="16" rx="2.5" fill="#231f20"/>
      <circle cx="8" cy="9" r="1.6" fill="#fff"/>
      <circle cx="12" cy="12" r="1.8" fill="#fff"/>
      <circle cx="16" cy="9" r="1.6" fill="#fff"/>
      <circle cx="16" cy="15.5" r="1.6" fill="#fff"/>
      <circle cx="8" cy="15.5" r="1.6" fill="#fff"/>
      <path stroke="#ff6600" stroke-width="1.4" d="M9.4 9.7 10.6 11.2M13.4 11.2 14.6 9.7M13.4 12.8 14.6 14.6M9.4 14.6 10.6 12.8"/>
    </svg>
  {:else}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <ellipse cx="12" cy="7" rx="7.5" ry="2.6" fill="#64748b"/>
      <path fill="#64748b" d="M4.5 7v8.5c0 1.4 3.4 2.6 7.5 2.6s7.5-1.2 7.5-2.6V7c0 1.4-3.4 2.6-7.5 2.6S4.5 8.4 4.5 7z"/>
      <ellipse cx="12" cy="15.5" rx="7.5" ry="2.6" fill="#475569"/>
    </svg>
  {/if}
</span>

<style>
  .database-type-icon {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .database-type-icon :global(svg) {
    width: 100%;
    height: 100%;
    display: block;
  }
</style>
