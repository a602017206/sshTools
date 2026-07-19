# 表结构字段元数据展示

## 背景

表结构详情此前只显示 DDL，无法直观查看字段的数据类型、长度和描述。JDBC 元数据已提供大部分字段信息，但 `REMARKS` 没有经过协议和界面传递。

## 范围

为所有使用 JDBC agent 的关系型数据库统一展示字段名、数据类型、长度、描述、可空和默认值。非 JDBC 的原生 NoSQL 面板不使用关系型表结构详情，未纳入本次范围。

## 修改文件

- `jdbc-agent/src/main/proto/jdbc_agent.proto`：增加字段描述协议字段。
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`：读取 JDBC `REMARKS`。
- `internal/service/jdbcproto/jdbc_agent*.go`：同步 protobuf Go 绑定。
- `internal/config/database.go`、`internal/service/jdbc_gateway.go`、`internal/service/jdbc_managed_gateway.go`、`internal/service/database_service.go`、`app.go`：映射并导出结构化字段元数据，并将当前数据库传给 JDBC `Catalog`。
- `frontend/src/components/TableStructurePanel.svelte`、`frontend/src/components/DatabaseTablePanel.svelte`、`frontend/src/lib/tableStructureMetadata.js`：在 DDL 前展示字段表格，并在表数据网格的列头下展示类型、长度和描述。
- `frontend/wailsjs/go/main/App.js`：同步 Wails 前端绑定。
- 对应 Java、Go 和前端测试：覆盖字段描述传输与显示格式。

## 验证

- 执行 JDBC agent `MetadataServiceImplTest`，验证 H2 字段注释通过 `REMARKS` 返回。
- 执行 Go JDBC 网关测试，验证描述映射到 `ColumnSchema`。
- 执行前端 Node 测试，验证类型、长度和描述格式。
- 执行前端生产构建，确认 Svelte 与 JDBC agent 可编译。

## 剩余风险

字段描述依赖数据库驱动是否实现 JDBC `REMARKS` 以及当前账户是否有读取元数据权限。缺失描述时界面显示 `-`，不会影响字段类型和长度展示。
