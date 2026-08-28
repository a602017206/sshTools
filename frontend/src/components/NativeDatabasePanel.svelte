<script>
  import RedisWorkspace from './workspaces/RedisWorkspace.svelte';
  import ElasticsearchWorkspace from './workspaces/ElasticsearchWorkspace.svelte';
  import KafkaWorkspace from './workspaces/KafkaWorkspace.svelte';
  import GenericNativeWorkspace from './workspaces/GenericNativeWorkspace.svelte';
  import { resolveNativeWorkspaceKind } from '../lib/nativeDatabaseWorkspace.js';

  export let sessionId = null;
  export let dbConfig = null;

  $: kind = resolveNativeWorkspaceKind(dbConfig?.metadata?.db_type || '');
</script>

{#if kind === 'redis'}
  <RedisWorkspace {sessionId} {dbConfig} />
{:else if kind === 'elasticsearch'}
  <ElasticsearchWorkspace {sessionId} {dbConfig} />
{:else if kind === 'kafka'}
  <KafkaWorkspace {sessionId} {dbConfig} />
{:else}
  <GenericNativeWorkspace {sessionId} {dbConfig} />
{/if}
