# JDBC 驱动管理 Wails API

## 背景

前端驱动管理页面需要通过 Wails 调用 Go 后端能力，获取驱动清单、安装状态、运行时状态，并触发离线导入、校验、删除和 agent 重启等操作。

## 范围

- 新增 `DriverView` 和 `RuntimeStatus` API 模型。
- `DriverCatalogService` 支持合并 manifest 和本地安装目录，返回驱动安装状态。
- `App` 新增 JDBC 驱动管理导出方法：
  - `ListJDBCDrivers`
  - `InstallJDBCDriver`
  - `ImportJDBCDriverPackage`
  - `ValidateJDBCDriver`
  - `RemoveJDBCDriver`
  - `GetJDBCRuntimeStatus`
  - `SetJDBCRuntimeMode`
  - `RestartJDBCAgent`
- 运行 Wails bindings 生成，更新 `frontend/wailsjs`。

## 修改文件

- `app.go`
- `internal/service/jdbc_api_models.go`
- `internal/service/jdbc_catalog.go`
- `internal/service/jdbc_catalog_test.go`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- `frontend/wailsjs/go/models.ts`
- `frontend/wailsjs/runtime/package.json`
- `frontend/wailsjs/runtime/runtime.d.ts`
- `frontend/wailsjs/runtime/runtime.js`

## 验证

- 已运行 `go test ./internal/service -run TestDriverManager -v`，结果通过。
- 已运行 `go test ./...`，结果通过。
- 已运行 `/Users/dingwei/go/bin/wails generate module`，结果通过。
- 已确认 `frontend/wailsjs/go/main/App.js` 和 `App.d.ts` 包含新的 JDBC 驱动管理方法。

## 剩余风险

- `InstallJDBCDriver` 仍返回在线安装未实现错误，首版主要支持离线导入。
- `RestartJDBCAgent` 当前会启动本地 agent 进程，但真实 gRPC client 尚未与进程端口绑定。
- Wails 生成时刷新了 runtime 文件，属于生成器输出变化，后续需要在前端任务中确认构建兼容性。
