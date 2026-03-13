# gopkg

A selfish Go package manager – a thin wrapper around standard Go tooling that provides a convenient, npm-style developer experience.

## Install

```sh
go install github.com/podhmo/gopkg@latest
```

## Usage

Run all commands from anywhere inside a Go module (gopkg walks up to the `go.mod` root automatically).

```
gopkg <subcommand> [flags]
```

### Subcommands

| Command | Description |
|---------|-------------|
| `gopkg install [--dev]` | Run `go mod tidy`; with `--dev` also `go install` every tool listed in the `tool` directive |
| `gopkg upgrade [--dev]` | Run `go get -u ./...`; with `--dev` also upgrade tools from the `tool` directive |
| `gopkg format [--fix]` | Run `goimports -w ./...`; with `--fix` run `go fix ./...` first |
| `gopkg lint` | Run `go vet ./...` |

### Examples

```sh
# Tidy dependencies
gopkg install

# Tidy + install dev tools declared in go.mod tool directives
gopkg install --dev

# Upgrade all dependencies
gopkg upgrade

# Upgrade dependencies AND dev tools
gopkg upgrade --dev

# Format code (requires goimports)
gopkg format

# Fix + format
gopkg format --fix

# Lint
gopkg lint
```

## Requirements

- Go 1.24+
- `goimports` on `$PATH` for `gopkg format` (install via `go install golang.org/x/tools/cmd/goimports@latest`)

## How it works

`gopkg` walks the directory tree from the current working directory toward the filesystem root, stopping at the first directory that contains a `go.mod` file.  
All tooling commands are executed with that directory as the working directory.

## License

MIT
