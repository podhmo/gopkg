# ADR-005: `gopkg lint` Subcommand Behavior

- Date: 2026-03-14
- Status: Accepted

## Context
`gopkg lint` intentionally provides a minimal and deterministic lint entrypoint with no local policy DSL.

## Decision
We standardize the following `lint` behavior:
- Always run `go vet ./...` from module root.
- Provide no command-specific lint configuration flags.
- Fixed and non-customizable in this command:
  - Lint engine is fixed to `go vet`.
  - Target scope is fixed to `./...`.
- Variable input:
  - None.

## Consequences
- Easier: predictable lint behavior tightly aligned with standard Go tooling.
- Harder: users requiring richer lint policy must use additional tools outside this subcommand.
