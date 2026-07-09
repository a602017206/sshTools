# JDBC 元数据模型和本地目录约定

## 背景

数据库模块将从 Go `database/sql` 直连模式迁移到本地 JDBC agent 架构。首批工作需要先定义 JDBC 驱动清单、驱动 profile、jar 校验信息和本地目录约定，供后续驱动导入、运行时选择和 agent 启动复用。

## 范围

- 新增 JDBC 驱动清单和 profile 数据模型。
- 新增 JDBC 本地目录路径构造。
- 新增驱动清单加载服务，并支持按驱动 ID 选择推荐 profile。
- 在数据库连接配置中预留 `DriverProfileID` 和额外连接属性。

## 修改文件

- `internal/config/jdbc.go`
- `internal/config/database.go`
- `internal/service/jdbc_paths.go`
- `internal/service/jdbc_catalog.go`
- `internal/service/jdbc_catalog_test.go`

## 验证

- 已运行 `go test ./internal/service -run TestDriverCatalogLoadsManifestAndSelectsRecommendedProfile -v`，结果通过。
- 首次运行测试时受到 Go 构建缓存目录权限限制，最小修复方案是授权 `go test` 写入用户级构建缓存后重跑。

## 剩余风险

- 当前仅加载主清单，尚未合并本地已安装驱动状态。
- profile 选择逻辑只覆盖推荐版本和首个 profile 回退，后续需要结合安装状态和前端选择策略完善。
