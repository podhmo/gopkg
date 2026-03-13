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

// runFormat runs `goimports -w ./...` and, when fix is true, runs `go fix
// ./...` first.
func runFormat(fix bool) error {
	root, err := findProjectRoot()
	if err != nil {
		return err
	}

	if fix {
		if err := run(root, "go", "fix", "./..."); err != nil {
			return err
		}
	}

	return run(root, "goimports", "-w", "./...")
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
