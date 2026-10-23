# 变更：Docker 容器连接设计规格

## 背景

Docker 域与添加连接页签已预留，但缺少本机 / SSH 关联 / TCP 三种接入与第一期工作区边界的书面规格；现有表单字段易误导（主机密码套用到 Docker）。

## 范围

- 新增 brainstorming 确认后的 Docker 连接设计规格（资产模型、三种接入、第一期列表工作区、非目标）。
- 不包含实现代码。

## 修改文件

- `docs/superpowers/specs/2026-10-23-docker-connection-design.md`（正文）
- `docs/designs/2026-10-23-docker-connection-design.md`（入口镜像）
- `docs/changes/features/2026-10-23-docker-connection-design.md`（本文件）

## 验证

- 规格与已拍板口径一致：三种入口、SSH 隧道优先 + CLI 回退、侧栏+SSH 右键、本机检测后确认添加、第一期仅列容器。
- 自检：无 TBD/占位；与 `ConnectionConfig.type=docker` 及多域导航「不内嵌引擎」一致。

## 剩余风险

- 实现前若产品改口径（例如取消 CLI 回退），需先改规格再开计划。
