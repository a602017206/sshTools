# Native Data Platform Operations Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 Redis、Elasticsearch 和 Kafka 原生连接提供受限、只读的资源详情与诊断能力。

**Architecture:** 扩展统一 `NativeDatabaseClient` 详情契约，并由三类 provider 分别实现协议访问。App 暴露单一 Wails 方法，Svelte 工作区在资源选中后调用并呈现标准化 JSON 详情。

**Tech Stack:** Go、go-redis/v9、go-elasticsearch/v8、franz-go、Wails、Svelte。

---

### Task 1: 定义统一详情契约

**Files:**
- Modify: `internal/service/native_database.go`
- Test: `internal/service/native_database_test.go`

**Step 1:** 写失败测试，要求服务按会话返回 `NativeResourceDetails` 并委托到客户端。

**Step 2:** 运行 `go test ./internal/service -run TestNativeDatabaseService.*Describe -v`，确认因缺少 API 失败。

**Step 3:** 添加 `NativeResourceDetails` 与 `DescribeResource`，并在 service 中包装错误。

**Step 4:** 重跑上述测试，确认通过。

### Task 2: 实现三类 provider 的详情

**Files:**
- Modify: `internal/service/native_redis.go`
- Modify: `internal/service/native_elasticsearch.go`
- Modify: `internal/service/native_kafka.go`
- Test: `internal/service/native_redis_test.go`
- Test: `internal/service/native_elasticsearch_test.go`
- Test: `internal/service/native_kafka_test.go`

**Step 1:** 分别写失败测试：Redis 键预览、ES 索引文档预览、Kafka Topic 分区元数据。

**Step 2:** 运行相应 `go test`，确认测试因接口不存在失败。

**Step 3:** 最小化扩展客户端适配器与会话实现；限制 Redis/ES 返回量，Kafka 仅读元数据。

**Step 4:** 运行 `go test ./internal/service -run 'Test(Redis|Elasticsearch|Kafka)Native' -v`，确认通过。

### Task 3: 暴露并显示详情

**Files:**
- Modify: `app.go`
- Modify: `frontend/src/components/NativeDatabasePanel.svelte`
- Modify: `frontend/src/lib/nativeDatabaseWorkspace.js`
- Test: `frontend/src/lib/nativeDatabaseWorkspace.test.js`

**Step 1:** 写前端配置失败测试，声明支持资源详情。

**Step 2:** 运行 `npm test -- nativeDatabaseWorkspace.test.js`，确认失败。

**Step 3:** 添加 Wails 详情方法；点击资源时加载并显示 JSON 详情与加载/错误状态。

**Step 4:** 运行前端测试和构建。

### Task 4: 验证与记录

**Files:**
- Create: `docs/changes/features/2026-08-21-native-data-platform-operations.md`
- Create: `docs/development/2026-08-21-native-data-platform-operations.md`

**Step 1:** 运行 `go test ./internal/service -v`、`go test ./...` 与 `cd frontend && npm run build`。

**Step 2:** 在变更记录中写明背景、范围、修改文件、验证和剩余风险。

**Step 3:** 在实现记录中说明接口、限制和验证结果。
