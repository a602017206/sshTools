# 变更：Docker 容器 Exec 实现计划

## 背景

三期规格刀 2（exec 新页签）已评审通过；刀 1 侧栏可拖已落地。需在写代码前落成可执行分步计划。

## 范围

- 新增刀 2 实现计划：API/CLI exec、DockerService 生命周期、Wails 导出、前端 docker-exec 页签。
- 不含刀 3 镜像/网络/卷。

## 修改文件

- `docs/superpowers/plans/2026-10-27-docker-exec.md`（新增）
- `docs/plans/2026-10-27-docker-exec.md`（新增，同步副本）
- `docs/changes/process/2026-10-27-docker-exec-plan.md`（本文件）

## 验证

- 对照规格刀 2：交互、Wails API、API/CLI 路径、关闭清理、Terminal 复用均有对应 Task。

## 剩余风险

- StartExec HTTP hijack 在不同 transport 上实现复杂度较高，计划已给出注记；实现阶段可能需微调测试策略。
- 手工验收依赖本机/SSH Docker 环境。
