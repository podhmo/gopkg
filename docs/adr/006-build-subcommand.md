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

## Consequences
- Easier: fast incremental local builds via Go cache and stable binary placement in `.local/gobin`.
- Harder: default output location is opinionated; consumers expecting direct artifact path must use `-o`.
