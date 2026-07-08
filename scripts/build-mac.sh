#!/bin/bash

set -e

echo "=================================="
echo "Building AHaSSHTools for macOS"
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
rm -rf build/bin/AHaSSHTools.app

echo "$VERSION"

echo "Building application..."

wails build -clean "$VERSION"

APP_PATH="./build/bin/AHaSSHTools.app"

sleep 5

if [ ! -d "$APP_PATH" ]; then
    echo "❌ Error: Build failed, app not found at $APP_PATH"
    exit 1
fi

echo ""
echo "Post-processing application..."

echo "→ Removing quarantine attributes..."
xattr -cr "$APP_PATH"

echo "→ Applying ad-hoc signature..."
codesign --force --deep --sign - "$APP_PATH" 2>/dev/null || {
    echo "⚠️  Warning: codesign failed, but app should still work"
}

echo "→ Verifying signature..."
codesign -dvv "$APP_PATH" 2>&1 | head -5

echo ""
echo "✅ Build complete!"
echo ""
echo "App location: $APP_PATH"
echo ""
echo "=================================="
echo "Distribution Instructions"
echo "=================================="
echo ""
echo "1. Compress app:"
echo "   cd build/bin && zip -r AHaSSHTools.zip AHaSSHTools.app"
echo ""
echo "2. Tell users to extract and run:"
echo "   xattr -cr AHaSSHTools.app"
echo "   open AHaSSHTools.app"
echo ""
echo "3. Or users can allow in System Preferences:"
echo "   System Preferences → Security & Privacy → General"
echo "   Click 'Open Anyway' after first launch attempt"
echo ""
