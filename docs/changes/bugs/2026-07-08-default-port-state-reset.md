# Bug Fix: Default Port State Reset

## Background

The add/edit connection dialog reused local state across SSH, database, and Docker connection types. After opening an existing database connection, a new SSH or Docker connection could inherit port `3306`; after opening an SSH connection, database or Docker entries could inherit port `22`.

## Root Cause

The form reset calculated the default port before resetting `assetType` back to SSH. Connection type buttons also only changed `assetType`, so the port field was not refreshed when users selected a different type.

## Scope

This fix is limited to the connection add/edit dialog state and default port calculation.

## Modified Files

- `frontend/src/components/AddAssetDialog.svelte`

## Behavior

- New SSH connections default to port `22`.
- New MySQL connections default to port `3306`.
- New PostgreSQL connections default to port `5432`.
- New Docker connections default to port `2375`.
- Editing an existing connection keeps its saved port instead of overwriting it with a type default.

## Verification

- `npm run build` passed.
- `GOCACHE="$(pwd)/.cache/go-build" go test ./...` passed.
- `GOCACHE="$(pwd)/.cache/go-build" go vet ./...` passed.

## Residual Risks

- No dedicated frontend unit test harness exists for the Svelte dialog state flow. The fix is verified by build coverage and targeted state-flow review.
- Existing Svelte accessibility and chunk-size warnings remain unrelated to this bug.
