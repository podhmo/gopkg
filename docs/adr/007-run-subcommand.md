# ADR-007: `gopkg run` Subcommand Behavior

- Date: 2026-03-14
- Status: Accepted

## Context
`gopkg run` is implemented as build-then-execute using the same binary production path as `gopkg build` (without `-o`) and strict argument splitting around `--`.

## Decision
We standardize the following `run` behavior:
- Parse command args by splitting at first `--`; left side is build args, right side is runtime args.
- Build using `runBuildFrom(..., output=\"\")`, which installs binaries into `<moduleRoot>/.local/gobin`.
- Determine executable name from first package (or `.` default), with module-name fallback for root package and `.exe` suffix on Windows.
- Execute resulting binary with current process stdin/stdout/stderr and current working directory as execution directory.
- Fixed and non-customizable in this command:
  - Build phase always uses the same default install target `<moduleRoot>/.local/gobin`.
  - Arg split protocol is fixed to `--`.
  - Binary name derivation policy is fixed to package/module naming rules in implementation.
- Variable input:
  - Build-side package args and runtime args can be provided by caller.

## Consequences
- Easier: run behavior remains consistent with build behavior and supports CLI runtime args reliably via `--`.
- Harder: runtime always depends on the opinionated `.local/gobin` build path and naming strategy.
