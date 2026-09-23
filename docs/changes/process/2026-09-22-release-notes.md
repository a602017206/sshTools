# 变更：发布说明补充 vim 终端修复

## 背景

SSH 里 vim 状态栏散成碎片的问题已修复。自动发布工作流的更新说明需同步该修复项。

## 范围

- 更新 `.github/workflows/release.yml` 的「修复问题」列表
- 不改构建矩阵、签名步骤或附件路径

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-09-22-release-notes.md`（本文）

## 验证

- 检查 YAML 缩进与 `body: |` 列表结构
- 执行 `git diff --check -- .github/workflows/release.yml docs/changes/process/2026-09-22-release-notes.md`

## 剩余风险

- 发布说明面向即将打出的下一标签；若发版前还有其它合入，需再同步更新
