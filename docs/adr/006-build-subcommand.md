# ADR-006: `gopkg build` Subcommand Behavior

- Date: 2026-03-14
- Status: Accepted

## Context
`gopkg build` chooses a cache-friendly default strategy and a separate explicit-output strategy. It fixes binary installation location and `GOBIN` handling for deterministic local workflows.

## Decision
We standardize the following `build` behavior:
- If `-o` is provided, run `go build -o <output> [pkgs...]`.
- If `-o` is omitted, run `go install [pkgs...]` with `GOBIN=<abs(moduleRoot/.local/gobin)>`.
- Auto-create `.local/gobin` under module root when needed.
- Default package target is `.` when no packages are provided.
- Pass through optional `-v` to underlying build/install command.
- Fixed and non-customizable in this command:
  - Default binary install directory is fixed to `<moduleRoot>/.local/gobin`.
  - `GOBIN` derivation is fixed to the absolute path of that directory.
  - Default package fallback is fixed to `.`.
- Variable input:
  - `-o` switches to explicit output mode.
  - Package list and `-v` verbosity are caller-provided.

## Consequences
- Easier: fast incremental local builds via Go cache and stable binary placement in `.local/gobin`.
- Harder: default output location is opinionated; consumers expecting direct artifact path must use `-o`.
