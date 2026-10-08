# 变更：按实际完成情况重写 README

## 背景

根目录 `README.md` 仍按早期「SSH + SFTP + 监控 + DevTools」描述，与当前产品差距很大：

- 关系库（JDBC）、Redis / ES、Kafka / RocketMQ / RabbitMQ、Copilot、会话日志、远程编辑、命令提示等已上线，却未写入或仍标在「计划中」。
- 「计划中」里会话日志、主题、URL 工具、性能监控等已完成，造成误导。
- 安装命令仍指向已删除的 `frontend_old`；Go 版本写成 1.25，与 `go.mod`（1.24）不符。
- 使用指南过长且与 `QUICK_START.md` 重复；项目结构树严重过时。

## 范围

- 重写 README：徽章与下载表保留并并入「下载」节；按域与成熟度（完整 / 可用 / 雏形）梳理功能。
- 明确 Docker、隧道、跳板机等未完成项；注明 `DEVELOPMENT_PLAN.md` 可能滞后。
- 精简使用说明，指向 `QUICK_START.md`；修正开发命令与技术栈版本。
- 保留赞助、安全说明、文档入口与 License。

## 修改文件

- `README.md`
- `docs/changes/process/2026-10-08-readme-restructure.md`（本文）

## 验证

- 对照 `AddAssetDialog` 数据库类型、`DevToolsPanel` 工具列表、`app.go` 会话日志 / Copilot API、`nativeDatabaseWorkspace.js` 成熟度描述进行人工核对。
- 未重新跑应用冒烟；未改代码逻辑。

## 剩余风险

- 国产库与部分原生库能力随版本变化，README 成熟度标签需随发版同步。
- MongoDB 等「雏形」类型若后续加深，需更新表格，避免长期写死为只读。
- `DEVELOPMENT_PLAN.md` 本身未在本次同步，仍可能与 README 不一致。
