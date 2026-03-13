package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	if err := runInitFrom(root, "example.com/testinit", false); err != nil {
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

	if err := runInitFrom(root, "", false); err != nil {
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

	err := runInitFrom(root, "", false)
	if err == nil {
		t.Fatal("expected error when directory path has no github.com prefix, got nil")
	}
}

// TestRunInitFrom_CI verifies that runInitFrom with ci=true creates
// .github/workflows/ci.yml containing the current Go version and a
// pull_request trigger with the expected event types.
func TestRunInitFrom_CI(t *testing.T) {
	root := t.TempDir()

	if err := runInitFrom(root, "example.com/testci", true); err != nil {
		t.Fatalf("runInitFrom: %v", err)
	}

	workflowPath := filepath.Join(root, ".github", "workflows", "ci.yml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("expected ci.yml to exist: %v", err)
	}
	content := string(data)

	// Must contain the pull_request trigger with the required event types.
	for _, want := range []string{
		"pull_request",
		"opened",
		"synchronize",
		"reopened",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("ci.yml missing %q; content:\n%s", want, content)
		}
	}

	// Must contain the current Go version.
	goVersion := strings.TrimPrefix(runtime.Version(), "go")
	if !strings.Contains(content, goVersion) {
		t.Errorf("ci.yml missing go-version %q; content:\n%s", goVersion, content)
	}
}

// TestCIWorkflowContent verifies that ciWorkflowContent embeds the provided
// Go version and includes the expected pull_request event types.
func TestCIWorkflowContent(t *testing.T) {
	content := ciWorkflowContent("1.24.0")

	for _, want := range []string{
		"pull_request",
		"opened",
		"synchronize",
		"reopened",
		`go-version: "1.24.0"`,
	} {
		if !strings.Contains(content, want) {
			t.Errorf("ciWorkflowContent missing %q; content:\n%s", want, content)
		}
	}
}

// TestRunInitFrom_GoModAlreadyExists verifies that runInitFrom is idempotent:
// when go.mod already exists in the target directory the call succeeds without
// modifying anything.
func TestRunInitFrom_GoModAlreadyExists(t *testing.T) {
	root := t.TempDir()

	// Create a minimal go.mod so the module is already initialised.
	existingContent := "module example.com/existing\n\ngo 1.21\n"
	modPath := filepath.Join(root, "go.mod")
	if err := os.WriteFile(modPath, []byte(existingContent), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Running runInitFrom must not return an error.
	if err := runInitFrom(root, "example.com/shouldbeskipped", false); err != nil {
		t.Fatalf("runInitFrom returned unexpected error: %v", err)
	}

	// The original go.mod must be unchanged.
	data, err := os.ReadFile(modPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != existingContent {
		t.Errorf("go.mod was modified; got:\n%s\nwant:\n%s", data, existingContent)
	}
}

// TestRunInitFrom_NoCIWorkflow verifies that ci=false does NOT create
// .github/workflows/ci.yml.
func TestRunInitFrom_NoCIWorkflow(t *testing.T) {
	root := t.TempDir()

	if err := runInitFrom(root, "example.com/testno", false); err != nil {
		t.Fatalf("runInitFrom: %v", err)
	}

	workflowPath := filepath.Join(root, ".github", "workflows", "ci.yml")
	if _, err := os.Stat(workflowPath); err == nil {
		t.Fatalf("ci.yml should not exist when ci=false, but it does")
	}
}
