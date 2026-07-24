function quoteIdentifier(value, databaseType) {
  const quote = String(databaseType).toLowerCase() === 'mysql' ? '`' : '"';
  return `${quote}${String(value).replaceAll(quote, `${quote}${quote}`)}${quote}`;
}

function qualifiedTableName({ databaseType, databaseName = '', schemaName = '', tableName }) {
  const dialect = String(databaseType).toLowerCase();
  const parts = ['postgresql', 'kingbase'].includes(dialect) ? [schemaName, tableName] : [databaseName, tableName];
  return parts.filter(Boolean).map(part => quoteIdentifier(part, dialect)).join('.');
}

function supported(databaseType) {
  return ['mysql', 'postgresql', 'kingbase'].includes(String(databaseType).toLowerCase());
}

export function buildDropTableSQL({ databaseType, databaseName = '', schemaName = '', tableName }) {
  if (!supported(databaseType) || !String(tableName || '').trim()) return '';
  return `DROP TABLE ${qualifiedTableName({ databaseType, databaseName, schemaName, tableName: tableName.trim() })};`;
}

export function buildCopyTableStatements({ databaseType, databaseName = '', schemaName = '', sourceTable, targetTable, includeData = false }) {
  if (!supported(databaseType) || !String(sourceTable || '').trim() || !String(targetTable || '').trim()) return [];
  const dialect = String(databaseType).toLowerCase();
  const source = qualifiedTableName({ databaseType: dialect, databaseName, schemaName, tableName: sourceTable.trim() });
  const target = qualifiedTableName({ databaseType: dialect, databaseName, schemaName, tableName: targetTable.trim() });
  const create = dialect === 'mysql'
    ? `CREATE TABLE ${target} LIKE ${source};`
    : `CREATE TABLE ${target} (LIKE ${source} INCLUDING ALL);`;
  return includeData ? [create, `INSERT INTO ${target} SELECT * FROM ${source};`] : [create];
}
