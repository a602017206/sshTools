const MYSQL_TYPES = [
  'TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'BIGINT', 'DECIMAL', 'FLOAT', 'DOUBLE', 'BIT', 'BOOLEAN',
  'CHAR', 'VARCHAR', 'TINYTEXT', 'TEXT', 'MEDIUMTEXT', 'LONGTEXT',
  'BINARY', 'VARBINARY', 'TINYBLOB', 'BLOB', 'MEDIUMBLOB', 'LONGBLOB',
  'DATE', 'TIME', 'DATETIME', 'TIMESTAMP', 'YEAR', 'JSON'
];

const POSTGRESQL_TYPES = [
  'SMALLINT', 'INTEGER', 'BIGINT', 'SERIAL', 'BIGSERIAL', 'NUMERIC', 'DECIMAL', 'REAL', 'DOUBLE PRECISION', 'BOOLEAN',
  'CHAR', 'VARCHAR', 'TEXT', 'BYTEA',
  'DATE', 'TIME', 'TIMESTAMP', 'TIMESTAMPTZ', 'INTERVAL', 'JSON', 'JSONB', 'UUID'
];

const ORACLE_TYPES = [
  'NUMBER', 'INTEGER', 'FLOAT', 'BINARY_FLOAT', 'BINARY_DOUBLE',
  'CHAR', 'VARCHAR2', 'NCHAR', 'NVARCHAR2', 'CLOB', 'NCLOB', 'BLOB', 'RAW',
  'DATE', 'TIMESTAMP'
];

const GENERIC_TYPES = ['BIGINT', 'INT', 'NUMBER', 'VARCHAR', 'VARCHAR2', 'TEXT', 'CLOB', 'DECIMAL', 'TIMESTAMP', 'DATE', 'BOOLEAN'];

const LENGTH_TYPE_PATTERN = /^(VARCHAR2?|CHAR|NCHAR|NVARCHAR2?|DECIMAL|NUMERIC|NUMBER|INT|INTEGER|BIGINT|SMALLINT|TINYINT|MEDIUMINT|BIT|BINARY|VARBINARY|RAW)$/i;

function baseTypes(databaseType) {
  switch (String(databaseType || '').toLowerCase()) {
    case 'mysql':
      return MYSQL_TYPES;
    case 'postgresql':
    case 'kingbase':
    case 'opengauss':
      return POSTGRESQL_TYPES;
    case 'oracle':
      return ORACLE_TYPES;
    default:
      return GENERIC_TYPES;
  }
}

export function columnTypeOptions(databaseType, currentType = '') {
  const types = baseTypes(databaseType);
  const current = String(currentType || '').trim().toUpperCase();
  return current && !types.includes(current) ? [current, ...types] : types;
}

export function typeAcceptsLength(type) {
  return LENGTH_TYPE_PATTERN.test(String(type || '').trim());
}
