# 变更：Docker 三期设计规格（侧栏可拖 / Exec / 资源）

## 背景

二期后侧栏不可拖导致日志过窄；用户需要继续 exec 与镜像/网络/卷管理。经 brainstorming 拍板交付顺序与范围，落成规格文档后再写实现计划。

## 范围

- 新增三期设计规格：刀 1 侧栏可拖、刀 2 exec 新页签、刀 3 镜像/网络/卷只读+删除。
- 明确非目标：attach、pull/create/run、Compose、侧栏宽度落盘、关父会话级联关 exec。

## 修改文件

- `docs/superpowers/specs/2026-10-27-docker-phase3-exec-resources-design.md`（新增）
- `docs/designs/2026-10-27-docker-phase3-exec-resources-design.md`（新增，与 specs 同步）
- `docs/changes/features/2026-10-27-docker-phase3-exec-resources-design.md`（本文件）

## 验证

- 与用户确认的口径对照：展示方式 A、顺序 1、exec 页签 A、资源深度 A、方案 1 三刀切开；规格正文无占位符与自相矛盾。

## 剩余风险

- 规格尚未经用户审阅文件终稿；审阅后可能微调宽度上下限或 exec 生命周期与父会话关系。
- 实现计划与代码尚未开始。
