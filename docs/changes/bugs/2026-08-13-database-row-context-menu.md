# 修复数据库数据行右键菜单定位和关闭行为

## 背景

数据库数据行右键菜单在复杂表格容器中使用视口坐标定位，受父级定位上下文影响会偏离被选中单元格；菜单打开后点击其他位置也无法关闭。

## 范围

仅调整数据库表数据行的右键菜单。复制行、复制 INSERT 和删除记录的原有行为不变。

## 修改文件

- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/lib/databaseRowContextMenu.js`
- `frontend/test/databaseRowContextMenu.test.js`
- `.github/workflows/release.yml`

## 验证

- 运行 `cd frontend && node --test test/databaseRowContextMenu.test.js`。
- 运行 `cd frontend && ./node_modules/.bin/vite build --logLevel error`。

## 剩余风险

菜单以触发单元格下边缘定位；在窗口底部空间不足时浏览器仍可能需要滚动查看菜单。
