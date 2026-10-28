# 变更：Release 更新说明补充 Docker 与下载进度

## 背景

本轮已落地 Docker 容器连接第一期，以及 SFTP 下载进度条修复，需写入发布流程中的更新说明，便于打 tag 发版时用户可见。

## 范围

- 更新 `.github/workflows/release.yml` 中 `body` 的「新增功能 / 修复问题」条目。
- 不调整构建矩阵或产物路径。

## 修改文件

- `.github/workflows/release.yml`
- `docs/changes/process/2026-10-23-release-notes-docker-download.md`（本文件）

## 验证

- 目视核对 release body：含 Docker 三种接入与下载进度修复说明，原有状态墙与批量执行条目保留。

## 剩余风险

- 更新说明为静态文案，与实际 commit 是否全部合入需发版前人工确认。
