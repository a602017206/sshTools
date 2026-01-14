# Repository Guidelines

## Project Structure & Module Organization
- `main.go` and `app.go` bootstrap the Wails desktop app and bind Go services.
- `internal/` contains core backend packages (`ssh/`, `service/`, `api/`, `config/`, `store/`, `terminal/`).
- `cmd/apiserver/` hosts the standalone API server entrypoint.
- `frontend/` is the Svelte + Vite UI, with components in `frontend/src/components/` and stores in `frontend/src/stores/`.
- `flutter_ui/` is an alternate Flutter client for multi-platform builds.
- `build/` and `assets/` hold packaging configs and static resources.

## Build, Test, and Development Commands
- `go install github.com/wailsapp/wails/v2/cmd/wails@latest` installs the Wails CLI.
- `cd frontend && npm install` installs Svelte frontend dependencies.
- `wails dev` runs the desktop app with hot reload.
- `wails build` produces a production build in `build/bin/` (see `wails.json` for targets).
- `go test ./internal/service -v` runs backend unit tests.
- `cd flutter_ui && flutter pub get` sets up the Flutter client.
- `cd flutter_ui && flutter run -d macos` runs the Flutter app (swap device as needed).
- `cd flutter_ui && flutter test` runs Flutter tests.

## Coding Style & Naming Conventions
- Go code follows standard `gofmt`; keep exported symbols in `CamelCase` and unexported in `camelCase`.
- Svelte components use `PascalCase` filenames (e.g., `ConnectionManager.svelte`); store modules stay lower camel case (e.g., `devtools.js`).
- Keep new UI tools in `frontend/src/components/tools/` and register them in `frontend/src/tools/index.js`.

## Testing Guidelines
- Go tests live alongside code as `*_test.go` (see `internal/service/devtools_service_test.go`).
- Flutter tests live under `flutter_ui/test/`.
- For UI regressions, follow the manual checklist in `TESTING_GUIDE.md`.

## Commit & Pull Request Guidelines
- Commit messages follow Conventional Commits (e.g., `feat: add devtools panel`, `fix: handle empty config`).
- PRs should include a short summary, test commands run, and screenshots for UI changes.
- Link related issues or notes when behavior changes impact users.

## Security & Configuration Tips
- App configuration and encrypted credentials are stored in `~/.sshtools/`; do not commit real configs or secrets.
