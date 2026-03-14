# ADR-002: `gopkg install` Subcommand Behavior

- Date: 2026-03-14
- Status: Accepted

## Context
`gopkg install` is defined as dependency normalization, not arbitrary package installation. It intentionally wraps core Go module maintenance and optional dev-tool setup from `go.mod` tool directives.

## Decision
We standardize the following `install` behavior:
- Always run `go mod tidy` at module root.
- With `--dev`, parse `tool` directives in `go.mod` and run `go install <tool>` for each declared tool.
- Stop on first command error.
- Fixed and non-customizable in this command:
  - Base operation is fixed to `go mod tidy`.
  - Tool source is fixed to `go.mod` `tool` directives only (no separate config file).
- Variable input:
  - `--dev` toggles whether tool installation is also performed.

## Consequences
- Easier: consistent dependency tidy flow and reproducible dev-tool installation from module metadata.
- Harder: command scope is intentionally narrow; it does not support ad-hoc package installation semantics.
