# ADR Workflow Standard

This document defines how AI agents must handle Architecture Decision Records (ADRs) and enforce them in this repository.

## 1. Immutability Principle (不変の原則)
- Existing files in `docs/adr/` are **IMMUTABLE**. Never modify the core content of a past decision.
- When an architectural decision is updated or replaced, you MUST:
  1. Create a new ADR file (e.g., `docs/adr/008-new-database.md`).
  2. Update the `Status` of the old ADR to `Superseded by [ADR-008]`.

## 2. ADR Structure (ADRの必須フォーマット)
When generating a new ADR, strictly follow this structure:
- **Date**: YYYY-MM-DD
- **Status**: Proposed / Accepted / Superseded by [ADR-XXX] / Deprecated
- **Context**: Why are we making this decision? (The context that cannot be captured by tests).
- **Decision**: What is the exact decision?
- **Consequences**: What becomes easier? What becomes harder?

Required template:
```md
# ADR-XXX: <short title>

- Date: YYYY-MM-DD
- Status: Proposed / Accepted / Superseded by [ADR-XXX] / Deprecated

## Context
...

## Decision
...

## Consequences
...
```
The `Status` line above shows alternatives; each ADR must choose exactly one value from the enum.

### Why these fields are mandatory
- **Date** is required to keep decision chronology explicit.
- **Status** is required so readers can safely know whether the ADR is active.
- **Context** is required to preserve non-obvious constraints that tests cannot encode.
- **Decision** is required to make the chosen rule explicit and reviewable.
- **Consequences** is required to document trade-offs and future operational cost.

### Status enum (closed set)
Only these values are allowed for `Status`:
- `Proposed`
- `Accepted`
- `Superseded by [ADR-XXX]`
- `Deprecated`

No other status strings are allowed. This keeps ADR lifecycle machine-checkable and consistent.

## 3. The Archgate Pattern (実行可能ルールへの結合)
Architectural decisions are useless unless enforced deterministically. When you create an ADR, you MUST pair it with an executable check.

- **Create Linter Rules/Tests**: If an ADR restricts dependencies (e.g., "Service A cannot access Layer B"), write an AST-based custom linter rule or an architectural test to enforce it.
- **Error Messages as Prompts**: Any linter rule generated for an ADR MUST output errors in the following format:
  ```text
  ERROR: [What is wrong]
    WHY: [Why this rule exists. MUST include link to ADR, e.g., "See docs/adr/007-xxx.md"]
    FIX: [Exact steps to fix the issue]
  ```
