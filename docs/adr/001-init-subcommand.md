# ADR-001: `gopkg init` Subcommand Behavior

- Date: 2026-03-14
- Status: Accepted

## Context
`gopkg init` is opinionated and does more than `go mod init`. The implementation includes automatic module-path inference, optional CI workflow generation, and automatic setup of formatting tooling.

## Decision
We standardize the following `init` behavior:
- If `--ci` is specified, write `.github/workflows/ci.yml` using the current runtime Go version and a fixed `go test ./...` workflow.
- If `go.mod` already exists, return success without making further changes (idempotent behavior).
- If module path argument is omitted, infer from current directory only when the path contains `github.com/`; otherwise return an explicit error.
- After `go mod init`, install `golang.org/x/tools/cmd/goimports` as a Go tool dependency.
- Fixed and non-customizable in this command:
  - CI workflow path is fixed at `.github/workflows/ci.yml`.
  - Auto-installed formatter tool is fixed as `golang.org/x/tools/cmd/goimports`.
  - Module inference rule is fixed to `github.com/`-based directory detection only.
- Variable input:
  - Explicit module path argument can be provided by caller.
  - `--ci` toggles whether workflow file generation is executed.

## Consequences
- Easier: new projects get a consistent bootstrap flow, including optional CI and formatter tool setup.
- Harder: workflows and module-path inference are intentionally constrained; non-`github.com/` inference requires explicit module input.
