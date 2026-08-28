# Sonar 高优先级问题修复实施记录

## 实施原则

每个源文件按“测试基线、最小改动、包级验证”执行。对没有可独立自动化测试入口的展示层代码，使用前端构建和既有测试作为回归检查。

## 验证记录

已完成以下低风险规则修复：

- `S1192`：统一 API 请求体错误文案、JDBC Agent 不可用文案、Copilot schema reader 文案、SFTP 进度事件前缀、JDBC Agent supervisor 未初始化文案和 JDBC `COLUMN_DEF` 列名。
- `S1186`：为 Windows 命令配置空实现、macOS 菜单回调与 Java 测试观察者的空完成回调补充原因说明。
- `S7761`：将外观设置中的 `data-*` 属性操作改为 `dataset` 操作。

已执行的验证：

- `go test ./internal/api/handlers ./internal/service ./internal/service/copilot ./internal/ssh`：通过。
- `jdbc-agent` 的 `MetadataServiceImplTest`、`QueryServiceImplTest`、`HealthServiceImplTest` 定向 Gradle 测试：通过。
- 根包定向测试仍失败，失败项为既有的 `TestAddAssetDialogPersistsSelectedJDBCProfile` 与 `TestAppPassesSavedJDBCProfileToDatabaseConnection`；两项均检查当前工作区前端源码中缺失的 JDBC Profile 代码，不属于本次修改范围。

尚未处理的 33 条 `S3776` 均涉及 SSH、SFTP、数据库、配置加载、SQL 生成或应用编排的复杂控制流，需要先补足对应行为测试后再拆分。

## 第二轮实施

新扫描仅保留 35 条 `S3776`。本轮已先处理具备既有回归测试的低耦合服务：

- 将开发工具的 AES-GCM、AES-CBC、SM4-CBC 加解密分派到算法专用辅助函数，并复用 CBC 操作；输入校验、密文格式和错误文案保持不变。
- 将 tar.gz 解压的读取、大小校验、条目分派和普通文件写入拆分为私有函数；路径穿越、符号链接和展开大小限制保持不变。
- 将制品下载拆分为地址校验、HTTP 响应校验和临时文件提交步骤；HTTPS/重定向、大小限制、哈希校验和权限收紧逻辑保持不变。
- 将 JSON 格式化、压缩和转义的表驱动测试断言提取为共享辅助函数；测试用例与覆盖的行为保持不变。
- 将 Adoptium 运行时发现拆分为平台解析、URL 构造、响应解码和候选包选择；HTTP 请求头、查询条件及无匹配时的错误保持不变。

已执行的验证：

- `go test ./internal/service -run 'TestEncryptDecrypt(AESGCM|AESCBC|SM4CBC)$'`：通过。
- `go test ./internal/service -run '^TestArtifactDownloader'`：通过。
- `go test ./internal/service -run '^TestArchiveExtractor'`：通过。
- `go test ./internal/service -run '^(TestFormatJSON|TestMinifyJSON|TestEscapeJSON)$'`：通过。
- `go test ./internal/service -run '^TestAdoptiumRuntimeProvider'`：通过。
