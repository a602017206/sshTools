#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PROTO_DIR="$ROOT_DIR/jdbc-agent/src/main/proto"
OUT_DIR="$ROOT_DIR/internal/service/jdbcproto"

mkdir -p "$OUT_DIR"

PROTOC_BIN="${PROTOC:-protoc}"
export PATH="$PATH:$(go env GOPATH)/bin"

"$PROTOC_BIN" \
  --proto_path="$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" \
  --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/jdbc_agent.proto"
