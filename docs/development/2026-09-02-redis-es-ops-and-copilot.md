# 实现：Redis / ES 运维工作区与 AI 模块

## 摘要

按 C+C 规划完成 P0/P1：两边工作区升级为可日常查改的最小集；Copilot 原生工具支持只读查询与变更提案，写入由前端确认执行。

## 要点

- `NativeResourcePage` + Redis `SCAN MATCH` 分页
- Redis `ExecuteQuery` CLI 白名单；`create_key` / `delete_keys`
- ES Dev Tools 受限路径；`create_index` / `delete_index` / `refresh_index`；查询 from/size clamp
- Artifact 扩展 `native_query` / `native_mutation`
