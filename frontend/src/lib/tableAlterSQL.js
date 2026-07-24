function dialectOf(databaseType) {
  const dialect = String(databaseType).toLowerCase();
  if (dialect === 'kingbase') return 'postgresql';
  return ['mysql', 'postgresql'].includes(dialect) ? dialect : '';
}

function quoteIdentifier(value, dialect) {
  const quote = dialect === 'mysql' ? '`' : '"';
  return `${quote}${String(value).replaceAll(quote, `${quote}${quote}`)}${quote}`;
}

function qualifiedTableName({ databaseType, databaseName = '', schemaName = '', tableName }) {
  const dialect = dialectOf(databaseType);
  const parts = dialect === 'postgresql' ? [schemaName, tableName] : [databaseName, tableName];
  return parts.filter(Boolean).map(part => quoteIdentifier(part, dialect)).join('.');
}

function sqlLiteral(value) {
  return `'${String(value).replaceAll("'", "''")}'`;
}

function fieldType(field, dialect) {
  const type = String(field.type || 'VARCHAR').trim().toUpperCase();
  const length = String(field.length || '').trim();
  if (!length || (dialect === 'postgresql' && /^(BIGINT|INT|INTEGER|SMALLINT|TIMESTAMP|DATE|BOOLEAN|TEXT)$/i.test(type))) return type;
  return /^(VARCHAR|CHAR|DECIMAL|NUMERIC|INT|INTEGER|BIGINT|SMALLINT)$/i.test(type) ? `${type}(${length})` : type;
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
      statements.push(`ALTER TABLE ${table} ADD COLUMN ${fieldDefinition(field, dialect)};`);
      if (dialect === 'postgresql' && String(field.comment || '').trim()) {
        statements.push(`COMMENT ON COLUMN ${table}.${quoteIdentifier(field.name.trim(), dialect)} IS ${sqlLiteral(field.comment)};`);
      }
      continue;
    }

    retainedOriginalNames.add(originalName);
    const renamed = originalName !== String(field.name || '').trim();
    const changed = fieldChanged(original, field);
    if (dialect === 'mysql' && (renamed || changed)) {
      const keyword = renamed ? 'CHANGE COLUMN' : 'MODIFY COLUMN';
      const source = renamed ? `${quoteIdentifier(originalName, dialect)} ` : '';
      const includeComment = String(original.comment || '').trim() !== String(field.comment || '').trim() || Boolean(String(field.comment || '').trim());
      statements.push(`ALTER TABLE ${table} ${keyword} ${source}${fieldDefinition(field, dialect, { includeComment })};`);
    }
    if (dialect === 'postgresql') {
      if (renamed) statements.push(`ALTER TABLE ${table} RENAME COLUMN ${quoteIdentifier(originalName, dialect)} TO ${quoteIdentifier(field.name.trim(), dialect)};`);
      const column = quoteIdentifier(field.name.trim(), dialect);
      if (!same(original.type, field.type) || !same(original.length, field.length)) statements.push(`ALTER TABLE ${table} ALTER COLUMN ${column} TYPE ${fieldType(field, dialect)};`);
      if (Boolean(original.nullable) !== Boolean(field.nullable)) statements.push(`ALTER TABLE ${table} ALTER COLUMN ${column} ${field.nullable ? 'DROP NOT NULL' : 'SET NOT NULL'};`);
      if (String(original.defaultValue || '').trim() !== String(field.defaultValue || '').trim()) statements.push(`ALTER TABLE ${table} ALTER COLUMN ${column} ${String(field.defaultValue || '').trim() ? `SET DEFAULT ${String(field.defaultValue).trim()}` : 'DROP DEFAULT'};`);
      if (String(original.comment || '').trim() !== String(field.comment || '').trim()) statements.push(`COMMENT ON COLUMN ${table}.${column} IS ${String(field.comment || '').trim() ? sqlLiteral(field.comment) : 'NULL'};`);
    }
  }

  for (const original of originalFields) {
    const originalName = String(original._originalName || original.name || '').trim();
    if (originalName && !retainedOriginalNames.has(originalName)) statements.push(`ALTER TABLE ${table} DROP COLUMN ${quoteIdentifier(originalName, dialect)};`);
  }

  const originalPrimary = currentPrimaryNames(originalFields, fields);
  const nextPrimary = fields.filter(field => field.primary && String(field.name || '').trim()).map(field => field.name.trim());
  if (!sameNameSet(originalPrimary, nextPrimary)) {
    if (dialect === 'mysql') {
      if (originalPrimary.length) statements.push(`ALTER TABLE ${table} DROP PRIMARY KEY;`);
      if (nextPrimary.length) statements.push(`ALTER TABLE ${table} ADD PRIMARY KEY (${nextPrimary.map(name => quoteIdentifier(name, dialect)).join(', ')});`);
    } else {
      if (originalPrimary.length) statements.push(`ALTER TABLE ${table} DROP CONSTRAINT IF EXISTS ${quoteIdentifier(`${tableName}_pkey`, dialect)};`);
      if (nextPrimary.length) statements.push(`ALTER TABLE ${table} ADD PRIMARY KEY (${nextPrimary.map(name => quoteIdentifier(name, dialect)).join(', ')});`);
    }
  }

  return statements;
}
