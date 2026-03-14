# ADR-009: ADR Workflow Standard Adoption

- Date: 2026-03-14
- Status: Accepted

## Context
The repository needs deterministic governance for architectural decisions. Without explicit ADR workflow rules, ADRs can drift, be overwritten, or remain unenforced. We also need agent-routing guidance so contributors consult architecture policy before making structural changes.

## Decision
We adopt and document ADR workflow standards in:
- `docs/adr-workflow.md` (requested top-level ADR workflow document)

We require:
- immutability of existing ADR content in `docs/adr/` with supersession process,
- fixed ADR metadata/section structure (Date, Status, Context, Decision, Consequences),
- explicit status enum as a closed set (`Proposed`, `Accepted`, `Superseded by [ADR-XXX]`, `Deprecated`),
- executable enforcement expectations (archgate pattern) for architectural rules.

The workflow document explicitly states why each mandatory item exists so authors know what must be written and why.

`AGENTS.md` is updated to route contributors to the workflow standard first and to prohibit ADR overwrite/invention of rules outside ADR documents.

## Consequences
- Easier: architecture history is traceable, stable, and reviewable; agent behavior is better aligned with repository governance.
- Harder: process overhead increases because changes to architecture must be recorded as new ADRs and paired with enforceable checks when applicable.
