# Bug Fix: Close Tab On Shell Exit

## Background

Typing `exit` in an SSH shell prints `logout`, but the terminal tab stayed open. The backend read loop returned on EOF/read error without notifying the frontend that the session had ended.

## Scope

This change adds backend session-close events for SSH and local shell sessions and makes the frontend remove the matching tab when those events arrive.

## Modified Files

- `app.go`
- `internal/api/handlers/session.go`
- `internal/service/session_service.go`
- `internal/ssh/manager.go`
- `internal/ssh/manager_test.go`
- `frontend/src/components/TerminalPanel.svelte`

## Verification

- Added backend tests for remote-exit cleanup and duplicate-notification prevention.
- Verified the new tests failed before implementation because the cleanup helper did not exist.
- Verified the new tests passed after implementation.
- `GOCACHE="$(pwd)/.cache/go-build" go test ./...` passed.
- `GOCACHE="$(pwd)/.cache/go-build" go vet ./...` passed.
- `npm run build` passed with existing Svelte accessibility and chunk-size warnings.

## Residual Risks

- The automated test covers backend cleanup semantics. Full UI behavior still depends on Wails event delivery and is verified through the frontend build plus manual app testing.
- Existing frontend Svelte accessibility warnings remain unrelated to this fix.
