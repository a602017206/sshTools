function quoteIdentifier(value, databaseType) {
  const quote = String(databaseType).toLowerCase() === 'mysql' ? '`' : '"';
  return `${quote}${String(value).replaceAll(quote, `${quote}${quote}`)}${quote}`;
}

function isSchemaScoped(databaseType) {
  return ['postgresql', 'kingbase', 'opengauss', 'oracle'].includes(String(databaseType).toLowerCase());
}

function qualifiedTableName({ databaseType, databaseName, schemaName, tableName }) {
  const parts = isSchemaScoped(databaseType) ? [schemaName, tableName] : [databaseName, tableName];
  return parts.filter(Boolean).map(part => quoteIdentifier(part, databaseType)).join('.');
}

function fieldType(field, databaseType) {
  let type = String(field.type || 'VARCHAR').toUpperCase();
  const length = String(field.length || '').trim();
  const dialect = String(databaseType).toLowerCase();
  if (dialect === 'oracle') {
    if (type === 'VARCHAR' || type === 'NVARCHAR') type = 'VARCHAR2';
    if (type === 'TEXT') type = 'CLOB';
    if (/^(BIGINT|INT|INTEGER|SMALLINT|BOOLEAN|DECIMAL|NUMERIC)$/i.test(type)) type = 'NUMBER';
  }
  const skipLength = (['postgresql', 'kingbase', 'opengauss'].includes(dialect)
    && /^(BIGINT|INT|INTEGER|SMALLINT|TIMESTAMP|DATE|BOOLEAN|TEXT)$/i.test(type))
    || (dialect === 'oracle' && /^(TIMESTAMP|DATE|CLOB|BLOB)$/i.test(type));
  if (!length || skipLength) return type;
  return /^(VARCHAR2?|CHAR|NVARCHAR2?|DECIMAL|NUMERIC|NUMBER|INT|INTEGER|BIGINT|SMALLINT)$/i.test(type) ? `${type}(${length})` : type;
}

export function buildCreateTableSQL({ databaseType, databaseName = '', schemaName = '', tableName, fields = [] }) {
  const dialect = String(databaseType).toLowerCase();
  if (!['mysql', 'postgresql', 'kingbase', 'oracle'].includes(dialect)) return '';
  const validFields = fields.filter(field => String(field?.name || '').trim());
  if (!String(tableName || '').trim() || !validFields.length) return '';

  const definitions = validFields.map(field => `${quoteIdentifier(field.name.trim(), dialect)} ${fieldType(field, dialect)}${field.nullable ? '' : ' NOT NULL'}${String(field.defaultValue || '').trim() ? ` DEFAULT ${String(field.defaultValue).trim()}` : ''}`);
  const primary = validFields.filter(field => field.primary).map(field => quoteIdentifier(field.name.trim(), dialect));
  if (primary.length) definitions.push(`PRIMARY KEY (${primary.join(', ')})`);
  return `CREATE TABLE ${qualifiedTableName({ databaseType: dialect, databaseName, schemaName, tableName: tableName.trim() })} (\n${definitions.map(definition => `  ${definition}`).join(',\n')}\n);`;
}
