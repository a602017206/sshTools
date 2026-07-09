# JDBC 驱动管理设计

日期：2026-07-09

## 背景

当前数据库模块通过 Go `database/sql` 连接数据库，只注册了 MySQL 和 PostgreSQL 驱动。如果把所有已知 Go 数据库驱动都编译进主程序，会增加二进制体积；驱动变更也必须跟随应用发版；同时对企业级数据库和国产数据库的覆盖仍然有限。

目标方向是全部 JDBC 化：Go/Wails 应用负责产品状态、凭据、安装、进程生命周期；Java JDBC agent 负责驱动加载、数据库会话、SQL 执行、元数据访问。这样可以形成类似 DBeaver/DBX 的驱动模型：驱动和 Java 运行时只在需要时安装。

## 目标

- 支持更多数据库类型，不把所有驱动编译进 Go 主程序。
- 数据库模块统一通过 JDBC 执行，不再维护 Go 原生和 JDBC 两套连接路径。
- JRE 和 JDBC 驱动同时支持在线安装和离线导入。
- 普通用户创建连接时只需选择数据库类型，不需要理解驱动细节。
- 高级用户可以覆盖 driver class、JDBC URL template、版本、Maven 坐标和额外 jar。
- 尽量保持现有 Wails 数据库 API 的概念稳定，减少前端数据库面板的改动范围。

## 非目标

- 数据编辑、导入导出、结构对比、ER 图、迁移工具不进入首版范围。
- Go 原生数据库驱动不进入本设计，现有 Go `database/sql` 路径后续应替换或旁路。
- JDBC agent 不支持公网访问。

## 推荐架构

应用采用全部 JDBC agent 模型。

```text
Svelte UI
  -> Wails App API
  -> Go JdbcGatewayService
  -> Go AgentProcessManager
  -> Java JDBC Agent
  -> JDBC Driver jar
  -> Database
```

Go 应用仍是产品外壳和设置源。Java agent 是由 Go 启动的本地工作进程。Go 与 Java agent 通过绑定到 `127.0.0.1` 的 gRPC 通信。

运行时策略：

- 默认：首次使用数据库功能时自动下载并安装托管 JRE。
- 离线：允许从本地导入 JRE 包。
- 高级：允许指定系统 Java。
- 生命周期：默认按需启动 agent；设置里允许常驻；空闲超时后自动退出。

驱动策略：

- 官方驱动清单是主要来源。
- 自定义 Maven 坐标和手动导入 jar 是高级能力。
- 官方下载包和离线包都必须做 checksum 校验。
- 驱动和运行时存储在 `~/.sshtools/` 下。

首批 JDBC 数据库：

- MySQL
- PostgreSQL
- SQLite
- Oracle
- SQL Server
- 达梦 DM
- 人大金仓 Kingbase
- openGauss

首版数据库能力：

- 测试连接
- 打开和关闭会话
- 执行 SQL
- 返回查询列、行、影响行数
- 列出 catalog/database、schema、table、column

## 文件系统布局

```text
~/.sshtools/
  config.json
  credentials.enc
  runtimes/
    jre-21-<os>-<arch>/
  drivers/
    manifest.json
    mysql/
      8.4.0/
        mysql-connector-j.jar
        driver.json
    oracle/
      23.x/
        ojdbc11.jar
        driver.json
  agent/
    jdbc-agent.jar
  logs/
    jdbc-agent.log
    driver-install.log
    runtime-install.log
```

## 驱动离线包格式

离线驱动包使用 zip 格式：

```text
driver-package.zip
  package.json
  jars/*.jar
  checksums.sha256
  licenses/*
```

`package.json` 示例：

```json
{
  "id": "oracle",
  "name": "Oracle",
  "version": "23.5",
  "driverClass": "oracle.jdbc.OracleDriver",
  "urlTemplate": "jdbc:oracle:thin:@//{host}:{port}/{database}",
  "defaultPort": 1521,
  "jre": ">=17",
  "jars": ["ojdbc11.jar"]
}
```

## Go 模块边界

`DriverCatalogService`

- 读取官方 manifest。
- 读取本地已安装 manifest。
- 读取用户自定义 profile。
- 回答有哪些驱动、推荐哪些版本、哪些版本已安装。

`DriverInstallService`

- 下载官方驱动包。
- 导入离线包。
- 校验 checksum。
- 安装、卸载、导出、回滚驱动文件。
- 发出安装任务状态。

`RuntimeService`

- 检测托管 JRE、导入 JRE、系统 Java。
- 选择当前运行时。
- 安装和卸载托管运行时文件。
- 校验 Java 版本兼容性。

`AgentProcessManager`

- 按需启动 Java agent。
- 传递本地 gRPC 端口和 token。
- 执行健康检查。
- 重启或停止 agent。
- 应用空闲退出和常驻设置。

`JdbcGatewayService`

- 保持面向 Wails 的数据库 API 稳定。
- 把连接、查询、元数据、关闭调用转换成 gRPC 请求。
- 把 gRPC/agent 错误映射成前端可处理的数据库错误分类。

`DatabaseProfileService`

- 渲染 JDBC URL template。
- 提供默认端口、driver class、连接属性。
- 支持高级用户覆盖配置。

## Java Agent 模块边界

`DriverLoader`

- 为每个 driver profile 创建隔离 classloader。
- 加载 JDBC driver jar。
- 避免不同厂商驱动依赖冲突。

`ConnectionRegistry`

- 维护 `sessionID` 到 JDBC `Connection` 的映射。
- 负责连接关闭和清理。

`QueryService`

- 执行 SQL。
- 区分返回结果集的查询和更新语句。
- 返回列、行、影响行数。

`MetadataService`

- 使用 `DatabaseMetaData` 列出 catalog、schema、table、column。
- 对 JDBC 元数据不完整的数据库保留专用适配空间。

`HealthService`

- 返回 agent 版本、运行时版本、已安装驱动可见性、活跃会话。

## 驱动管理界面

使用双栏驱动管理页。

左栏：

- 搜索框。
- 过滤项：已安装、可安装、可更新、离线包、自定义 JDBC。
- 数据库列表，每项显示紧凑状态：未安装、已安装、可更新、校验失败、缺依赖。

右栏：

- 顶部显示数据库名、推荐版本、安装状态、license、包大小、来源。
- 版本区显示推荐版本、已安装版本、历史版本、切换、回退。
- 文件区显示 jar、checksum、安装路径、依赖 jar。
- 高级区显示 driver class、URL template、默认端口、连接属性、Maven 坐标、额外 jar。
- 操作区提供安装、重新安装、卸载、导入离线包、导出离线包、校验、测试驱动。
- 底部任务条显示下载、解压、checksum、安装、重试、失败状态。

JRE 管理放在同一个驱动管理区域，可以作为顶部状态条或设置子区。内容包括托管 JRE、系统 Java、离线导入、版本检测、切换、卸载。

普通用户路径：

```text
打开驱动管理 -> 选择数据库 -> 安装推荐驱动 -> 新建连接 -> 测试 -> 连接
```

高级用户路径：

```text
打开驱动管理 -> 新增自定义 JDBC profile -> 导入 jar 或填写 Maven 坐标 -> 校验 -> 创建连接并覆盖 profile
```

## 连接配置

连接 profile 存储：

- `db_type`
- 主机
- 端口
- 数据库名或服务名
- 用户名
- 加密密码引用
- 可选 `driver_profile_id`
- 可选连接属性

普通用户只选择数据库类型。应用把每种数据库类型绑定到推荐 driver profile。高级设置允许覆盖 profile、版本、class、URL template、jar。

## 错误处理

错误需要分类，并在前端提供对应行动按钮。

- `RUNTIME_MISSING`：无可用 JRE。操作：安装托管 JRE、导入离线 JRE、选择系统 Java。
- `DRIVER_MISSING`：驱动未安装。操作：安装推荐驱动、导入离线包。
- `DRIVER_INVALID`：jar 缺失、checksum 不匹配、class 加载失败。操作：重新安装、查看文件、删除。
- `AGENT_UNAVAILABLE`：启动失败、gRPC 不可用、版本不兼容。操作：重启 agent、查看日志。
- `DB_CONNECT_FAILED`：认证、网络、URL、数据库拒绝。操作：编辑连接、测试网络、查看原始错误。

SQL 错误保留数据库厂商原始 message，但密码、token、完整 JDBC URL 必须脱敏。

## 安全

- agent 只监听 `127.0.0.1`。
- Go 每次启动 agent 时生成一次性 token；每个 gRPC 请求都必须携带 token。
- Java agent 不直接读取 `credentials.enc`、SSH key、应用配置。
- Go 解密凭据后，只把创建会话所需的密码发送给 agent。
- 密码只在 agent 会话内存中存在。
- 官方包安装必须 checksum 校验。
- 离线包导入必须 checksum 校验。
- 自定义 Maven 坐标标记为未验证来源，默认不自动更新。
- 卸载驱动前检查连接 profile 依赖，并展示影响列表。

## 测试策略

Go 单元测试：

- manifest 解析。
- 版本选择。
- checksum 校验。
- 驱动路径计算。
- 离线包导入失败回滚。
- 运行时选择优先级。
- agent 启动、健康检查、空闲停止、重启。
- `JdbcGatewayService` 的 gRPC 错误映射。

Java agent 单元测试：

- classloader 隔离。
- JDBC URL template 渲染。
- select 和 update 的查询结果映射。
- 使用 H2 或 SQLite JDBC 测试元数据列表。

集成测试：

- 使用 H2 或 SQLite JDBC 作为无外部服务依赖的测试数据库。
- Go 启动 Java agent。
- 安装本地测试驱动包。
- 连接、查询、列表、列字段、关闭会话。
- 杀掉 agent 后，Go 能标记会话断开并返回重连操作。

手工验收：

- 无 Java 环境的机器首次使用数据库功能时，能引导安装托管 JRE。
- 离线机器能导入 JRE 包和 driver-package.zip。
- Oracle、SQL Server、达梦 DM、人大金仓 Kingbase、openGauss 至少能完成连接测试和简单查询。

## 推进计划

1. 引入驱动和运行时元数据模型、本地 manifest 存储。
2. 构建 Java agent，提供 gRPC health、connect、query、metadata、close 方法。
3. 增加 Go agent 进程管理器和 gateway service。
4. 用 gateway 调用替换当前 Go `database/sql` 数据库调用。
5. 增加驱动管理界面和 JRE 管理。
6. 增加离线导入和导出流程。
7. 对首批数据库做验证。

## 剩余风险

- JDBC 元数据行为因厂商而异，可能需要数据库专用适配器。
- 部分厂商驱动有 license 限制，可能不能随应用分发。
- Java runtime 下载引入供应链和镜像可用性风险。
- gRPC 进程管理增加跨平台生命周期复杂度。
- agent 崩溃会让活跃 session 失效，用户需要重连。
