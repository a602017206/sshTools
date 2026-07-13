# Release Workflow 的 Gradle Java 路径失败

## 背景

GitHub Actions 打包在 Gradle 初始化阶段失败：`org.gradle.java.home` 固定为开发机的 macOS JDK 路径 `/Library/Java/JavaVirtualMachines/jdk-21.jdk/Contents/Home`。该路径在 GitHub Runner，尤其是 Windows Runner 上无效。

## 范围

删除机器专属 Gradle Java 路径，release workflow 显式安装 Temurin JDK 21，并增加静态检查，防止重新提交固定本机 Java 路径。

## 修改文件

- `jdbc-agent/gradle.properties`
- `.github/workflows/release.yml`
- `scripts/test-release-workflow.sh`
- 本变更记录。

## 验证

运行 `./scripts/test-release-workflow.sh`；在本机运行 `./jdbc-agent/gradlew test` 验证 Gradle 使用环境提供的 JDK 21。

## Gradle 工具链阻塞与最小修复

阻塞点是 Runner 没有开发机的绝对 JDK 路径，而不是 Gradle 8.5 解压或 Svelte 警告。最小修复是移除 `org.gradle.java.home`，由 `actions/setup-java` 写入跨平台的 `JAVA_HOME`；不修改 Gradle、protoc 或 gRPC 版本。

## 剩余风险

当前无法在本机直接执行 GitHub Runner 的 Windows 和 macOS 矩阵。配置检查只能保证 JDK 注入和绝对路径约束，完整证据仍需要重新触发 GitHub Actions。
