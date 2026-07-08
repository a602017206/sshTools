# Codebase Scan - 2026-07-08

## Scope

Scanned the Go/Wails SSH tool repository with emphasis on build/test health, configuration persistence, documentation sprawl, and currently visible working tree changes.

## Findings

### Fixed: fallback config manager could not save in memory

`app.go` falls back to `config.NewFallbackConfigManager()` when disk-backed config initialization fails. The fallback manager uses an empty `configPath`, but `ConfigManager.Save()` still attempted to write to that empty path. Any connection or settings save in fallback mode failed with `open : no such file or directory`.

Resolution: `Save()` now treats an empty `configPath` as non-persistent in-memory mode and returns success after updating the in-memory config.

### Documentation sprawl

The repository root contained active guides mixed with temporary debugging notes and completed implementation summaries. This made the active documentation entry points harder to identify.

Resolution: one-off and historical documents were moved to `docs/archive/`; `docs/README.md` now defines the documentation structure.

### Existing uncommitted changes

The working tree already contained changes in frontend, SFTP/session tracking, config, and tests before this scan. Those changes were preserved and not reverted.

### Known frontend build warnings

`npm run build` succeeds, but Svelte reports existing accessibility and unused export warnings in several components:

- `frontend/src/components/AssetList.svelte`
- `frontend/src/components/FileManager.svelte`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/src/components/DevToolsPanel.svelte`
- `frontend/src/components/ui/Dialog.svelte`
- `frontend/src/components/JsonFormatter.svelte`
- `frontend/src/components/Base64Tool.svelte`
- `frontend/src/components/HashTool.svelte`
- `frontend/src/components/TimestampTool.svelte`
- `frontend/src/components/UuidTool.svelte`
- `frontend/src/components/UrlTool.svelte`

The warnings are primarily non-interactive elements with click handlers, labels without associated controls, and unused `themeStore` exports. Vite also reports a large `index.js` chunk and mixed dynamic/static imports of Wails runtime modules.

These warnings were inventoried but not fixed in this pass because they cut across many UI components and should be handled as a dedicated accessibility/build hygiene change.

## Verification

- Added a regression test for fallback config in-memory saves.
- Verified the regression test failed before the fix.
- Verified the regression test passed after the fix.
- `GOCACHE="$(pwd)/.cache/go-build" go test ./...` passed.
- `GOCACHE="$(pwd)/.cache/go-build" go vet ./...` passed.
- `npm run build` passed with the known frontend warnings listed above.

## Follow-up Recommendations

- Review existing frontend build artifacts under `frontend/build/` and `build/frontend/` to decide whether generated assets should remain tracked.
- Add focused tests around SSH session cwd tracking if that behavior continues to evolve.
- Consider consolidating `CLAUDE.md` and `AGENTS.md` if both are meant to describe the same agent workflow.
- Address Svelte accessibility warnings in a dedicated UI quality pass.
