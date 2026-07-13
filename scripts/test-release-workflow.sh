#!/usr/bin/env bash
set -euo pipefail

if rg -q '^org\.gradle\.java\.home=' jdbc-agent/gradle.properties; then
  echo 'Gradle 配置不能固定机器专属 org.gradle.java.home' >&2
  exit 1
fi

rg -U 'uses: actions/setup-java@v4\n[\s\S]*distribution: temurin\n[\s\S]*java-version: .21.' .github/workflows/release.yml >/dev/null || {
  echo 'Release workflow 必须安装 Temurin JDK 21' >&2
  exit 1
}
