# 表数据工作区改造

## 背景

用户需要表数据页在视觉和工作流上接近 Navicat 的数据浏览页面，而不是通用 SQL 编辑器。

## 范围

重构 `DatabaseTablePanel` 为数据网格优先的表数据工作区，提供数据和 SQL 模式、工具栏、行号、单元格选择、对象信息、查询历史和状态栏。筛选与排序采用类似 Navicat 的条件编辑器：可选择字段、`AND`/`OR`、比较方式和值，并可按多个字段指定升序或降序。

条件支持包含、不包含、等于、不等于、大于、大于等于、小于、小于等于、为 NULL、不为 NULL、在列表中和不在列表中。应用条件时仅生成受控的表浏览查询；手写 SQL 仍保持原样执行。

## 修改文件

- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/lib/tableQueryBuilder.js`
- `frontend/test/databaseTableWorkspace.test.js`
- `frontend/test/tableQueryBuilder.test.js`
- `docs/designs/2026-07-18-table-data-workspace.md`

## 验证

执行表数据工作区结构测试、查询生成器单元测试、前端构建和桌面应用构建。手工确认打开表后默认显示数据网格，SQL 模式可编辑并执行查询，条件编辑器可按字段组合筛选并多字段排序，网格可以在当前结果中查找和滚动。

### 构建环境记录

首次桌面打包在 Wails 的“编译应用”阶段中断，且没有生成最终二进制。详细日志显示 Wails 在内部 `npm run build` 子步骤后未继续输出，而独立前端构建已成功生成资源。最小修复方案是以前端独立构建结果为输入，使用 Wails `-s` 跳过重复前端步骤，只执行 Go 编译和应用打包；不跳过任何未验证资源，也不替换构建链路。

## 剩余风险

后端尚未提供字段类型、可编辑行、主键识别和游标分页能力，因此本次页面仅提供只读数据浏览和 SQL 执行。条件值统一按 SQL 文本字面量传递，数值、日期和数据库专有类型的精确类型转换仍由 JDBC 驱动负责。
