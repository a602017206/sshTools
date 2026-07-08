# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

sshTools is a cross-platform SSH desktop client built with Go and Wails. It combines a Go backend for SSH functionality with a Svelte frontend for the UI.

## Development Commands

### Wails Development
```bash
# Install frontend dependencies
cd frontend && npm install

# Run in development mode (hot reload)
wails dev

# Build for production
wails build

# Build for specific platform
wails build -platform darwin/arm64
wails build -platform windows/amd64
wails build -platform linux/amd64

# Build for macOS with ad-hoc signing (recommended)
./scripts/build-mac.sh
```

### Go Backend
```bash
# Build Go backend only
go build -o sshTools .

# Run tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Vet code
go vet ./...

# Tidy dependencies
go mod tidy
```

### Frontend Development
```bash
cd frontend

# Install dependencies
npm install

# Run frontend in dev mode
npm run dev

# Build frontend
npm run build
```

## Architecture

### Backend Structure (`/internal`)

- **ssh/**: Core SSH functionality
  - `client.go`: SSH client implementation with connection management
  - `session.go`: SSH session handling, PTY support, and terminal I/O
  - `sftp.go`: SFTP implementation (TODO)
  - `tunnel.go`: Port forwarding/tunneling (TODO)

- **config/**: Configuration management
  - `config.go`: Manages connection configs, app settings, and persistence to ~/.sshtools/config.json

- **store/**: Secure credential storage
  - `credentials.go`: In-memory credential store (TODO: integrate with system keychain)

- **terminal/**: Terminal emulator
  - `terminal.go`: Terminal buffer and rendering (TODO: ANSI escape sequence support)

- **crypto/**: Cryptographic operations (TODO)

### Frontend Structure (`/frontend`)

- Built with Svelte + Vite + xterm.js
- Located in `/frontend/src`
- Communicates with Go backend through Wails bindings and events
- Access backend methods via `wailsjs/go/main/App` imports

**Components:**
- `Terminal.svelte`: xterm.js terminal component with PTY support
- `ConnectionManager.svelte`: SSH connection management UI
- `App.svelte`: Main application layout and session orchestration

### Main Application (`app.go`)

The `App` struct is the main application controller that:
- Manages SSH clients and sessions
- Handles configuration through ConfigManager
- Stores credentials through CredentialStore
- Exposes methods to frontend via Wails bindings

Key methods exposed to frontend:
- `GetConnections()`: Returns all saved SSH connections
- `AddConnection(conn)`: Adds a new connection configuration
- `RemoveConnection(id)`: Removes a connection
- `TestConnection(host, port, user, password)`: Tests SSH connection
- `ConnectSSH(sessionID, host, port, user, password, cols, rows)`: Creates and starts SSH session
- `SendSSHData(sessionID, data)`: Sends data to SSH session
- `ResizeSSH(sessionID, cols, rows)`: Resizes terminal
- `CloseSSH(sessionID)`: Closes SSH session

Events emitted to frontend:
- `ssh:output:{sessionID}`: SSH session output data

## Configuration

- App config stored at: `~/.sshtools/config.json`
- Contains: connection configurations and application settings
- Format: JSON with connections array and settings object

## Key Dependencies

- `github.com/wailsapp/wails/v2`: Desktop application framework
- `golang.org/x/crypto/ssh`: SSH protocol implementation
- Svelte + Vite: Frontend framework and build tool

## Development Notes

- The application uses Wails v2 which embeds the frontend into the Go binary
- Backend methods must be exported (capitalized) to be callable from frontend
- Use `wails dev` for development with hot reload
- Frontend bindings are auto-generated in `frontend/wailsjs/`

## Distribution

### macOS Application Signing

macOS apps require proper handling to avoid "app is damaged" errors on other computers.

**Quick build with ad-hoc signing:**
```bash
./scripts/build-mac.sh
```

This script:
1. Builds the application with `wails build`
2. Removes quarantine attributes with `xattr -cr`
3. Applies ad-hoc code signing with `codesign`

**Distribution:**
```bash
cd build/bin
zip -r sshTools.zip sshTools.app
```

**User instructions:**
After downloading, users should run:
```bash
xattr -cr sshTools.app
open sshTools.app
```

See `MACOS_SIGNING.md` for detailed information about:
- Ad-hoc signing vs. official signing
- Apple Developer Program requirements
- Notarization process
- Common troubleshooting
