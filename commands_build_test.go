package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

const simpleMain = `package main

func main() {}
`

// TestRunBuildFrom_NoOutput verifies that, when no -o flag is given, the binary
// is installed into <root>/.local/gobin via go install with GOBIN set.
func TestRunBuildFrom_NoOutput(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testbuild\n\ngo 1.21\n")
	writeFile(t, filepath.Join(root, "main.go"), simpleMain)

	if err := runBuildFrom(root, "", false, nil); err != nil {
		t.Fatalf("runBuildFrom: %v", err)
	}

	gobinDir := filepath.Join(root, ".local", "gobin")
	entries, err := os.ReadDir(gobinDir)
	if err != nil {
		t.Fatalf("reading .local/gobin: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected a binary in .local/gobin but the directory is empty")
	}

	// Verify the installed binary is executable.
	binPath := filepath.Join(gobinDir, entries[0].Name())
	if err := exec.Command(binPath).Run(); err != nil {
		t.Errorf("installed binary %s failed to run: %v", binPath, err)
	}
}

// TestRunBuildFrom_WithOutput verifies that, when -o is given, go build places
// the binary at the specified path.
func TestRunBuildFrom_WithOutput(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testbuild\n\ngo 1.21\n")
	writeFile(t, filepath.Join(root, "main.go"), simpleMain)

	outName := "mybinary"
	if runtime.GOOS == "windows" {
		outName += ".exe"
	}
	outPath := filepath.Join(root, outName)

	if err := runBuildFrom(root, outPath, false, nil); err != nil {
		t.Fatalf("runBuildFrom: %v", err)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected output binary at %s: %v", outPath, err)
	}

	// Verify the built binary is executable.
	if err := exec.Command(outPath).Run(); err != nil {
		t.Errorf("built binary %s failed to run: %v", outPath, err)
	}
}

// TestRunBuildFrom_GobinIsAbsolutePath verifies that GOBIN is always set to an
// absolute path by runBuildFrom, even when root is a relative path.
// go install rejects relative GOBIN values.
func TestRunBuildFrom_GobinIsAbsolutePath(t *testing.T) {
	// Build a minimal module in a temp dir so we can derive a relative root.
	abs := t.TempDir()
	writeFile(t, filepath.Join(abs, "go.mod"), "module example.com/testbuild\n\ngo 1.21\n")
	writeFile(t, filepath.Join(abs, "main.go"), simpleMain)

	// Change into the temp dir so that "." is a valid relative root.
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(abs); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) }) //nolint:errcheck

	// Pass "." as root – this is a relative path.
	if err := runBuildFrom(".", "", false, nil); err != nil {
		t.Fatalf("runBuildFrom with relative root: %v", err)
	}

	// The binary must have landed in .local/gobin.
	gobinDir := filepath.Join(abs, ".local", "gobin")
	entries, err := os.ReadDir(gobinDir)
	if err != nil {
		t.Fatalf("reading .local/gobin: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected a binary in .local/gobin but the directory is empty")
	}
}

// TestRunBuildFrom_VerboseFlagPassedToGo verifies that when verbose is true,
// runBuildFrom passes -v to the underlying go install/build command.
func TestRunBuildFrom_VerboseFlagPassedToGo(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testbuild\n\ngo 1.21\n")
	writeFile(t, filepath.Join(root, "main.go"), simpleMain)

	// Capture stdout to verify -v appears in the logged command invocation.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	runBuildFrom(root, "", true, nil) //nolint:errcheck

	w.Close()
	os.Stdout = origStdout
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading stdout: %v", err)
	}

	if !strings.Contains(buf.String(), " -v") {
		t.Errorf("expected -v flag in logged command, got: %q", buf.String())
	}
}
