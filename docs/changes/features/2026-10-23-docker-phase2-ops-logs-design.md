# 变更：Docker 二期（操作 / 详情 / 日志）设计规格

## 背景

第一期仅列容器；二期需要启停删、详情与日志，且 CLI 回退路径需同等可用。exec 留后续。

## 范围

- 新增 brainstorming 确认后的二期设计规格。
- 不含实现代码。

## 修改文件

- `docs/superpowers/specs/2026-10-23-docker-phase2-ops-logs-design.md`（正文）
- `docs/designs/2026-10-23-docker-phase2-ops-logs-design.md`（入口镜像）
- `docs/changes/features/2026-10-23-docker-phase2-ops-logs-design.md`（本文件）

## 验证

- 规格与拍板一致：范围 B、CLI 全支持、侧栏日志 Tab、统一 ContainerOps、不做 exec。
- 自检：无 TBD；与一期连接模型无冲突。

## 剩余风险

- 实现前若砍掉 CLI 写操作或改为底部日志面板，需先改规格再开计划。
