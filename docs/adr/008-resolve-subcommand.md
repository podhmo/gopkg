# ADR-008: `gopkg resolve` Subcommand Behavior

- Date: 2026-03-14
- Status: Accepted

## Context
`gopkg resolve` exists to normalize relative package references into module import paths for downstream command composition.

## Decision
We standardize the following `resolve` behavior:
- For each input argument, convert only relative package paths (`.`, `./...`, `../...`) into full module import paths.
- Resolve against current working directory, then rebase to module root and prefix module name from `go.mod`.
- If path escapes module root or cannot be relativized, keep the original argument unchanged.
- Pass through flags (e.g., `-all`) and already absolute/non-relative arguments unchanged.
- Print one resolved argument per line to stdout.
- Fixed and non-customizable in this command:
  - Resolution logic is fixed to relative-path-only conversion.
  - Output format is fixed to one result per line.
- Variable input:
  - Arbitrary argument list can be provided by caller.

## Consequences
- Easier: deterministic conversion from local paths to canonical import paths.
- Harder: conversion policy is intentionally conservative and keeps unsupported/escaping paths unchanged.
