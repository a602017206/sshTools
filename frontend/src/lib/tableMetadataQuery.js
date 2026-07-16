function escapeSqlLiteral(value) {
  return String(value || '').replace(/'/g, "''");
}

export function buildTableMetadataQuery(databaseType, databaseName) {
  const dbType = String(databaseType || '').toLowerCase();
  const escapedDatabase = escapeSqlLiteral(databaseName);

  if (dbType === 'postgresql' || dbType === 'kingbase') {
    return `
      SELECT
        c.relname AS table_name,
        c.reltuples::bigint AS table_rows,
        pg_total_relation_size(c.oid) AS data_length,
        '' AS engine,
        NULL::text AS create_time,
        NULL::text AS update_time,
        pg_catalog.obj_description(c.oid, 'pg_class') AS table_comment,
        '' AS table_collation
      FROM pg_catalog.pg_class c
      JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
      WHERE c.relkind = 'r'
        AND n.nspname = 'public'
      ORDER BY c.relname;
    `;
  }

  return `
    SELECT
      TABLE_NAME AS table_name,
      TABLE_ROWS AS table_rows,
      DATA_LENGTH AS data_length,
      ENGINE AS engine,
      CREATE_TIME AS create_time,
      UPDATE_TIME AS update_time,
      TABLE_COLLATION AS table_collation,
      TABLE_COMMENT AS table_comment
    FROM information_schema.tables
    WHERE TABLE_SCHEMA = '${escapedDatabase}'
    ORDER BY TABLE_NAME;
  `;
}

export function tableMetadataFromNames(tableNames) {
  return (tableNames || []).map(tableName => ({
    tableName: String(tableName || ''),
    rowCount: null,
    dataLength: null,
    engine: '-',
    createTime: '-',
    updateTime: '-',
    collation: '-',
    comment: '-'
  })).filter(table => table.tableName);
}
