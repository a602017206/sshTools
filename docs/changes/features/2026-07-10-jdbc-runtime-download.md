# JDBC 托管运行时在线安装

## 背景

离线 JRE 导入完成后，普通用户仍需要手动准备归档。设计要求首次使用数据库功能时可以安装托管 Java 21 运行时，并对供应链内容做 checksum 校验。

## 范围

- 新增通用 HTTPS artifact 下载器，限制响应大小并原子提交。
- 使用 Eclipse Adoptium 官方 API 查询当前操作系统和架构的 Temurin Java 21 JRE。
- 使用 API 返回的 SHA-256 校验归档，再复用离线导入流程安装。
- 向 Wails 暴露 `InstallJDBCManagedRuntime`。

## 修改文件

- `app.go`
- `internal/service/artifact_download.go`
- `internal/service/artifact_download_test.go`
- `internal/service/jdbc_runtime.go`
- `internal/service/jdbc_runtime_test.go`
- `internal/service/jdbc_runtime_provider.go`
- `internal/service/jdbc_runtime_provider_test.go`
- `docs/changes/features/2026-07-10-jdbc-runtime-download.md`

## 验证

- 红灯：定向测试在实现前因下载器、provider、运行时包类型和安装方法未定义而失败。
- 绿灯：`go test ./internal/service -run 'TestArtifactDownloader|TestAdoptiumRuntimeProvider|TestRuntimeServiceInstallsManagedRuntime' -v` 通过。
- 回归：`go test ./...` 通过。

## 剩余风险

- 在线安装依赖 Adoptium API 和其下载镜像可用性；离线归档导入仍是必要后备路径。
- 下载器默认上限为 1 GiB，未来超大运行时包需要明确调整，不能静默绕过。
- 官方 API 契约参考 `https://github.com/adoptium/api.adoptium.net/blob/main/docs/cookbook.adoc`；API 变更需要同步更新解析测试。
