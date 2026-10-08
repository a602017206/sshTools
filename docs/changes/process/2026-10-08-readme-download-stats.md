# 变更：README 增加下载与访问统计

## 背景

需要在 README 中直观展示应用发布后的下载量和访问量，不额外维护统计脚本。

## 范围

- 标题下增加徽章：最新版本、总下载量、最新版下载量、访问次数。
- 新增「下载统计」表格，按 Windows、macOS（Apple Silicon）、macOS（Intel）展示最新版各安装包下载量。
- 下载数据来自 shields.io 实时读取 GitHub Releases 的 `download_count`；访问次数使用 [hits.sh](https://hits.sh) 计数徽章（原先用的 `visitor-badge.laobi.icu` 对中文标签宽度计算错误，文字会叠在一起，已更换）。

## 修改文件

- `README.md`
- `docs/changes/process/2026-10-08-readme-download-stats.md`（本文）

## 验证

- 请求 shields.io 下载量徽章，返回 `downloads: 9`、`downloads@latest: 0`、`downloads@latest: 0 [AHaSSHTools.exe]`，与 GitHub API 统计一致。
- 访问计数徽章返回 200。
- 版本号徽章在本地请求超时（shields.io 响应慢），未确认渲染效果。

## 剩余风险

- 下载量包含重复下载和自动化拉取，仅作参考；Source code 压缩包不计入。
- 访问次数只统计 README 图片被加载的次数（含 GitHub 图片缓存影响），并非 GitHub Insights 中的真实访问量；hits.sh 服务停止后徽章将无法显示。更换服务后计数会重新累计，不会继承旧徽章的数值。
- 安装包文件名变更后，表格中按文件名统计的徽章需要同步修改。
