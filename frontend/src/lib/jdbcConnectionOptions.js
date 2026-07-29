export function buildJDBCConnectionOptions(databaseType, database, oracleConnectionMode = 'service', sqlServerInstanceName = '') {
  const type = String(databaseType || '').toLowerCase();
  const properties = {};

  if (type === 'oracle') {
    properties.oracleConnectionMode = oracleConnectionMode === 'sid' ? 'sid' : 'service';
  }
  if (type === 'sqlserver' && String(sqlServerInstanceName || '').trim()) {
    properties.instanceName = String(sqlServerInstanceName).trim();
  }

  return { database: String(database || '').trim(), properties };
}
