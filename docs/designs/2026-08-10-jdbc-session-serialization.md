# JDBC Session 串行化与 Oracle 元数据 Catalog

## 问题

同一 JDBC session 上并发执行元数据查询与 SQL 查询，会违反 JDBC Connection 线程安全约定。Oracle 场景下常表现为 gRPC `DeadlineExceeded`，前端误报为连接失败。

## 决策

1. Agent 侧为每个 session 增加互斥锁，所有依赖 Connection 的 RPC 经 `withConnection` 串行执行。
2. 前端打开表数据时先完成列元数据再执行默认查询，避免主动制造并发。
3. Oracle/达梦元数据请求不传服务名作为 catalog；无 schema 时跳过列元数据，避免全库扫描。
4. 将 deadline/timeout 映射为独立错误码 `QUERY_TIMEOUT`，文案明确指向超时而非连接失败。

## 非目标

- 不引入连接池或多 Connection 每 session
- 不调整默认 30 秒查询超时
