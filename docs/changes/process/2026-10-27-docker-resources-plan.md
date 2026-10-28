# 变更：Docker 镜像/网络/卷实现计划

## 背景

三期规格刀 3（镜像/网络/卷只读+删除）已评审；刀 1、刀 2 已落地。需在写代码前落成可执行分步计划。

## 范围

- 新增刀 3 实现计划：API/CLI 列表与删除、DockerService、Wails 导出、工作区资源 Tab 与详情侧栏。
- 不含 pull/create/run/Compose。

## 修改文件

- `docs/superpowers/plans/2026-10-27-docker-resources.md`（新增）
- `docs/plans/2026-10-27-docker-resources.md`（新增，同步副本）
- `docs/changes/process/2026-10-27-docker-resources-plan.md`（本文件）

## 验证

- 对照规格刀 3：资源 Tab、字段、删除确认、Wails API、API+CLI 均有对应 Task。

## 剩余风险

- CLI `docker images` Size 人类可读字符串需粗解析，可能与 API 字节数不完全一致。
- 手工验收依赖本机/SSH Docker 环境。
