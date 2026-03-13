package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunFormatFrom_GoimportsNotATool verifies that when goimports is not
// listed as a tool dependency in go.mod, runFormatFrom returns an error and
// prints a hint to stderr explaining how to add it.
func TestRunFormatFrom_GoimportsNotATool(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testformat\n\ngo 1.24\n")
	writeFile(t, filepath.Join(root, "main.go"), "package main\n\nfunc main() {}\n")

	// Capture stderr to verify the hint message.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w
	defer func() { os.Stderr = origStderr }()

	formatErr := runFormatFrom(root, false, false, nil)

	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading stderr: %v", err)
	}

	if formatErr == nil {
		t.Fatal("expected an error when goimports is not a tool dependency, got nil")
	}

	hint := buf.String()
	if !strings.Contains(hint, "go get -tool") {
		t.Errorf("expected hint containing 'go get -tool' in stderr, got: %q", hint)
	}
	if !strings.Contains(hint, goimportsTool) {
		t.Errorf("expected hint containing %q in stderr, got: %q", goimportsTool, hint)
	}
}

// TestResolvePattern verifies that resolvePattern correctly converts package
// patterns to goimports-compatible directory paths.
func TestResolvePattern(t *testing.T) {
	const mod = "github.com/example/mymod"
	tests := []struct {
		pattern string
		want    string
	}{
		// "./..." is normalised to "." (goimports walks recursively).
		{"./...", "."},
		// "./foo/..." is normalised to "./foo".
		{"./foo/...", "./foo"},
		// Relative paths without "..." pass through unchanged.
		{"./foo", "./foo"},
		{"../other", "../other"},
		// Module root → ".".
		{mod, "."},
		// Sub-package with wildcard → relative directory.
		{mod + "/foo/...", "./foo"},
		// Sub-package without wildcard → relative directory.
		{mod + "/pkg", "./pkg"},
		// Unrecognised import path passes through stripped of "/...".
		{"github.com/other/pkg/...", "github.com/other/pkg"},
	}
	for _, tc := range tests {
		got := resolvePattern(tc.pattern, mod)
		if got != tc.want {
			t.Errorf("resolvePattern(%q, %q) = %q, want %q", tc.pattern, mod, got, tc.want)
		}
	}
}

// TestResolveFormatPatterns_Empty verifies that an empty pattern list defaults
// to "." (the project root, walked recursively by goimports).
func TestResolveFormatPatterns_Empty(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testformat\n\ngo 1.24\n")

	got, err := resolveFormatPatterns(root, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "." {
		t.Errorf("got %v, want [.]", got)
	}
}

// TestResolveFormatPatterns_ImportPath verifies that an absolute import path
// is converted to a relative directory path.
func TestResolveFormatPatterns_ImportPath(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module github.com/example/mymod\n\ngo 1.24\n")

	got, err := resolveFormatPatterns(root, []string{"github.com/example/mymod/sub/..."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0] != "./sub" {
		t.Errorf("got %v, want [./sub]", got)
	}
}

// TestRunFormatFrom_LocalFlagPassedToGoimports verifies that runFormatFrom
// invokes goimports with the -local flag set to the module name from go.mod.
func TestRunFormatFrom_LocalFlagPassedToGoimports(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/mymod\n\ngo 1.24\n")
	writeFile(t, filepath.Join(root, "main.go"), "package main\n\nfunc main() {}\n")

	// Capture stdout to verify the -local flag is logged in the command invocation.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	// runFormatFrom will fail because goimports is not a tool dependency, but
	// the run() helper logs the command to stdout before executing it.
	runFormatFrom(root, false, false, nil) //nolint:errcheck

	w.Close()
	os.Stdout = origStdout
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading stdout: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "-local") {
		t.Errorf("expected -local flag in logged command, got: %q", output)
	}
	if !strings.Contains(output, "example.com/mymod") {
		t.Errorf("expected module name in logged command, got: %q", output)
	}
}

// TestRunFormatFrom_VerboseFlagPassedToGoimports verifies that when verbose is
// true, runFormatFrom passes -v to goimports in the logged command.
func TestRunFormatFrom_VerboseFlagPassedToGoimports(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/mymod\n\ngo 1.24\n")
	writeFile(t, filepath.Join(root, "main.go"), "package main\n\nfunc main() {}\n")

	// Capture stdout to verify the -v flag appears in the logged command.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	// runFormatFrom will fail because goimports is not a tool dependency, but
	// the run() helper logs the command to stdout before executing it.
	runFormatFrom(root, false, true, nil) //nolint:errcheck

	w.Close()
	os.Stdout = origStdout
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading stdout: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, " -v") {
		t.Errorf("expected -v flag in logged command, got: %q", output)
	}
}
