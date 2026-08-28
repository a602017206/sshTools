const DATABASE_TYPE_LABELS = {
  mysql: 'MySQL',
  postgresql: 'PostgreSQL',
  sqlite: 'SQLite',
  oracle: 'Oracle',
  sqlserver: 'SQL Server',
  dm: '达梦 DM',
  kingbase: '人大金仓 Kingbase',
  opengauss: 'openGauss',
  redis: 'Redis / KeyDB',
  mongodb: 'MongoDB',
  elasticsearch: 'Elasticsearch / OpenSearch',
  memcached: 'Memcached',
  cassandra: 'Cassandra / ScyllaDB',
  couchbase: 'Couchbase',
  influxdb: 'InfluxDB',
  neo4j: 'Neo4j',
  kafka: 'Kafka'
};

const KNOWN_DATABASE_TYPES = Object.keys(DATABASE_TYPE_LABELS);

export function resolveDatabaseType(source) {
  if (!source) return '';
  return String(source.metadata?.db_type || source.dbType || source.db_type || '').toLowerCase();
}

export function resolveDatabaseTypeIconKind(databaseType) {
  const type = String(databaseType || '').toLowerCase();
  return KNOWN_DATABASE_TYPES.includes(type) ? type : 'database';
}

export function databaseTypeLabel(databaseType) {
  const type = String(databaseType || '').toLowerCase();
  return DATABASE_TYPE_LABELS[type] || (type ? type.toUpperCase() : 'Database');
}
