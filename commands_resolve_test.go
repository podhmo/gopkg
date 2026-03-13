package main

import (
	"bytes"
	"path/filepath"
	"strings"
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

// TestRunResolveFrom verifies that runResolveFrom writes the resolved import
// path to the writer when given a relative path argument.
func TestRunResolveFrom(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testmod\n\ngo 1.21\n")

	var buf bytes.Buffer
	if err := runResolveFrom(&buf, root, root, []string{"./greet"}); err != nil {
		t.Fatalf("runResolveFrom: %v", err)
	}

	got := strings.TrimSpace(buf.String())
	want := "example.com/testmod/greet"
	if got != want {
		t.Errorf("runResolveFrom output = %q, want %q", got, want)
	}
}

// TestRunResolveFrom_MultipleArgs verifies that runResolveFrom writes one
// resolved path per line for multiple arguments.
func TestRunResolveFrom_MultipleArgs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testmod\n\ngo 1.21\n")

	var buf bytes.Buffer
	args := []string{".", "./cmd/tool", "github.com/other/pkg"}
	if err := runResolveFrom(&buf, root, root, args); err != nil {
		t.Fatalf("runResolveFrom: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	wants := []string{
		"example.com/testmod",
		"example.com/testmod/cmd/tool",
		"github.com/other/pkg",
	}
	if len(lines) != len(wants) {
		t.Fatalf("runResolveFrom output lines = %d, want %d\ngot: %q", len(lines), len(wants), buf.String())
	}
	for i, want := range wants {
		if lines[i] != want {
			t.Errorf("line[%d] = %q, want %q", i, lines[i], want)
		}
	}
}
