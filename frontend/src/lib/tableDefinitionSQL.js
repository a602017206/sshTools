function quoteIdentifier(value, databaseType) {
  const quote = String(databaseType).toLowerCase() === 'mysql' ? '`' : '"';
  return `${quote}${String(value).replaceAll(quote, `${quote}${quote}`)}${quote}`;
}

function qualifiedTableName({ databaseType, databaseName, schemaName, tableName }) {
  const postgreSQL = ['postgresql', 'kingbase'].includes(String(databaseType).toLowerCase());
  const parts = postgreSQL ? [schemaName, tableName] : [databaseName, tableName];
  return parts.filter(Boolean).map(part => quoteIdentifier(part, databaseType)).join('.');
}

function fieldType(field, databaseType) {
  const type = String(field.type || 'VARCHAR').toUpperCase();
  const length = String(field.length || '').trim();
  const postgreSQL = ['postgresql', 'kingbase'].includes(String(databaseType).toLowerCase());
  if (!length || (postgreSQL && /^(BIGINT|INT|INTEGER|SMALLINT|TIMESTAMP|DATE|BOOLEAN|TEXT)$/i.test(type))) return type;
  return /^(VARCHAR|CHAR|DECIMAL|NUMERIC|INT|BIGINT)$/i.test(type) ? `${type}(${length})` : type;
}

export function buildCreateTableSQL({ databaseType, databaseName = '', schemaName = '', tableName, fields = [] }) {
  const dialect = String(databaseType).toLowerCase();
  if (!['mysql', 'postgresql', 'kingbase'].includes(dialect)) return '';
  const validFields = fields.filter(field => String(field?.name || '').trim());
  if (!String(tableName || '').trim() || !validFields.length) return '';

  const definitions = validFields.map(field => `${quoteIdentifier(field.name.trim(), dialect)} ${fieldType(field, dialect)}${field.nullable ? '' : ' NOT NULL'}${String(field.defaultValue || '').trim() ? ` DEFAULT ${String(field.defaultValue).trim()}` : ''}`);
  const primary = validFields.filter(field => field.primary).map(field => quoteIdentifier(field.name.trim(), dialect));
  if (primary.length) definitions.push(`PRIMARY KEY (${primary.join(', ')})`);
  return `CREATE TABLE ${qualifiedTableName({ databaseType: dialect, databaseName, schemaName, tableName: tableName.trim() })} (\n${definitions.map(definition => `  ${definition}`).join(',\n')}\n);`;
}
