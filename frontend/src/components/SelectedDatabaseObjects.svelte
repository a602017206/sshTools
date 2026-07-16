<script>
  import { createEventDispatcher } from 'svelte';
  import { databaseNavigationStore } from '../stores.js';
  import { databaseSidebarCategories, defaultDatabaseObjectCategory } from '../lib/databaseObjectTree.js';

  export let sessionId = null;
  export let dbConfig = null;

  let objects = {};
  let errors = {};
  let loadingCategory = '';
  let activeCategoryId = defaultDatabaseObjectCategory();
  let searchText = '';
  let selectionSeen = '';
  const dispatch = createEventDispatcher();

  const categories = databaseSidebarCategories();
  $: selected = $databaseNavigationStore[sessionId] || { databaseName: dbConfig?.metadata?.database || '', schemaName: '' };
  $: databaseType = dbConfig?.metadata?.db_type || dbConfig?.dbType || '';
  $: selectionKey = `${sessionId || ''}:${selected.databaseName}:${selected.schemaName}`;
  $: activeCategory = categories.find(category => category.id === activeCategoryId) || categories[0];
  $: activeObjects = objects[activeCategoryId] || [];
  $: filteredObjects = activeObjects.filter(name => name.toLowerCase().includes(searchText.trim().toLowerCase()));
  $: if (sessionId && selected.databaseName && selectionKey !== selectionSeen) {
    selectionSeen = selectionKey;
    activeCategoryId = defaultDatabaseObjectCategory();
    searchText = '';
    objects = {};
    errors = {};
    loadCategory(defaultDatabaseObjectCategory());
  }

  async function loadCategory(categoryID = activeCategoryId, force = false) {
    if (!sessionId || !selected.databaseName) return;
    if (!force && objects[categoryID]) return;

    const category = categories.find(item => item.id === categoryID);
    if (!category) return;
    loadingCategory = categoryID;
    errors = { ...errors, [categoryID]: '' };
    try {
      const names = category.types
        ? await window.wailsBindings.ListDatabaseObjects(sessionId, selected.databaseName, selected.schemaName, category.types)
        : await window.wailsBindings.ListDatabaseRoutines(sessionId, selected.databaseName, selected.schemaName, category.functions);
      objects = { ...objects, [categoryID]: names || [] };
    } catch (error) {
      objects = { ...objects, [categoryID]: [] };
      errors = { ...errors, [categoryID]: error?.message || String(error || `加载${category.label}失败`) };
    } finally {
      if (loadingCategory === categoryID) loadingCategory = '';
    }
  }

  function selectCategory(categoryID) {
    activeCategoryId = categoryID;
    searchText = '';
    loadCategory(categoryID);
  }

  function refresh() {
    loadCategory(activeCategoryId, true);
  }

  function openTableStructure(tableName) {
    dispatch('open-table-structure', {
      sessionId,
      databaseName: selected.databaseName,
      schemaName: selected.schemaName,
      tableName
    });
  }
</script>

<div class="object-browser">
  <header class="object-browser__header">
    <div class="object-browser__location">
      <span class="object-browser__eyebrow">对象</span>
      <strong>{selected.databaseName || '请选择数据库'}</strong>
      {#if selected.schemaName}<span class="object-browser__schema">/ {selected.schemaName}</span>{/if}
    </div>
    <div class="object-browser__tabs" aria-label="数据库对象类型">
      {#each categories as category}
        <button
          type="button"
          class:object-browser__tab--active={activeCategoryId === category.id}
          class="object-browser__tab"
          on:click={() => selectCategory(category.id)}
        >
          <span aria-hidden="true">{category.icon}</span>{category.label}
        </button>
      {/each}
    </div>
  </header>

  <div class="object-browser__body">
    <main class="object-browser__main">
      <div class="object-browser__toolbar">
        <button type="button" class="object-browser__icon-button" title="刷新当前列表" on:click={refresh} disabled={loadingCategory === activeCategoryId}>↻</button>
        <span class="object-browser__toolbar-title">{activeCategory?.label}</span>
        <label class="object-browser__search">
          <span aria-hidden="true">⌕</span>
          <input bind:value={searchText} placeholder={`搜索${activeCategory?.label || '对象'}`} aria-label={`搜索${activeCategory?.label || '对象'}`} />
        </label>
      </div>

      <div class="object-browser__table" role="table" aria-label={`${activeCategory?.label || '对象'}列表`}>
        <div class="object-browser__table-head" role="row">
          <span role="columnheader">名称</span>
          <span role="columnheader">类型</span>
        </div>
        {#if !selected.databaseName}
          <div class="object-browser__empty">请在左侧选择数据库</div>
        {:else if loadingCategory === activeCategoryId}
          <div class="object-browser__empty">正在加载{activeCategory.label}...</div>
        {:else if errors[activeCategoryId]}
          <div class="object-browser__error">加载失败：{errors[activeCategoryId]}</div>
        {:else if !filteredObjects.length}
          <div class="object-browser__empty">{searchText ? '没有匹配的对象' : `暂无${activeCategory.label}`}</div>
        {:else}
          {#each filteredObjects as name}
            {#if activeCategoryId === 'tables'}
              <button type="button" class="object-browser__row object-browser__row--button" role="row" on:click={() => openTableStructure(name)}>
                <span role="cell"><span class="object-browser__item-icon" aria-hidden="true">▦</span>{name}</span>
                <span role="cell">表</span>
              </button>
            {:else}
              <div class="object-browser__row" role="row">
                <span role="cell"><span class="object-browser__item-icon" aria-hidden="true">{activeCategory.icon}</span>{name}</span>
                <span role="cell">{activeCategory.label}</span>
              </div>
            {/if}
          {/each}
        {/if}
      </div>
      <footer class="object-browser__status">{filteredObjects.length} 个{activeCategory?.label || '对象'}</footer>
    </main>

    <aside class="object-browser__details">
      <div class="object-browser__database-mark" aria-hidden="true">▰</div>
      <h2>{selected.databaseName || '未选择数据库'}</h2>
      <p>数据库</p>
      <dl>
        <div><dt>类型</dt><dd>{String(databaseType || 'JDBC').toUpperCase()}</dd></div>
        <div><dt>主机</dt><dd>{dbConfig?.host || dbConfig?.metadata?.host || '当前连接'}</dd></div>
        {#if dbConfig?.port}<div><dt>端口</dt><dd>{dbConfig.port}</dd></div>{/if}
        {#if selected.schemaName}<div><dt>Schema</dt><dd>{selected.schemaName}</dd></div>{/if}
      </dl>
    </aside>
  </div>
</div>

<style>
  .object-browser { height: 100%; display: flex; flex-direction: column; background: var(--bg-primary); color: var(--text-primary); }
  .object-browser__header { display: flex; align-items: end; gap: 24px; min-height: 88px; padding: 16px 24px 0; border-bottom: 1px solid var(--border-primary); }
  .object-browser__location { min-width: 160px; padding-bottom: 15px; display: flex; flex-direction: column; gap: 3px; }
  .object-browser__location strong { font-size: 16px; font-weight: 650; }
  .object-browser__eyebrow { font-size: 12px; color: var(--text-secondary); }
  .object-browser__schema { font-size: 12px; color: var(--text-secondary); }
  .object-browser__tabs { display: flex; gap: 2px; align-self: stretch; }
  .object-browser__tab { min-width: 72px; height: 56px; border: 0; border-bottom: 3px solid transparent; background: transparent; color: var(--text-secondary); font-size: 13px; cursor: pointer; display: inline-flex; align-items: center; justify-content: center; gap: 6px; }
  .object-browser__tab:hover { color: var(--text-primary); background: var(--bg-secondary); }
  .object-browser__tab--active { color: #1687d4; border-bottom-color: #1687d4; font-weight: 600; }
  .object-browser__body { min-height: 0; flex: 1; display: grid; grid-template-columns: minmax(0, 1fr) 232px; }
  .object-browser__main { min-width: 0; display: flex; flex-direction: column; }
  .object-browser__toolbar { min-height: 54px; padding: 0 18px; display: flex; align-items: center; gap: 12px; border-bottom: 1px solid var(--border-primary); }
  .object-browser__icon-button { width: 28px; height: 28px; border: 0; background: transparent; color: var(--text-secondary); font-size: 21px; line-height: 1; cursor: pointer; }
  .object-browser__icon-button:hover:not(:disabled) { color: #1687d4; background: var(--bg-secondary); }
  .object-browser__icon-button:disabled { opacity: .45; cursor: wait; }
  .object-browser__toolbar-title { font-size: 14px; font-weight: 600; }
  .object-browser__search { margin-left: auto; width: min(300px, 46%); height: 32px; padding: 0 9px; display: flex; align-items: center; gap: 7px; border: 1px solid var(--border-primary); color: var(--text-secondary); }
  .object-browser__search:focus-within { border-color: #1687d4; }
  .object-browser__search input { width: 100%; border: 0; outline: 0; background: transparent; color: inherit; font-size: 13px; }
  .object-browser__table { overflow: auto; flex: 1; }
  .object-browser__table-head, .object-browser__row { display: grid; grid-template-columns: minmax(240px, 1fr) 140px; align-items: center; min-height: 38px; }
  .object-browser__table-head { position: sticky; top: 0; z-index: 1; color: var(--text-secondary); background: var(--bg-secondary); border-bottom: 1px solid var(--border-primary); font-size: 12px; }
  .object-browser__table-head span, .object-browser__row span { padding: 0 18px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .object-browser__row { border-bottom: 1px solid var(--border-primary); font-size: 14px; }
  .object-browser__row--button { width: 100%; border-left: 0; border-right: 0; border-top: 0; background: transparent; color: inherit; text-align: left; cursor: pointer; }
  .object-browser__row:hover { background: color-mix(in srgb, #1687d4 7%, transparent); }
  .object-browser__item-icon { display: inline-block; width: 25px; padding: 0 !important; color: #1687d4; }
  .object-browser__empty, .object-browser__error { padding: 28px 18px; font-size: 13px; color: var(--text-secondary); }
  .object-browser__error { color: #dc2626; }
  .object-browser__status { min-height: 28px; padding: 6px 18px; border-top: 1px solid var(--border-primary); color: var(--text-secondary); font-size: 12px; }
  .object-browser__details { padding: 28px 22px; border-left: 1px solid var(--border-primary); background: var(--bg-secondary); overflow: auto; }
  .object-browser__database-mark { color: #0a9d5b; font-size: 52px; line-height: 1; }
  .object-browser__details h2 { margin: 16px 0 4px; font-size: 18px; overflow-wrap: anywhere; }
  .object-browser__details > p { margin: 0 0 24px; color: var(--text-secondary); font-size: 13px; }
  .object-browser__details dl { margin: 0; display: grid; gap: 17px; }
  .object-browser__details dl div { display: grid; gap: 3px; }
  .object-browser__details dt { font-size: 12px; color: var(--text-secondary); }
  .object-browser__details dd { margin: 0; font-size: 13px; overflow-wrap: anywhere; }
  @media (max-width: 900px) { .object-browser__body { grid-template-columns: 1fr; } .object-browser__details { display: none; } .object-browser__header { gap: 8px; padding-left: 16px; } .object-browser__location { display: none; } .object-browser__tab { min-width: 58px; } }
</style>
