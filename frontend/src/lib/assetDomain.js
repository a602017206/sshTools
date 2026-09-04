export const ASSET_DOMAINS = [
  { id: 'all', label: '全部', shortLabel: '全部' },
  { id: 'ssh', label: 'SSH', shortLabel: 'SSH' },
  { id: 'database', label: '数据库', shortLabel: '库' },
  { id: 'cache', label: '缓存', shortLabel: '缓存' },
  { id: 'search', label: '搜索', shortLabel: '搜索' },
  { id: 'mq', label: '消息队列', shortLabel: '消息' },
  { id: 'docker', label: 'Docker', shortLabel: '容器' }
];

const CACHE_DB_TYPES = new Set(['redis', 'memcached', 'keydb']);
const SEARCH_DB_TYPES = new Set(['elasticsearch', 'opensearch']);
const MQ_DB_TYPES = new Set(['kafka', 'rocketmq', 'rabbitmq']);
const DATABASE_DB_TYPES = new Set([
  'mysql',
  'postgresql',
  'sqlite',
  'oracle',
  'sqlserver',
  'dm',
  'kingbase',
  'opengauss',
  'mongodb',
  'cassandra',
  'couchbase',
  'influxdb',
  'neo4j'
]);

export function resolveAssetDbType(asset) {
  return String(asset?.metadata?.db_type || asset?.dbType || asset?.db_type || '').toLowerCase();
}

export function resolveAssetDomain(asset) {
  if (!asset) return 'ssh';

  const explicit = String(asset.metadata?.domain || asset.domain || '').toLowerCase();
  if (ASSET_DOMAINS.some((domain) => domain.id === explicit && domain.id !== 'all')) {
    return explicit;
  }

  const type = String(asset.type || 'ssh').toLowerCase();
  if (type === 'docker') return 'docker';
  if (type === 'ssh' || type === 'server') return 'ssh';

  if (type === 'database') {
    const dbType = resolveAssetDbType(asset);
    if (CACHE_DB_TYPES.has(dbType)) return 'cache';
    if (SEARCH_DB_TYPES.has(dbType)) return 'search';
    if (MQ_DB_TYPES.has(dbType)) return 'mq';
    if (DATABASE_DB_TYPES.has(dbType) || !dbType) return 'database';
    return 'database';
  }

  return 'ssh';
}

export function filterAssetsByDomain(assets, domainId) {
  const domain = String(domainId || 'all').toLowerCase();
  if (!domain || domain === 'all') return assets || [];
  return (assets || []).filter((asset) => resolveAssetDomain(asset) === domain);
}

export function defaultAssetTypeForDomain(domainId) {
  switch (String(domainId || 'all').toLowerCase()) {
    case 'ssh':
      return { assetType: 'ssh', dbType: '' };
    case 'cache':
      return { assetType: 'database', dbType: 'redis' };
    case 'search':
      return { assetType: 'database', dbType: 'elasticsearch' };
    case 'mq':
      return { assetType: 'database', dbType: 'kafka' };
    case 'docker':
      return { assetType: 'docker', dbType: '' };
    case 'database':
      return { assetType: 'database', dbType: 'mysql' };
    default:
      return { assetType: 'ssh', dbType: '' };
  }
}

export function domainLabel(domainId) {
  return ASSET_DOMAINS.find((domain) => domain.id === domainId)?.label || domainId;
}

/** 新建连接对话框中数据库类型的域分组（不含 all/ssh/docker） */
export function groupDatabaseTypesByDomain(databaseTypes) {
  const groups = [
    { id: 'database', label: '数据库', types: [] },
    { id: 'cache', label: '缓存', types: [] },
    { id: 'search', label: '搜索', types: [] },
    { id: 'mq', label: '消息队列', types: [] }
  ];
  const byId = Object.fromEntries(groups.map((group) => [group.id, group]));

  for (const item of databaseTypes || []) {
    const domain = resolveAssetDomain({
      type: 'database',
      metadata: { db_type: item.value }
    });
    if (byId[domain]) {
      byId[domain].types.push(item);
    } else {
      byId.database.types.push(item);
    }
  }

  return groups.filter((group) => group.types.length > 0);
}
