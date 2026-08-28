# Elasticsearch 面板元信息与交互增强

## 背景

Elasticsearch 工作区左右分栏不可拖拽，索引列表缺少搜索，对象信息区默认展示文档预览，缺少 mapping / 数据量；连接头部也未展示集群节点等基础信息。

## 范围

- `DescribeIndex` 改为返回索引 stats + mapping，不再预览文档
- 新增 `DescribeNativeDatabaseSession`，返回集群健康、节点数、版本等概览
- 前端：索引名称搜索、可拖拽调整对象信息宽度、头部展示集群信息、右侧展示 mapping 与数据量

## 修改文件

- `internal/service/native_elasticsearch.go`
- `internal/service/native_database_operations.go`
- `internal/service/native_elasticsearch_test.go`
- `app.go`
- `frontend/src/components/NativeDatabasePanel.svelte`
- `frontend/src/lib/nativeDatabaseOperations.js`
- `frontend/src/lib/nativeDatabaseWorkspace.js`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- 相关测试

## 验证

- [x] `go test ./internal/service/... -run Elasticsearch`
- [x] `node --test frontend/test/nativeDatabaseOperations.test.js frontend/test/nativeDatabaseWorkspace.test.js`
- [ ] 手工：拖拽分栏、搜索索引、查看 mapping/数据量与集群节点信息

## 剩余风险

- 部分 ES 版本/权限可能拒绝 `_cat/nodes` 或 `_mapping`；错误会显示在对象信息区，不影响查询页签
- 超大 mapping 仍有 2MB 读取上限
