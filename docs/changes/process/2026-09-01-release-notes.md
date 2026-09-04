# 变更：发布说明补充 09-01 会话改动

## 背景

自动发布工作流的更新说明需覆盖本次会话完成的数据库 Schema 右键菜单、运行 SQL 文件、Copilot 当前打开对象上下文，以及 SSH 终端编码热切换。

## 范围

- 更新 `.github/workflows/release.yml` 的发布说明正文
- 记录 2026 年 09 月 01 日会话中的新增功能

不改构建矩阵、签名步骤或附件路径。

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-09-01-release-notes.md`（本文）

## 验证

- 检查 YAML 缩进与 `body: |` 列表结构
- 执行 `git diff --check -- .github/workflows/release.yml docs/changes/process/2026-09-01-release-notes.md`

## 剩余风险

- 发布说明面向即将打出的下一标签；若发版前还有其它合入，需再同步更新
- `更新时间` 仍使用 `github.event.release.published_at`，标签推送触发时该字段可能为空
