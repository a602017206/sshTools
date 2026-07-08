# Bug Fix: Add Dialog After Edit

## Background

After editing an existing connection, opening the add connection dialog could keep the previous connection ID in the form state. Submitting the dialog then updated the previously edited connection instead of creating a new one.

## Scope

This fix is limited to the connection add/edit dialog state transition.

## Modified Files

- `frontend/src/components/AddAssetDialog.svelte`

## Verification

- `npm run build` passed.
- `GOCACHE="$(pwd)/.cache/go-build" go test ./...` passed.
- `GOCACHE="$(pwd)/.cache/go-build" go vet ./...` passed.

## Residual Risks

- No dedicated frontend unit test harness exists in this project. The fix is verified by build coverage and the corrected state logic.
- Existing Svelte accessibility and chunk-size warnings remain unrelated to this bug.
