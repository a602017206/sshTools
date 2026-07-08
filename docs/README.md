# Documentation Map

This directory separates active planning, design, implementation notes, audits, change records, and historical archives.

## Active Directories

- `audits/` - codebase scans, review reports, issue inventories, and verification summaries.
- `changes/features/` - requirement and feature change records.
- `changes/bugs/` - bug fix change records.
- `changes/process/` - repository, documentation, workflow, and process change records.
- `designs/` - design proposals, trade-offs, architecture decisions, and UX/system design notes.
- `development/` - implementation notes, execution summaries, rollout notes, and follow-up engineering details.
- `plans/` - step-by-step implementation plans.
- `archive/` - historical one-off summaries, temporary debugging notes, and superseded reports.

## Change Documentation Rules

Every modification must have a corresponding change document under `docs/changes/`.

Use `features/` for new behavior or changed requirements. Use `bugs/` for defects and regressions. Use `process/` for documentation structure, workflow, build process, or repository governance.

Each change document should include:

- Background
- Scope
- Modified files
- Verification
- Residual risks

Design and development must stay separate. Put proposed architecture and trade-offs in `docs/designs/`; put implementation details and execution notes in `docs/development/`.

## Root Documentation Policy

Keep root-level Markdown limited to active entry points and long-lived guides. Archive temporary reports and completed implementation summaries instead of leaving them in the root.
