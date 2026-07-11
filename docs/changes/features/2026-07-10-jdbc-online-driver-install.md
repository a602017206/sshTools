# JDBC 推荐驱动在线安装

## 背景

JDBC 驱动管理此前仅支持离线压缩包导入，内置推荐驱动清单和在线安装入口尚未落地。用户首次启动时也可能没有本地清单文件，导致驱动管理页面无法列出支持的数据库类型。

## 范围

- 内嵌首批八类数据库的推荐驱动清单，并在用户清单不存在时自动写入。
- 为 MySQL、PostgreSQL、SQLite、SQL Server 和 openGauss 配置 Maven Central 固定版本下载地址及 SHA-256。
- Oracle、达梦数据库和人大金仓保留推荐 profile，但不配置在线地址，安装时明确提示使用离线导入。
- 在线安装先在临时版本目录下载并校验全部 jar，全部成功后再原子提交目标版本目录。
- 应用层 `InstallJDBCDriver` 按数据库类型和版本解析 profile，并调用真实在线安装服务。

## 修改文件

- `internal/service/jdbc_builtin_manifest.json`
- `internal/service/jdbc_catalog.go`
- `internal/service/jdbc_catalog_test.go`
- `internal/service/jdbc_install.go`
- `internal/service/jdbc_install_test.go`
- `app.go`

## 验证

- `go test ./internal/service -run 'TestDriverCatalogBootstraps|TestDriverInstallDownloads|TestDriverInstallRejects' -v`
- `go test ./...`
- 固定版本 jar 从 Maven Central 下载后使用 `shasum -a 256` 计算清单校验值。
- openGauss 驱动类和 URL 前缀依据 openGauss 官方文档确认，公开构件依据 Maven Central 元数据确认。

## 剩余风险

- 内置版本为经过校验的固定版本，不会自动跟随上游新版本；升级时必须重新核对许可证、下载地址和 SHA-256。
- Maven Central 或网络不可用时，公开驱动在线安装仍会失败，用户需要改用离线驱动包。
- 受厂商分发条款限制的驱动不提供在线下载，离线包的取得和授权由用户负责。
