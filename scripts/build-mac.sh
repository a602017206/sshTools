#!/bin/bash

# Build script for macOS with proper signing
# This script builds the app and applies ad-hoc signing

set -e

echo "=================================="
echo "Building AHaSSHTools for macOS"
echo "=================================="

# Clean previous build
echo "Cleaning previous build..."
rm -rf build/bin/AHaSSHTools.app

# Build the app
echo "Building application..."
wails build -clean

# Post-build processing
APP_PATH="./build/bin/AHaSSHTools.app"

if [ ! -d "$APP_PATH" ]; then
    echo "❌ Error: Build failed, app not found at $APP_PATH"
    exit 1
fi

echo ""
echo "Post-processing application..."

# Remove quarantine attributes
echo "→ Removing quarantine attributes..."
xattr -cr "$APP_PATH"

# Ad-hoc sign the app
echo "→ Applying ad-hoc signature..."
codesign --force --deep --sign - "$APP_PATH" 2>/dev/null || {
    echo "⚠️  Warning: codesign failed, but app should still work"
}

# Verify signature
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
echo "1. Compress the app:"
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
