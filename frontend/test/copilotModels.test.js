import test from 'node:test';
import assert from 'node:assert/strict';
import {
  availableFetchedModels,
  createProviderFromPreset,
  flattenCopilotModels,
  mergeFetchedModels,
  mirrorLegacyCopilotFields,
  providerKindAllowsEmptyAPIKey,
  resolveActiveModelId
} from '../src/lib/copilotModels.js';

test('DeepSeek preset seeds official base URL and common models', () => {
  const provider = createProviderFromPreset('deepseek');
  assert.equal(provider.kind, 'deepseek');
  assert.equal(provider.name, 'DeepSeek');
  assert.equal(provider.base_url, 'https://api.deepseek.com/v1');
  assert.ok(provider.id);
  assert.deepEqual(provider.models.map((item) => item.model_id), ['deepseek-chat', 'deepseek-reasoner']);
  assert.ok(provider.models.every((item) => item.id && item.name));
});

test('flattenCopilotModels keeps provider metadata for the picker', () => {
  const models = flattenCopilotModels([
    {
      id: 'p1',
      name: 'DeepSeek',
      kind: 'deepseek',
      base_url: 'https://api.deepseek.com/v1',
      models: [{ id: 'm1', name: 'Chat', model_id: 'deepseek-chat' }]
    }
  ]);
  assert.equal(models.length, 1);
  assert.equal(models[0].provider_name, 'DeepSeek');
  assert.equal(models[0].provider_id, 'p1');
  assert.equal(models[0].model_id, 'deepseek-chat');
});

test('resolveActiveModelId keeps a valid id and falls back to the first model', () => {
  const providers = [{
    id: 'p1',
    name: 'OpenAI',
    models: [
      { id: 'm1', name: 'GPT-4o', model_id: 'gpt-4o' },
      { id: 'm2', name: 'Mini', model_id: 'gpt-4o-mini' }
    ]
  }];
  assert.equal(resolveActiveModelId(providers, 'm2'), 'm2');
  assert.equal(resolveActiveModelId(providers, 'gone'), 'm1');
  assert.equal(resolveActiveModelId([], 'm1'), '');
});

test('Ollama allows an empty API key while cloud presets do not', () => {
  assert.equal(providerKindAllowsEmptyAPIKey('ollama'), true);
  assert.equal(providerKindAllowsEmptyAPIKey('deepseek'), false);
});

test('mergeFetchedModels keeps custom names and only appends missing ids', () => {
  const merged = mergeFetchedModels(
    [
      { id: 'local-1', name: '对话', model_id: 'deepseek-chat' },
      { id: 'local-2', name: '', model_id: '' }
    ],
    [
      { id: 'deepseek-chat' },
      { id: 'deepseek-reasoner', name: 'Reasoner' },
      { id: '  ' }
    ]
  );
  assert.equal(merged.length, 2);
  assert.equal(merged[0].id, 'local-1');
  assert.equal(merged[0].name, '对话');
  assert.equal(merged[0].model_id, 'deepseek-chat');
  assert.equal(merged[1].id, 'local-2');
  assert.equal(merged[1].model_id, 'deepseek-reasoner');
  assert.equal(merged[1].name, 'Reasoner');
});

test('mergeFetchedModels appends when there is no empty placeholder', () => {
  const merged = mergeFetchedModels(
    [{ id: 'local-1', name: '对话', model_id: 'deepseek-chat' }],
    [{ id: 'deepseek-reasoner' }]
  );
  assert.equal(merged.length, 2);
  assert.equal(merged[1].model_id, 'deepseek-reasoner');
  assert.equal(merged[1].name, 'deepseek-reasoner');
  assert.ok(merged[1].id);
});

test('availableFetchedModels hides ids already on the provider', () => {
  const remaining = availableFetchedModels(
    [{ model_id: 'deepseek-chat' }],
    [{ id: 'deepseek-chat' }, { id: 'deepseek-reasoner', name: 'Reasoner' }]
  );
  assert.deepEqual(remaining, [{ id: 'deepseek-reasoner', name: 'Reasoner' }]);
});

test('mirrorLegacyCopilotFields copies the active model into old settings keys', () => {
  const mirrored = mirrorLegacyCopilotFields([
    {
      id: 'p1',
      base_url: 'https://api.deepseek.com/v1',
      models: [
        { id: 'm1', model_id: 'deepseek-chat' },
        { id: 'm2', model_id: 'deepseek-reasoner' }
      ]
    }
  ], 'm2');
  assert.equal(mirrored.copilot_base_url, 'https://api.deepseek.com/v1');
  assert.equal(mirrored.copilot_model, 'deepseek-reasoner');
});
