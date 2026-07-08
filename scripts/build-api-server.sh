#!/bin/bash

# Build API server script

set -e

echo "Building sshTools API Server..."

# Navigate to project root
cd "$(dirname "$0")/.."

# Build API server
go build -o build/bin/sshtools-api cmd/apiserver/main.go

echo "✓ API server built successfully: build/bin/sshtools-api"
echo ""
echo "To run the server:"
echo "  ./build/bin/sshtools-api"
echo ""
echo "Server will start on http://localhost:8080"
echo "WebSocket endpoint: ws://localhost:8080/api/v1/ws"
