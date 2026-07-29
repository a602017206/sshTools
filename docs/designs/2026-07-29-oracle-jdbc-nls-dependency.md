# Oracle JDBC 字符集依赖设计

## 背景

Oracle 数据库使用 `ZHS16GBK` 等非基础字符集时，单独加载 `ojdbc11.jar` 会触发 ORA-17056。Oracle 将 NLS/全球化支持作为独立的 `orai18n.jar` 发布。

## 方案

- Oracle 推荐 profile 固定安装同版本的 `ojdbc11` 与 `orai18n`。
- 内置清单版本升级，使既有本地清单合并新的必需 jar。
- 安装状态不能只检查版本目录存在；必须检查该 profile 定义的每个 jar 都存在且为普通文件。
- 缺失 NLS jar 的旧安装显示为未安装，用户点击安装即可执行原有的原子重装流程。

## 风险

字符集扩展 jar 体积会增加 Oracle 驱动下载量。profile 使用固定 SHA-256 校验；上游内容与记录不一致时将拒绝安装而非加载未经验证的 jar。
