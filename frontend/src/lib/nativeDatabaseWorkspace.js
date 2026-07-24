const workspaceByType = {
  redis: {
    title: 'Redis 键空间',
    resourceLabel: '逻辑库',
    childLabel: '键',
    description: '展开逻辑库可浏览其中的键；当前为只读浏览。',
    canExpand: true
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
    description: '当前可浏览索引；文档查询与编辑尚未提供。',
    canExpand: false
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
    description: '当前可浏览 Topic；消息消费与生产尚未提供。',
    canExpand: false
  }
};

const fallbackWorkspace = {
  title: '原生数据库资源',
  resourceLabel: '资源',
  childLabel: '',
  description: '当前为只读资源浏览。',
  canExpand: false
};

export function nativeDatabaseWorkspace(databaseType) {
  return workspaceByType[String(databaseType || '').toLowerCase()] || fallbackWorkspace;
}
