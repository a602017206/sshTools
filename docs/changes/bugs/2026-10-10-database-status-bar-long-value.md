# 缺陷：点击长文本单元格后状态栏撑满整个表格

## 背景

表数据页点击 `LONGTEXT` 等长文本单元格后，大段 HTML 文本铺满整个工作区，表格无法继续操作，也无法复制。

此前两次修复都针对单元格的 `title` 悬停提示，但真正原因是底部状态栏：它把选中单元格的完整值原样显示为「当前值：…」，长文本自动换行后，状态栏高度被撑到覆盖整个表格。

## 范围

- 新增 `formatStatusValue`：状态栏当前值压缩空白为单行，超过 80 个字符截断并加 `…`；空值显示 `NULL`。
- 状态栏取值区域强制单行、溢出省略，其余状态信息不被挤压。
- 单元格本身仍是单行输入框：点击后可在格内移动光标、选中和复制，交互与 Navicat 的行内编辑一致。

## 修改文件

- `frontend/src/lib/tableGridColumns.js`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/test/tableGridColumns.test.js`
- `.github/workflows/release.yml`（发布说明）
- 本文档

## 验证

```bash
cd frontend && node --test test/tableGridColumns.test.js && npx vite build
```

单元测试 6 个全部通过，前端构建成功。未在桌面端手工点击验证。

## 剩余风险

- 状态栏只显示前 80 个字符；查看完整长文本需要在单元格内滚动光标，暂未提供 Navicat 那样的独立大文本查看/编辑窗口。
