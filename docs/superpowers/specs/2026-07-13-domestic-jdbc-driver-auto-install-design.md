# 国产 JDBC 驱动自动安装设计

## 背景

达梦和人大金仓在内置 JDBC 清单中仅配置离线导入提示，未提供下载地址与 SHA-256 校验值。驱动管理页因此无法执行自动安装。

达梦官方文档确认 `com.dameng:DmJdbcDriver8` 已发布到 Maven Central。人大金仓官方文档确认 `cn.com.kingbase:kingbase8` 已发布到 Maven Central。两者均可通过 HTTPS 下载并由应用进行 SHA-256 完整性校验。

人大金仓 V8 与 V9 的 JDBC 常规接口大体兼容，但读写分离场景存在 `nodelist` 参数差异，且 V9 有别名驱动类和 URL 形态。不能将 V9 profile 静默替代 V8 profile。

## 目标

- 达梦 DM8 驱动可从 Maven Central 自动安装。
- 人大金仓 V8 和 V9 驱动可分别安装、校验、卸载与选择。
- 数据库连接保存所选 profile，连接时使用该 profile 而非总是使用推荐版本。
- 每个在线安装 jar 都有 HTTPS 来源和固定 SHA-256；离线导入继续保留。

## 来源与版本

| 数据库 | Profile | Maven 坐标 | 用途 |
| --- | --- | --- | --- |
| 达梦 | DM8 `8.1.5.45` | `com.dameng:DmJdbcDriver8:8.1.5.45` | DM8 默认自动安装版本 |
| 人大金仓 | V8 `8.6.1` | `cn.com.kingbase:kingbase8:8.6.1` | 默认推荐，适配现有 V8 连接 |
| 人大金仓 | V9 `9.0.1` | `cn.com.kingbase:kingbase8:9.0.1` | V9 服务端的独立可选 profile |

应用使用 Java 21，因此选用无 `.jre6`、`.jre7` 后缀的 JDK 8 及以上 jar。

## 架构与数据流

1. 内置 manifest 为上述 profile 写入 Maven Central jar 地址、SHA-256、驱动类和 URL 模板。
2. 驱动管理页沿用现有 profile 选择控件，按选中的 profile 安装、校验或卸载。
3. 数据库连接表单读取已安装 profile，并让用户选择与服务端一致的 profile。
4. `DatabaseConfig.DriverProfileID` 保存该选择；连接建立时 profile resolver 优先使用该 ID，缺失时才使用驱动的推荐 profile。
5. 安装器下载到临时文件，校验 SHA-256 后原子提交；下载、校验或驱动加载失败时保留可读错误并支持离线导入。

## 兼容性策略

- 人大金仓 V8 默认使用 V8 profile，不因 V9 profile 已安装而自动升级。
- V9 profile 仅在用户明确选择 V9 时使用。
- V8/V9 使用各自的 driver class 与 URL template，以 manifest 定义为准。
- 达梦只提供 DM8 profile；若服务端与该驱动不兼容，用户通过离线导入厂商对应版本处理。

## 测试与验收

- 清单测试：三个 profile 都包含 HTTPS URL、64 位 SHA-256 和正确的驱动元数据。
- 安装器测试：使用本地 HTTP 测试服务验证下载、校验和安装目录。
- 连接配置测试：指定 `DriverProfileID` 时 resolver 返回对应 profile；未指定时返回推荐 profile。
- 前端契约测试：连接表单包含已安装 profile 的选择入口，并在驱动未安装时提示正确操作。
- 验收：驱动管理页可分别安装金仓 V8、V9 与达梦 DM8；新建连接可选择 V8 或 V9，建立请求带入所选 profile。

## 非目标

- 不自动探测服务端是人大金仓 V8 还是 V9。
- 不修改 gRPC/proto 或 JDBC agent。
- 不将需要登录、验证码或授权的厂商下载页作为无人值守安装源。
