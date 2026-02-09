# Build Scripts

This directory contains platform-specific build scripts for AHaSSHTools.

## Scripts

### macOS

- **build-mac.sh** - Build for macOS with ad-hoc signing and quarantine removal
  ```bash
  ./scripts/build-mac.sh
  ```

### Windows

- **build-windows.bat** - Build for Windows
  ```cmd
  scripts\build-windows.bat
  ```

### Linux

- **build-linux.sh** - Build for Linux
  ```bash
  ./scripts/build-linux.sh
  ```

## Version Handling

All build scripts automatically detect and inject version from git tags:

1. **With git tag**: Uses latest tag (e.g., `v0.0.8`)
2. **Without git tag**: Uses default `"dev"` version

To create a new release:

```bash
git tag v1.0.0
git push origin v1.0.0
# Then run the build script
```

## GitHub Actions

When pushing tags to GitHub, the CI/CD workflow automatically:
- Detects the tag name
- Injects version via `-ldflags`
- Builds for Windows, macOS (arm64/amd64), and Linux
- Creates GitHub release with artifacts

## Post-Build Scripts

- **post-build-darwin.sh** - Remove quarantine attributes and ad-hoc sign macOS builds
  ```bash
  ./scripts/post-build-darwin.sh ./build/bin/AHaSSHTools.app
  ```
