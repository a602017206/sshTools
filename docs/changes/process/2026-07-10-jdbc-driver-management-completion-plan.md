# JDBC 驱动管理闭环计划记录

## 背景

首轮 JDBC 驱动管理实施完成后，最终验收确认在线驱动安装、托管 JRE 安装和应用内真实 agent/gRPC 重连仍未闭环。进一步审计还发现生产包没有明确部署 agent jar，数据库测试连接仍可能走旧 Go 驱动路径，需要统一收尾计划。

## 范围

- 新增 JDBC 驱动管理闭环实施计划。
- 把 agent 资源打包、supervisor、会话恢复、JRE 在线/离线安装、推荐驱动在线安装、UI 操作和最终验收拆成八个独立任务。
- 为每个行为任务规定失败测试、红灯确认、最小实现、绿灯确认、中文变更记录和独立提交。

## 修改文件

- `docs/plans/2026-07-10-jdbc-driver-management-completion.md`
- `docs/changes/process/2026-07-10-jdbc-driver-management-completion-plan.md`

## 验证

- 已对照 `docs/superpowers/specs/2026-07-09-jdbc-driver-management-design.md` 和首轮最终验证记录检查目标范围。
- 计划覆盖首轮记录中的三项缺口，并补充生产 agent 资源部署与 JDBC 测试连接两项必要前置。
- 所有任务都包含明确文件、测试命令、预期红灯、绿灯和提交边界。

## 剩余风险

- Adoptium API 和公开 JDBC 驱动版本在实施时必须使用官方来源核验，不能仅依赖计划中的接口假设。
- Oracle、达梦、人大金仓等驱动受授权限制，可能只能提供离线导入，不能承诺在线安装。
- agent 崩溃恢复会丢失事务、临时表和进程内数据库状态；首版只自动恢复连接，不重放未提交事务。
