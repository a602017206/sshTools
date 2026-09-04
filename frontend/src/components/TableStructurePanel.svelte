<script>
  import { buildCreateTableSQL } from '../lib/tableDefinitionSQL.js';
  import { buildAlterTableStatements } from '../lib/tableAlterSQL.js';

  export let sessionId = null;
  export let dbConfig = null;
  export let databaseName = '';
  export let schemaName = '';
  export let tableName = '';
  export let mode = 'design';

  let ddlData = null;
  let schemaData = null;
  let isLoading = false;
  let isSaving = false;
  let errorMessage = '';
  let successMessage = '';
  let copied = false;
  let loadedRequest = '';
  let draftTableName = tableName || 'new_table';
  let fieldDrafts = [];
  let originalFields = [];

  const newField = (name = '') => ({ _originalName: '', name, type: 'VARCHAR', length: name === 'id' ? '20' : '255', nullable: name !== 'id', primary: name === 'id', defaultValue: '', comment: '' });
  $: isCreateMode = mode === 'create';
  $: databaseType = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toLowerCase();
  $: supportsCreate = ['mysql', 'postgresql', 'kingbase', 'oracle'].includes(databaseType);
  $: titleName = isCreateMode ? '新建表' : (schemaName && tableName ? `${schemaName}.${tableName}` : (tableName || '设计表'));
  $: dbTypeLabel = String(dbConfig?.metadata?.db_type || dbConfig?.dbType || '').toUpperCase();
  $: requestKey = `${sessionId || ''}:${databaseName || ''}:${schemaName || ''}:${tableName || ''}:${mode}`;
  $: alterStatements = buildAlterTableStatements({ databaseType, databaseName, schemaName, tableName, originalFields, fields: fieldDrafts });
  $: ddlPreview = isCreateMode
    ? buildCreateTableSQL({ databaseType, databaseName, schemaName, tableName: draftTableName, fields: fieldDrafts })
    : alterStatements.join('\n');
  $: displayedDDL = isCreateMode ? ddlPreview : (alterStatements.length ? ddlPreview : (ddlData?.ddl || ''));
  $: fieldsReadOnly = !supportsCreate || isSaving;
  $: if (isCreateMode && requestKey !== loadedRequest) {
    loadedRequest = requestKey;
    draftTableName = tableName || 'new_table';
    fieldDrafts = [newField('id'), newField('name'), { ...newField('created_at'), type: 'TIMESTAMP', length: '', nullable: false, primary: false }];
    ddlData = null;
    schemaData = null;
    originalFields = [];
  } else if (!isCreateMode && sessionId && tableName && requestKey !== loadedRequest) {
    loadedRequest = requestKey;
    loadDDL();
  }

  function populateFields(columns = []) {
    const hydratedFields = columns.map(column => ({
      _originalName: column.name,
      name: column.name,
      type: String(column.type || 'VARCHAR').toUpperCase(),
      length: String(column.column_size || ''),
      nullable: Boolean(column.nullable),
      primary: Boolean(column.is_primary_key || column.primary_key || column.isPrimaryKey),
      defaultValue: column.has_default ? String(column.default_value || '') : '',
      comment: String(column.description || '')
    }));
    fieldDrafts = hydratedFields;
    originalFields = hydratedFields.map(field => ({ ...field }));
  }

  async function loadDDL() {
    if (!sessionId || !tableName || !window.wailsBindings) return;
    isLoading = true;
    errorMessage = '';
    try {
      const schema = await window.wailsBindings.GetTableSchemaInSchema(sessionId, databaseName, schemaName, tableName);
      const ddl = schemaName
        ? await window.wailsBindings.GetTableDDLInSchema(sessionId, databaseName, schemaName, tableName)
        : await window.wailsBindings.GetTableDDL(sessionId, databaseName, tableName);
      ddlData = ddl;
      schemaData = schema;
      draftTableName = tableName;
      populateFields(schema?.columns || []);
    } catch (error) {
      errorMessage = `加载表结构失败: ${error?.message || String(error || '未知错误')}`;
    } finally { isLoading = false; }
  }

  function updateField(index, patch) {
    fieldDrafts = fieldDrafts.map((field, fieldIndex) => fieldIndex === index ? { ...field, ...patch } : field);
  }

  async function saveNewTable() {
    if (!ddlPreview || !window.wailsBindings || !sessionId) return;
    isSaving = true;
    errorMessage = '';
    successMessage = '';
    try {
      await window.wailsBindings.ExecuteDatabaseQuery(sessionId, ddlPreview);
      successMessage = `已创建表 ${draftTableName}`;
    } catch (error) {
      errorMessage = `创建表失败: ${error?.message || String(error || '未知错误')}`;
    } finally { isSaving = false; }
  }

  async function saveStructure() {
    if (!alterStatements.length || !window.wailsBindings || !sessionId) return;
    isSaving = true;
    errorMessage = '';
    successMessage = '';
    try {
      for (const statement of alterStatements) await window.wailsBindings.ExecuteDatabaseQuery(sessionId, statement);
      successMessage = `已保存表 ${tableName}`;
      await loadDDL();
    } catch (error) {
      errorMessage = `保存失败：${error?.message || String(error || '未知错误')}`;
    } finally { isSaving = false; }
  }

  async function copyDDL() {
    if (!displayedDDL) return;
    await navigator.clipboard?.writeText(displayedDDL);
    copied = true;
    setTimeout(() => copied = false, 1600);
  }
</script>

<section class="table-designer" aria-label={isCreateMode ? '新建表' : '设计表'}>
  <header class="table-designer__header">
    <div><span>{dbTypeLabel || 'JDBC'} · {isCreateMode ? 'TABLE DESIGNER' : 'TABLE STRUCTURE'}</span><h2>{titleName}</h2></div>
    <div class="table-designer__actions">
      {#if !isCreateMode}<button type="button" on:click={loadDDL} disabled={isLoading}>刷新</button>{/if}
      <button type="button" on:click={copyDDL} disabled={!displayedDDL}>{copied ? '已复制' : '复制 DDL'}</button>
      {#if isCreateMode}<button type="button" class="table-designer__save" on:click={saveNewTable} disabled={!ddlPreview || isSaving}>{isSaving ? '保存中...' : '保存'}</button>{/if}
      {#if !isCreateMode}<button type="button" class="table-designer__save" on:click={saveStructure} disabled={!alterStatements.length || fieldsReadOnly}>{isSaving ? '保存中...' : '保存'}</button>{/if}
    </div>
  </header>

  {#if errorMessage}<div class="table-designer__notice table-designer__notice--error">{errorMessage}</div>{/if}
  {#if successMessage}<div class="table-designer__notice table-designer__notice--success">{successMessage}</div>{/if}
  {#if !supportsCreate}<div class="table-designer__notice table-designer__notice--error">当前仅支持 MySQL、PostgreSQL、人大金仓和 Oracle 的表结构修改。</div>{/if}

  <div class="table-designer__body">
    <label class="table-designer__name">表名<input bind:value={draftTableName} disabled={!isCreateMode} /></label>
    <div class="table-designer__section-head"><strong>字段</strong>{#if !fieldsReadOnly}<button type="button" on:click={() => fieldDrafts = [...fieldDrafts, newField()]}>添加字段</button>{/if}</div>
    <div class="table-designer__grid-wrap">
      <table class="table-designer__grid"><thead><tr><th>字段名</th><th>类型</th><th>长度</th><th>非空</th><th>主键</th><th>默认值</th><th>注释</th>{#if !fieldsReadOnly}<th></th>{/if}</tr></thead>
        <tbody>{#each fieldDrafts as field, index}<tr>
          <td><input value={field.name} on:input={(event) => updateField(index, { name: event.currentTarget.value })} disabled={fieldsReadOnly} /></td>
          <td><select value={field.type} on:change={(event) => updateField(index, { type: event.currentTarget.value })} disabled={fieldsReadOnly}>{#each ['BIGINT', 'INT', 'NUMBER', 'VARCHAR', 'VARCHAR2', 'TEXT', 'CLOB', 'DECIMAL', 'TIMESTAMP', 'DATE', 'BOOLEAN'] as type}<option value={type}>{type}</option>{/each}</select></td>
          <td><input value={field.length} on:input={(event) => updateField(index, { length: event.currentTarget.value })} disabled={fieldsReadOnly} /></td>
          <td><input type="checkbox" checked={!field.nullable} on:change={(event) => updateField(index, { nullable: !event.currentTarget.checked })} disabled={fieldsReadOnly} /></td>
          <td><input type="checkbox" checked={field.primary} on:change={(event) => updateField(index, { primary: event.currentTarget.checked })} disabled={fieldsReadOnly} /></td>
          <td><input value={field.defaultValue} on:input={(event) => updateField(index, { defaultValue: event.currentTarget.value })} disabled={fieldsReadOnly} /></td>
          <td><input value={field.comment} on:input={(event) => updateField(index, { comment: event.currentTarget.value })} disabled={fieldsReadOnly} /></td>
          {#if !fieldsReadOnly}<td><button type="button" class="table-designer__remove" title="删除字段" on:click={() => fieldDrafts = fieldDrafts.filter((_, rowIndex) => rowIndex !== index)}>×</button></td>{/if}
        </tr>{/each}</tbody>
      </table>
    </div>
    {#if !isCreateMode && schemaData?.columns?.length === 0 && !isLoading}<p class="table-designer__empty">暂无字段信息</p>{/if}
    <section class="table-designer__ddl"><div><strong>DDL 预览</strong><span>{isCreateMode ? '保存时将执行以下语句' : (alterStatements.length ? '保存时将执行以下语句' : '当前表定义')}</span></div><pre>{displayedDDL}</pre></section>
  </div>
</section>

<style>
  .table-designer { height: 100%; display: flex; flex-direction: column; overflow: hidden; background: #f7f8f5; color: #1d2935; font: 12px "PingFang SC", "Hiragino Sans GB", sans-serif; }
  .table-designer__header { min-height: 66px; display: flex; align-items: center; justify-content: space-between; gap: 10px; padding: 0 14px; background: #fff; border-bottom: 1px solid #d9e0e4; }
  .table-designer__header span { color: #0e6674; font-size: 10px; font-weight: 700; letter-spacing: .08em; }.table-designer__header h2 { margin: 3px 0 0; font-size: 14px; }.table-designer__actions { display: flex; gap: 5px; }.table-designer button { min-height: 28px; padding: 0 8px; border: 1px solid #cfd8dc; border-radius: 3px; background: #fff; color: #31414d; cursor: pointer; font: inherit; }.table-designer button:disabled { cursor: not-allowed; opacity: .5; }.table-designer__actions .table-designer__save { border-color: #0e6674; background: #0e6674; color: #fff; }
  .table-designer__notice { padding: 8px 14px; border-bottom: 1px solid #d9e0e4; }.table-designer__notice--error { background: #fff1f1; color: #b42318; }.table-designer__notice--success { background: #eff8f5; color: #067647; }.table-designer__body { overflow: auto; padding: 14px; }.table-designer__name { display: grid; gap: 5px; max-width: 320px; color: #52606d; font-weight: 650; }.table-designer input, .table-designer select { box-sizing: border-box; width: 100%; min-height: 28px; border: 1px solid #cfd8dc; border-radius: 3px; background: #fff; color: #1d2935; font: inherit; }.table-designer input:disabled, .table-designer select:disabled { border-color: transparent; background: transparent; color: #52606d; }.table-designer__section-head { display: flex; align-items: center; justify-content: space-between; margin: 20px 0 7px; }.table-designer__grid-wrap { overflow: auto; border: 1px solid #d9e0e4; background: #fff; }.table-designer__grid { width: 100%; min-width: 660px; border-collapse: collapse; }.table-designer__grid th { padding: 7px; background: #f1f4f3; color: #52606d; font-size: 11px; font-weight: 650; text-align: left; white-space: nowrap; }.table-designer__grid td { min-width: 58px; padding: 3px; border-top: 1px solid #e1e6e8; }.table-designer__grid td:nth-child(1) { min-width: 110px; }.table-designer__grid td:nth-child(2) { min-width: 90px; }.table-designer__grid input[type="checkbox"] { width: 16px; min-height: auto; }.table-designer__remove { color: #b42318 !important; border: 0 !important; font-size: 18px !important; }.table-designer__ddl { margin-top: 18px; }.table-designer__ddl > div { display: flex; justify-content: space-between; color: #52606d; }.table-designer__ddl span { color: #7b8791; font-size: 11px; }.table-designer__ddl pre { margin: 7px 0 0; padding: 12px; overflow: auto; border: 1px solid #d9e0e4; background: #fff; color: #31414d; font: 11px "SFMono-Regular", Menlo, monospace; line-height: 1.6; white-space: pre-wrap; }.table-designer__empty { color: #7b8791; }
</style>
