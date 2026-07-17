# MySQL 系统表与表数据浏览修复

## 背景

对象浏览器为“表”分类显式请求了 `TABLE` 类型，覆盖 JDBC agent 默认的 `TABLE` 与 `SYSTEM TABLE` 类型集合。MySQL 驱动会把 `information_schema` 中的部分对象标记为 `SYSTEM TABLE`，导致左侧和主列表看不到表。对象浏览器中的表点击也被错误地连接到表结构，而非表数据浏览标签。

## 范围

将 `SYSTEM TABLE` 纳入表分类；表项点击创建数据浏览标签并默认执行 `LIMIT 100` 查询。对于 PostgreSQL、Kingbase 和 openGauss，数据浏览查询使用 `schema.table`，其余 JDBC 数据库使用 `database.table`。

## 修改文件

- `frontend/src/lib/databaseObjectTree.js`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/test/databaseObjectTree.test.js`

## 验证

执行对象分类单元测试、`cd frontend && npm run build` 以及桌面应用构建。手工确认 MySQL 的 `information_schema` 显示系统表，点击表后进入包含数据结果的标签。

### 构建环境记录

首次执行 Wails 打包时，当前 shell 的 `PATH` 未包含 Go 编译器，Wails 报告 `unable to find compiler: go`。补齐 Go 路径后，Wails 又因系统路径缺少 `/usr/sbin/sysctl` 停止。最小修复方案是在重试同一构建命令时加入 `/usr/local/go/bin`、`/usr/sbin` 和 `/usr/bin`，不修改 Wails 配置或跳过桌面打包。

## 剩余风险

系统表的实际可见性仍受 MySQL 账户权限和驱动实现影响。表名包含非标准字符时，浏览器会使用数据库类型对应的标识符引号处理。
