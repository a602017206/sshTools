#!/bin/bash

APP_PATH="${1:-./build/bin/AHaSSHTools.app}"

if [ -d "$APP_PATH" ]; then
    echo "Removing quarantine attributes from $APP_PATH..."
    xattr -cr "$APP_PATH"

    echo "Ad-hoc signing the app..."
    codesign --force --deep --sign - "$APP_PATH"

    echo "✓ Post-build processing complete"
    echo ""
    echo "To distribute this app, tell users to run:"
    echo "  xattr -cr /path/to/AHaSSHTools.app"
    echo "  sudo spctl --master-disable  # (or allow in System Preferences)"
else
    echo "Error: App not found at $APP_PATH"
    exit 1
fi
