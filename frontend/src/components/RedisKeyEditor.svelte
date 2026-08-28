<script>
  export let state = null;
  export let ttlInput = '';
  export let saving = false;
  export let onSave = () => {};
  export let onDelete = () => {};

  function addHashField() {
    state = { ...state, fields: [...(state.fields || []), { field: '', value: '' }] };
  }

  function removeHashField(index) {
    state = { ...state, fields: state.fields.filter((_, i) => i !== index) };
  }

  function addListItem() {
    state = { ...state, items: [...(state.items || []), ''] };
  }

  function removeListItem(index) {
    state = { ...state, items: state.items.filter((_, i) => i !== index) };
  }

  function moveListItem(index, direction) {
    const items = [...state.items];
    const target = index + direction;
    if (target < 0 || target >= items.length) return;
    [items[index], items[target]] = [items[target], items[index]];
    state = { ...state, items };
  }

  function addSetMember() {
    state = { ...state, members: [...(state.members || []), ''] };
  }

  function removeSetMember(index) {
    state = { ...state, members: state.members.filter((_, i) => i !== index) };
  }

  function addZSetEntry() {
    state = { ...state, entries: [...(state.entries || []), { member: '', score: 0 }] };
  }

  function removeZSetEntry(index) {
    state = { ...state, entries: state.entries.filter((_, i) => i !== index) };
  }
</script>

{#if state}
  <div class="redis-key-editor">
    <div class="redis-key-editor__type">类型：<strong>{state.type}</strong></div>
    {#if state.truncated}
      <p class="redis-key-editor__hint">预览已截断，保存时会写入当前编辑区中的完整内容。</p>
    {/if}

    {#if state.type === 'string'}
      <label>
        <span>值</span>
        <textarea bind:value={state.value} rows="6"></textarea>
      </label>
    {:else if state.type === 'hash'}
      <div class="redis-key-editor__table">
        <div class="redis-key-editor__table-head"><span>字段</span><span>值</span><span></span></div>
        {#each state.fields as row, index}
          <div class="redis-key-editor__table-row">
            <input bind:value={row.field} placeholder="field" />
            <input bind:value={row.value} placeholder="value" />
            <button type="button" class="danger" on:click={() => removeHashField(index)}>删</button>
          </div>
        {/each}
      </div>
      <button type="button" on:click={addHashField}>添加字段</button>
    {:else if state.type === 'list'}
      <div class="redis-key-editor__table">
        <div class="redis-key-editor__table-head"><span>#</span><span>元素</span><span></span></div>
        {#each state.items as item, index}
          <div class="redis-key-editor__table-row">
            <span class="redis-key-editor__index">{index}</span>
            <input bind:value={state.items[index]} placeholder="list item" />
            <div class="redis-key-editor__row-actions">
              <button type="button" on:click={() => moveListItem(index, -1)} disabled={index === 0}>↑</button>
              <button type="button" on:click={() => moveListItem(index, 1)} disabled={index === state.items.length - 1}>↓</button>
              <button type="button" class="danger" on:click={() => removeListItem(index)}>删</button>
            </div>
          </div>
        {/each}
      </div>
      <button type="button" on:click={addListItem}>添加元素</button>
    {:else if state.type === 'set'}
      <div class="redis-key-editor__table">
        {#each state.members as member, index}
          <div class="redis-key-editor__table-row redis-key-editor__table-row--compact">
            <input bind:value={state.members[index]} placeholder="member" />
            <button type="button" class="danger" on:click={() => removeSetMember(index)}>删</button>
          </div>
        {/each}
      </div>
      <button type="button" on:click={addSetMember}>添加成员</button>
    {:else if state.type === 'zset'}
      <div class="redis-key-editor__table">
        <div class="redis-key-editor__table-head"><span>成员</span><span>分数</span><span></span></div>
        {#each state.entries as entry, index}
          <div class="redis-key-editor__table-row">
            <input bind:value={entry.member} placeholder="member" />
            <input type="number" step="any" bind:value={entry.score} />
            <button type="button" class="danger" on:click={() => removeZSetEntry(index)}>删</button>
          </div>
        {/each}
      </div>
      <button type="button" on:click={addZSetEntry}>添加成员</button>
    {:else}
      <p class="redis-key-editor__hint">暂不支持编辑 {state.type} 类型。</p>
    {/if}

    <label>
      <span>TTL（秒，可选）</span>
      <input bind:value={ttlInput} placeholder="留空表示永不过期" />
    </label>

    <div class="redis-key-editor__actions">
      {#if ['string', 'hash', 'list', 'set', 'zset'].includes(state.type)}
        <button type="button" on:click={onSave} disabled={saving}>保存</button>
      {/if}
      <button type="button" class="danger" on:click={onDelete} disabled={saving}>删除键</button>
    </div>
  </div>
{/if}

<style>
  .redis-key-editor { display: grid; gap: 10px; margin-top: 14px; }
  .redis-key-editor__type { font-size: 12px; color: #31414d; }
  .redis-key-editor__hint { margin: 0; font-size: 11px; color: #6d7783; }
  .redis-key-editor label { display: grid; gap: 4px; font-size: 11px; color: #6d7783; }
  .redis-key-editor input, .redis-key-editor textarea {
    width: 100%; border: 1px solid #d9e0e4; border-radius: 4px; padding: 8px;
    font: 12px/1.5 ui-monospace, SFMono-Regular, Menlo, monospace; box-sizing: border-box;
  }
  .redis-key-editor__table { display: grid; gap: 6px; }
  .redis-key-editor__table-head, .redis-key-editor__table-row {
    display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr) auto; gap: 6px; align-items: center;
  }
  .redis-key-editor__table-head { font-size: 11px; color: #7b8791; }
  .redis-key-editor__index { font-size: 11px; color: #7b8791; text-align: center; }
  .redis-key-editor__table-row--compact { grid-template-columns: minmax(0, 1fr) auto; }
  .redis-key-editor__actions { display: flex; gap: 8px; flex-wrap: wrap; }
  button.danger { color: #b91c1c; }
</style>
