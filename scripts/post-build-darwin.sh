#!/bin/bash

# Post-build script for macOS to remove quarantine attributes
# This helps prevent "app is damaged" error on other Macs

APP_PATH="./build/bin/sshTools.app"

if [ -d "$APP_PATH" ]; then
    echo "Removing quarantine attributes from $APP_PATH..."
    xattr -cr "$APP_PATH"

    # Ad-hoc sign the app (works without Apple Developer account)
    echo "Ad-hoc signing the app..."
    codesign --force --deep --sign - "$APP_PATH"

    echo "✓ Post-build processing complete"
    echo ""
    echo "To distribute this app, tell users to run:"
    echo "  xattr -cr /path/to/sshTools.app"
    echo "  sudo spctl --master-disable  # (or allow in System Preferences)"
else
    echo "Error: App not found at $APP_PATH"
    exit 1
fi
