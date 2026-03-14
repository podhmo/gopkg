# AGENTS.md

## Development Rules

### Tests Are Mandatory

- Writing tests is **required** for all implementations.
- An implementation without tests is not considered complete.

### Definition of Done

- "Done" means **all tests pass**.
- Finishing the implementation alone is not enough.
- Run `go test ./...` and confirm every test passes before declaring completion.

## 📍 Pointers & Routing
- **Architecture & Decisions**: Read `docs/adr-workflow.md` FIRST. All architectural history is in `docs/adr/`.

## 🚫 Strict Prohibitions (NEVER DO THESE)

1. **DO NOT overwrite/delete existing ADRs**. If a decision changes, create a new ADR and mark the old one as `Superseded`.
2. **DO NOT invent architectural rules**. Rely on `docs/adr/` for structural facts. If tests/linters fail, follow the `FIX` instructions in the error message.
