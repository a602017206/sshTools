# JDBC 在线驱动重新安装

## 背景

驱动管理界面在已安装版本上提供“重新安装”操作，但在线安装服务检测到目标版本目录已存在时直接返回错误，导致界面操作无法完成。

## 范围

- 在线驱动的全部 jar 下载并通过 SHA-256 后，允许替换同版本现有目录。
- 提交前把旧版本移动到同一父目录的唯一备份路径。
- 新版本目录提交失败时恢复旧版本；下载或校验失败时完全不触碰旧版本。
- 重新安装成功后清除旧目录中的遗留文件。

## 修改文件

- `internal/service/jdbc_install.go`
- `internal/service/jdbc_install_test.go`

## 验证

- `go test ./internal/service -run TestDriverInstallDownloadsProfileJarsAtomically -v`
- `go test ./...`

## 剩余风险

- 目录替换依赖安装目录位于同一文件系统，以保证 `os.Rename` 的原子目录移动语义。
- 进程在旧目录移到备份后、提交新目录前被强制终止时，备份目录需要人工恢复；正常错误路径会自动恢复。
