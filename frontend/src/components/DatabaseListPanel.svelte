<script>
  import { onMount } from 'svelte';

  export let sessionId = null;
  export let dbConfig = null;

  let databases = [];
  let tableMetaByDatabase = {};
  let selectedDatabase = '';
  let databaseSearchQuery = '';
  let tableSearchQuery = '';
  let sortKey = 'tableName';
  let sortDirection = 'desc';
  const sortableColumns = [
    { key: 'tableName', label: '表名' },
    { key: 'rowCount', label: '行数' },
    { key: 'dataLength', label: '数据长度' },
    { key: 'engine', label: '引擎' },
    { key: 'createTime', label: '创建时间' },
    { key: 'updateTime', label: '修改时间' },
    { key: 'collation', label: '排序规则' },
    { key: 'comment', label: '注释' }
  ];
  let loadingDatabases = false;
  let loadingTables = false;
  let errorMessage = '';

  onMount(async () => {
    await loadDatabases();
  });

  async function loadDatabases() {
    if (!window.wailsBindings || !sessionId) return;

    loadingDatabases = true;
    errorMessage = '';

    try {
      const list = await window.wailsBindings.ListDatabases(sessionId);
      databases = (list || []).slice().sort();

      if (!selectedDatabase && databases.length > 0) {
        selectedDatabase = dbConfig?.metadata?.database || databases[0];
      }

      if (selectedDatabase) {
        await loadTableMetadata(selectedDatabase);
      }
    } catch (error) {
      errorMessage = `加载数据库失败: ${error.message || '未知错误'}`;
    } finally {
      loadingDatabases = false;
    }
  }

  function escapeSqlLiteral(value) {
    return String(value || '').replace(/'/g, "''");
  }

  function buildTableMetadataQuery(databaseName) {
    const dbType = (dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toLowerCase();
    const escapedDatabase = escapeSqlLiteral(databaseName);

    if (dbType === 'postgresql') {
      return `
        SELECT
          c.relname AS table_name,
          c.reltuples::bigint AS table_rows,
          pg_total_relation_size(c.oid) AS data_length,
          '' AS engine,
          NULL::text AS create_time,
          NULL::text AS update_time,
          pg_catalog.obj_description(c.oid, 'pg_class') AS table_comment,
          '' AS table_collation
        FROM pg_catalog.pg_class c
        JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
        WHERE c.relkind = 'r'
          AND n.nspname = 'public'
        ORDER BY c.relname;
      `;
    }

    return `
      SELECT
        TABLE_NAME AS table_name,
        TABLE_ROWS AS table_rows,
        DATA_LENGTH AS data_length,
        ENGINE AS engine,
        CREATE_TIME AS create_time,
        UPDATE_TIME AS update_time,
        TABLE_COLLATION AS table_collation,
        TABLE_COMMENT AS table_comment
      FROM information_schema.tables
      WHERE TABLE_SCHEMA = '${escapedDatabase}'
      ORDER BY TABLE_NAME;
    `;
  }

  function normalizeTableMetadata(columns, row) {
    const record = {};
    columns.forEach((col, index) => {
      record[String(col || '').toLowerCase()] = row[index];
    });

    return {
      tableName: String(record.table_name || ''),
      rowCount: record.table_rows,
      dataLength: record.data_length,
      engine: record.engine || '-',
      createTime: record.create_time || '-',
      updateTime: record.update_time || '-',
      collation: record.table_collation || '-',
      comment: record.table_comment || '-'
    };
  }

  function formatNumber(value) {
    const num = Number(value);
    if (Number.isNaN(num)) return '-';
    return num.toLocaleString('en-US');
  }

  function formatBytes(value) {
    const num = Number(value);
    if (Number.isNaN(num) || num < 0) return '-';
    if (num < 1024) return `${num} B`;
    if (num < 1024 * 1024) return `${(num / 1024).toFixed(1)} KB`;
    if (num < 1024 * 1024 * 1024) return `${(num / (1024 * 1024)).toFixed(1)} MB`;
    return `${(num / (1024 * 1024 * 1024)).toFixed(1)} GB`;
  }

  function formatDate(value) {
    if (!value || value === '-') return '-';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return String(value);
    return date.toLocaleString('zh-CN');
  }

  function toNumberOrNull(value) {
    const num = Number(value);
    if (Number.isNaN(num)) return null;
    return num;
  }

  function toTimestampOrNull(value) {
    if (!value || value === '-') return null;
    const ts = Date.parse(value);
    if (Number.isNaN(ts)) return null;
    return ts;
  }

  function getSortValue(meta, key) {
    if (key === 'tableName') return String(meta.tableName || '').toLowerCase();
    if (key === 'rowCount') return toNumberOrNull(meta.rowCount);
    if (key === 'dataLength') return toNumberOrNull(meta.dataLength);
    if (key === 'createTime') return toTimestampOrNull(meta.createTime);
    if (key === 'updateTime') return toTimestampOrNull(meta.updateTime);
    if (key === 'engine') return String(meta.engine || '').toLowerCase();
    if (key === 'collation') return String(meta.collation || '').toLowerCase();
    if (key === 'comment') return String(meta.comment || '').toLowerCase();
    return null;
  }

  function handleSort(nextKey) {
    if (sortKey === nextKey) {
      sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
      return;
    }
    sortKey = nextKey;
    sortDirection = 'desc';
  }

  async function loadTableMetadata(databaseName) {
    if (!window.wailsBindings || !sessionId || !databaseName) return;

    loadingTables = true;
    errorMessage = '';

    try {
      const query = buildTableMetadataQuery(databaseName);
      const response = await window.wailsBindings.ExecuteDatabaseQuery(sessionId, query);
      const parsed = JSON.parse(response || '{}');
      const columns = parsed?.columns || [];
      const rows = parsed?.rows || [];
      const metaList = rows
        .map(row => normalizeTableMetadata(columns, row))
        .filter(item => item.tableName);

      tableMetaByDatabase = {
        ...tableMetaByDatabase,
        [databaseName]: metaList
      };
    } catch (error) {
      errorMessage = `加载表失败: ${error.message || '未知错误'}`;
    } finally {
      loadingTables = false;
    }
  }

  async function handleDatabaseClick(databaseName) {
    selectedDatabase = databaseName;
    tableSearchQuery = '';
    if (!tableMetaByDatabase[databaseName]) {
      await loadTableMetadata(databaseName);
    }
  }

  function handleTableOpen(tableName) {
    window.dispatchEvent(new CustomEvent('database:table-select', {
      detail: {
        sessionId,
        databaseName: selectedDatabase,
        tableName
      }
    }));
  }

  $: selectedTableMeta = selectedDatabase ? (tableMetaByDatabase[selectedDatabase] || []) : [];
  $: normalizedDatabaseSearchQuery = databaseSearchQuery.trim().toLowerCase();
  $: filteredDatabases = !normalizedDatabaseSearchQuery
    ? databases
    : databases.filter(dbName => String(dbName || '').toLowerCase().includes(normalizedDatabaseSearchQuery));

  $: normalizedSearchQuery = tableSearchQuery.trim().toLowerCase();
  $: filteredTableMeta = !normalizedSearchQuery
    ? selectedTableMeta
    : selectedTableMeta.filter(meta => {
      const name = String(meta.tableName || '').toLowerCase();
      const comment = String(meta.comment || '').toLowerCase();
      return name.includes(normalizedSearchQuery) || comment.includes(normalizedSearchQuery);
    });

  $: sortedTableMeta = filteredTableMeta.slice().sort((a, b) => {
    const va = getSortValue(a, sortKey);
    const vb = getSortValue(b, sortKey);

    if (va === null && vb === null) return 0;
    if (va === null) return 1;
    if (vb === null) return -1;

    if (typeof va === 'string' && typeof vb === 'string') {
      const diff = va.localeCompare(vb);
      return sortDirection === 'asc' ? diff : -diff;
    }

    const diff = Number(va) - Number(vb);
    return sortDirection === 'asc' ? diff : -diff;
  });
</script>

<div class="h-full flex bg-white dark:bg-gray-800">
  <aside class="w-64 border-r border-gray-200 dark:border-gray-700 flex flex-col">
    <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">数据库</h3>
      <button class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700" on:click={loadDatabases}>刷新</button>
    </div>
    <div class="px-2 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
      <input
        type="text"
        bind:value={databaseSearchQuery}
        placeholder="搜索数据库..."
        class="w-full text-xs px-3 py-2 rounded border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
      />
    </div>
    <div class="flex-1 overflow-y-auto p-2">
      {#if loadingDatabases}
        <div class="text-xs text-gray-500 dark:text-gray-400 p-2">加载中...</div>
      {:else if databases.length === 0}
        <div class="text-xs text-gray-500 dark:text-gray-400 p-2">暂无数据库</div>
      {:else if filteredDatabases.length === 0}
        <div class="text-xs text-gray-500 dark:text-gray-400 p-2">未匹配到数据库</div>
      {:else}
        {#each filteredDatabases as dbName}
          <button
            type="button"
            class="w-full text-left text-sm px-3 py-2 rounded mb-1 transition-colors {selectedDatabase === dbName ? 'bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300' : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-200'}"
            on:click={() => handleDatabaseClick(dbName)}
          >
            {dbName}
          </button>
        {/each}
      {/if}
    </div>
  </aside>

  <section class="flex-1 min-w-0 flex flex-col">
    <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
      <h3 class="text-sm font-semibold text-gray-900 dark:text-white">表信息列表 {selectedDatabase ? `(${selectedDatabase})` : ''}</h3>
      {#if selectedDatabase}
        <button class="text-xs px-2 py-1 rounded bg-gray-100 dark:bg-gray-700" on:click={() => loadTableMetadata(selectedDatabase)}>刷新</button>
      {/if}
    </div>

    {#if selectedDatabase}
      <div class="px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
        <input
          type="text"
          bind:value={tableSearchQuery}
          placeholder="搜索表名或注释..."
          class="w-full text-xs px-3 py-2 rounded border border-gray-200 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-800 dark:text-gray-100"
        />
      </div>
    {/if}

    <div class="flex-1 overflow-y-auto p-3">
      {#if errorMessage}
        <div class="mb-2 text-xs text-red-500">{errorMessage}</div>
      {/if}

      {#if !selectedDatabase}
        <div class="text-sm text-gray-500 dark:text-gray-400">请选择一个数据库</div>
      {:else if loadingTables}
        <div class="text-sm text-gray-500 dark:text-gray-400">表加载中...</div>
      {:else if selectedTableMeta.length === 0}
        <div class="text-sm text-gray-500 dark:text-gray-400">暂无表</div>
      {:else if filteredTableMeta.length === 0}
        <div class="text-sm text-gray-500 dark:text-gray-400">未匹配到表，请调整搜索关键词</div>
      {:else}
        <div class="overflow-auto border border-gray-200 dark:border-gray-700 rounded">
          <table class="w-full text-xs border-collapse">
            <thead class="bg-gray-50 dark:bg-gray-700/60">
              <tr>
                {#each sortableColumns as column}
                  <th class="text-left px-3 py-2 border-b border-gray-200 dark:border-gray-700">
                    <button
                      type="button"
                      class="inline-flex items-center gap-1 select-none hover:text-blue-600 dark:hover:text-blue-300 {sortKey === column.key ? 'text-blue-600 dark:text-blue-300' : ''}"
                      on:click={() => handleSort(column.key)}
                    >
                      <span>{column.label}</span>
                      {#if sortKey === column.key}
                        <span aria-hidden="true">{sortDirection === 'asc' ? '↑' : '↓'}</span>
                      {/if}
                    </button>
                  </th>
                {/each}
              </tr>
            </thead>
            <tbody>
              {#each sortedTableMeta as meta}
                <tr
                  class="hover:bg-blue-50 dark:hover:bg-blue-900/20 cursor-pointer"
                  on:click={() => handleTableOpen(meta.tableName)}
                >
                  <td class="px-3 py-2 border-b border-gray-100 dark:border-gray-700 text-blue-700 dark:text-blue-300">{meta.tableName}</td>
                  <td class="px-3 py-2 border-b border-gray-100 dark:border-gray-700">{formatNumber(meta.rowCount)}</td>
                  <td class="px-3 py-2 border-b border-gray-100 dark:border-gray-700">{formatBytes(meta.dataLength)}</td>
                  <td class="px-3 py-2 border-b border-gray-100 dark:border-gray-700">{meta.engine || '-'}</td>
                  <td class="px-3 py-2 border-b border-gray-100 dark:border-gray-700">{formatDate(meta.createTime)}</td>
                  <td class="px-3 py-2 border-b border-gray-100 dark:border-gray-700">{formatDate(meta.updateTime)}</td>
                  <td class="px-3 py-2 border-b border-gray-100 dark:border-gray-700">{meta.collation || '-'}</td>
                  <td class="px-3 py-2 border-b border-gray-100 dark:border-gray-700">{meta.comment || '-'}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </div>
  </section>
</div>
