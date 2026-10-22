# 变更：发布说明补充命令提示交互调整

## 背景

常用命令提示已改为默认仅鼠标点击采纳，并新增「Tab 键采纳提示」开关，浮层位置改为贴近光标（见 `docs/changes/features/2026-10-10-command-suggest-mouse-accept.md`）。自动发布工作流的更新说明需同步这项改动。

## 范围

- 更新 `.github/workflows/release.yml` 的「优化体验」列表
- 不改构建矩阵、签名步骤或附件路径

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-10-10-release-notes.md`（本文）

## 验证

- 检查 YAML 缩进与 `body: |` 列表结构
- 使用 Ruby `YAML.load_file` 解析 `.github/workflows/release.yml`

## 剩余风险

- 发布说明面向即将打出的下一标签；若发版前还有其它合入，需再同步更新
