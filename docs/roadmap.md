# Roadmap

## v0.1.0 – MVP

- [x] Project root exploration (`go.mod` / `.git` heuristic)
- [x] `gopkg install` – `go mod tidy` + optional dev-tool install
- [x] `gopkg upgrade` – `go get -u ./...` + optional dev-tool upgrade
- [x] `gopkg format` – `goimports -w ./...` + optional `go fix`
- [x] `gopkg lint` – `go vet ./...`

## v0.2.0 – Quality of life

- [ ] Colour-coded output / progress messages
- [ ] `gopkg run <tool> [args]` – run a tool from the `tool` directive without a full install
- [ ] Config file support (`.gopkg.toml`) for per-project defaults
- [ ] Shell completion (bash / zsh / fish)

## v0.3.0 – Extended tooling

- [ ] `gopkg check` – run `staticcheck` or other static-analysis tools
- [ ] `gopkg test` – thin wrapper around `go test ./...` with sensible defaults
- [ ] `gopkg doc` – open `pkg.go.dev` documentation for a package
- [ ] Plugin system for community-contributed subcommands

## Ideas / Backlog

- Workspace (`go.work`) awareness
- `gopkg outdated` – list dependencies that have newer versions available
- CI mode: machine-readable JSON output
