# 数据库对象浏览器布局

## 背景

数据库对象主界面原先同时显示表、视图、存储过程和函数四个区块，信息密度低，也无法形成类似桌面数据库客户端的连续浏览流程。

## 范围

将主界面调整为单一对象浏览器：表为默认选中分类，用户切换分类后才加载并展示对应对象；提供搜索、刷新和当前数据库连接信息栏。表对象保留打开表结构行为。

## 修改文件

- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/lib/databaseObjectTree.js`
- `frontend/test/databaseObjectTree.test.js`
- `docs/designs/2026-07-16-database-sidebar-navigation.md`

## 验证

执行对象分类单元测试、`cd frontend && npm run build` 和桌面应用构建。手工确认默认进入表列表，切换其他分类后列表发生替换。

## 剩余风险

JDBC 元数据不提供统一的对象行数、引擎和体积统计，当前列表只显示可靠的对象名称。不同数据库驱动的例程元数据支持差异仍会导致过程或函数为空列表。
