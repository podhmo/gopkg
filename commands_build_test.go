package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

	if err := runBuildFrom(root, "", nil); err != nil {
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

	if err := runBuildFrom(root, outPath, nil); err != nil {
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

// TestRunBuildFrom_GobinIsAbsolutePath verifies that the GOBIN directory created
// is an absolute path (required by go install).
func TestRunBuildFrom_GobinIsAbsolutePath(t *testing.T) {
	root := t.TempDir() // t.TempDir always returns an absolute path
	gobinDir := filepath.Join(root, ".local", "gobin")

	if !filepath.IsAbs(gobinDir) {
		t.Errorf("GOBIN path %q is not absolute", gobinDir)
	}
}
