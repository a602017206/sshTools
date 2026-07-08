# Process Change: Documentation Governance

## Background

The repository had many root-level Markdown files, including active guides, temporary debugging notes, completed implementation summaries, and historical reports. Future changes also needed a clear rule that every modification creates a corresponding change document.

## Scope

This process change defines the documentation layout, separates feature changes from bug fixes, and separates design records from development records.

## Modified Files

- `AGENTS.md`
- `README.md`
- `frontend/README.md`
- `docs/README.md`
- `docs/archive/*`
- `docs/audits/2026-07-08-codebase-scan.md`
- `docs/designs/2026-07-08-documentation-governance-design.md`
- `docs/development/2026-07-08-documentation-governance-implementation.md`

## Verification

- Searched for stale links to archived documents and updated active references.
- Confirmed archived documents exist under `docs/archive/`.
- `GOCACHE="$(pwd)/.cache/go-build" go test ./...` passed.
- `GOCACHE="$(pwd)/.cache/go-build" go vet ./...` passed.
- `npm run build` passed with existing frontend warnings documented in `docs/audits/2026-07-08-codebase-scan.md`.

## Residual Risks

- Some archived documents may contain relative links that were correct only at their original root location. They are retained as historical records rather than active navigation documents.
