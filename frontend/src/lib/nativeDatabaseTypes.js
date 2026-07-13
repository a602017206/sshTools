const nativeTypes = {
  redis: { port: '6379', resourceLabel: '键' },
  mongodb: { port: '27017', resourceLabel: '集合' },
  elasticsearch: { port: '9200', resourceLabel: '索引' },
  memcached: { port: '11211', resourceLabel: '统计项' },
  cassandra: { port: '9042', resourceLabel: '表' },
  couchbase: { port: '8091', resourceLabel: '集合' },
  influxdb: { port: '8086', resourceLabel: '资源' }
};

export function isNativeDatabaseType(databaseType) {
  return Object.prototype.hasOwnProperty.call(nativeTypes, databaseType);
}

export function databaseTypeConfig(databaseType) {
  return nativeTypes[databaseType] || null;
}
