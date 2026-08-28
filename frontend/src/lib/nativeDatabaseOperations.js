export const NATIVE_DB_OPERATIONS = {
  REDIS_SET: 'set',
  REDIS_SAVE: 'save',
  REDIS_DELETE: 'delete',
  ES_INDEX: 'index_document',
  ES_UPDATE: 'update_document',
  ES_DELETE: 'delete_document'
};

export function defaultElasticsearchQuery(size = 20) {
  return JSON.stringify({ query: { match_all: {} }, size }, null, 2);
}

export function buildRedisSetPayload(value, ttlSeconds) {
  const payload = { type: 'string', value: String(value ?? '') };
  const ttl = Number(ttlSeconds);
  if (Number.isFinite(ttl) && ttl > 0) {
    payload.ttlSeconds = ttl;
  }
  return JSON.stringify(payload);
}

export function createRedisEditorState(content) {
  const parsed = parseNativeResourceContent(content);
  const type = parsed.type || 'unknown';
  return {
    type,
    value: parsed.value ?? '',
    fields: Object.entries(parsed.fields || {}).map(([field, value]) => ({ field, value: String(value ?? '') })),
    items: [...(parsed.items || [])].map((item) => String(item ?? '')),
    members: [...(parsed.members || [])].map((member) => String(member ?? '')),
    entries: (parsed.entries || []).map((entry) => ({
      member: String(entry?.member ?? ''),
      score: Number(entry?.score ?? 0)
    })),
    ttlSeconds: parsed.ttlSeconds ?? -1,
    truncated: Boolean(parsed.truncated),
    length: parsed.length
  };
}

export function buildRedisSavePayload(state, ttlInput) {
  const payload = { type: state.type };
  switch (state.type) {
    case 'string':
      payload.value = state.value ?? '';
      break;
    case 'hash':
      payload.fields = Object.fromEntries(
        (state.fields || [])
          .filter((row) => String(row.field || '').trim())
          .map((row) => [String(row.field).trim(), String(row.value ?? '')])
      );
      break;
    case 'list':
      payload.items = (state.items || []).map((item) => String(item ?? ''));
      break;
    case 'set':
      payload.members = (state.members || []).map((member) => String(member ?? '')).filter(Boolean);
      break;
    case 'zset':
      payload.entries = (state.entries || [])
        .filter((entry) => String(entry.member || '').trim())
        .map((entry) => ({ member: String(entry.member).trim(), score: Number(entry.score ?? 0) }));
      break;
    default:
      throw new Error(`不支持的 Redis 类型: ${state.type}`);
  }
  const ttl = Number(ttlInput);
  if (Number.isFinite(ttl) && ttl > 0) {
    payload.ttlSeconds = ttl;
  }
  return JSON.stringify(payload);
}

export function redisInspectorState(content) {
  const state = createRedisEditorState(content);
  return {
    type: state.type,
    editable: ['string', 'hash', 'list', 'set', 'zset'].includes(state.type),
    value: state.value,
    ttlSeconds: state.ttlSeconds,
    truncated: state.truncated,
    length: state.length
  };
}

export function buildElasticsearchDocumentPayload(id, document, operation = NATIVE_DB_OPERATIONS.ES_INDEX) {
  let parsedDocument = document;
  if (typeof document === 'string') {
    parsedDocument = JSON.parse(document);
  }
  const payload = { document: parsedDocument };
  if (id) {
    payload.id = id;
  }
  if (operation === NATIVE_DB_OPERATIONS.ES_DELETE) {
    return JSON.stringify({ id });
  }
  return JSON.stringify(payload);
}

export function parseNativeResourceContent(content) {
  try {
    return JSON.parse(content || '{}');
  } catch {
    return {};
  }
}

export function parseElasticsearchQueryHits(content) {
  const parsed = parseNativeResourceContent(content);
  const hits = Array.isArray(parsed.hits) ? parsed.hits : [];
  return hits.map((hit) => {
    const source = typeof hit === 'string' ? parseNativeResourceContent(hit) : hit;
    const id = source?._id || source?.id || '';
    const document = source?._source ?? source?.document ?? source;
    return { id, raw: source, document };
  });
}

export function formatMutationMessage(result) {
  if (!result) return '';
  return result.summary || '操作已完成';
}

export function redisDatabaseOptions(databases) {
  return (databases || []).map((item) => ({
    value: String(item.name),
    label: `DB ${item.name}`
  }));
}

export function filterNativeResources(resources, searchTerm) {
  const keyword = String(searchTerm || '').trim().toLowerCase();
  if (!keyword) return resources || [];
  return (resources || []).filter((resource) => String(resource?.name || '').toLowerCase().includes(keyword));
}

export function parseElasticsearchClusterOverview(content) {
  const parsed = parseNativeResourceContent(content);
  return {
    clusterName: parsed.clusterName || '',
    version: parsed.version || '',
    health: parsed.health || 'unknown',
    nodeCount: Number(parsed.nodeCount ?? parsed.numberOfNodes ?? 0) || 0,
    dataNodeCount: Number(parsed.numberOfDataNodes ?? 0) || 0,
    activeShards: Number(parsed.activeShards ?? 0) || 0,
    unassignedShards: Number(parsed.unassignedShards ?? 0) || 0,
    nodes: Array.isArray(parsed.nodes) ? parsed.nodes : []
  };
}

export function parseElasticsearchIndexMetadata(content) {
  const parsed = parseNativeResourceContent(content);
  const stats = parsed.stats || {};
  return {
    health: stats.health || 'unknown',
    status: stats.status || '',
    docsCount: stats.docsCount || '0',
    docsDeleted: stats.docsDeleted || '0',
    storeSize: stats.storeSize || '-',
    priStoreSize: stats.priStoreSize || '-',
    primaries: stats.primaries || '-',
    replicas: stats.replicas || '-',
    mapping: parsed.mapping || {}
  };
}
