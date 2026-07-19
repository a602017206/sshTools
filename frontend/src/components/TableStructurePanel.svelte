<script>
  import { formatColumnDescription, formatColumnLength, formatColumnType } from '../lib/tableStructureMetadata.js';

  export let sessionId = null;
  export let dbConfig = null;
  export let databaseName = '';
  export let schemaName = '';
  export let tableName = '';

  let ddlData = null;
  let schemaData = null;
  let isLoading = false;
  let errorMessage = '';
  let copied = false;
  let loadedRequest = '';

  $: titleName = schemaName && tableName ? `${schemaName}.${tableName}` : (tableName || '表结构');
  $: dbTypeLabel = dbConfig?.metadata?.db_type ? dbConfig.metadata.db_type.toUpperCase() : '';
  $: requestKey = `${sessionId || ''}:${databaseName || ''}:${schemaName || ''}:${tableName || ''}`;
  $: if (sessionId && tableName && requestKey !== loadedRequest) {
    loadedRequest = requestKey;
    loadDDL();
  }

  async function loadDDL() {
    if (!sessionId || !tableName) return;
    if (!window.wailsBindings) return;

    isLoading = true;
    errorMessage = '';
    schemaData = null;

    try {
      const [ddl, schema] = await Promise.all([
        schemaName
          ? window.wailsBindings.GetTableDDLInSchema(sessionId, databaseName, schemaName, tableName)
          : window.wailsBindings.GetTableDDL(sessionId, databaseName, tableName),
        window.wailsBindings.GetTableSchemaInSchema(sessionId, databaseName, schemaName, tableName)
      ]);
      ddlData = ddl;
      schemaData = schema;
    } catch (error) {
      console.error('Failed to load table DDL:', error);
      const detail = error?.message || String(error || '未知错误');
      errorMessage = `加载表结构失败: ${detail}`;
    } finally {
      isLoading = false;
    }
  }

  async function copyDDL() {
    if (!ddlData?.ddl) return;

    try {
      await navigator.clipboard.writeText(ddlData.ddl);
      copied = true;
      setTimeout(() => {
        copied = false;
      }, 2000);
    } catch (error) {
      console.error('Failed to copy:', error);
    }
  }
</script>

<div class="h-full flex flex-col bg-white dark:bg-gray-800">
  <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between gap-3">
    <div class="min-w-0">
      <div class="text-sm font-semibold text-gray-900 dark:text-white truncate">{titleName}</div>
      <div class="text-xs text-gray-500 dark:text-gray-400">{dbTypeLabel ? `${dbTypeLabel} · ` : ''}表结构定义</div>
    </div>
    <div class="flex items-center gap-2">
      <button
        class="px-3 py-1.5 text-xs rounded bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600"
        on:click={loadDDL}
        disabled={isLoading}
      >
        刷新
      </button>
      <button
        class="px-3 py-1.5 text-xs rounded {copied ? 'bg-green-500 text-white' : 'bg-blue-600 text-white'} hover:brightness-95 disabled:opacity-50"
        on:click={copyDDL}
        disabled={!ddlData?.ddl || isLoading}
      >
        {copied ? '已复制' : '复制 DDL'}
      </button>
    </div>
  </div>

  {#if errorMessage}
    <div class="px-4 py-2 text-xs text-red-600 bg-red-50 dark:bg-red-900/20 border-b border-red-100 dark:border-red-800">
      {errorMessage}
    </div>
  {/if}

  <div class="flex-1 overflow-auto p-4">
    {#if isLoading}
      <div class="flex items-center justify-center h-full">
        <div class="text-sm text-gray-500 dark:text-gray-400">加载中...</div>
      </div>
    {:else if schemaData?.columns?.length || ddlData?.ddl}
      {#if schemaData?.columns?.length}
        <section class="mb-5">
          <h2 class="mb-2 text-xs font-semibold text-gray-700 dark:text-gray-200">字段</h2>
          <div class="overflow-x-auto border border-gray-200 dark:border-gray-700">
            <table class="min-w-full text-left text-xs text-gray-700 dark:text-gray-200">
              <thead class="bg-gray-50 dark:bg-gray-900 text-gray-500 dark:text-gray-400">
                <tr>
                  <th class="px-3 py-2 font-medium">字段</th>
                  <th class="px-3 py-2 font-medium">数据类型</th>
                  <th class="px-3 py-2 font-medium">长度</th>
                  <th class="px-3 py-2 font-medium">描述</th>
                  <th class="px-3 py-2 font-medium">可空</th>
                  <th class="px-3 py-2 font-medium">默认值</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
                {#each schemaData.columns as column}
                  <tr>
                    <td class="px-3 py-2 font-mono whitespace-nowrap">{column.name}</td>
                    <td class="px-3 py-2 whitespace-nowrap">{formatColumnType(column)}</td>
                    <td class="px-3 py-2 whitespace-nowrap">{formatColumnLength(column)}</td>
                    <td class="px-3 py-2 min-w-[120px]">{formatColumnDescription(column)}</td>
                    <td class="px-3 py-2 whitespace-nowrap">{column.nullable ? '是' : '否'}</td>
                    <td class="px-3 py-2 font-mono whitespace-nowrap">{column.has_default ? column.default_value : '-'}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </section>
      {/if}
      {#if ddlData?.ddl}
        <section>
          <h2 class="mb-2 text-xs font-semibold text-gray-700 dark:text-gray-200">DDL</h2>
          <pre class="text-xs font-mono whitespace-pre-wrap break-all bg-gray-50 dark:bg-gray-900 p-4 rounded-lg border border-gray-200 dark:border-gray-700 text-gray-800 dark:text-gray-200 overflow-auto">{ddlData.ddl}</pre>
        </section>
      {/if}
    {:else}
      <div class="flex items-center justify-center h-full">
        <div class="text-sm text-gray-500 dark:text-gray-400">暂无表结构数据</div>
      </div>
    {/if}
  </div>
</div>
