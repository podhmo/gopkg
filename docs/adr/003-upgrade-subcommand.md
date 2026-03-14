# ADR-003: `gopkg upgrade` Subcommand Behavior

- Date: 2026-03-14
- Status: Accepted

## Context
`gopkg upgrade` is opinionated as a full-module upgrade operation, aligned with Go toolchain commands and optional dev-tool upgrades.

## Decision
We standardize the following `upgrade` behavior:
- Always run `go get -u ./...` at module root.
- With `--dev`, parse `tool` directives in `go.mod` and run `go get -u <tool>` for each declared tool.
- Stop on first command error.

## Consequences
- Easier: single command upgrades module dependencies and optionally dev tools in one predictable flow.
- Harder: no selective/partial upgrade policy is provided by this command.
