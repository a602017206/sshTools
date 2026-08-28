# Phase 2 实现计划：工作区拆分与新建对话框域分组

## 目标

1. 从 `NativeDatabasePanel` 拆出 `RedisWorkspace` / `ElasticsearchWorkspace`
2. 新增 `KafkaWorkspace` 骨架（Topic 列表 + Describe）
3. 新建连接对话框数据库类型按域分组

## 步骤

1. 增加 `resolveNativeWorkspaceKind(dbType)`
2. 抽取 Redis / ES / Kafka / Generic 四个工作区组件
3. `NativeDatabasePanel` 改为按 kind 路由
4. 更新 layout 测试
5. `AddAssetDialog` 使用 optgroup 按域分组
6. 补充变更文档并跑相关测试

## 验收

- Redis 打开不再混入 ES 查询空态
- Kafka 面板文案为 Topic 元数据，无 DSL/键编辑
- 新建对话框类型列表按「数据库 / 缓存 / 搜索 / 消息队列」分组
