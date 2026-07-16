# 人大金仓表列表加载失败

## 背景

数据库列表页面会根据数据库类型生成表元数据查询。此前仅将 `postgresql` 识别为 PostgreSQL 兼容类型，人大金仓 `kingbase` 错误进入 MySQL 查询分支，访问了其不提供的 `TABLE_ROWS`、`ENGINE` 等列，导致加载表失败。

## 范围

将表元数据 SQL 提取为可测试函数，并让 `kingbase` 与 `postgresql` 使用 PostgreSQL 系统目录查询。MySQL 及其他现有类型继续使用原来的 `information_schema.tables` 查询。

## 修改文件

- `frontend/src/lib/tableMetadataQuery.js`
- `frontend/src/components/DatabaseListPanel.svelte`
- `frontend/test/tableMetadataQuery.test.js`
- 本变更记录。

## 验证

执行 `node --test test/tableMetadataQuery.test.js`，验证人大金仓使用 PostgreSQL 兼容查询、MySQL 保持原查询；执行前端生产构建验证组件编译。

首次执行 `npm run build` 时，JDBC agent 暂存阶段的 Gradle wrapper 无法访问 `~/.gradle/wrapper/dists/.../gradle-8.5-bin.zip.lck`，报错为 `Operation not permitted`。最小修复方案是仅允许该构建访问既有 Gradle wrapper 缓存后重试相同命令，不跳过 JDBC agent 构建或替换暂存脚本。

## 剩余风险

当前查询只读取 `public` schema 中的普通表。使用自定义 schema、分区表、外部表或权限受限目录的部署需要后续扩展 schema 筛选与对象类型选择。
