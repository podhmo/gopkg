package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestModulePathFromDir verifies that modulePathFromDir correctly infers a
// module path from a directory path that contains "github.com/".
func TestModulePathFromDir(t *testing.T) {
	tests := []struct {
		dir     string
		want    string
		wantErr bool
	}{
		{
			dir:  "/home/user/go/src/github.com/myorg/myrepo",
			want: "github.com/myorg/myrepo",
		},
		{
			dir:  "/home/user/go/src/github.com/myorg/myrepo/subpkg",
			want: "github.com/myorg/myrepo/subpkg",
		},
		{
			dir:     "/home/user/go/src/example.com/myrepo",
			wantErr: true,
		},
		{
			dir:     "/tmp/myproject",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		got, err := modulePathFromDir(tc.dir)
		if tc.wantErr {
			if err == nil {
				t.Errorf("modulePathFromDir(%q): expected error, got nil (result %q)", tc.dir, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("modulePathFromDir(%q): unexpected error: %v", tc.dir, err)
			continue
		}
		if got != tc.want {
			t.Errorf("modulePathFromDir(%q) = %q, want %q", tc.dir, got, tc.want)
		}
	}
}

// TestRunInitFrom_ExplicitModulePath verifies that runInitFrom initializes a
// Go module with the provided module path and adds goimports as a tool.
func TestRunInitFrom_ExplicitModulePath(t *testing.T) {
	root := t.TempDir()

	if err := runInitFrom(root, "example.com/testinit"); err != nil {
		t.Fatalf("runInitFrom: %v", err)
	}

	// go.mod must exist.
	modPath := filepath.Join(root, "go.mod")
	if _, err := os.Stat(modPath); err != nil {
		t.Fatalf("expected go.mod to exist: %v", err)
	}

	// The module declaration must match the provided path.
	modName, err := readModuleName(modPath)
	if err != nil {
		t.Fatalf("readModuleName: %v", err)
	}
	if modName != "example.com/testinit" {
		t.Errorf("module name = %q, want %q", modName, "example.com/testinit")
	}

	// goimports must be listed as a tool directive.
	tools, err := readToolDirectives(modPath)
	if err != nil {
		t.Fatalf("readToolDirectives: %v", err)
	}
	found := false
	for _, tool := range tools {
		if tool == goimportsTool {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected %q in tool directives, got: %v", goimportsTool, tools)
	}
}

// TestRunInitFrom_InferredModulePath verifies that runInitFrom infers the
// module path from the directory when no explicit path is given.
func TestRunInitFrom_InferredModulePath(t *testing.T) {
	// Create a temp directory whose path includes "github.com/" so that the
	// auto-detection logic can extract a module path.
	base := t.TempDir()
	root := filepath.Join(base, "github.com", "testorg", "testrepo")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	if err := runInitFrom(root, ""); err != nil {
		t.Fatalf("runInitFrom: %v", err)
	}

	modPath := filepath.Join(root, "go.mod")
	modName, err := readModuleName(modPath)
	if err != nil {
		t.Fatalf("readModuleName: %v", err)
	}

	want := "github.com/testorg/testrepo"
	if modName != want {
		t.Errorf("module name = %q, want %q", modName, want)
	}
}

// TestRunInitFrom_NoGithubInPath verifies that runInitFrom returns an error
// when the directory path cannot be used to infer a module path.
func TestRunInitFrom_NoGithubInPath(t *testing.T) {
	root := t.TempDir()

	err := runInitFrom(root, "")
	if err == nil {
		t.Fatal("expected error when directory path has no github.com prefix, got nil")
	}
}
