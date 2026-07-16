const postgreSQLCategories = [
  { id: 'tables', label: '表', icon: '▦' },
  { id: 'views', label: '视图', icon: '◉' },
  { id: 'materialized_views', label: '物化视图', icon: '◉' },
  { id: 'procedures', label: '存储过程', icon: '▤' },
  { id: 'functions', label: '函数', icon: 'ƒ' },
  { id: 'extensions', label: '扩展', icon: '◇' }
];

const mySQLCategories = [
  { id: 'tables', label: '表', icon: '▦' },
  { id: 'views', label: '视图', icon: '◉' },
  { id: 'procedures', label: '存储过程', icon: '▤' },
  { id: 'functions', label: '函数', icon: 'ƒ' },
  { id: 'events', label: '事件', icon: '◷' }
];

export function isPostgreSQLCompatible(databaseType) {
  return ['postgresql', 'kingbase', 'opengauss'].includes(String(databaseType || '').toLowerCase());
}

export function databaseObjectCategories(databaseType) {
  const normalizedType = String(databaseType || '').toLowerCase();
  if (isPostgreSQLCompatible(normalizedType)) return postgreSQLCategories;
  if (normalizedType === 'mysql') return mySQLCategories;
  return [{ id: 'tables', label: '表', icon: '▦' }];
}

export function buildPostgreSQLSchemaQuery() {
  return `
    SELECT nspname
    FROM pg_catalog.pg_namespace
    WHERE nspname NOT LIKE 'pg_toast%'
    ORDER BY nspname
  `;
}

function escapeSqlLiteral(value) {
  return String(value || '').replace(/'/g, "''");
}

export function buildPostgreSQLObjectQuery(schema, category) {
  const escapedSchema = escapeSqlLiteral(schema);
  const queries = {
    tables: `SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = '${escapedSchema}' ORDER BY tablename`,
    views: `SELECT viewname FROM pg_catalog.pg_views WHERE schemaname = '${escapedSchema}' ORDER BY viewname`,
    materialized_views: `SELECT matviewname FROM pg_catalog.pg_matviews WHERE schemaname = '${escapedSchema}' ORDER BY matviewname`,
    procedures: `SELECT proname FROM pg_catalog.pg_proc p JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname = '${escapedSchema}' AND p.prokind = 'p' ORDER BY proname`,
    functions: `SELECT proname FROM pg_catalog.pg_proc p JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace WHERE n.nspname = '${escapedSchema}' AND p.prokind = 'f' ORDER BY proname`,
    extensions: `SELECT extname FROM pg_catalog.pg_extension ORDER BY extname`
  };
  return queries[category] || queries.tables;
}

export function buildMySQLObjectQuery(database, category) {
  const escapedDatabase = escapeSqlLiteral(database);
  const queries = {
    tables: `SELECT TABLE_NAME FROM information_schema.tables WHERE TABLE_SCHEMA = '${escapedDatabase}' AND TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_NAME`,
    views: `SELECT TABLE_NAME FROM information_schema.views WHERE TABLE_SCHEMA = '${escapedDatabase}' ORDER BY TABLE_NAME`,
    procedures: `SELECT ROUTINE_NAME FROM information_schema.routines WHERE ROUTINE_SCHEMA = '${escapedDatabase}' AND ROUTINE_TYPE = 'PROCEDURE' ORDER BY ROUTINE_NAME`,
    functions: `SELECT ROUTINE_NAME FROM information_schema.routines WHERE ROUTINE_SCHEMA = '${escapedDatabase}' AND ROUTINE_TYPE = 'FUNCTION' ORDER BY ROUTINE_NAME`,
    events: `SELECT EVENT_NAME FROM information_schema.events WHERE EVENT_SCHEMA = '${escapedDatabase}' ORDER BY EVENT_NAME`
  };
  return queries[category] || queries.tables;
}

export function queryFirstColumn(resultJSON) {
  const result = JSON.parse(resultJSON || '{}');
  return (result.rows || []).map(row => String(row?.[0] || '')).filter(Boolean);
}
