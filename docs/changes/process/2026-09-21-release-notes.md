# 变更：发布说明补充 Copilot 多模型

## 背景

自动发布工作流的更新说明已覆盖会话日志与常用命令提示，尚未写入 Copilot 多模型管理、按服务商获取模型列表，以及对应的设置刷新与绑定生成修复。

## 范围

- 更新 `.github/workflows/release.yml` 的「更新内容」正文
- 保留已有会话日志、命令提示、目录跟踪条目，只做补充

不改构建矩阵、签名步骤或附件路径。

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-09-21-release-notes.md`（本文）

## 验证

- 检查 YAML 缩进与 `body: |` 列表结构
- 执行 `git diff --check -- .github/workflows/release.yml docs/changes/process/2026-09-21-release-notes.md`

## 剩余风险

- 发布说明面向即将打出的下一标签；若发版前还有其它合入，需再同步更新
- `更新时间` 仍使用 `github.event.release.published_at`，标签推送触发时该字段可能为空
