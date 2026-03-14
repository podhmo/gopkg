# ADR-004: `gopkg format` Subcommand Behavior

- Date: 2026-03-14
- Status: Accepted

## Context
Formatting is a major opinionated area in `gopkg`: it combines `goimports`, module-local import grouping, optional `go fix`, and a post-pass that corrects import aliases when package names differ from path-derived names.

## Decision
We standardize the following `format` behavior:
- Optional `--fix` runs `go fix ./...` before formatting.
- Use `go tool golang.org/x/tools/cmd/goimports -local <moduleName> -w` for formatting.
- Default format target is project root (`.`) when no package patterns are provided.
- Convert `./...`-style and module import-path patterns to directory roots accepted by `goimports`.
- After `goimports`, run import-alias correction as best-effort; failures emit warning and do not fail the whole format step.
- On goimports execution failure, print hint to add tool dependency via `go get -tool golang.org/x/tools/cmd/goimports@latest`.
- Fixed and non-customizable in this command:
  - Formatter is fixed to `goimports` (not `gofumpt` and not user-selectable).
  - Import grouping is fixed to module-local grouping via `-local <moduleName>`.
  - Post-format alias-fix pass is always attempted (best-effort warning on failure).
- Variable input:
  - `--fix` toggles pre-format `go fix`.
  - Package patterns can narrow formatting scope.

## Consequences
- Easier: consistent formatting, local import grouping, and automatic alias normalization.
- Harder: behavior is intentionally coupled to goimports/tool-directive workflow and specific pattern resolution semantics.
