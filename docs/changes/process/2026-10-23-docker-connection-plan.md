# 变更：Docker 容器连接实现计划

## 背景

规格 `docs/superpowers/specs/2026-10-23-docker-connection-design.md` 已确认，进入实现前需落地分任务计划。

## 范围

- 新增实现计划文档（TDD 分步、文件映射、接口约定）。
- 不含业务代码。

## 修改文件

- `docs/superpowers/plans/2026-10-23-docker-connection.md`
- `docs/changes/process/2026-10-23-docker-connection-plan.md`（本文件）

## 验证

- 计划覆盖规格：三种 transport、隧道+CLI 回退、本机探测提示、工作区列表、非目标边界。
- 无 TBD 占位；任务接口命名前后一致。

## 剩余风险

- 实现中若发现 Wails 导出或顶栏 mode 与现有双模式冲突，需在 Task 5/7 微调 UI 挂载，不改变规格能力边界。
