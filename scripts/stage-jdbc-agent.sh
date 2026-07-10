#!/bin/sh

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
AGENT_DIR="$ROOT_DIR/jdbc-agent"
SOURCE_JAR="$AGENT_DIR/build/libs/sshtools-jdbc-agent-all.jar"
TARGET_JAR="$ROOT_DIR/frontend/build/jdbc-agent.jar"

(
  cd "$AGENT_DIR"
  ./gradlew shadowJar
)

if [ ! -f "$SOURCE_JAR" ]; then
  echo "未找到 JDBC agent 构建产物: $SOURCE_JAR" >&2
  exit 1
fi

mkdir -p "$(dirname -- "$TARGET_JAR")"
cp "$SOURCE_JAR" "$TARGET_JAR"
echo "已暂存 JDBC agent: $TARGET_JAR"
