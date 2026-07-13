# Windows Release 的 JDBC Agent 暂存失败

## 背景

Windows GitHub Runner 已完成 Vite 构建，但 `frontend/package.json` 使用 `../scripts/stage-jdbc-agent.sh` 暂存 JDBC agent。Windows 的命令解释器不支持 `.sh` 脚本和 `../` 命令，因此报出 ` '..' is not recognized as an internal or external command`。

## 范围

将 JDBC agent 暂存改为 Node 脚本，根据当前平台调用 `gradlew` 或 `gradlew.bat`，并使用 Node 文件 API 复制 jar。保留现有 shell 脚本供手工 POSIX 调用，但前端构建不再依赖它。

## 修改文件

- `frontend/package.json`
- `scripts/stage-jdbc-agent.mjs`
- `frontend/test/stageJdbcAgent.test.js`
- 本变更记录。

## 验证

执行前端 Node 测试和 `npm run build`。完整 Windows 矩阵需要重新触发 GitHub Actions 验证。

## 工具链阻塞与最小修复

阻塞点是 Windows `cmd.exe` 无法执行 POSIX shell 脚本，不是 Vite 警告或 Gradle 编译失败。最小修复是使用 Node 18（release workflow 已安装）执行跨平台 staging，不修改 Gradle、protoc 或 gRPC 版本。

## 剩余风险

当前环境无法直接运行 Windows Runner；Windows 的 `gradlew.bat` 和文件复制路径仍需由 GitHub Actions 矩阵确认。
