package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveDocArg_RelativePaths verifies that relative paths are converted
// to full module import paths.
func TestResolveDocArg_RelativePaths(t *testing.T) {
	root := "/home/user/github.com/example/mymod"
	pwd := root
	moduleName := "github.com/example/mymod"

	tests := []struct {
		arg  string
		want string
	}{
		// Relative paths should be resolved.
		{".", moduleName},
		{"./foo", moduleName + "/foo"},
		{"./foo/bar", moduleName + "/foo/bar"},
		// Flags and symbols pass through unchanged.
		{"-all", "-all"},
		{"-u", "-u"},
		{"Symbol", "Symbol"},
		// Absolute import paths pass through unchanged.
		{"github.com/other/pkg", "github.com/other/pkg"},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got := resolveDocArg(tt.arg, pwd, root, moduleName)
			if got != tt.want {
				t.Errorf("resolveDocArg(%q) = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}

// TestResolveDocArg_PwdSubdir verifies that paths are resolved correctly when
// the working directory is a subdirectory of the module root.
func TestResolveDocArg_PwdSubdir(t *testing.T) {
	root := "/home/user/github.com/example/mymod"
	pwd := filepath.Join(root, "cmd", "tool")
	moduleName := "github.com/example/mymod"

	tests := []struct {
		arg  string
		want string
	}{
		// "." resolves to the package at pwd.
		{".", moduleName + "/cmd/tool"},
		// "./sub" resolves relative to pwd.
		{"./sub", moduleName + "/cmd/tool/sub"},
		// "../other" goes up from pwd.
		{"../other", moduleName + "/cmd/other"},
		// Flags still pass through.
		{"-all", "-all"},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			got := resolveDocArg(tt.arg, pwd, root, moduleName)
			if got != tt.want {
				t.Errorf("resolveDocArg(%q, pwd=%q) = %q, want %q", tt.arg, pwd, got, tt.want)
			}
		})
	}
}

// TestResolveDocArg_OutsideModule verifies that paths outside the module root
// are passed through unchanged.
func TestResolveDocArg_OutsideModule(t *testing.T) {
	root := "/home/user/github.com/example/mymod"
	pwd := root
	moduleName := "github.com/example/mymod"

	// "../../sibling" escapes the module root – should pass through.
	arg := "../../sibling"
	got := resolveDocArg(arg, pwd, root, moduleName)
	if got != arg {
		t.Errorf("resolveDocArg(%q) = %q, want %q (pass-through)", arg, got, arg)
	}
}

// TestRunDocFrom verifies that runDocFrom invokes go doc with the converted
// import path when given a relative path argument.
func TestRunDocFrom(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testdoc\n\ngo 1.21\n")

	// Create a sub-package with an exported symbol so that go doc has something
	// to show.
	subDir := filepath.Join(root, "greet")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	writeFile(t, filepath.Join(subDir, "greet.go"), `// Package greet says hello.
package greet

// Hello returns a greeting.
func Hello() string { return "hello" }
`)

	// Run gopkg doc ./greet from the module root; the relative path should be
	// resolved to example.com/testdoc/greet.
	if err := runDocFrom(root, root, []string{"./greet"}); err != nil {
		t.Fatalf("runDocFrom: %v", err)
	}
}

// TestRunDocFrom_WithFlag verifies that flags such as -all are forwarded to
// go doc unchanged.
func TestRunDocFrom_WithFlag(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testdoc\n\ngo 1.21\n")
	subDir := filepath.Join(root, "greet")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	writeFile(t, filepath.Join(subDir, "greet.go"), `// Package greet says hello.
package greet

// Hello returns a greeting.
func Hello() string { return "hello" }
`)

	// -all is a valid go doc flag and should not cause an error.
	if err := runDocFrom(root, root, []string{"-all", "./greet"}); err != nil {
		t.Fatalf("runDocFrom with -all: %v", err)
	}
}
