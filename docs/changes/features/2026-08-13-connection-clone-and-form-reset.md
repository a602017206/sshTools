# 连接克隆与新增表单重置

## 背景

新增 SSH 连接时，表单可能保留上一次填写的连接信息，容易将旧主机、账号等配置误用于新连接。与此同时，用户缺少基于已有连接快速创建相似配置的入口。

## 范围

- 打开“新增连接”时重置表单，确保以空白连接配置开始填写。
- 在连接项右键菜单中增加“克隆”功能，基于原连接创建一份可编辑副本。
- 克隆后的连接名称追加 `copy` 与时间戳，避免与原名称混淆；用户仍可自行修改名称。
- 克隆仅复制连接配置，不复制密码、私钥口令等凭据。

## 修改文件

- `frontend/src/components/ConnectionManager.svelte`：处理新增连接的表单重置，并提供右键克隆入口。
- `frontend/src/lib/connectionFormData.js`：整理新增和克隆连接使用的表单数据。
- `frontend/test/connectionFormData.test.js`：覆盖表单重置与克隆数据行为。
- `docs/designs/2026-08-13-connection-clone-and-form-reset.md`：记录方案和设计取舍。
- `docs/plans/2026-08-13-connection-clone-and-form-reset.md`：记录实施步骤。

## 验证

- `node --test test/connectionFormData.test.js`
- `npm run build`
- `git diff --check`

## 剩余风险

- 右键菜单及快捷操作的最终体验仍需在桌面应用中手工确认。
- 克隆不会携带凭据；首次使用克隆连接时，用户需按实际认证方式重新填写或选择凭据。
