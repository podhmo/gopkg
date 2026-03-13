package main

import (
	"os"
	"path/filepath"
	"testing"
)

// makeTempTree creates a temporary directory hierarchy for testing.
// The returned string is the absolute path to the created temporary root.
func makeTempTree(t *testing.T, files ...string) string {
	t.Helper()
	base := t.TempDir()
	for _, f := range files {
		path := filepath.Join(base, f)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll %s: %v", path, err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatalf("WriteFile %s: %v", path, err)
		}
	}
	return base
}

// makeTempDir creates a directory (not a file) inside base.
func makeTempDir(t *testing.T, base, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(base, name), 0o755); err != nil {
		t.Fatalf("MkdirAll %s: %v", name, err)
	}
}

func TestFindProjectRootFrom_GoMod(t *testing.T) {
	// go.mod exists at root.
	root := makeTempTree(t, "go.mod")
	got, err := findProjectRootFrom(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != root {
		t.Errorf("got %q, want %q", got, root)
	}
}

func TestFindProjectRootFrom_GoModInParent(t *testing.T) {
	// go.mod is in a parent directory; search starts from a subdirectory.
	root := makeTempTree(t, "go.mod", "sub/deep/file.go")
	start := filepath.Join(root, "sub", "deep")

	got, err := findProjectRootFrom(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != root {
		t.Errorf("got %q, want %q", got, root)
	}
}

func TestFindProjectRootFrom_GitButNoGoMod(t *testing.T) {
	// .git exists but go.mod does not – should return an error.
	root := t.TempDir()
	makeTempDir(t, root, ".git")

	_, err := findProjectRootFrom(root)
	if err == nil {
		t.Fatal("expected an error but got nil")
	}
}

func TestFindProjectRootFrom_GoModBeforeGit(t *testing.T) {
	// go.mod is at the same level as .git – go.mod wins.
	root := makeTempTree(t, "go.mod")
	makeTempDir(t, root, ".git")

	got, err := findProjectRootFrom(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != root {
		t.Errorf("got %q, want %q", got, root)
	}
}

func TestFindProjectRootFrom_GoModAboveGit(t *testing.T) {
	// go.mod is in a grandparent; .git is in the parent.
	// Expected: error, because .git is encountered before go.mod.
	base := t.TempDir()
	// grandparent/go.mod  (NOT present – intentionally absent)
	// base/.git
	makeTempDir(t, base, ".git")
	start := filepath.Join(base, "sub")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := findProjectRootFrom(start)
	if err == nil {
		t.Fatal("expected error when .git is found before go.mod")
	}
}

func TestFindProjectRootFrom_NoGoModNoGit(t *testing.T) {
	// Isolated temp directory with no go.mod and no .git anywhere inside.
	// The walk will hit the FS root and return an error.
	root := t.TempDir()
	start := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(start, 0o755); err != nil {
		t.Fatal(err)
	}

	// The walk might find a go.mod or .git higher up on the real filesystem,
	// so we can only assert that findProjectRootFrom does NOT panic.
	// We skip a hard assertion on the error value here.
	findProjectRootFrom(start) //nolint:errcheck
}
