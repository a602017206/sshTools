const nullOperations = new Set(['is_null', 'is_not_null']);
const listOperations = new Set(['in', 'not_in']);
const schemaScopedTypes = new Set(['postgresql', 'kingbase', 'opengauss', 'oracle', 'sqlserver', 'dm']);

export const tableFilterOperations = [
  { value: 'contains', label: '包含' },
  { value: 'not_contains', label: '不包含' },
  { value: 'equals', label: '等于' },
  { value: 'not_equals', label: '不等于' },
  { value: 'greater_than', label: '大于' },
  { value: 'greater_or_equal', label: '大于等于' },
  { value: 'less_than', label: '小于' },
  { value: 'less_or_equal', label: '小于等于' },
  { value: 'is_null', label: '为 NULL' },
  { value: 'is_not_null', label: '不为 NULL' },
  { value: 'in', label: '在列表中' },
  { value: 'not_in', label: '不在列表中' }
];

export function operationNeedsValue(operation) {
  return !nullOperations.has(operation);
}

export function operationUsesList(operation) {
  return listOperations.has(operation);
}

function normalizeDatabaseType(databaseType) {
  return String(databaseType || '').toLowerCase();
}

function quoteIdentifier(identifier, databaseType) {
  const dialect = normalizeDatabaseType(databaseType);
  if (dialect === 'mysql') {
    return `\`${String(identifier).replaceAll('`', '``')}\``;
  }
  if (dialect === 'sqlserver') {
    return `[${String(identifier).replaceAll(']', ']]')}]`;
  }
  return `"${String(identifier).replaceAll('"', '""')}"`;
}

export function buildQualifiedTableName({ databaseType, databaseName = '', schemaName = '', tableName = '' } = {}) {
  if (!tableName) return '';
  const dialect = normalizeDatabaseType(databaseType);
  const parts = schemaScopedTypes.has(dialect) ? [schemaName, tableName] : [databaseName, tableName];
  return parts.filter(Boolean).map(part => quoteIdentifier(part, dialect)).join('.');
}

function quoteLiteral(value) {
  return `'${String(value).replaceAll("'", "''")}'`;
}

function listValues(value) {
  return String(value || '').split(/[\n,]/).map(item => item.trim()).filter(Boolean);
}

function buildPredicate(rule, databaseType) {
  if (!rule?.field || !rule.operation) return '';
  if (operationNeedsValue(rule.operation) && !String(rule.value ?? '').trim()) return '';

  const field = quoteIdentifier(rule.field, databaseType);
  const value = String(rule.value ?? '');
  const operatorMap = {
    contains: `${field} LIKE ${quoteLiteral(`%${value}%`)}`,
    not_contains: `${field} NOT LIKE ${quoteLiteral(`%${value}%`)}`,
    equals: `${field} = ${quoteLiteral(value)}`,
    not_equals: `${field} <> ${quoteLiteral(value)}`,
    greater_than: `${field} > ${quoteLiteral(value)}`,
    greater_or_equal: `${field} >= ${quoteLiteral(value)}`,
    less_than: `${field} < ${quoteLiteral(value)}`,
    less_or_equal: `${field} <= ${quoteLiteral(value)}`,
    is_null: `${field} IS NULL`,
    is_not_null: `${field} IS NOT NULL`
  };
  if (listOperations.has(rule.operation)) {
    const values = listValues(value);
    if (!values.length) return '';
    return `${field} ${rule.operation === 'in' ? 'IN' : 'NOT IN'} (${values.map(quoteLiteral).join(', ')})`;
  }
  return operatorMap[rule.operation] || '';
}

function buildLimitClause(databaseType, limit, offset, hasOrderBy) {
  const dialect = normalizeDatabaseType(databaseType);
  const normalizedLimit = Number.isInteger(limit) && limit > 0 ? limit : 100;
  const normalizedOffset = Number.isInteger(offset) && offset > 0 ? offset : 0;

  if (dialect === 'oracle' || dialect === 'dm') {
    if (normalizedOffset > 0) {
      return ` OFFSET ${normalizedOffset} ROWS FETCH NEXT ${normalizedLimit} ROWS ONLY`;
    }
    return ` FETCH FIRST ${normalizedLimit} ROWS ONLY`;
  }

  if (dialect === 'sqlserver') {
    const orderPrefix = hasOrderBy ? '' : ' ORDER BY (SELECT NULL)';
    return `${orderPrefix} OFFSET ${normalizedOffset} ROWS FETCH NEXT ${normalizedLimit} ROWS ONLY`;
  }

  return ` LIMIT ${normalizedLimit}${normalizedOffset ? ` OFFSET ${normalizedOffset}` : ''}`;
}

export function buildTableBrowseSQL({ fromSQL, databaseType, filters = [], sorters = [], limit = 100, offset = 0 }) {
  if (!fromSQL) return '';
  const predicates = filters.map(rule => ({ rule, predicate: buildPredicate(rule, databaseType) })).filter(item => item.predicate);
  const whereClause = predicates.length
    ? ` WHERE ${predicates.map((item, index) => `${index ? (item.rule.connector === 'OR' ? 'OR' : 'AND') : ''} ${item.predicate}`.trim()).join(' ')}`
    : '';
  const orderItems = sorters.filter(item => item?.field).map(item => `${quoteIdentifier(item.field, databaseType)} ${item.direction === 'DESC' ? 'DESC' : 'ASC'}`);
  const orderClause = orderItems.length ? ` ORDER BY ${orderItems.join(', ')}` : '';
  const limitClause = buildLimitClause(databaseType, limit, offset, orderItems.length > 0);
  return `SELECT * FROM ${fromSQL}${whereClause}${orderClause}${limitClause};`;
}
