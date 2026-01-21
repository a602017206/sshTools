# CODEBUDDY.md
This file provides guidance to CodeBuddy when working with code in this repository.

## Common Development Commands

### Application Development
- `go install github.com/wailsapp/wails/v2/cmd/wails@latest` - Install Wails CLI (required for building the desktop app)
- `cd frontend && npm install` - Install Svelte frontend dependencies
- `wails dev` - Run desktop app in development mode with hot reload
- `wails build` - Build production binary to `build/bin/`
- `wails build -platform darwin/arm64` - Build for macOS Apple Silicon; swap target for other platforms

### Backend Testing
- `go test ./...` - Run all Go tests
- `go test -v ./internal/service` - Run service tests with verbose output
- `go test -v ./internal/service -run TestFormatJSON` - Run specific test function
- `go test ./internal/service -cover` - Run tests with coverage report

### Go Utilities
- `go fmt ./...` - Format Go code using gofmt
- `go vet ./...` - Run Go vet for code analysis
- `go mod tidy` - Clean up Go module dependencies

### macOS Distribution
- `./scripts/build-mac.sh` - Build with ad-hoc signing for distribution (removes quarantine attributes)

### Flutter Client
- `cd flutter_ui && flutter pub get` - Install Flutter dependencies
- `cd flutter_ui && flutter run -d macos` - Run Flutter app on macOS (change device as needed)
- `cd flutter_ui && flutter test` - Run Flutter tests

## Architecture Overview

This is a cross-platform SSH desktop client built with Go backend and Wails framework. The app combines SSH terminal emulation, SFTP file management, system monitoring, and an extensible developer tools toolkit in a unified UI.

### Backend Architecture (`internal/`)

**SSH Core (`ssh/`)**: Handles all SSH protocol operations
- `client.go` - SSH client with connection management and auth (password, keyboard-interactive, key)
- `session.go` - PTY session handling, terminal I/O, and bidirectional communication
- `manager.go` - Manages multiple SSH sessions concurrently
- `sftp.go` - SFTP client for file operations (upload, download, delete, rename, mkdir)
- `transfer.go` - File transfer task management with progress tracking and cancellation
- `monitor.go` - Real-time system performance monitoring (CPU, memory, disk, network)

**Service Layer (`service/`)**: Business logic exposed to frontend
- `devtools_service.go` - Developer toolkit backend with JSON formatting/validation/minification/escaping
- `devtools_service_test.go` - 24 unit tests for devtools functions

**Configuration & Storage**:
- `config/` - Manages connection configs and app settings, persists to `~/.sshtools/config.json`
- `store/` - In-memory credential storage; passwords encrypted with AES-256-GCM, stored at `~/.sshtools/credentials.enc`
- `crypto/` - Cryptographic operations (AES-GCM encryption, key derivation from machine features)

**Terminal (`terminal/`)**: Terminal emulation layer for xterm.js integration

### Frontend Architecture (`frontend/src/`)

Built with Svelte + Vite, communicating with Go backend via Wails bindings.

**Component Structure**:
- `App.svelte` - Main layout orchestrating tabs, connection sidebar, and right-side panels
- `TabBar.svelte` - Browser-style horizontal tab bar with renaming, close confirmations, auto-switching
- `Terminal.svelte` - xterm.js terminal emulator with PTY support and real-time output
- `ConnectionManager.svelte` - SSH connection CRUD operations with test/save/delete
- `FileManager.svelte` - Collapsible right-side SFTP panel with breadcrumb nav, file operations, transfer progress
- `MonitorPanel.svelte` - Real-time system metrics (CPU per-core, memory, disk partitions, network I/O)
- `DevToolsPanel.svelte` - Collapsible toolkit panel with tool registration system
- `tools/JsonFormatter.svelte` - JSON formatting tool with validation, syntax highlighting, minification

**State Management (`stores/`)**:
- `theme.js` - Light/dark theme state
- `fileManager.js` - File manager panel state (collapsible, width, current path, file list)
- `monitor.js` - Monitor panel state and metrics data
- `devtools.js` - Toolkit state with tool registry system

**Tool Extension System**:
- New tools registered in `frontend/src/tools/index.js`
- Each tool is a Svelte component in `frontend/src/components/tools/`
- Tools are discovered and rendered dynamically in DevToolsPanel
- Registration requires: id, name, icon, component, category, order

### Application Entry (`app.go`)

The `App` struct serves as the main controller:
- Exports Go methods to frontend via Wails bindings (must be capitalized)
- Manages SSH session lifecycle through SessionManager
- Handles ConfigManager for config persistence
- Provides CredentialStore for encrypted password storage
- Emits events to frontend: `ssh:output:{sessionID}` for terminal data

Key exported methods include: GetConnections, AddConnection, RemoveConnection, TestConnection, ConnectSSH, SendSSHData, ResizeSSH, CloseSSH, and devtools methods (FormatJSON, ValidateJSON, MinifyJSON, EscapeJSON).

### Configuration & Data Persistence

Stored at `~/.sshtools/`:
- `config.json` - Connection configs, app settings (theme, sidebar/panel widths), never commit this file
- `credentials.enc` - AES-GCM encrypted passwords with machine-bound key derivation

## Development Guidelines

### Adding New Developer Tools
1. Create Svelte component in `frontend/src/components/tools/YourTool.svelte`
2. Add backend methods in `internal/service/devtools_service.go`
3. Expose methods in `app.go` with exported names
4. Register tool in `frontend/src/tools/index.js` with id, name, icon, component, order
5. Frontend auto-regenerates bindings on `wails dev`

### Code Style
- Go: Exported symbols in `CamelCase`, unexported in `camelCase`, follow `gofmt`
- Svelte: Components use `PascalCase` filenames (e.g., `ConnectionManager.svelte`), store modules use `camelCase`

### Testing
- Go unit tests alongside code as `*_test.go` files
- Run `go test ./internal/service -v` before committing backend changes
- UI regressions checked manually per `TESTING_GUIDE.md`

## Platform-Specific Notes

### macOS Distribution
To avoid "app is damaged" errors when distributing:
- Run `./scripts/build-mac.sh` which applies ad-hoc signing and removes quarantine attributes
- Build output at `build/bin/AHaSSHTools.app`
- Distribute as zip; users run `xattr -cr sshTools.app && open sshTools.app` after download
- For official distribution, requires Apple Developer Program signing and notarization (see `MACOS_SIGNING.md`)

## Security Considerations
- Never commit `~/.sshtools/` directory contents (contains real configs and encrypted credentials)
- SSH key passphrases are never stored to disk
- Password encryption uses AES-256-GCM with machine-derived key
- All communication happens locally via Wails bindings (no network exposure)
