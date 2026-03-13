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

	formatErr := runFormatFrom(root, false)

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
	runFormatFrom(root, false) //nolint:errcheck

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
