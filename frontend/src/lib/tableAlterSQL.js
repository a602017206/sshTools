function dialectOf(databaseType) {
  const dialect = String(databaseType).toLowerCase();
  if (dialect === 'kingbase' || dialect === 'opengauss') return 'postgresql';
  if (dialect === 'oracle') return 'oracle';
  return ['mysql', 'postgresql'].includes(dialect) ? dialect : '';
}

function quoteIdentifier(value, dialect) {
  const quote = dialect === 'mysql' ? '`' : '"';
  return `${quote}${String(value).replaceAll(quote, `${quote}${quote}`)}${quote}`;
}

function qualifiedTableName({ databaseType, databaseName = '', schemaName = '', tableName }) {
  const dialect = dialectOf(databaseType);
  const parts = dialect === 'postgresql' || dialect === 'oracle' ? [schemaName, tableName] : [databaseName, tableName];
  return parts.filter(Boolean).map(part => quoteIdentifier(part, dialect)).join('.');
}

function sqlLiteral(value) {
  return `'${String(value).replaceAll("'", "''")}'`;
}

function fieldType(field, dialect) {
  const type = String(field.type || 'VARCHAR').trim().toUpperCase();
  const length = String(field.length || '').trim();
  const skipLength = (dialect === 'postgresql' || dialect === 'oracle') && /^(BIGINT|INT|INTEGER|SMALLINT|TIMESTAMP|DATE|BOOLEAN|TEXT|CLOB|BLOB)$/i.test(type);
  if (!length || skipLength) return type;
  return /^(VARCHAR2?|CHAR|NVARCHAR2?|DECIMAL|NUMERIC|NUMBER|INT|INTEGER|BIGINT|SMALLINT)$/i.test(type) ? `${type}(${length})` : type;
}

function defaultClause(field) {
  const value = String(field.defaultValue || '').trim();
  return value ? ` DEFAULT ${value}` : '';
}

function fieldDefinition(field, dialect, { includeComment = false } = {}) {
  const name = quoteIdentifier(String(field.name || '').trim(), dialect);
  const notNull = field.nullable ? '' : ' NOT NULL';
  const comment = dialect === 'mysql' && includeComment ? ` COMMENT ${sqlLiteral(field.comment || '')}` : '';
  return `${name} ${fieldType(field, dialect)}${notNull}${defaultClause(field)}${comment}`;
}

function same(valueA, valueB) {
  return String(valueA || '').trim().toUpperCase() === String(valueB || '').trim().toUpperCase();
}

function fieldChanged(original, field) {
  return !same(original.type, field.type)
    || !same(original.length, field.length)
    || Boolean(original.nullable) !== Boolean(field.nullable)
    || String(original.defaultValue || '').trim() !== String(field.defaultValue || '').trim()
    || String(original.comment || '').trim() !== String(field.comment || '').trim();
}

function currentPrimaryNames(originalFields, fields) {
  return originalFields
    .filter(field => field.primary)
    .map(field => fields.find(candidate => candidate._originalName === field._originalName)?.name || field.name)
    .filter(Boolean);
}

function sameNameSet(left, right) {
  return left.length === right.length && left.every((name, index) => String(name).toLowerCase() === String(right[index]).toLowerCase());
}

function appendAddedColumn(statements, table, field, dialect) {
  if (dialect === 'oracle') {
    statements.push(`ALTER TABLE ${table} ADD (${fieldDefinition(field, dialect)});`);
  } else {
    statements.push(`ALTER TABLE ${table} ADD COLUMN ${fieldDefinition(field, dialect)};`);
  }
  if ((dialect === 'postgresql' || dialect === 'oracle') && String(field.comment || '').trim()) {
    statements.push(`COMMENT ON COLUMN ${table}.${quoteIdentifier(field.name.trim(), dialect)} IS ${sqlLiteral(field.comment)};`);
  }
}

function appendMySQLColumnChanges(statements, table, originalName, original, field, dialect) {
  const renamed = originalName !== String(field.name || '').trim();
  if (!renamed && !fieldChanged(original, field)) return;
  const keyword = renamed ? 'CHANGE COLUMN' : 'MODIFY COLUMN';
  const source = renamed ? `${quoteIdentifier(originalName, dialect)} ` : '';
  const commentChanged = String(original.comment || '').trim() !== String(field.comment || '').trim();
  const includeComment = commentChanged || Boolean(String(field.comment || '').trim());
  statements.push(`ALTER TABLE ${table} ${keyword} ${source}${fieldDefinition(field, dialect, { includeComment })};`);
}

function appendOracleColumnChanges(statements, table, originalName, original, field, dialect) {
  const name = String(field.name || '').trim();
  if (originalName !== name) {
    statements.push(`ALTER TABLE ${table} RENAME COLUMN ${quoteIdentifier(originalName, dialect)} TO ${quoteIdentifier(name, dialect)};`);
  }
  if (!same(original.type, field.type) || !same(original.length, field.length) || Boolean(original.nullable) !== Boolean(field.nullable) || String(original.defaultValue || '').trim() !== String(field.defaultValue || '').trim()) {
    statements.push(`ALTER TABLE ${table} MODIFY (${fieldDefinition({ ...field, name }, dialect)});`);
  }
  if (String(original.comment || '').trim() !== String(field.comment || '').trim()) {
    statements.push(`COMMENT ON COLUMN ${table}.${quoteIdentifier(name, dialect)} IS ${String(field.comment || '').trim() ? sqlLiteral(field.comment) : 'NULL'};`);
  }
}

function appendPostgreSQLColumnChanges(statements, table, originalName, original, field, dialect) {
  const name = String(field.name || '').trim();
  if (originalName !== name) {
    statements.push(`ALTER TABLE ${table} RENAME COLUMN ${quoteIdentifier(originalName, dialect)} TO ${quoteIdentifier(name, dialect)};`);
  }
  const column = quoteIdentifier(name, dialect);
  if (!same(original.type, field.type) || !same(original.length, field.length)) statements.push(`ALTER TABLE ${table} ALTER COLUMN ${column} TYPE ${fieldType(field, dialect)};`);
  if (Boolean(original.nullable) !== Boolean(field.nullable)) statements.push(`ALTER TABLE ${table} ALTER COLUMN ${column} ${field.nullable ? 'DROP NOT NULL' : 'SET NOT NULL'};`);
  if (String(original.defaultValue || '').trim() !== String(field.defaultValue || '').trim()) statements.push(`ALTER TABLE ${table} ALTER COLUMN ${column} ${String(field.defaultValue || '').trim() ? `SET DEFAULT ${String(field.defaultValue).trim()}` : 'DROP DEFAULT'};`);
  if (String(original.comment || '').trim() !== String(field.comment || '').trim()) statements.push(`COMMENT ON COLUMN ${table}.${column} IS ${String(field.comment || '').trim() ? sqlLiteral(field.comment) : 'NULL'};`);
}

function appendPrimaryKeyChanges(statements, table, tableName, dialect, originalFields, fields) {
  const originalPrimary = currentPrimaryNames(originalFields, fields);
  const nextPrimary = fields.filter(field => field.primary && String(field.name || '').trim()).map(field => field.name.trim());
  if (sameNameSet(originalPrimary, nextPrimary)) return;
  if (dialect === 'mysql' || dialect === 'oracle') {
    if (originalPrimary.length) statements.push(`ALTER TABLE ${table} DROP PRIMARY KEY;`);
  } else if (originalPrimary.length) {
    statements.push(`ALTER TABLE ${table} DROP CONSTRAINT IF EXISTS ${quoteIdentifier(`${tableName}_pkey`, dialect)};`);
  }
  if (nextPrimary.length) statements.push(`ALTER TABLE ${table} ADD PRIMARY KEY (${nextPrimary.map(name => quoteIdentifier(name, dialect)).join(', ')});`);
}

export function buildAlterTableStatements({ databaseType, databaseName = '', schemaName = '', tableName, originalFields = [], fields = [] }) {
  const dialect = dialectOf(databaseType);
  if (!dialect || !String(tableName || '').trim()) return [];

  const table = qualifiedTableName({ databaseType: dialect, databaseName, schemaName, tableName: tableName.trim() });
  const statements = [];
  const originalsByName = new Map(originalFields.map(field => [field._originalName || field.name, field]));
  const retainedOriginalNames = new Set();

  for (const field of fields.filter(field => String(field?.name || '').trim())) {
    const originalName = String(field._originalName || '').trim();
    const original = originalsByName.get(originalName);
    if (!original) {
      appendAddedColumn(statements, table, field, dialect);
      continue;
    }

    retainedOriginalNames.add(originalName);
    if (dialect === 'mysql') appendMySQLColumnChanges(statements, table, originalName, original, field, dialect);
    if (dialect === 'postgresql') {
      appendPostgreSQLColumnChanges(statements, table, originalName, original, field, dialect);
    }
    if (dialect === 'oracle') {
      appendOracleColumnChanges(statements, table, originalName, original, field, dialect);
    }
  }

  for (const original of originalFields) {
    const originalName = String(original._originalName || original.name || '').trim();
    if (originalName && !retainedOriginalNames.has(originalName)) statements.push(`ALTER TABLE ${table} DROP COLUMN ${quoteIdentifier(originalName, dialect)};`);
  }

  appendPrimaryKeyChanges(statements, table, tableName, dialect, originalFields, fields);

  return statements;
}
