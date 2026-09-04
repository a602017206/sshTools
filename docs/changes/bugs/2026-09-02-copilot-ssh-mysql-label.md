# 修复 SSH 场景 AI 助手显示 mysql

## 背景

Shell 助手标题正确，但「当前：」显示 `mysql`。SSH 资产保存时表单默认 `dbType=mysql` 被写入 `metadata.db_type`；上下文构建用 `session.type === 'database' || dbType`，有 db_type 就当 JDBC。

## 范围

- SSH / local / mode=ssh 强制走 SSH 上下文，标签显示 `user@host`
- 新建/更新 SSH 资产不再写入 `db_type` 等数据库字段

## 修改文件

- `frontend/src/lib/copilotContext.js`
- `frontend/test/copilotContext.test.js`
- `frontend/src/components/AIPanel.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `.github/workflows/release.yml`
- `docs/changes/bugs/2026-09-02-copilot-ssh-mysql-label.md`（本文）

## 验证

- `cd frontend && node --test test/copilotContext.test.js`
- 手工：打开 SSH 会话的 Shell 助手，「当前」为 `user@host` 而非 mysql

## 剩余风险

- 旧 SSH 资产 metadata 里可能仍有 `db_type`，已由上下文判定忽略
