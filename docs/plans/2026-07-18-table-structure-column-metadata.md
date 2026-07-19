# 表结构字段元数据展示实施计划

> **执行说明：** 本计划按测试驱动方式逐项实施并在当前工作区验证。

**目标：** 在所有 JDBC 数据库的表结构详情中展示字段类型、长度和描述。

**架构：** JDBC agent 从 `DatabaseMetaData.getColumns` 读取 `REMARKS`，经 protobuf、Go 网关和 Wails 绑定传递给 `TableStructurePanel`。前端使用统一字段表格渲染，不按数据库产品分支。

**技术栈：** Java gRPC、protobuf、Go、Wails、Svelte、Node.js test。

---

### 任务 1：定义字段描述的传输契约

**文件：**
- 修改：`jdbc-agent/src/main/proto/jdbc_agent.proto`
- 修改：`jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`
- 测试：`jdbc-agent/src/test/java/com/sshtools/jdbcagent/MetadataServiceImplTest.java`

1. 先增加断言，要求列元数据返回数据库字段备注。
2. 执行 JDBC agent 测试，确认在现状下失败。
3. 增加 `description` 协议字段并映射 JDBC `REMARKS`。
4. 重新生成 protobuf 代码并执行测试。

### 任务 2：将结构化字段元数据暴露给表结构面板

**文件：**
- 修改：`internal/config/database.go`
- 修改：`internal/service/jdbc_gateway.go`
- 修改：`internal/service/database_service.go`
- 修改：`app.go`
- 测试：`internal/service/jdbc_gateway_test.go`

1. 先测试 Go 网关将字段描述映射到 `ColumnSchema`。
2. 增加 schema 范围的公共服务与 Wails 导出方法。
3. 执行 Go service 测试。

### 任务 3：在表结构详情渲染字段表头

**文件：**
- 新增：`frontend/src/lib/tableStructureMetadata.js`
- 新增：`frontend/test/tableStructureMetadata.test.js`
- 修改：`frontend/src/components/TableStructurePanel.svelte`

1. 先测试字段类型、长度和空描述的显示格式。
2. 实现格式化帮助函数。
3. 表结构面板加载结构化字段，并在 DDL 前渲染字段表格。
4. 执行前端测试和生产构建。

### 任务 4：记录与验证

**文件：**
- 新增：`docs/changes/features/2026-07-18-table-structure-column-metadata.md`

1. 记录背景、范围、修改文件、验证和剩余风险。
2. 检查 diff、生成物和构建输出。
