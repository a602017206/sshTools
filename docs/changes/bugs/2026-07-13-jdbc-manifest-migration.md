# JDBC 历史驱动清单未更新修复

## 背景

JDBC 驱动清单首次创建后持久化到 `~/.sshtools/drivers/manifest.json`。旧版本应用写入的达梦和人大金仓 profile 不含在线下载地址和 SHA-256。后续应用升级时，catalog 仅在清单不存在时写入内置定义，因此用户点击安装仍会看到旧的离线导入 profile。

## 范围

- 内置 JDBC manifest 版本升级到 `2`。
- 低版本历史清单启动时自动合并最新内置 driver/profile。
- 保留未知自定义 driver，以及内置 driver 下不与新 profile ID 冲突的自定义 profile。
- 使用临时文件和原子重命名写回，权限保持 `0600`。

## 修改文件

- `internal/service/jdbc_builtin_manifest.json`
- `internal/service/jdbc_catalog.go`
- `internal/service/jdbc_catalog_test.go`
- `docs/changes/bugs/2026-07-13-jdbc-manifest-migration.md`

## 验证

- `go test ./internal/service -run TestDriverCatalogMigratesOutdatedBuiltinProfiles -v`
- `go test ./internal/service -run 'TestDriverCatalog(MigratesOutdatedBuiltinProfiles|ProvidesVerifiedDomesticOnlineProfiles)' -v`
- `go test ./...`

## 剩余风险

自定义 profile 若与未来内置 profile 使用相同 ID，会以内置 profile 为准。需要保留同 ID 自定义定义时，应使用不同 profile ID。
