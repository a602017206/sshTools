# JDBC 离线驱动包导入

## 背景

JDBC agent 架构需要支持在无外部网络依赖的环境中安装数据库驱动。离线驱动包导入能力用于把 profile 元数据、jar 文件和 checksum 一起安装到本地驱动目录，作为后续连接和 agent classloader 加载的基础。

## 范围

- 新增 `DriverInstallService`，支持导入 zip 格式离线驱动包。
- 读取 `package.json` 获取驱动 ID、版本、驱动类、URL 模板、JRE 要求和 jar 列表。
- 读取 `checksums.sha256` 并校验包内 jar 文件内容。
- 安装 jar 到 `drivers/<id>/<version>/jars/`，并写入 `driver.json`。
- checksum 校验或安装失败时清理目标目录。

## 修改文件

- `internal/service/jdbc_install.go`
- `internal/service/jdbc_install_test.go`

## 验证

- 已运行 `go test ./internal/service -run 'TestDriverInstallImportsOfflinePackageAndValidatesChecksum|TestDriverInstallRollsBackOnChecksumMismatch' -v`，结果通过。

## 剩余风险

- 当前离线包格式只覆盖首版最小字段，尚未与主驱动清单合并。
- 当前安装前会清理同 ID、同版本目标目录，后续需要结合版本升级和用户确认策略完善。
- zip 路径安全策略仅按计划实现最小导入路径，后续导入真实用户包前应增加路径穿越防护和包格式校验。
