# 数据库对象列表滚动修复

## 背景

对象浏览器的列表位于 flex 和 grid 嵌套布局中，但主内容子项没有允许收缩的高度约束。表数量较多时，列表会向下撑出可视区域，无法滚动到后续对象；macOS 自动隐藏滚动条也降低了可发现性。

## 范围

限制对象浏览器主内容区的高度，列表在自身区域内纵向和横向滚动；预留滚动条槽位，并为 Chromium/Wails 显示可见的滚动轨道和滑块。

## 修改文件

- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/test/objectBrowserLayout.test.js`

## 验证

执行对象浏览器布局测试、前端构建和桌面应用构建。手工确认大量表时可拖动右侧滚动条至列表末尾。

## 剩余风险

滚动条的精确外观由操作系统和 Chromium 版本决定；非 WebKit 环境会保留原生滚动条，但仍可滚动。
