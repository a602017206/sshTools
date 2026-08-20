# 变更：发布说明补充 AI Copilot

## 背景

`feat/ai-copilot-sql-shell` 已合入 `master`，自动发布工作流的更新说明仍为空或仍为上一版内容，需要写入本版本 AI Copilot 相关的新增、修复与体验说明。

## 范围

- 更新 `.github/workflows/release.yml` 的发布说明正文。
- 记录相对 `v0.0.18` 之后的 AI Copilot 功能、修复与体验优化。

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-08-20-release-notes-ai-copilot.md`（本文档）

## 验证

- 检查 YAML 缩进与发布说明列表结构。
- 执行 `git diff --check -- .github/workflows/release.yml`。

## 剩余风险

- 发布说明面向即将打出的下一标签；若发版前还有其它合入，需再同步更新。
- 规格第 9/10 节的 `wails dev` 手工回归仍未在桌面环境完成，说明文案与实现一致，但不替代手工验收。
