#!/bin/bash

set -e

echo "=================================="
echo "Building AHaSSHTools for Linux"
echo "=================================="

GIT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
if [ -z "$GIT_VERSION" ]; then
    echo "No git tag found, using default version"
    VERSION=""
else
    VERSION="-ldflags=-X main.Version=$GIT_VERSION"
    echo "Using version from git tag: $GIT_VERSION"
fi

echo "Cleaning previous build..."
rm -rf build/bin/AHaSSHTools

echo "Building application..."
wails build -platform linux/amd64 -clean $VERSION

if [ -f "build/bin/AHaSSHTools" ]; then
    echo ""
    echo "=================================="
    echo "Build complete!"
    echo "Binary location: build/bin/AHaSSHTools"
    echo "=================================="
else
    echo ""
    echo "Error: Build failed, binary not found"
    exit 1
fi
