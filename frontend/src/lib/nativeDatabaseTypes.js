const nativeTypes = {
  redis: { port: '6379', resourceLabel: '键' },
  mongodb: { port: '27017', resourceLabel: '集合' },
  elasticsearch: { port: '9200', resourceLabel: '索引' }
};

export function isNativeDatabaseType(databaseType) {
  return Object.prototype.hasOwnProperty.call(nativeTypes, databaseType);
}

export function databaseTypeConfig(databaseType) {
  return nativeTypes[databaseType] || null;
}
