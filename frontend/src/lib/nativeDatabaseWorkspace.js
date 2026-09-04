import { isNativeDatabaseType } from './nativeDatabaseTypes.js';

const workspaceByType = {
  redis: {
    title: 'Redis 键空间',
    resourceLabel: '逻辑库',
    childLabel: '键',
    description: '选择逻辑库后浏览键；支持 string/hash/list/set/zset 预览与编辑。',
    dbSelector: 'dropdown',
    canExpand: false,
    canDescribe: true,
    canWrite: true,
    canDelete: true
  },
  mongodb: {
    title: 'MongoDB 数据库',
    resourceLabel: '数据库',
    childLabel: '集合',
    description: '展开数据库可浏览其中的集合；当前为只读浏览。',
    canExpand: true
  },
  elasticsearch: {
    title: 'Elasticsearch 索引',
    resourceLabel: '索引',
    childLabel: '',
    description: '选择索引查看 mapping 与数据量；查询页签可执行 DSL。',
    canExpand: false,
    canDescribe: true,
    canQuery: true,
    canWrite: true,
    canDelete: true,
    canSearchResources: true,
    canResizeInspector: true,
    showSessionOverview: true
  },
  memcached: {
    title: 'Memcached 统计',
    resourceLabel: '统计项',
    childLabel: '',
    description: '当前显示服务统计项；不提供键枚举或写入操作。',
    canExpand: false
  },
  cassandra: {
    title: 'Cassandra 键空间',
    resourceLabel: '键空间',
    childLabel: '表',
    description: '展开键空间可浏览其中的表；当前为只读浏览。',
    canExpand: true
  },
  couchbase: {
    title: 'Couchbase Bucket',
    resourceLabel: 'Bucket',
    childLabel: '集合',
    description: '展开 Bucket 可浏览 Scope 与集合；当前为只读浏览。',
    canExpand: true
  },
  influxdb: {
    title: 'InfluxDB Bucket',
    resourceLabel: 'Bucket',
    childLabel: '',
    description: '当前可浏览 Bucket；数据查询与编辑尚未提供。',
    canExpand: false
  },
  neo4j: {
    title: 'Neo4j 数据库',
    resourceLabel: '数据库',
    childLabel: '',
    description: '当前可浏览数据库；图查询与编辑尚未提供。',
    canExpand: false
  },
  kafka: {
    title: 'Kafka Topic',
    resourceLabel: 'Topic',
    childLabel: '',
    description: '可查看 Topic 分区元数据；不消费或生产消息。',
    canExpand: false,
    canDescribe: true
  },
  rocketmq: {
    title: 'RocketMQ Topic',
    resourceLabel: 'Topic',
    childLabel: '',
    description: '通过 NameServer 浏览 Topic；不生产或消费消息。',
    canExpand: false,
    canDescribe: true
  },
  rabbitmq: {
    title: 'RabbitMQ 队列',
    resourceLabel: 'Queue',
    childLabel: '',
    description: '浏览 Queue 与基础指标；不生产或消费消息。',
    canExpand: false,
    canDescribe: true
  }
};

const fallbackWorkspace = {
  title: '原生数据库资源',
  resourceLabel: '资源',
  childLabel: '',
  description: '当前为只读资源浏览。',
  canExpand: false,
  canDescribe: false
};

export function nativeDatabaseWorkspace(databaseType) {
  const type = String(databaseType || '').toLowerCase();
  if (type === 'opensearch') return workspaceByType.elasticsearch;
  return workspaceByType[type] || fallbackWorkspace;
}

/** @returns {'redis'|'elasticsearch'|'kafka'|'generic'} */
export function resolveNativeWorkspaceKind(databaseType) {
  const type = String(databaseType || '').toLowerCase();
  if (type === 'redis') return 'redis';
  if (type === 'elasticsearch' || type === 'opensearch') return 'elasticsearch';
  if (type === 'kafka' || type === 'rocketmq' || type === 'rabbitmq') return 'kafka';
  return 'generic';
}

/** 会话标签文案：原生类型用工作区标题，避免一律写成「数据库」。 */
export function databaseSessionTabLabel(asset) {
  const name = String(asset?.name || '').trim() || '连接';
  const dbType = String(asset?.metadata?.db_type || asset?.dbType || asset?.db_type || '').toLowerCase();
  if (isNativeDatabaseType(dbType)) {
    return `${name} · ${nativeDatabaseWorkspace(dbType).title}`;
  }
  return `${name} · 数据库`;
}
