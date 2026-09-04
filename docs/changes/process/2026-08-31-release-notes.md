# 变更：发布说明补充 08-31 会话改动

## 背景

自动发布工作流的更新说明仍为空模板，未覆盖本次会话完成的资产分组、标签批量关闭、终端主题、Oracle 表结构、文件夹上传与上传冲突处理。

## 范围

- 更新 `.github/workflows/release.yml` 的发布说明正文
- 记录 2026 年 08 月 31 日会话中的新增功能、缺陷修复与体验优化

不改构建矩阵、签名步骤或附件路径。

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-08-31-release-notes.md`（本文）

## 验证

- 检查 YAML 缩进与 `body: |` 列表结构
- 执行 `git diff --check -- .github/workflows/release.yml docs/changes/process/2026-08-31-release-notes.md`

## 剩余风险

- 发布说明面向即将打出的下一标签；若发版前还有其它合入，需再同步更新
- `更新时间` 仍使用 `github.event.release.published_at`，标签推送触发时该字段可能为空
