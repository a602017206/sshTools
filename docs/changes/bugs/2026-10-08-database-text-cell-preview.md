# 缺陷：text 列悬停预览妨碍双击复制

## 背景

数据库表数据网格中，每个单元格都通过 `title` 显示完整内容。`text` 等长文本列（如存放 HTML 的 `content_`）单击或悬停时，macOS 会弹出铺满窗口的原生提示框，遮挡表格，导致无法双击选中、复制，也无法继续操作。

首版修复只按列类型判断，依赖表结构元数据；在人大金仓表上实测提示框仍会弹出，因此补充了按内容判断的兜底。

## 范围

- 以下单元格不再设置 `title` 悬停预览：
  - 字段类型包含 `text`、`json`、`xml`、`clob`、`blob`、`bytea` 的列；
  - 内容超过 100 个字符的单元格（不依赖列元数据）；
  - 内容包含换行的单元格。
- 其他短内容保持原有提示；`NULL` 值仍提示 `NULL`。

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

单元测试 5 个全部通过，前端构建成功。未在桌面端手工验证。

## 剩余风险

超过 100 个字符的普通字段（如长 `VARCHAR`）也不再显示悬停提示，需要查看完整内容时可点进单元格查看或复制。
