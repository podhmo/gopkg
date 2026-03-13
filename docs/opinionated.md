# Opinionated Behaviors

`gopkg` is intentionally "selfish" – it makes a number of fixed, non-configurable decisions so that the common case works out of the box. This document enumerates every behavior that is hardcoded or deliberately constrained.

---

## Project Root Discovery (`root.go`)

1. **Walk upward from the current working directory.**
   `gopkg` always walks up the directory tree starting from `os.Getwd()`. There is no flag to specify the project root explicitly.

2. **Stop at the first `go.mod` file.**
   The first directory that contains a `go.mod` file is unconditionally treated as the project root. Nested `go.mod` files (e.g. inside a sub-module) are never used.

5. **All commands execute in the project root.**
   Every subcommand runs its underlying Go toolchain commands with `cmd.Dir` set to the discovered root directory.

---

## Module Path Inference (`commands.go` → `modulePathFromDir`)

1. **GitHub-only auto-detection.**
   When `gopkg init` is called without an explicit module path, the module path is inferred by searching the current directory path for the substring `"github.com/"`. Everything from `"github.com/"` onward becomes the module path.

2. **Non-GitHub paths require an explicit argument.**
   If the directory path does not contain `"github.com/"`, `gopkg init` fails with:
   ```
   cannot infer module path from "<dir>": path does not contain "github.com/"; please provide a module path explicitly
   ```

---

## `gopkg init` Behavior (`commands.go` → `runInitFrom`)

1. **Idempotent: silently succeeds if `go.mod` already exists.**
   When a `go.mod` file is already present, `gopkg init` does nothing and exits with success. This makes it safe to call repeatedly (e.g. in CI scripts) without modifying the existing module.
   *Note:* The CI workflow file (if `-ci` is given) is still written/overwritten even when `go.mod` already exists.

2. **`goimports` is always installed as a tool dependency.**
   After `go mod init`, `gopkg init` unconditionally runs:
   ```
   go get -tool golang.org/x/tools/cmd/goimports@latest
   ```
   This makes `gopkg format` work immediately after initialization.

3. **The CI workflow file is always written to `.github/workflows/ci.yml`.**
   When `-ci` is given, this hardcoded path is used. The directory is created with permissions `0o755` and the file is written with permissions `0o644`.

4. **The CI workflow template is hardcoded.**
   The generated `.github/workflows/ci.yml` contains:
   - Workflow name: `CI`
   - Trigger: `pull_request` with types `[opened, synchronize, reopened]` only
   - Job name: `test`
   - Runner: `ubuntu-latest`
   - Steps: `actions/checkout@v4`, `actions/setup-go@v5`, then `go test ./...`
   - Go version: taken from `runtime.Version()` of the `gopkg` binary itself (leading `"go"` stripped)
   - No `push` trigger, no matrix strategy, no caching step

---

## `gopkg format` Behavior (`commands.go` → `runFormatFrom`)

1. **`goimports` is the only supported formatter.**
   `gopkg format` runs `go tool golang.org/x/tools/cmd/goimports -w`. There is no option to use `gofmt` or any other formatter.

2. **`goimports` must be listed as a tool directive in `go.mod`.**
   If `goimports` is not declared and the command fails, a hint is printed to stderr:
   ```
   hint: to use gopkg format, add goimports as a tool dependency:
     go get -tool golang.org/x/tools/cmd/goimports@latest
   ```

3. **`-local <module-name>` is always passed to `goimports`.**
   The module name is read from `go.mod` and passed as `-local` so that the project's own packages are grouped separately in import blocks (standard library → third-party → local).

4. **Pattern conversion: `/...` wildcard is stripped.**
   `goimports` walks directories recursively on its own, so the trailing `/...` is removed from every pattern before it is passed to `goimports`.

5. **Absolute import-path patterns are converted to relative directory paths.**
   For example, `github.com/owner/repo/pkg/...` becomes `./pkg`. The module prefix is stripped and a leading `./` is prepended.

6. **Default pattern is `.` (project root directory).**
   When no patterns are supplied, `gopkg format` formats the entire project by passing `.` to `goimports`.

7. **`--fix` runs `go fix ./...` before goimports.**
   There is no option to skip the `go fix` step once `--fix` is set, and the fix target is always `./...`.

---

## `gopkg build` Behavior (`commands.go` → `runBuildFrom`)

1. **Default output directory: `<module-root>/.local/gobin`.**
   When `-o` is not given, binaries are installed into `<module-root>/.local/gobin` by setting `GOBIN` to that absolute path and running `go install`. This leverages the Go build cache (unlike `go build -o`).

2. **`GOBIN` must be absolute.**
   The path is resolved with `filepath.Abs` before being passed to `go install`. `go install` itself rejects relative `GOBIN` values.

3. **The `.local/gobin` directory is created automatically with mode `0o755`.**
   If the directory does not exist it is created. No user configuration controls the permissions.

4. **Default package when none is specified: `.` (the current module root).**
   If no package arguments are given, `gopkg build` builds the package in the project root directory.

5. **With `-o`, `go build -o` is used instead of `go install`.**
   This bypasses the build cache for the final link step but allows writing the binary to an arbitrary path.

---

## `gopkg install` and `gopkg upgrade` Behavior

1. **`install` always runs `go mod tidy` first.**
   There is no flag to skip `go mod tidy`.

2. **`upgrade` always runs `go get -u ./...` first.**
   There is no flag to scope the upgrade to specific packages.

3. **`--dev` reads tool paths exclusively from `go.mod` `tool` directives.**
   No other source (e.g. a config file) is consulted.

4. **`install --dev` uses `go install <tool>` (no version suffix).**
   Tools are installed at the version recorded in `go.mod`.

5. **`upgrade --dev` uses `go get -u <tool>` individually for each tool.**
   Each tool is upgraded in a separate `go get` invocation.

---

## `gopkg lint` Behavior

1. **`go vet ./...` is the only linter.**
   There is no option to run third-party linters (e.g. `staticcheck`, `golangci-lint`).

2. **The scope is always `./...` (the entire module).**
   There is no flag to restrict linting to a subset of packages.

---

## Command Execution Output (`commands.go` → `run` / `runWithEnv`)

1. **Every command is printed to stdout before execution.**
   The format is `  → <name> [args...]` (two spaces, a Unicode right arrow, then the command). This is always printed and cannot be suppressed.

2. **`stdout` and `stderr` of all child processes are wired to the parent process.**
   All output is streamed in real time; there is no buffering, filtering, or quiet mode.

---

## File System Permissions

| Resource | Permission |
|---|---|
| Directories created by `gopkg` (e.g. `.local/gobin`, `.github/workflows`) | `0o755` |
| Files written by `gopkg` (e.g. `ci.yml`) | `0o644` |

These values are hardcoded and not configurable.

---

## `go.mod` Tool Directive Parsing (`modfile.go`)

1. **Two formats are supported: single-line and block.**
   ```
   tool golang.org/x/tools/cmd/goimports
   ```
   ```
   tool (
       golang.org/x/tools/cmd/goimports
       github.com/some/other/tool
   )
   ```
   No other format is accepted.

2. **Inline comments (`// ...`) are stripped.**
   Everything from `//` to end-of-line is ignored when parsing tool entries.

3. **Empty lines inside blocks are silently ignored.**

---

## No Configuration File

`gopkg` has no configuration file (no `.gopkg.toml`, `gopkg.yaml`, etc.). All behavior is controlled exclusively by command-line flags. This is a deliberate design decision; per-project configuration is listed as a future backlog item.
