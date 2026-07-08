# Feature: Ops Console Visual Refresh

## Background

The previous interface felt more like a generic desktop panel layout than a daily operations console. The requested direction is terminal-first, practical, compact, and theme-aware.

## Scope

This update refreshes the main Wails/Svelte frontend layout and styling without changing backend connection behavior.

## Modified Areas

- App shell and top bar.
- Left connection tree.
- Terminal panel tabs and toolbar.
- Right panel collapsed state and tool rail.
- Shared dialog frame.
- Global design tokens and default accent color.

## Behavior Changes

- SSH connections keep the right panel collapsed by default.
- The right side exposes a narrow tool rail while collapsed.
- The left connection list keeps its existing width but uses compact single-line asset rows.
- New default accent color is blue for fresh/default appearance settings.

## Verification

- `npm run build`
- `GOCACHE="$(pwd)/.cache/go-build" go test ./...`
- `GOCACHE="$(pwd)/.cache/go-build" go vet ./...`

## Notes

- Existing Svelte accessibility warnings remain in several components. The touched asset group row was converted to a button, but this change does not attempt a full accessibility cleanup.
- Existing generated frontend build output is updated by the Vite build.
