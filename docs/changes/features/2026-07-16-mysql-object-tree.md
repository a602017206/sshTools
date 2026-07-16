# MySQL 对象树

## 背景

数据库浏览需要接近 Navicat 的对象层级，MySQL 原有界面只显示表列表，无法展示视图、例程和事件。

## 范围

新增 MySQL 懒加载对象树，提供数据库下的表、视图、存储过程、函数和事件分类；表保留单击查看 DDL、双击查看数据的操作。PostgreSQL 兼容对象树也补充双击打开数据浏览的操作。

## 修改文件

- `frontend/src/lib/databaseObjectTree.js`
- `frontend/src/components/MySQLObjectTree.svelte`
- `frontend/src/components/PostgreSQLObjectTree.svelte`
- `frontend/src/components/TerminalPanel.svelte`
- `frontend/test/databaseObjectTree.test.js`
- `docs/designs/2026-07-16-mysql-object-tree.md`
- 本变更记录。

## 验证

执行 `node --test test/databaseObjectTree.test.js` 验证 MySQL 系统目录查询、数据库名转义和对象分类。首次执行 `npm run build` 时，Vite 编译通过，但 JDBC agent 暂存被 Gradle 缓存锁文件 `/Users/dingwei/.gradle/wrapper/dists/gradle-8.5-bin/.../gradle-8.5-bin.zip.lck` 的沙箱权限阻断。

## 剩余风险

SQLite、Oracle、SQL Server、达梦等 JDBC 数据库仍使用原列表界面，后续需要依据其系统目录分别扩展对象分类。当前未连接外部 MySQL 实例进行权限差异验证。Gradle 阻塞的最小修复方案是在受控本机环境中允许访问既有 `~/.gradle` 包装器缓存后重跑原 `npm run build`，不修改构建脚本或绕过 JDBC agent 暂存。
