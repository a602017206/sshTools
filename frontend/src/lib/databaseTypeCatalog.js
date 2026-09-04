/** 连接类型细分组（对齐 dbx 侧栏概念；与资产域 cache/search/mq 并存）。 */
export const DATABASE_TYPE_CATEGORIES = [
  { id: 'relational', label: '关系型数据库' },
  { id: 'domestic', label: '国产数据库' },
  { id: 'lightweight', label: '轻量数据库' },
  { id: 'document_cache_search', label: '文档 / 缓存 / 检索' },
  { id: 'graph_timeseries', label: '图谱 / 时序' },
  { id: 'mq', label: '消息队列' }
];

const CATEGORY_BY_TYPE = {
  mysql: 'relational',
  postgresql: 'relational',
  oracle: 'relational',
  sqlserver: 'relational',
  dm: 'domestic',
  kingbase: 'domestic',
  opengauss: 'domestic',
  sqlite: 'lightweight',
  mongodb: 'document_cache_search',
  redis: 'document_cache_search',
  keydb: 'document_cache_search',
  memcached: 'document_cache_search',
  elasticsearch: 'document_cache_search',
  opensearch: 'document_cache_search',
  cassandra: 'document_cache_search',
  couchbase: 'document_cache_search',
  neo4j: 'graph_timeseries',
  influxdb: 'graph_timeseries',
  kafka: 'mq',
  rocketmq: 'mq',
  rabbitmq: 'mq'
};

export function categoryForDbType(databaseType) {
  const type = String(databaseType || '').toLowerCase();
  return CATEGORY_BY_TYPE[type] || 'relational';
}

export function categoryLabel(categoryId) {
  return DATABASE_TYPE_CATEGORIES.find((item) => item.id === categoryId)?.label || categoryId;
}

/** JDBC 驱动管理侧栏分类（不含消息队列原生类型）。 */
export const JDBC_DRIVER_CATEGORIES = [
  { id: 'all', label: '全部驱动' },
  { id: 'relational', label: '关系型数据库' },
  { id: 'domestic', label: '国产数据库' },
  { id: 'lightweight', label: '轻量数据库' }
];

export function groupDatabaseTypesByCategory(databaseTypes) {
  const groups = DATABASE_TYPE_CATEGORIES.map((category) => ({
    id: category.id,
    label: category.label,
    types: []
  }));
  const byId = Object.fromEntries(groups.map((group) => [group.id, group]));

  for (const item of databaseTypes || []) {
    const categoryId = item.category || categoryForDbType(item.value);
    if (byId[categoryId]) {
      byId[categoryId].types.push(item);
    } else {
      byId.relational.types.push(item);
    }
  }

  return groups.filter((group) => group.types.length > 0);
}
