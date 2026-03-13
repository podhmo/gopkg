package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// runFormat runs `go tool golang.org/x/tools/cmd/goimports -w ./...` and,
// when fix is true, runs `go fix ./...` first.
func runFormat(fix bool) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}
	return runFormatFrom(root, fix)
}

// runFormatFrom is the testable core of runFormat.
func runFormatFrom(root string, fix bool) error {
	if fix {
		if err := run(root, "go", "fix", "./..."); err != nil {
			return err
		}
	}

	moduleName, err := readModuleName(filepath.Join(root, "go.mod"))
	if err != nil {
		return fmt.Errorf("reading module name: %w", err)
	}

	goimportsArgs := []string{"tool", goimportsTool}
	if moduleName != "" {
		goimportsArgs = append(goimportsArgs, "-local", moduleName)
	}
	goimportsArgs = append(goimportsArgs, "-w", "./...")

	if err := run(root, "go", goimportsArgs...); err != nil {
		fmt.Fprintf(os.Stderr, "\nhint: to use gopkg format, add goimports as a tool dependency:\n  go get -tool %s@latest\n", goimportsTool)
		return err
	}
	return nil
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

// runBuild builds the packages.  When output is empty it uses go install with
// GOBIN set to <root>/.local/gobin so that the build cache is leveraged.
// When output is non-empty it uses go build -o <output>.
func runBuild(output string, pkgs []string) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}
	return runBuildFrom(root, output, pkgs)
}

// runBuildFrom is the testable core of runBuild.
func runBuildFrom(root, output string, pkgs []string) error {
	if output != "" {
		goArgs := append([]string{"build", "-o", output}, pkgs...)
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
	return runWithEnv(root, map[string]string{"GOBIN": gobin}, "go", append([]string{"install"}, pkgs...)...)
}
