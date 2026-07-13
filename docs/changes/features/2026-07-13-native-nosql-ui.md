# 原生 NoSQL 连接界面

## 背景

Redis、MongoDB、Elasticsearch/OpenSearch 已有原生后端 API，但现有界面只会创建 JDBC 数据库会话并显示 SQL 表面板。

## 范围

连接表单新增已实现的原生类型并隐藏 JDBC profile 选择；连接、关闭和资源浏览按原生 API 路由；新增只读资源面板。关系型数据库的 SQL 面板不改变。

## 修改文件

- `frontend/src/lib/nativeDatabaseTypes.js`
- `frontend/src/components/NativeDatabasePanel.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/App.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/test/nativeDatabaseTypes.test.js`
- 本变更记录。

## 验证

- `node --test frontend/test/nativeDatabaseTypes.test.js frontend/test/jdbcDriverProfileState.test.js` 通过。
- `npm run build` 的 Vite 编译通过。

## Gradle 工具链阻塞

`npm run build` 的 Vite 阶段完成后，`scripts/stage-jdbc-agent.sh` 启动 Gradle wrapper 时无法访问 `~/.gradle/wrapper/dists/gradle-8.5-bin/.../gradle-8.5-bin.zip.lck`，错误为 `Operation not permitted`。

最小修复方案：以已授权的受限构建命令重跑 `npm run build`，允许 Gradle 访问本机 wrapper 分发目录；不修改 Gradle 版本、wrapper 配置或 Java agent 源码。

## 剩余风险

当前界面只列出已实现 provider。Memcached、Cassandra、Couchbase、InfluxDB、Neo4j 和 Kafka 的 provider 与注册完成后，会在同一类型配置中新增选项。
