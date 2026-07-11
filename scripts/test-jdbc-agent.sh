#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

(cd jdbc-agent && ./gradlew clean test shadowJar)

H2_JAR="$(cd jdbc-agent && ./gradlew -q printH2Jar)"
if [[ -z "$H2_JAR" ]]; then
  echo "未找到 H2 JDBC jar，请先运行 jdbc-agent Gradle 测试下载依赖。" >&2
  exit 1
fi

JDBC_AGENT_JAR="$PWD/jdbc-agent/build/libs/sshtools-jdbc-agent-all.jar" \
H2_JAR="$H2_JAR" \
go test -tags=integration ./internal/service -run 'TestJDBCAgent(H2EndToEnd|RecoversSessionAfterCrash)' -v
