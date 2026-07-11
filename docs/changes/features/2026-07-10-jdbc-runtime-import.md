# JDBC 托管运行时离线导入

## 背景

`RuntimeService.ImportRuntimeArchive` 原先只返回“暂未实现”，离线环境无法导入 JRE。直接解压第三方归档还存在路径穿越、符号链接和残留半安装目录风险。

## 范围

- 支持 `.zip`、`.tar.gz` 和 `.tgz` JRE 归档。
- 只允许普通文件和目录，拒绝绝对路径、父目录穿越、符号链接及其他特殊文件。
- 限制归档展开总量为 2 GiB。
- 自动定位唯一 `bin/java`，设置可执行权限并原子提交到平台相关目录。
- 无效归档或解压失败时删除全部临时内容。

## 修改文件

- `internal/service/archive_extract.go`
- `internal/service/archive_extract_test.go`
- `internal/service/jdbc_runtime.go`
- `internal/service/jdbc_runtime_test.go`
- `docs/changes/features/2026-07-10-jdbc-runtime-import.md`

## 验证

- 红灯：定向测试在实现前因 `ExtractArchive` 未定义而失败。
- 绿灯：`go test ./internal/service -run 'TestRuntimeServiceImports|TestArchiveExtractor|TestRuntimeServiceRollsBack' -v` 通过。
- 回归：`go test ./internal/service` 通过。

## 剩余风险

- 当前要求归档中只有一个 `bin/java`；同时携带多个可执行运行时的复合 JDK 包会被拒绝，需要用户导入单一 JRE/JDK 包。
- Windows 文件可执行权限语义与 Unix 不同，需要在 Windows 构建机补充归档导入验收。
