# JDBC 驱动 profile 操作状态

## 背景

驱动管理器按整个驱动是否存在任一已安装版本决定操作按钮。V8 已安装、V9 未安装时，选择 V9 仍会显示校验和卸载，导致卸载按钮操作错误。

## 范围

操作按钮改为仅依据当前选中的 profile 安装状态决定；被引用而禁止卸载时显示专用中文提示，并允许查看后端返回的连接或会话信息。

## 修改文件

- `frontend/src/lib/jdbcDriverProfileState.js`
- `frontend/src/components/JDBCDriverManager.svelte`
- `frontend/test/jdbcDriverProfileState.test.js`
- `app.go`
- 本变更记录。

## 验证

执行 `node --test frontend/test/jdbcDriverProfileState.test.js`，覆盖同一驱动不同 profile 的状态分离。

## 剩余风险

驱动列表筛选仍以驱动整体安装状态展示，表示该驱动是否至少安装过一个版本；这不影响所选 profile 的操作状态。
