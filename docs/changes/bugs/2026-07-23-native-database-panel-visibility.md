# 原生数据库面板可见性与资源浏览修复

## 背景

Redis、Elasticsearch 等原生数据库面板嵌入终端专用深色容器，但自身没有建立主题内容背景。浅色主题下，深色文字显示在黑色背景上，资源树看起来像空白页面。原有界面也没有区分可展开资源和仅能列出的叶子资源，容易误以为连接后没有内容。

## 范围

为全部原生数据库类型建立独立的主题工作区和统一资源展示模型。Redis、MongoDB、Cassandra、Couchbase 的一级资源支持展开；Elasticsearch、Memcached、InfluxDB、Neo4j、Kafka 使用明确的只读叶子资源呈现。Redis 会明确提示逻辑库与键的关系，Elasticsearch 会明确提示当前仅支持索引浏览。

本次不增加 Redis 键值读取/写入、Elasticsearch 文档查询/编辑或任何写操作。

## 修改文件

- `frontend/src/lib/nativeDatabaseWorkspace.js`
- `frontend/src/components/NativeDatabasePanel.svelte`
- `frontend/test/nativeDatabaseWorkspace.test.js`
- `frontend/test/nativeDatabasePanelLayout.test.js`
- `docs/designs/2026-07-23-native-database-workspace.md`
- `docs/development/2026-07-23-native-database-workspace.md`
- 本变更记录。

## 验证

- `cd frontend && node --test test/nativeDatabaseTypes.test.js test/nativeDatabaseWorkspace.test.js test/nativeDatabasePanelLayout.test.js`：3 项测试通过。
- `cd frontend && npm run build`：通过，包含 Vite 构建和 JDBC agent 打包。构建输出有既存组件的 Svelte 无障碍提示及大 chunk 提示，本次改动未新增编译警告。

## 剩余风险

资源数量仍受后端 provider 的限制，例如 Redis 单次最多枚举 1000 个键。原生数据库仍是资源浏览首版，不具备值、文档、消息或图数据的详情查看与编辑能力。
