# 表数据工作区改造

## 背景

用户需要表数据页在视觉和工作流上接近 Navicat 的数据浏览页面，而不是通用 SQL 编辑器。

## 范围

重构 `DatabaseTablePanel` 为数据网格优先的表数据工作区，提供数据和 SQL 模式、工具栏、筛选、行号、单元格选择、对象信息、查询历史和状态栏。

## 修改文件

- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/test/databaseTableWorkspace.test.js`
- `docs/designs/2026-07-18-table-data-workspace.md`

## 验证

执行表数据工作区结构测试、前端构建和桌面应用构建。手工确认打开表后默认显示数据网格，SQL 模式可编辑并执行查询，网格可以排序、筛选和滚动。

### 构建环境记录

首次桌面打包在 Wails 的“编译应用”阶段中断，且没有生成最终二进制。详细日志显示 Wails 在内部 `npm run build` 子步骤后未继续输出，而独立前端构建已成功生成资源。最小修复方案是以前端独立构建结果为输入，使用 Wails `-s` 跳过重复前端步骤，只执行 Go 编译和应用打包；不跳过任何未验证资源，也不替换构建链路。

## 剩余风险

后端尚未提供可编辑行、主键识别和游标分页能力，因此本次页面仅提供只读数据浏览和 SQL 执行。
