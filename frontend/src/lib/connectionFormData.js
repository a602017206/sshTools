const pad = (value) => String(value).padStart(2, '0');

// 连接配置来自 Wails 的 JSON 数据，使用 JSON 深拷贝以兼容不支持 structuredClone 的运行时。
const cloneConfig = (value) => (value ? JSON.parse(JSON.stringify(value)) : {});

function cloneTimestamp(date) {
  return [
    date.getFullYear(),
    pad(date.getMonth() + 1),
    pad(date.getDate()),
    pad(date.getHours()),
    pad(date.getMinutes()),
    pad(date.getSeconds()),
  ].join('');
}

export function createBlankConnectionFormData() {
  return {
    id: '',
    name: '',
    host: '',
    port: '22',
    username: '',
    password: '',
    savePassword: false,
    keyPath: '',
    passphrase: '',
    group: '',
    encoding: 'utf-8',
    dbType: 'mysql',
    driverProfileID: '',
    database: '',
    oracleConnectionMode: 'service',
    sqlServerInstanceName: '',
  };
}

export function createClonedConnectionFormData(source, date = new Date(), credentials = {}) {
  const metadata = cloneConfig(source.metadata);
  const settings = cloneConfig(source.settings);
  const tags = [...(source.tags || [])];

  return {
    ...createBlankConnectionFormData(),
    name: `${source.name || ''} copy ${cloneTimestamp(date)}`,
    host: source.host || '',
    port: source.port?.toString() || '22',
    username: source.username || source.user || '',
    password: credentials.password || '',
    savePassword: Boolean(credentials.savePassword && credentials.password),
    keyPath: source.keyPath || source.key_path || '',
    group: source.group || tags[0] || '',
    encoding: source.encoding || metadata.encoding || 'utf-8',
    dbType: source.dbType || metadata.db_type || 'mysql',
    driverProfileID: source.driverProfileID || metadata.driver_profile_id || '',
    database: source.database || metadata.database || '',
    oracleConnectionMode: source.oracleConnectionMode || (metadata.oracle_connection_mode === 'sid' ? 'sid' : 'service'),
    sqlServerInstanceName: source.sqlServerInstanceName || metadata.sqlserver_instance_name || '',
    type: source.type || 'ssh',
    authType: source.authType || source.auth_type || 'password',
    tags,
    metadata,
    settings,
  };
}
