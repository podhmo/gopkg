# TODO

## In progress

- [ ] Full test coverage for all subcommands

## Done

- [x] `go.mod` / module scaffold
- [x] Project-root exploration logic (`root.go`)
- [x] `go.mod` tool-directive parser (`modfile.go`)
- [x] `gopkg install` subcommand
- [x] `gopkg upgrade` subcommand
- [x] `gopkg format` subcommand
- [x] `gopkg lint` subcommand
- [x] CLI wiring (`main.go`)
- [x] `README.md`
- [x] `AGENTS.md`
- [x] `docs/roadmap.md`
- [x] Unit tests for root exploration
- [x] Unit tests for modfile parsing

## Backlog

- [ ] Colour-coded / verbose log output
- [ ] `gopkg run <tool>` subcommand
- [ ] Shell completion scripts
- [ ] `gopkg check` (staticcheck integration)
- [ ] `gopkg outdated`
- [ ] Config file support (`.gopkg.toml`)
