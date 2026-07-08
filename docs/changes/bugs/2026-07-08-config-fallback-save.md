# Bug Fix: Fallback Config Save

## Background

When the disk-backed config manager cannot initialize, the app creates an in-memory fallback config manager. The fallback has no file path, but `Save()` still attempted to write to disk.

## Scope

This fix only changes fallback save behavior. Disk-backed config persistence remains unchanged.

## Modified Files

- `internal/config/config.go`
- `internal/config/config_test.go`

## Verification

- Red test: `GOCACHE="$(pwd)/.cache/go-build" go test ./internal/config -run TestFallbackConfigManagerSavesInMemory -v` failed with `open : no such file or directory`.
- Green test: the same command passed after updating `Save()`.

## Residual Risks

- Fallback mode is intentionally non-persistent. User changes made while disk config is unavailable will not survive app restart.
