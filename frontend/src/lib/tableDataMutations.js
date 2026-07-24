function quoteIdentifier(value, databaseType) {
  const quote = String(databaseType).toLowerCase() === 'mysql' ? '`' : '"';
  return `${quote}${String(value).replaceAll(quote, `${quote}${quote}`)}${quote}`;
}

function valueSQL(value) {
  if (value === null || value === undefined) return 'NULL';
  if (typeof value === 'number' || typeof value === 'bigint') return String(value);
  if (typeof value === 'boolean') return value ? 'TRUE' : 'FALSE';
  return `'${String(value).replaceAll("'", "''")}'`;
}

function primaryWhere({ databaseType, columns, row, primaryKeys }) {
  const predicates = primaryKeys.map(key => {
    const index = columns.findIndex(column => String(column).toLowerCase() === String(key).toLowerCase());
    const column = columns[index];
    return index < 0 ? '' : `${quoteIdentifier(column, databaseType)} = ${valueSQL(row[index])}`;
  }).filter(Boolean);
  if (predicates.length) return predicates.join(' AND ');
  const operator = String(databaseType).toLowerCase() === 'mysql' ? '<=>' : 'IS NOT DISTINCT FROM';
  return columns.map((column, index) => `${quoteIdentifier(column, databaseType)} ${operator} ${valueSQL(row[index])}`).join(' AND ');
}

export function buildInsertSQL({ databaseType, table, columns, row }) {
  return `INSERT INTO ${table} (${columns.map(column => quoteIdentifier(column, databaseType)).join(', ')}) VALUES (${row.map(valueSQL).join(', ')});`;
}

export function buildUpdateSQL({ databaseType, table, columns, row, primaryKeys, changes }) {
  const where = primaryWhere({ databaseType, columns, row, primaryKeys });
  const assignments = Object.entries(changes).map(([column, value]) => `${quoteIdentifier(column, databaseType)} = ${valueSQL(value)}`);
  return where && assignments.length ? `UPDATE ${table} SET ${assignments.join(', ')} WHERE ${where};` : '';
}

export function buildDeleteSQL({ databaseType, table, columns, row, primaryKeys }) {
  const where = primaryWhere({ databaseType, columns, row, primaryKeys });
  return where ? `DELETE FROM ${table} WHERE ${where};` : '';
}
