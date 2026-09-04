const nativeTypes = {
  redis: { port: '6379', resourceLabel: '键', requiresUsername: false },
  mongodb: { port: '27017', resourceLabel: '集合' },
  elasticsearch: { port: '9200', resourceLabel: '索引' },
  memcached: { port: '11211', resourceLabel: '统计项' },
  cassandra: { port: '9042', resourceLabel: '表' },
  couchbase: { port: '8091', resourceLabel: '集合' },
  influxdb: { port: '8086', resourceLabel: '资源' },
  neo4j: { port: '7687', resourceLabel: '资源' },
  // Kafka / RocketMQ / RabbitMQ 常见无认证或密码可选；表单支持认证=无
  kafka: { port: '9092', resourceLabel: '主题', requiresUsername: false, requiresPassword: false, supportsAuthNone: true },
  rocketmq: { port: '9876', resourceLabel: '主题', requiresUsername: false, requiresPassword: false, supportsAuthNone: true },
  rabbitmq: { port: '5672', resourceLabel: '队列', requiresUsername: false, requiresPassword: false, supportsAuthNone: true }
};

function normalizeDatabaseType(databaseType) {
  return String(databaseType || '').toLowerCase();
}

export function isNativeDatabaseType(databaseType) {
  const type = normalizeDatabaseType(databaseType);
  if (type === 'opensearch') return true;
  return Object.prototype.hasOwnProperty.call(nativeTypes, type);
}

export function databaseTypeConfig(databaseType) {
  const type = normalizeDatabaseType(databaseType);
  if (type === 'opensearch') return nativeTypes.elasticsearch;
  return nativeTypes[type] || null;
}

export function databaseTypeRequiresUsername(databaseType) {
  return databaseTypeConfig(databaseType)?.requiresUsername !== false;
}

/** 密码是否为连接必填；Kafka 等无认证场景为 false。 */
export function databaseTypeRequiresPassword(databaseType) {
  return databaseTypeConfig(databaseType)?.requiresPassword !== false;
}

/** 是否支持「认证=无」（参考 DataGrip Kafka Authentication: None）。 */
export function databaseTypeSupportsAuthNone(databaseType) {
  return databaseTypeConfig(databaseType)?.supportsAuthNone === true;
}

/** 连接资产是否允许空密码（无认证或类型本身不要求密码）。 */
export function connectionAllowsEmptyPassword(assetOrType, authType) {
  const dbType = typeof assetOrType === 'string'
    ? assetOrType
    : (assetOrType?.metadata?.db_type || assetOrType?.dbType || assetOrType?.db_type || '');
  const auth = String(authType || assetOrType?.auth_type || assetOrType?.authType || 'password').toLowerCase();
  if (auth === 'none') return true;
  return !databaseTypeRequiresPassword(dbType);
}

/** 仅 JDBC 关系库在资产树展开库/Schema；缓存、搜索、消息等原生类型走独立工作区。 */
export function assetSupportsJdbcSidebar(asset) {
  if (!asset || normalizeDatabaseType(asset.type) !== 'database') return false;
  const dbType = normalizeDatabaseType(
    asset?.metadata?.db_type || asset?.dbType || asset?.db_type || ''
  );
  return !isNativeDatabaseType(dbType);
}
