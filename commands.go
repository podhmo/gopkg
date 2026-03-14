package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// run executes a command with the given arguments in dir, wiring stdout/stderr
// to the current process so that the user sees real-time output.
func run(dir string, name string, args ...string) error {
	fmt.Fprintf(os.Stdout, "  → %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runWithEnv is like run but merges the provided key=value pairs into the
// current process environment before executing the command.
func runWithEnv(dir string, env map[string]string, name string, args ...string) error {
	fmt.Fprintf(os.Stdout, "  → %s %v\n", name, args)
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmdEnv := os.Environ()
	for k, v := range env {
		cmdEnv = append(cmdEnv, k+"="+v)
	}
	cmd.Env = cmdEnv
	return cmd.Run()
}

// runInstall runs `go mod tidy` and, when dev is true, installs every tool
// listed in the go.mod tool directive.
func runInstall(dev bool) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}

	if err := run(root, "go", "mod", "tidy"); err != nil {
		return err
	}

	if dev {
		if err := installDevTools(root); err != nil {
			return err
		}
	}

	return nil
}

// runUpgrade runs `go get -u ./...` and, when dev is true, also upgrades the
// tools listed in the go.mod tool directive.
func runUpgrade(dev bool) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}

	if err := run(root, "go", "get", "-u", "./..."); err != nil {
		return err
	}

	if dev {
		if err := upgradeDevTools(root); err != nil {
			return err
		}
	}

	return nil
}

const goimportsTool = "golang.org/x/tools/cmd/goimports"

// runFormat runs `go tool golang.org/x/tools/cmd/goimports -w <pkgs>` and,
// when fix is true, runs `go fix ./...` first.
// pkgs defaults to []string{"./..."} when empty.
func runFormat(fix, verbose bool, pkgs []string) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}
	return runFormatFrom(root, fix, verbose, pkgs)
}

// runFormatFrom is the testable core of runFormat.
func runFormatFrom(root string, fix, verbose bool, pkgs []string) error {
	if fix {
		if err := run(root, "go", "fix", "./..."); err != nil {
			return err
		}
	}

	moduleName, err := readModuleName(filepath.Join(root, "go.mod"))
	if err != nil {
		return fmt.Errorf("reading module name: %w", err)
	}

	patterns, err := resolveFormatPatterns(root, pkgs)
	if err != nil {
		return err
	}

	// Before running goimports, add explicit aliases to imports whose declared
	// package name differs from the last non-version component of their path
	// (e.g. package "bar" imported from "github.com/foo/go-bar").  This must
	// happen first so that goimports can group the aliased imports correctly.
	if err := runFixImportAliasesFrom(root, patterns); err != nil {
		// Best-effort: alias fixing must not break the overall format step.
		fmt.Fprintf(os.Stderr, "warning: fixing import aliases: %v\n", err)
	}

	args := []string{"tool", goimportsTool, "-local", moduleName, "-w"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, patterns...)
	if err := run(root, "go", args...); err != nil {
		fmt.Fprintf(os.Stderr, "\nhint: to use gopkg format, add goimports as a tool dependency:\n  go get -tool %s@latest\n", goimportsTool)
		return err
	}
	return nil
}

// resolveFormatPatterns converts Go package patterns to directory paths that
// goimports accepts.  Relative patterns (e.g. "./...", "./foo/...") are
// stripped of their trailing "/..." wildcard and returned as directory paths.
// Absolute import-path patterns (e.g. "github.com/podhmo/gopkg/foo/...") are
// stripped of the module prefix and converted to the equivalent relative
// directory path.  When patterns is empty the project root (".") is returned.
func resolveFormatPatterns(root string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		return []string{"."}, nil
	}

	moduleName, err := readModuleName(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("reading module name: %w", err)
	}

	resolved := make([]string, len(patterns))
	for i, p := range patterns {
		resolved[i] = resolvePattern(p, moduleName)
	}
	return resolved, nil
}

// resolvePattern converts a single package pattern to a directory path that
// goimports understands.  The trailing "/..." wildcard is stripped because
// goimports already walks directories recursively.  Import paths that match
// moduleName are converted to relative paths (e.g.
// "github.com/foo/bar/pkg/..." → "./pkg").
func resolvePattern(pattern, moduleName string) string {
	// Strip the "/..." recursive wildcard – goimports walks directories
	// recursively by itself, so we only need the root directory.
	p := strings.TrimSuffix(pattern, "/...")

	// Already a relative path – pass through.
	if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") || p == "." {
		return p
	}
	// Exact module root.
	if p == moduleName {
		return "."
	}
	// Import path that starts with the module root followed by a "/".
	if strings.HasPrefix(p, moduleName+"/") {
		return "." + p[len(moduleName):]
	}
	// Unrecognised – pass through as-is.
	return p
}

// runLint runs `go vet ./...`.
func runLint() error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}

	return run(root, "go", "vet", "./...")
}

// installDevTools reads go.mod tool directives and runs `go install` for each.
func installDevTools(root string) error {
	tools, err := readToolDirectives(filepath.Join(root, "go.mod"))
	if err != nil {
		return fmt.Errorf("reading tool directives: %w", err)
	}
	for _, tool := range tools {
		if err := run(root, "go", "install", tool); err != nil {
			return err
		}
	}
	return nil
}

// upgradeDevTools reads go.mod tool directives and runs `go get -u` for each.
func upgradeDevTools(root string) error {
	tools, err := readToolDirectives(filepath.Join(root, "go.mod"))
	if err != nil {
		return fmt.Errorf("reading tool directives: %w", err)
	}
	for _, tool := range tools {
		if err := run(root, "go", "get", "-u", tool); err != nil {
			return err
		}
	}
	return nil
}

// ciWorkflowContent returns the content of the GitHub Actions CI workflow file
// for the given Go version (e.g. "1.24.0").
func ciWorkflowContent(goVersion string) string {
	return `name: CI

on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
      - uses: actions/setup-go@v6
        with:
          go-version: "` + goVersion + `"
      - run: go test ./...
`
}

// runInit initializes a new Go module in the current working directory.
// When modulePath is empty the path is inferred from the working directory:
// if the directory path contains "github.com/" the module path is taken as
// everything from "github.com/" onwards.
func runInit(modulePath string, ci bool) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	return runInitFrom(dir, modulePath, ci)
}

// runInitFrom is the testable core of runInit.
func runInitFrom(dir, modulePath string, ci bool) error {
	if ci {
		// runtime.Version() returns e.g. "go1.24" or "go1.24.1"; strip the leading "go".
		goVersion := strings.TrimPrefix(runtime.Version(), "go")
		if err := writeCIWorkflow(dir, goVersion); err != nil {
			return err
		}
	}

	// If go.mod already exists, the module is already initialised – skip all
	// work and succeed so that repeated calls (e.g. in CI) are idempotent.
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		return nil
	}

	if modulePath == "" {
		var err error
		modulePath, err = modulePathFromDir(dir)
		if err != nil {
			return err
		}
	}

	if err := run(dir, "go", "mod", "init", modulePath); err != nil {
		return err
	}

	// Install goimports as a tool dependency so that "gopkg format" works
	// out of the box.
	if err := run(dir, "go", "get", "-tool", goimportsTool+"@latest"); err != nil {
		return err
	}
	return nil
}

// writeCIWorkflow creates .github/workflows/ci.yml with a GitHub Actions CI
// configuration that runs tests on pull_request events.
func writeCIWorkflow(dir, goVersion string) error {
	workflowDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(workflowDir, 0o755); err != nil {
		return fmt.Errorf("creating workflow directory: %w", err)
	}
	workflowPath := filepath.Join(workflowDir, "ci.yml")
	fmt.Fprintf(os.Stdout, "  → writing %s\n", workflowPath)
	if err := os.WriteFile(workflowPath, []byte(ciWorkflowContent(goVersion)), 0o644); err != nil {
		return fmt.Errorf("writing ci.yml: %w", err)
	}
	return nil
}

// modulePathFromDir infers a Go module path from an absolute directory path.
// If the path contains "github.com/", the module path is everything from
// "github.com/" onwards.  Otherwise an error is returned asking the user to
// supply an explicit module path.
func modulePathFromDir(dir string) (string, error) {
	// Normalize to forward slashes for consistent substring search on all OSes.
	normalized := filepath.ToSlash(dir)
	const prefix = "github.com/"
	idx := strings.Index(normalized, prefix)
	if idx != -1 {
		return normalized[idx:], nil
	}
	return "", fmt.Errorf("cannot infer module path from %q: path does not contain %q; please provide a module path explicitly", dir, prefix)
}

// runBuild builds the packages.  When output is empty it uses go install with
// GOBIN set to <root>/.local/gobin so that the build cache is leveraged.
// When output is non-empty it uses go build -o <output>.
func runBuild(output string, verbose bool, pkgs []string) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}
	return runBuildFrom(root, output, verbose, pkgs)
}

// splitAtDashDash splits args at the first "--" separator.
// Everything before "--" is returned as before; everything after as after.
// If "--" is not present, all args are returned as before and after is nil.
func splitAtDashDash(args []string) (before, after []string) {
	for i, arg := range args {
		if arg == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

// runRun builds pkgs and then executes the resulting binary with runArgs,
// wiring stdin/stdout/stderr to the current process.
func runRun(verbose bool, pkgs []string, runArgs []string) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	return runRunFrom(root, pwd, verbose, pkgs, runArgs)
}

// runRunFrom is the testable core of runRun.  It builds the packages using the
// same mechanism as gopkg build (go install into <root>/.local/gobin), then
// locates the installed binary and executes it with runArgs.  The binary is
// left in place after execution.  pwd is the working directory for the
// executed binary.
func runRunFrom(root, pwd string, verbose bool, pkgs []string, runArgs []string) error {
	// Build using gopkg build's mechanism: go install → .local/gobin.
	if err := runBuildFrom(root, "", verbose, pkgs); err != nil {
		return err
	}

	binaryName, err := binaryNameForPackage(root, pkgs)
	if err != nil {
		return fmt.Errorf("determining binary name: %w", err)
	}
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	gobin := filepath.Join(root, ".local", "gobin")
	binPath := filepath.Join(gobin, binaryName)

	fmt.Fprintf(os.Stdout, "  → %s %v\n", binPath, runArgs)
	cmd := exec.Command(binPath, runArgs...)
	cmd.Dir = pwd
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// binaryNameForPackage returns the name of the binary that go install produces
// for the given package list.  Only the first package is considered; when pkgs
// is empty "." is assumed.
//
// For a relative path like "./cmd/app" the binary name is the last directory
// component ("app").  For "." the name is derived from the module name stored
// in go.mod (last "/" element).
func binaryNameForPackage(root string, pkgs []string) (string, error) {
	pkg := "."
	if len(pkgs) > 0 {
		pkg = pkgs[0]
	}

	// Relative paths: binary name == last directory component.
	base := filepath.Base(filepath.Clean(pkg))
	if base != "." {
		return base, nil
	}

	// "." (module root): derive from the module name.
	moduleName, err := readModuleName(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("reading module name: %w", err)
	}
	// Module paths use "/" separators; take the last element.
	parts := strings.Split(moduleName, "/")
	return parts[len(parts)-1], nil
}

// runResolve converts any relative-path argument to a full module import path
// and prints each resolved path to stdout.
func runResolve(args []string) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}
	return runResolveFrom(os.Stdout, root, pwd, args)
}

// runResolveFrom is the testable core of runResolve. It resolves each argument
// to a full module import path and writes one result per line to out.
// Arguments that are flags (starting with "-") or already absolute import paths
// are written unchanged.
func runResolveFrom(out io.Writer, root, pwd string, args []string) error {
	moduleName, err := readModuleName(filepath.Join(root, "go.mod"))
	if err != nil {
		return fmt.Errorf("reading module name: %w", err)
	}

	for _, arg := range args {
		resolved := resolveDocArg(arg, pwd, root, moduleName)
		fmt.Fprintln(out, resolved)
	}
	return nil
}

// resolveDocArg converts a single go doc argument to a full import path when
// it is a relative package path (".", "./foo", "../foo").  The path is resolved
// against pwd, expressed relative to root, and prefixed with moduleName.
// Arguments that are flags (starting with "-") or are already absolute import
// paths are returned unchanged.
func resolveDocArg(arg, pwd, root, moduleName string) string {
	if !strings.HasPrefix(arg, "./") && !strings.HasPrefix(arg, "../") && arg != "." {
		return arg
	}

	absPath := filepath.Join(pwd, arg)

	relPath, err := filepath.Rel(root, absPath)
	if err != nil {
		return arg
	}

	// If the path escapes the module root, pass through unchanged.
	if strings.HasPrefix(relPath, "..") {
		return arg
	}

	relPath = filepath.ToSlash(relPath)
	if relPath == "." {
		return moduleName
	}
	return moduleName + "/" + relPath
}

// runBuildFrom is the testable core of runBuild.
func runBuildFrom(root, output string, verbose bool, pkgs []string) error {
	if output != "" {
		goArgs := []string{"build"}
		if verbose {
			goArgs = append(goArgs, "-v")
		}
		goArgs = append(goArgs, "-o", output)
		goArgs = append(goArgs, pkgs...)
		return run(root, "go", goArgs...)
	}

	// Use go install with GOBIN pointing at <root>/.local/gobin so the Go
	// build cache is used (go build -o would bypass it for the final link).
	// GOBIN must be an absolute path – go install rejects relative paths.
	gobin, err := filepath.Abs(filepath.Join(root, ".local", "gobin"))
	if err != nil {
		return fmt.Errorf("resolving GOBIN path: %w", err)
	}
	if err := os.MkdirAll(gobin, 0o755); err != nil {
		return fmt.Errorf("creating GOBIN directory: %w", err)
	}

	if len(pkgs) == 0 {
		pkgs = []string{"."}
	}
	installArgs := []string{"install"}
	if verbose {
		installArgs = append(installArgs, "-v")
	}
	installArgs = append(installArgs, pkgs...)
	return runWithEnv(root, map[string]string{"GOBIN": gobin}, "go", installArgs...)
}
