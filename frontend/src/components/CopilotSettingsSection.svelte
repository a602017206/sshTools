<script>
  import {
    COPILOT_PROVIDER_PRESETS,
    availableFetchedModels,
    createEmptyModel,
    createProviderFromPreset,
    mergeFetchedModels,
    mirrorLegacyCopilotFields,
    providerKindAllowsEmptyAPIKey,
    resolveActiveModelId
  } from '../lib/copilotModels.js';

  export let draft;
  export let providerKeys = {};
  export let providerHasKey = {};
  export let providerKeyBusy = {};
  export let providerKeyError = {};
  export let onProviderKeyChange = () => {};
  export let onClearProviderKey = async () => {};

  // 必须把列表做成响应式变量。模板里调用 providers() 时 Svelte 订阅不到 draft，
  // 点添加只会改数据、不会重绘卡片。
  let providerList = [];
  let fetchBusy = {};
  let fetchError = {};
  let fetchedModels = {};
  let selectedFetchID = {};
  $: providerList = Array.isArray(draft?.copilot_providers) ? draft.copilot_providers : [];

  function commitProviders(next, extra = {}) {
    const providers = Array.isArray(next) ? next : [];
    const activeId = resolveActiveModelId(
      providers,
      extra.copilot_active_model_id ?? draft?.copilot_active_model_id
    );
    draft = {
      ...draft,
      ...extra,
      copilot_providers: providers,
      copilot_active_model_id: activeId,
      ...mirrorLegacyCopilotFields(providers, activeId)
    };
    providerList = providers;
  }

  function addProvider(kind) {
    const created = createProviderFromPreset(kind);
    const extra = (!draft?.copilot_active_model_id && created.models?.[0]?.id)
      ? { copilot_active_model_id: created.models[0].id }
      : {};
    commitProviders([...providerList, created], extra);
  }

  function removeProvider(providerId) {
    commitProviders(providerList.filter((item) => item.id !== providerId));
  }

  function addModel(providerId) {
    commitProviders(providerList.map((provider) => {
      if (provider.id !== providerId) return provider;
      return { ...provider, models: [...(provider.models || []), createEmptyModel()] };
    }));
  }

  function removeModel(providerId, modelId) {
    commitProviders(providerList.map((provider) => {
      if (provider.id !== providerId) return provider;
      return { ...provider, models: (provider.models || []).filter((model) => model.id !== modelId) };
    }));
  }

  function updateProvider(providerId, patch) {
    commitProviders(providerList.map((provider) => (
      provider.id === providerId ? { ...provider, ...patch } : provider
    )));
  }

  function updateModel(providerId, modelId, patch) {
    commitProviders(providerList.map((provider) => {
      if (provider.id !== providerId) return provider;
      return {
        ...provider,
        models: (provider.models || []).map((model) => (model.id === modelId ? { ...model, ...patch } : model))
      };
    }));
  }

  function canFetchModels(provider) {
    if (!String(provider?.base_url || '').trim()) return false;
    if (providerKindAllowsEmptyAPIKey(provider?.kind)) return true;
    return Boolean(String(providerKeys[provider.id] || '').trim() || providerHasKey[provider.id]);
  }

  function providerById(providerId) {
    return providerList.find((item) => item.id === providerId);
  }

  function remainingFetched(provider) {
    return availableFetchedModels(provider?.models, fetchedModels[provider?.id] || []);
  }

  function selectFirstRemaining(providerId, existing) {
    const remaining = availableFetchedModels(existing, fetchedModels[providerId] || []);
    selectedFetchID = { ...selectedFetchID, [providerId]: remaining[0]?.id || '' };
  }

  async function fetchProviderModels(provider) {
    const id = provider.id;
    fetchBusy = { ...fetchBusy, [id]: true };
    fetchError = { ...fetchError, [id]: '' };
    try {
      const api = window.wailsBindings || window.go?.main?.App || {};
      if (typeof api.ListCopilotProviderModels !== 'function') {
        throw new Error('当前版本不支持获取模型列表');
      }
      const list = await api.ListCopilotProviderModels(
        provider.base_url || '',
        providerKeys[id] || '',
        id,
        provider.kind || ''
      );
      const models = Array.isArray(list) ? list : [];
      fetchedModels = { ...fetchedModels, [id]: models };
      selectFirstRemaining(id, provider.models);
      if (!models.length) {
        fetchError = { ...fetchError, [id]: '接口没有返回模型' };
      }
    } catch (error) {
      console.error('Failed to list copilot provider models:', error);
      fetchedModels = { ...fetchedModels, [id]: [] };
      selectedFetchID = { ...selectedFetchID, [id]: '' };
      fetchError = { ...fetchError, [id]: String(error?.message || error) };
    } finally {
      fetchBusy = { ...fetchBusy, [id]: false };
    }
  }

  function applyFetched(providerId, items) {
    const provider = providerById(providerId);
    if (!provider) return;
    const merged = mergeFetchedModels(provider.models, items);
    commitProviders(providerList.map((item) => (
      item.id === providerId ? { ...item, models: merged } : item
    )));
    selectFirstRemaining(providerId, merged);
  }

  function addSelectedFetched(providerId) {
    const provider = providerById(providerId);
    if (!provider) return;
    const selected = String(selectedFetchID[providerId] || '').trim();
    const item = remainingFetched(provider).find((model) => model.id === selected);
    if (item) applyFetched(providerId, [item]);
  }

  function addAllFetched(providerId) {
    const provider = providerById(providerId);
    if (!provider) return;
    applyFetched(providerId, remainingFetched(provider));
  }

  function fetchButtonTitle(provider) {
    if (canFetchModels(provider)) return '从服务商接口拉取模型列表';
    if (providerKindAllowsEmptyAPIKey(provider?.kind)) return '请先填写 Base URL';
    return '请先填写 Base URL 和 API Key';
  }
</script>

<div class="space-y-6">
  <div class="rounded-xl border border-slate-200 dark:border-slate-700 p-4 bg-slate-50/70 dark:bg-slate-900/50 space-y-4">
    <div class="flex items-start justify-between gap-3">
      <div>
        <div class="text-sm font-semibold text-slate-900 dark:text-slate-100">模型服务商</div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">先添加服务商，再在下面挂多个可切换的模型。对话里会记住上次选用的模型。</p>
      </div>
      <div class="flex flex-wrap justify-end gap-1.5">
        {#each COPILOT_PROVIDER_PRESETS as preset}
          <button
            type="button"
            class="px-2.5 py-1.5 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200"
            on:click={() => addProvider(preset.kind)}
          >
            + {preset.name}
          </button>
        {/each}
      </div>
    </div>

    {#if providerList.length === 0}
      <div class="text-xs text-slate-500 dark:text-slate-400">还没有服务商。可先加 DeepSeek、OpenAI、Ollama，或自定义兼容接口。</div>
    {/if}

    {#each providerList as provider (provider.id)}
      {@const remaining = availableFetchedModels(provider.models, fetchedModels[provider.id] || [])}
      <div class="rounded-lg border border-slate-200 dark:border-slate-700 bg-white/80 dark:bg-slate-800/70 p-3 space-y-3">
        <div class="grid grid-cols-2 gap-3">
          <label class="space-y-1 block">
            <div class="text-xs font-semibold text-slate-700 dark:text-slate-200">显示名</div>
            <input
              type="text"
              value={provider.name}
              on:input={(event) => updateProvider(provider.id, { name: event.currentTarget.value })}
              class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm"
            />
          </label>
          <label class="space-y-1 block">
            <div class="text-xs font-semibold text-slate-700 dark:text-slate-200">Base URL</div>
            <input
              type="text"
              value={provider.base_url}
              on:input={(event) => updateProvider(provider.id, { base_url: event.currentTarget.value })}
              placeholder="https://api.example.com/v1"
              autocomplete="off"
              class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm"
            />
          </label>
        </div>

        <label class="space-y-1 block">
          <div class="text-xs font-semibold text-slate-700 dark:text-slate-200">API Key</div>
          <input
            type="password"
            value={providerKeys[provider.id] || ''}
            on:input={(event) => onProviderKeyChange(provider.id, event.currentTarget.value)}
            placeholder={providerHasKey[provider.id] ? '留空则保留已保存的密钥' : (provider.kind === 'ollama' ? 'Ollama 可留空' : '尚未保存密钥')}
            autocomplete="off"
            class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm"
          />
        </label>
        <div class="flex items-center justify-between gap-2">
          <span class="text-xs text-slate-500 dark:text-slate-400">
            {providerHasKey[provider.id] ? '已保存密钥' : (provider.kind === 'ollama' ? '本地模型可不填密钥' : '尚未保存密钥')}
          </span>
          <div class="flex gap-2">
            <button
              type="button"
              class="px-3 py-1.5 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200 disabled:opacity-50"
              disabled={!providerHasKey[provider.id] || providerKeyBusy[provider.id]}
              on:click={() => onClearProviderKey(provider.id)}
            >
              清除密钥
            </button>
            <button
              type="button"
              class="px-3 py-1.5 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-red-600 dark:text-red-300"
              on:click={() => removeProvider(provider.id)}
            >
              删除服务商
            </button>
          </div>
        </div>
        {#if providerKeyError[provider.id]}
          <p class="text-xs text-red-500">{providerKeyError[provider.id]}</p>
        {/if}

        <div class="space-y-2">
          <div class="flex items-center justify-between gap-2">
            <div class="text-xs font-semibold text-slate-700 dark:text-slate-200">模型</div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="text-xs text-slate-500 hover:text-slate-800 dark:hover:text-slate-100 disabled:opacity-50"
                disabled={fetchBusy[provider.id] || !canFetchModels(provider)}
                title={fetchButtonTitle(provider)}
                on:click={() => fetchProviderModels(provider)}
              >
                {fetchBusy[provider.id] ? '获取中…' : '获取模型'}
              </button>
              <button type="button" class="text-xs text-slate-500 hover:text-slate-800 dark:hover:text-slate-100" on:click={() => addModel(provider.id)}>+ 添加模型</button>
            </div>
          </div>
          {#if fetchError[provider.id]}
            <p class="text-xs text-red-500">{fetchError[provider.id]}</p>
          {/if}
          {#if remaining.length}
            <div class="flex flex-wrap items-center gap-2">
              <select
                value={selectedFetchID[provider.id] || remaining[0]?.id || ''}
                on:change={(event) => selectedFetchID = { ...selectedFetchID, [provider.id]: event.currentTarget.value }}
                class="min-w-0 flex-1 px-2.5 py-1.5 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-xs"
              >
                {#each remaining as item}
                  <option value={item.id}>{item.name || item.id}</option>
                {/each}
              </select>
              <button
                type="button"
                class="px-2 py-1.5 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200"
                on:click={() => addSelectedFetched(provider.id)}
              >
                加入选中
              </button>
              <button
                type="button"
                class="px-2 py-1.5 text-xs rounded-lg bg-slate-100 dark:bg-slate-700 text-slate-700 dark:text-slate-200"
                on:click={() => addAllFetched(provider.id)}
              >
                加入全部
              </button>
            </div>
            <p class="text-[11px] text-slate-500">已有模型会保留自定义名称，接口里多出来的可以按需加入。</p>
          {:else if (fetchedModels[provider.id] || []).length}
            <p class="text-[11px] text-slate-500">接口返回的模型都已加入。</p>
          {/if}
          {#each provider.models || [] as model (model.id)}
            <div class="grid grid-cols-[1fr_1fr_auto] gap-2 items-end">
              <label class="space-y-1 block">
                <div class="text-[11px] text-slate-500">显示名</div>
                <input
                  type="text"
                  value={model.name}
                  on:input={(event) => updateModel(provider.id, model.id, { name: event.currentTarget.value })}
                  placeholder="DeepSeek Chat"
                  class="w-full px-2.5 py-1.5 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm"
                />
              </label>
              <label class="space-y-1 block">
                <div class="text-[11px] text-slate-500">模型 ID</div>
                <input
                  type="text"
                  value={model.model_id}
                  on:input={(event) => updateModel(provider.id, model.id, { model_id: event.currentTarget.value })}
                  placeholder="deepseek-chat"
                  class="w-full px-2.5 py-1.5 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600 text-sm"
                />
              </label>
              <button type="button" class="px-2 py-1.5 text-xs rounded-lg text-slate-500 hover:text-red-500" on:click={() => removeModel(provider.id, model.id)}>删除</button>
            </div>
          {/each}
          {#if !(provider.models || []).length}
            <p class="text-xs text-slate-500">该服务商还没有模型。</p>
          {/if}
        </div>
      </div>
    {/each}
  </div>

  <div class="rounded-xl border border-slate-200 dark:border-slate-700 p-4 bg-slate-50/70 dark:bg-slate-900/50">
    <div class="grid grid-cols-2 gap-3">
      <label class="space-y-2 block">
        <div class="text-sm font-semibold">最大工具轮次</div>
        <input type="number" min="1" max="8" bind:value={draft.copilot_max_tool_rounds} class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600" />
      </label>
      <label class="space-y-2 block">
        <div class="text-sm font-semibold">单次工具结果上限</div>
        <input type="number" min="1000" max="20000" step="1000" bind:value={draft.copilot_max_tool_result_chars} class="w-full px-3 py-2 rounded-lg bg-white dark:bg-slate-700 border border-slate-200 dark:border-slate-600" />
      </label>
    </div>
  </div>
</div>
