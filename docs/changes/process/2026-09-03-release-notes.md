# 变更：发布说明补充 MQ 与驱动分组

## 背景

自动发布工作流的更新说明需覆盖 RocketMQ / RabbitMQ 原生连接、连接类型与 JDBC 驱动分组，以及近期 Redis/Kafka 相关修复。

## 范围

- 更新 `.github/workflows/release.yml` 的「更新内容」正文
- 不改构建矩阵、签名步骤或附件路径

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-09-03-release-notes.md`（本文）

## 验证

- 检查 YAML 缩进与 `body: |` 列表结构
- 执行 `git diff --check -- .github/workflows/release.yml docs/changes/process/2026-09-03-release-notes.md`

## 剩余风险

- 发布说明面向即将打出的下一标签；若发版前还有其它合入，需再同步更新
