# 变更：发版后清空发布说明

## 背景

上一版本已发布成功，`.github/workflows/release.yml` 中的「更新内容」属于已发布版本。为下一个版本重新累积更新说明，需要清空现有条目。

## 范围

- 清空「新增功能」「修复问题」「优化体验」下的全部条目，保留三个分类标题
- 不改版本号、更新时间、下载说明、构建矩阵、签名步骤或附件路径

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-10-08-release-notes-reset.md`（本文）

## 验证

- 使用 Ruby `YAML.load_file` 解析 `.github/workflows/release.yml`，格式正常

## 剩余风险

- 后续合入的功能和修复需要及时补写到发布说明，否则下次发版内容为空
