# Documentation Governance Design

## Goal

Make the documentation set easier to navigate and require future changes to leave a clear audit trail.

## Directory Model

Root Markdown files stay limited to active entry points and long-lived guides. Detailed change records and historical reports move under `docs/`.

The structure is:

- `docs/audits/` for scans and issue inventories.
- `docs/changes/features/` for requirement and feature changes.
- `docs/changes/bugs/` for bug fixes.
- `docs/changes/process/` for repository and workflow changes.
- `docs/designs/` for design proposals and architecture decisions.
- `docs/development/` for implementation and rollout notes.
- `docs/archive/` for historical summaries and temporary notes.

## Trade-offs

Archiving preserves history while reducing root noise. It is slightly more verbose than deleting obsolete documents, but safer for a project that has accumulated implementation context over time.

Splitting design and development creates more documents, but it avoids mixing intent, trade-offs, implementation details, and verification in one place.

## Decision

Use explicit category directories and require one matching change document for every future change.
