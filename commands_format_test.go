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
