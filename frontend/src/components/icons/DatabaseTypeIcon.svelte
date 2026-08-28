<script>
  import { databaseTypeLabel, resolveDatabaseTypeIconKind } from '../../lib/databaseTypeIcon.js';

  export let databaseType = '';
  export let size = 16;

  const uid = `db-${Math.random().toString(36).slice(2, 8)}`;
  $: icon = resolveDatabaseTypeIconKind(databaseType);
  $: title = databaseTypeLabel(databaseType);
</script>

<span class="database-type-icon" style={`width:${size}px;height:${size}px;`} {title}>
  {#if icon === 'mysql'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <ellipse cx="12" cy="7" rx="7.5" ry="2.6" fill="#00758f" />
      <path fill="#00758f" d="M4.5 7v8.5c0 1.4 3.4 2.6 7.5 2.6s7.5-1.2 7.5-2.6V7c0 1.4-3.4 2.6-7.5 2.6S4.5 8.4 4.5 7z" />
      <ellipse cx="12" cy="15.5" rx="7.5" ry="2.6" fill="#00546b" />
      <path fill="#f29111" d="M8.2 9.2c1.1-.4 2.4-.6 3.8-.6 1.4 0 2.7.2 3.8.6-.3 1.2-2 2.1-3.8 2.1s-3.5-.9-3.8-2.1z" />
    </svg>
  {:else if icon === 'postgresql'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <ellipse cx="12" cy="7" rx="7.5" ry="2.6" fill="#336791" />
      <path fill="#336791" d="M4.5 7v8.5c0 1.4 3.4 2.6 7.5 2.6s7.5-1.2 7.5-2.6V7c0 1.4-3.4 2.6-7.5 2.6S4.5 8.4 4.5 7z" />
      <ellipse cx="12" cy="15.5" rx="7.5" ry="2.6" fill="#254e70" />
      <path fill="#fff" opacity=".85" d="M10.2 9.4c.6-.8 1.5-1.2 2.8-1.2 1.1 0 2 .3 2.6.9-.5.7-1.5 1.1-2.6 1.1-1.1 0-2.1-.4-2.8-1.1z" />
    </svg>
  {:else if icon === 'sqlite'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#0f4c75" d="M6 4.5h9.2L18 7.3V19.5c0 .8-.7 1.5-1.5 1.5H6c-.8 0-1.5-.7-1.5-1.5V6c0-.8.7-1.5 1.5-1.5z" />
      <path fill="#3282b8" d="M15.2 4.5V7c0 .8.7 1.5 1.5 1.5H18z" />
      <ellipse cx="11.5" cy="12" rx="4.8" ry="1.6" fill="#bbe1fa" />
      <path fill="#bbe1fa" d="M6.7 12v3.8c0 .9 2.1 1.6 4.8 1.6s4.8-.7 4.8-1.6V12c0 .9-2.1 1.6-4.8 1.6S6.7 12.9 6.7 12z" />
    </svg>
  {:else if icon === 'oracle'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <ellipse cx="12" cy="7" rx="7.5" ry="2.6" fill="#f80000" />
      <path fill="#f80000" d="M4.5 7v8.5c0 1.4 3.4 2.6 7.5 2.6s7.5-1.2 7.5-2.6V7c0 1.4-3.4 2.6-7.5 2.6S4.5 8.4 4.5 7z" />
      <ellipse cx="12" cy="15.5" rx="7.5" ry="2.6" fill="#c40000" />
      <path fill="#fff" opacity=".9" d="M8.8 10.2h6.4c.8 0 1.4.6 1.4 1.4v.2c0 .8-.6 1.4-1.4 1.4H8.8c-.8 0-1.4-.6-1.4-1.4v-.2c0-.8.6-1.4 1.4-1.4z" />
    </svg>
  {:else if icon === 'sqlserver'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <ellipse cx="12" cy="7" rx="7.5" ry="2.6" fill="#cc2927" />
      <path fill="#cc2927" d="M4.5 7v8.5c0 1.4 3.4 2.6 7.5 2.6s7.5-1.2 7.5-2.6V7c0 1.4-3.4 2.6-7.5 2.6S4.5 8.4 4.5 7z" />
      <ellipse cx="12" cy="15.5" rx="7.5" ry="2.6" fill="#9e1f1d" />
      <path fill="#fff" opacity=".9" d="M9 10h6v1.2H9V10zm0 2.4h4.2v1.2H9v-1.2z" />
    </svg>
  {:else if icon === 'dm'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <ellipse cx="12" cy="7" rx="7.5" ry="2.6" fill="#007a33" />
      <path fill="#007a33" d="M4.5 7v8.5c0 1.4 3.4 2.6 7.5 2.6s7.5-1.2 7.5-2.6V7c0 1.4-3.4 2.6-7.5 2.6S4.5 8.4 4.5 7z" />
      <ellipse cx="12" cy="15.5" rx="7.5" ry="2.6" fill="#005824" />
      <path fill="#fff" opacity=".9" d="M9.2 9.8h1.8v4.4H9.2V9.8zm3.8 0H15v4.4h-2v-4.4z" />
    </svg>
  {:else if icon === 'kingbase'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <ellipse cx="12" cy="7" rx="7.5" ry="2.6" fill="#005bac" />
      <path fill="#005bac" d="M4.5 7v8.5c0 1.4 3.4 2.6 7.5 2.6s7.5-1.2 7.5-2.6V7c0 1.4-3.4 2.6-7.5 2.6S4.5 8.4 4.5 7z" />
      <ellipse cx="12" cy="15.5" rx="7.5" ry="2.6" fill="#004483" />
      <path fill="#ffd200" d="M9.5 10.2h1.4l1.1 2.2 1.1-2.2h1.4l-2 3.8v1.2h-1.4v-1.2l-2-3.8z" />
    </svg>
  {:else if icon === 'opengauss'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <ellipse cx="12" cy="7" rx="7.5" ry="2.6" fill="#009639" />
      <path fill="#009639" d="M4.5 7v8.5c0 1.4 3.4 2.6 7.5 2.6s7.5-1.2 7.5-2.6V7c0 1.4-3.4 2.6-7.5 2.6S4.5 8.4 4.5 7z" />
      <ellipse cx="12" cy="15.5" rx="7.5" ry="2.6" fill="#006b28" />
      <circle cx="12" cy="11.8" r="2.2" fill="#fff" opacity=".9" />
    </svg>
  {:else if icon === 'redis'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="4" y="5" width="16" height="14" rx="2.2" fill="#dc382d" />
      <path fill="#fff" opacity=".95" d="M8.2 8.4h7.6v1.4H8.2V8.4zm0 3h7.6v1.4H8.2v-1.4zm0 3h5.2v1.4H8.2v-1.4z" />
      <circle cx="17.2" cy="7.2" r="1.1" fill="#fff" opacity=".85" />
    </svg>
  {:else if icon === 'mongodb'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#47a248" d="M12 3.5c-2.8 2.1-4.6 5.4-4.6 9.2 0 2.2.8 4.2 2.1 5.8.6.8 1.4 1.5 2.5 1.5s1.9-.7 2.5-1.5c1.3-1.6 2.1-3.6 2.1-5.8 0-3.8-1.8-7.1-4.6-9.2z" />
      <path fill="#599e4e" d="M12 4.8c-.8 1.8-1.2 3.7-1.2 5.9 0 1.8.4 3.4 1.1 4.7.4.7.9 1.1 1.5 1.1V4.8z" opacity=".75" />
    </svg>
  {:else if icon === 'elasticsearch'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="4" y="5" width="16" height="14" rx="2.2" fill="#343741" />
      <circle cx="11" cy="12" r="3.6" fill="none" stroke="#fec514" stroke-width="1.8" />
      <path stroke="#fec514" stroke-width="1.8" stroke-linecap="round" d="M13.8 14.8 16.8 17.8" />
      <path fill="#fec514" d="M8.2 8.4h7.6v1.2H8.2V8.4z" opacity=".85" />
    </svg>
  {:else if icon === 'memcached'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="5" y="6" width="14" height="12" rx="1.8" fill="#008080" />
      <rect x="7" y="8.5" width="10" height="2" rx=".6" fill="#fff" opacity=".9" />
      <rect x="7" y="12" width="10" height="2" rx=".6" fill="#fff" opacity=".75" />
      <rect x="7" y="15.5" width="6.5" height="1.6" rx=".5" fill="#fff" opacity=".6" />
    </svg>
  {:else if icon === 'cassandra'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="7.2" fill="none" stroke="#1287b1" stroke-width="1.6" />
      <circle cx="12" cy="12" r="4.4" fill="none" stroke="#1287b1" stroke-width="1.4" />
      <circle cx="12" cy="12" r="1.8" fill="#1287b1" />
    </svg>
  {:else if icon === 'couchbase'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="4" y="5" width="16" height="14" rx="2.2" fill="#ea2328" />
      <path fill="#fff" d="M8.5 8.8c1.8-1.2 4.2-1.2 6 0-1.8 1.2-4.2 1.2-6 0zm0 3.2c1.8-1.2 4.2-1.2 6 0-1.8 1.2-4.2 1.2-6 0zm0 3.2c1.8-1.2 4.2-1.2 6 0-1.8 1.2-4.2 1.2-6 0z" opacity=".95" />
    </svg>
  {:else if icon === 'influxdb'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#22adf6" d="M12 4.5 18.8 8.2v7.6L12 19.5 5.2 15.8V8.2L12 4.5z" />
      <path fill="#fff" opacity=".9" d="M9.2 12.2c.8-1.6 2.4-2.6 4.3-2.6 1.2 0 2.3.4 3.1 1.1-.9.8-2.1 1.3-3.4 1.3-1.4 0-2.7-.5-3.6-1.3l-.4 1.5z" />
    </svg>
  {:else if icon === 'neo4j'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="8" cy="9" r="2.2" fill="#018bff" />
      <circle cx="16" cy="9" r="2.2" fill="#018bff" />
      <circle cx="12" cy="16" r="2.2" fill="#60a917" />
      <path stroke="#018bff" stroke-width="1.5" d="M9.8 9.8 10.8 14M14.2 9.8 13.2 14M10 9h4" />
    </svg>
  {:else if icon === 'kafka'}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <rect x="4" y="5" width="16" height="14" rx="2.2" fill="#231f20" />
      <path fill="#fff" d="M7.5 8.2h2.2v7.6H7.5V8.2zm3.4 1.8h2.2v5.8h-2.2V10zm3.4-1.8h2.2v7.6h-2.2V8.2z" opacity=".95" />
      <path fill="#ff6600" d="M7.5 8.2h9v1.4h-9V8.2z" />
    </svg>
  {:else}
    <svg viewBox="0 0 24 24" aria-hidden="true">
      <ellipse cx="12" cy="7" rx="7.5" ry="2.6" fill="#2563eb" />
      <path fill="#2563eb" d="M4.5 7v8.5c0 1.4 3.4 2.6 7.5 2.6s7.5-1.2 7.5-2.6V7c0 1.4-3.4 2.6-7.5 2.6S4.5 8.4 4.5 7z" />
      <ellipse cx="12" cy="15.5" rx="7.5" ry="2.6" fill="#1d4ed8" />
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
