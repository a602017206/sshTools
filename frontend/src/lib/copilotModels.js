function newCopilotId() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  return `id-${Math.random().toString(36).slice(2)}${Date.now().toString(36)}`;
}

export const COPILOT_PROVIDER_PRESETS = [
  {
    kind: 'deepseek',
    name: 'DeepSeek',
    base_url: 'https://api.deepseek.com/v1',
    models: [
      { name: 'DeepSeek Chat', model_id: 'deepseek-chat' },
      { name: 'DeepSeek Reasoner', model_id: 'deepseek-reasoner' }
    ]
  },
  {
    kind: 'openai',
    name: 'OpenAI',
    base_url: 'https://api.openai.com/v1',
    models: [
      { name: 'GPT-4o', model_id: 'gpt-4o' },
      { name: 'GPT-4o mini', model_id: 'gpt-4o-mini' }
    ]
  },
  {
    kind: 'ollama',
    name: 'Ollama',
    base_url: 'http://127.0.0.1:11434/v1',
    models: [
      { name: 'Llama 3.1', model_id: 'llama3.1' }
    ]
  },
  {
    kind: 'custom',
    name: '自定义',
    base_url: '',
    models: []
  }
];

function cloneModels(models) {
  return (Array.isArray(models) ? models : []).map((item) => ({
    id: item.id || newCopilotId(),
    name: String(item.name || item.model_id || '未命名模型'),
    model_id: String(item.model_id || '')
  }));
}

export function createProviderFromPreset(kind, overrides = {}) {
  const preset = COPILOT_PROVIDER_PRESETS.find((item) => item.kind === kind) || COPILOT_PROVIDER_PRESETS[COPILOT_PROVIDER_PRESETS.length - 1];
  return {
    id: overrides.id || newCopilotId(),
    kind: preset.kind,
    name: overrides.name || preset.name,
    base_url: overrides.base_url !== undefined ? overrides.base_url : preset.base_url,
    models: cloneModels(overrides.models || preset.models)
  };
}

export function createEmptyModel() {
  return { id: newCopilotId(), name: '', model_id: '' };
}

export function flattenCopilotModels(providers) {
  const list = [];
  for (const provider of Array.isArray(providers) ? providers : []) {
    for (const model of Array.isArray(provider?.models) ? provider.models : []) {
      list.push({
        id: model.id,
        name: model.name || model.model_id || '未命名模型',
        model_id: model.model_id || '',
        provider_id: provider.id,
        provider_name: provider.name || provider.kind || '服务商',
        provider_kind: provider.kind || 'custom',
        base_url: provider.base_url || ''
      });
    }
  }
  return list;
}

export function resolveActiveModelId(providers, activeId) {
  const models = flattenCopilotModels(providers);
  if (!models.length) return '';
  if (activeId && models.some((item) => item.id === activeId)) return activeId;
  return models[0].id;
}

export function providerKindAllowsEmptyAPIKey(kind) {
  return String(kind || '').toLowerCase() === 'ollama';
}

export function mirrorLegacyCopilotFields(providers, activeId) {
  const models = flattenCopilotModels(providers);
  const active = models.find((item) => item.id === activeId) || models[0];
  if (!active) {
    return { copilot_base_url: '', copilot_model: '' };
  }
  return {
    copilot_base_url: active.base_url || '',
    copilot_model: active.model_id || ''
  };
}

export function modelPickerLabel(model) {
  const name = String(model?.name || model?.model_id || '未命名模型');
  const provider = String(model?.provider_name || '').trim();
  return provider && provider !== name ? `${name} · ${provider}` : name;
}

function fetchedModelID(item) {
  return String(item?.id || item?.ID || item?.model_id || '').trim();
}

function fetchedModelName(item, id) {
  return String(item?.name || item?.Name || id).trim() || id;
}

export function availableFetchedModels(existing, fetched) {
  const have = new Set(
    (Array.isArray(existing) ? existing : [])
      .map((item) => String(item?.model_id || '').trim())
      .filter(Boolean)
  );
  const out = [];
  const seen = new Set();
  for (const item of Array.isArray(fetched) ? fetched : []) {
    const id = fetchedModelID(item);
    if (!id || have.has(id) || seen.has(id)) continue;
    seen.add(id);
    out.push({ id, name: fetchedModelName(item, id) });
  }
  return out;
}

export function mergeFetchedModels(existing, fetched) {
  const next = (Array.isArray(existing) ? existing : []).map((item) => ({
    id: item.id || newCopilotId(),
    name: String(item.name || ''),
    model_id: String(item.model_id || '')
  }));
  for (const item of availableFetchedModels(next, fetched)) {
    const emptyIdx = next.findIndex((model) => !model.model_id.trim());
    if (emptyIdx >= 0) {
      next[emptyIdx] = {
        ...next[emptyIdx],
        model_id: item.id,
        name: next[emptyIdx].name.trim() || item.name
      };
    } else {
      next.push({ id: newCopilotId(), name: item.name, model_id: item.id });
    }
  }
  return next;
}
