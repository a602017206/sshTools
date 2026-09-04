# 设计：数据库 Schema 右键菜单与运行 SQL 文件

## 背景

Navicat 可在库/Schema 节点上新建查询、刷新、断开，并直接执行上百 MB 的初始化脚本。当前应用里该节点没有右键；查询页和 `ExecuteDatabaseQuery` 按单条 SQL、30 秒、结果表格设计，不能承载 91MB 级脚本。

## 决策

- 菜单挂在左侧 `DatabaseSidebarTree` 的库（MySQL）或 Schema（Oracle/PostgreSQL）节点上。
- 菜单项：新建查询、运行 SQL 文件…、刷新、断开。不做转储、命令行、数据字典、逆向模型。
- **运行 SQL 文件**在选完系统文件后立即后台执行，不把内容填进查询编辑器。
- 文件由 Go 在本机读取并流式拆句，经现有 JDBC `ExecuteQuery` 逐条执行；进度用事件推送，Wails 调用立即返回。可取消。
- 执行前按方言切换当前库/Schema（MySQL `USE`、Oracle `ALTER SESSION SET CURRENT_SCHEMA`、PostgreSQL `SET search_path`）。
- 单条语句超时 5 分钟；整文件无 30 秒上限。gRPC 消息上限提高到 64MB，以容纳较大的多行 INSERT。
- 第一条失败即停止，进度里带出错语句摘要。存储过程体中夹杂分号的脚本第一版可能拆不准。

## 限制

JDBC agent 仍在本机；逐条 gRPC 会比 Navicat 进程内执行慢，但能跑 91MB 量级的建表 + INSERT 初始化。不把整份脚本载入前端。
