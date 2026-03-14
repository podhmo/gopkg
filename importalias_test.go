package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// isVersionComponent
// ---------------------------------------------------------------------------

func TestIsVersionComponent(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// Valid major-version components.
		{"v2", true},
		{"v3", true},
		{"v10", true},
		{"v100", true},
		{"v20", true},

		// v0 and v1 are NOT major-version components by Go convention.
		{"v0", false},
		{"v1", false},

		// Not a version component: missing prefix, extra chars, empty, etc.
		{"", false},
		{"v", false},
		{"2", false},
		{"V2", false},   // case-sensitive
		{"v2a", false},  // non-digit suffix
		{"v2.0", false}, // dot inside
		{"version", false},
		{"v02", false}, // leading zero
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := isVersionComponent(tc.input)
			if got != tc.want {
				t.Errorf("isVersionComponent(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// stripVersionSuffix
// ---------------------------------------------------------------------------

func TestStripVersionSuffix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// "/vN" style (Go module major-version convention).
		{"github.com/foo/bar/v2", "github.com/foo/bar"},
		{"github.com/foo/bar/v3", "github.com/foo/bar"},
		{"github.com/foo/bar/v10", "github.com/foo/bar"},
		{"github.com/foo/bar/v100", "github.com/foo/bar"},

		// v1 is NOT stripped.
		{"github.com/foo/bar/v1", "github.com/foo/bar/v1"},
		// v0 is NOT stripped.
		{"github.com/foo/bar/v0", "github.com/foo/bar/v0"},

		// No version suffix – unchanged.
		{"github.com/foo/bar", "github.com/foo/bar"},
		{"fmt", "fmt"},
		{"os/exec", "os/exec"},

		// ".vN" style (gopkg.in convention).
		{"gopkg.in/yaml.v3", "gopkg.in/yaml"},
		{"gopkg.in/check.v2", "gopkg.in/check"},
		{"gopkg.in/mgo.v2", "gopkg.in/mgo"},

		// gopkg.in v1 is NOT stripped.
		{"gopkg.in/check.v1", "gopkg.in/check.v1"},

		// String that looks like a version component but is not the last segment.
		{"github.com/v2/foo", "github.com/v2/foo"},

		// Token "v2bar" is not a valid version component.
		{"github.com/foo/v2bar", "github.com/foo/v2bar"},

		// Deep path with version.
		{"golang.org/x/text/v2", "golang.org/x/text"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := stripVersionSuffix(tc.input)
			if got != tc.want {
				t.Errorf("stripVersionSuffix(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// importPathLastComponent
// ---------------------------------------------------------------------------

func TestImportPathLastComponent(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// No version suffix.
		{"github.com/foo/bar", "bar"},
		{"github.com/foo/go-bar", "go-bar"},
		{"github.com/foo/bar-go", "bar-go"},
		{"fmt", "fmt"},
		{"os/exec", "exec"},

		// "/vN" stripped before taking last component.
		{"github.com/foo/bar/v2", "bar"},
		{"github.com/foo/bar/v3", "bar"},
		{"github.com/foo/go-bar/v2", "go-bar"},

		// v1 not stripped.
		{"github.com/foo/bar/v1", "v1"},

		// ".vN" stripped (gopkg.in style).
		{"gopkg.in/yaml.v3", "yaml"},
		{"gopkg.in/check.v2", "check"},

		// gopkg.in v1 not stripped.
		{"gopkg.in/check.v1", "check.v1"},

		// Deep paths.
		{"golang.org/x/tools/cmd/goimports", "goimports"},
		{"golang.org/x/tools/cmd/goimports/v2", "goimports"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := importPathLastComponent(tc.input)
			if got != tc.want {
				t.Errorf("importPathLastComponent(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// needsImportAlias
// ---------------------------------------------------------------------------

func TestNeedsImportAlias(t *testing.T) {
	tests := []struct {
		importPath  string
		packageName string
		want        bool
	}{
		// Package name matches last component exactly – no alias needed.
		{"github.com/foo/bar", "bar", false},
		{"fmt", "fmt", false},
		{"os/exec", "exec", false},

		// Package name matches last non-version component – no alias needed.
		{"github.com/foo/bar/v3", "bar", false},
		{"gopkg.in/yaml.v3", "yaml", false},
		{"github.com/foo/go-bar/v2", "go-bar", false},

		// Package name differs from last non-version component – alias needed.
		{"github.com/foo/go-bar", "bar", true},
		{"github.com/foo/bar-go", "bar", true},
		{"github.com/foo/baz-client", "client", true},
		{"github.com/foo/bar/v3", "baz", true},
		{"gopkg.in/yaml.v3", "myyaml", true},

		// Package name differs from path component in unusual ways.
		{"github.com/foo/bar", "baz", true},
		{"github.com/foo/bar", "Bar", true}, // case differs
	}

	for _, tc := range tests {
		name := tc.importPath + "/" + tc.packageName
		t.Run(name, func(t *testing.T) {
			got := needsImportAlias(tc.importPath, tc.packageName)
			if got != tc.want {
				t.Errorf("needsImportAlias(%q, %q) = %v, want %v",
					tc.importPath, tc.packageName, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// formatImportSpec
// ---------------------------------------------------------------------------

func TestFormatImportSpec(t *testing.T) {
	tests := []struct {
		importPath  string
		packageName string
		want        string
	}{
		// No alias needed.
		{"github.com/foo/bar", "bar", `"github.com/foo/bar"`},
		{"fmt", "fmt", `"fmt"`},
		{"github.com/foo/bar/v3", "bar", `"github.com/foo/bar/v3"`},
		{"gopkg.in/yaml.v3", "yaml", `"gopkg.in/yaml.v3"`},

		// Alias needed: package name ≠ last non-version component.
		{"github.com/foo/go-bar", "bar", `bar "github.com/foo/go-bar"`},
		{"github.com/foo/bar-go", "bar", `bar "github.com/foo/bar-go"`},
		{"github.com/foo/baz-client", "client", `client "github.com/foo/baz-client"`},
		{"github.com/foo/bar/v3", "baz", `baz "github.com/foo/bar/v3"`},
		{"github.com/foo/bar", "baz", `baz "github.com/foo/bar"`},

		// gopkg.in with mismatched package name.
		{"gopkg.in/yaml.v3", "myyaml", `myyaml "gopkg.in/yaml.v3"`},
	}

	for _, tc := range tests {
		name := tc.importPath + "+" + tc.packageName
		t.Run(name, func(t *testing.T) {
			got := formatImportSpec(tc.importPath, tc.packageName)
			if got != tc.want {
				t.Errorf("formatImportSpec(%q, %q) = %q, want %q",
					tc.importPath, tc.packageName, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// fixImportAliasesInFile
// ---------------------------------------------------------------------------

// TestFixImportAliasesInFile_AddsAlias verifies that an import whose package
// name differs from the last path component receives an explicit alias.
func TestFixImportAliasesInFile_AddsAlias(t *testing.T) {
	src := `package main

import "github.com/foo/go-bar"

func main() { _ = bar.X }
`
	path := filepath.Join(t.TempDir(), "main.go")
	writeFile(t, path, src)

	pkgNames := map[string]string{
		"github.com/foo/go-bar": "bar",
	}

	changed, err := fixImportAliasesInFile(path, pkgNames)
	if err != nil {
		t.Fatalf("fixImportAliasesInFile: %v", err)
	}
	if !changed {
		t.Fatal("expected file to be changed, but changed=false")
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `bar "github.com/foo/go-bar"`) {
		t.Errorf("expected alias in output:\n%s", got)
	}
}

// TestFixImportAliasesInFile_NoAlias_MatchingName verifies that an import
// whose package name already matches the last path component is untouched.
func TestFixImportAliasesInFile_NoAlias_MatchingName(t *testing.T) {
	src := `package main

import "github.com/foo/bar"

func main() { _ = bar.X }
`
	path := filepath.Join(t.TempDir(), "main.go")
	writeFile(t, path, src)

	pkgNames := map[string]string{
		"github.com/foo/bar": "bar",
	}

	changed, err := fixImportAliasesInFile(path, pkgNames)
	if err != nil {
		t.Fatalf("fixImportAliasesInFile: %v", err)
	}
	if changed {
		got, _ := os.ReadFile(path)
		t.Errorf("expected no change, but file was rewritten:\n%s", got)
	}
}

// TestFixImportAliasesInFile_NoAlias_VersionSuffix verifies that a versioned
// import whose package name matches the last non-version component is not
// aliased (the convention is already satisfied).
func TestFixImportAliasesInFile_NoAlias_VersionSuffix(t *testing.T) {
	src := `package main

import "github.com/foo/bar/v3"

func main() { _ = bar.X }
`
	path := filepath.Join(t.TempDir(), "main.go")
	writeFile(t, path, src)

	pkgNames := map[string]string{
		"github.com/foo/bar/v3": "bar",
	}

	changed, err := fixImportAliasesInFile(path, pkgNames)
	if err != nil {
		t.Fatalf("fixImportAliasesInFile: %v", err)
	}
	if changed {
		got, _ := os.ReadFile(path)
		t.Errorf("expected no change for version-suffix import, but file was rewritten:\n%s", got)
	}
}

// TestFixImportAliasesInFile_PreservesExistingAlias verifies that an import
// that already has an explicit alias is left unchanged, even if our heuristic
// would suggest a different one.
func TestFixImportAliasesInFile_PreservesExistingAlias(t *testing.T) {
	src := `package main

import myalias "github.com/foo/go-bar"

func main() { _ = myalias.X }
`
	path := filepath.Join(t.TempDir(), "main.go")
	writeFile(t, path, src)

	pkgNames := map[string]string{
		"github.com/foo/go-bar": "bar",
	}

	changed, err := fixImportAliasesInFile(path, pkgNames)
	if err != nil {
		t.Fatalf("fixImportAliasesInFile: %v", err)
	}
	if changed {
		got, _ := os.ReadFile(path)
		t.Errorf("expected no change when alias already present, but file was rewritten:\n%s", got)
	}
}

// TestFixImportAliasesInFile_PreservesBlankImport verifies that a blank import
// (_) is left untouched.
func TestFixImportAliasesInFile_PreservesBlankImport(t *testing.T) {
	src := `package main

import _ "github.com/foo/go-bar"
`
	path := filepath.Join(t.TempDir(), "main.go")
	writeFile(t, path, src)

	pkgNames := map[string]string{
		"github.com/foo/go-bar": "bar",
	}

	changed, err := fixImportAliasesInFile(path, pkgNames)
	if err != nil {
		t.Fatalf("fixImportAliasesInFile: %v", err)
	}
	if changed {
		got, _ := os.ReadFile(path)
		t.Errorf("expected blank import to be preserved, but file was rewritten:\n%s", got)
	}
}

// TestFixImportAliasesInFile_PreservesDotImport verifies that a dot import
// (.) is left untouched.
func TestFixImportAliasesInFile_PreservesDotImport(t *testing.T) {
	src := `package main

import . "github.com/foo/go-bar"
`
	path := filepath.Join(t.TempDir(), "main.go")
	writeFile(t, path, src)

	pkgNames := map[string]string{
		"github.com/foo/go-bar": "bar",
	}

	changed, err := fixImportAliasesInFile(path, pkgNames)
	if err != nil {
		t.Fatalf("fixImportAliasesInFile: %v", err)
	}
	if changed {
		got, _ := os.ReadFile(path)
		t.Errorf("expected dot import to be preserved, but file was rewritten:\n%s", got)
	}
}

// TestFixImportAliasesInFile_UnknownPackage verifies that an import not found
// in the pkgNames map is left unchanged.
func TestFixImportAliasesInFile_UnknownPackage(t *testing.T) {
	src := `package main

import "github.com/foo/go-bar"
`
	path := filepath.Join(t.TempDir(), "main.go")
	writeFile(t, path, src)

	changed, err := fixImportAliasesInFile(path, map[string]string{} /* empty */)
	if err != nil {
		t.Fatalf("fixImportAliasesInFile: %v", err)
	}
	if changed {
		got, _ := os.ReadFile(path)
		t.Errorf("expected no change for unknown package, but file was rewritten:\n%s", got)
	}
}

// TestFixImportAliasesInFile_MultipleImports verifies that multiple imports in
// the same file are handled independently.
func TestFixImportAliasesInFile_MultipleImports(t *testing.T) {
	src := `package main

import (
	"fmt"
	"github.com/foo/go-bar"
	"github.com/foo/bar"
	"github.com/foo/baz-client"
)
`
	path := filepath.Join(t.TempDir(), "main.go")
	writeFile(t, path, src)

	pkgNames := map[string]string{
		"github.com/foo/go-bar":    "bar",
		"github.com/foo/bar":       "bar",
		"github.com/foo/baz-client": "client",
	}

	changed, err := fixImportAliasesInFile(path, pkgNames)
	if err != nil {
		t.Fatalf("fixImportAliasesInFile: %v", err)
	}
	if !changed {
		t.Fatal("expected file to be changed")
	}

	got := string(mustReadFile(t, path))
	// "go-bar" needs alias "bar".
	if !strings.Contains(got, `bar "github.com/foo/go-bar"`) {
		t.Errorf("expected alias for go-bar:\n%s", got)
	}
	// "baz-client" needs alias "client".
	if !strings.Contains(got, `client "github.com/foo/baz-client"`) {
		t.Errorf("expected alias for baz-client:\n%s", got)
	}
	// "bar" does NOT need an alias.
	if strings.Contains(got, `bar "github.com/foo/bar"`) {
		t.Errorf("unexpected alias for bar:\n%s", got)
	}
	// Standard library "fmt" is not in pkgNames but should still be present.
	if !strings.Contains(got, `"fmt"`) {
		t.Errorf("fmt import missing:\n%s", got)
	}
}

// TestFixImportAliasesInFile_InvalidGoFile verifies that the function returns
// false and no error for a file that is not valid Go source.
func TestFixImportAliasesInFile_InvalidGoFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.go")
	writeFile(t, path, "this is not valid go source code }{{{")

	changed, err := fixImportAliasesInFile(path, map[string]string{"x": "x"})
	if err != nil {
		t.Fatalf("expected no error for invalid Go file, got: %v", err)
	}
	if changed {
		t.Error("expected changed=false for invalid Go file")
	}
}

// TestFixImportAliasesInFile_GopkgInStyle verifies gopkg.in-style versioned
// imports where the package name matches the non-version component (no alias
// needed) and where it does not (alias needed).
func TestFixImportAliasesInFile_GopkgInStyle(t *testing.T) {
	t.Run("matching", func(t *testing.T) {
		src := `package main

import "gopkg.in/yaml.v3"
`
		path := filepath.Join(t.TempDir(), "main.go")
		writeFile(t, path, src)

		pkgNames := map[string]string{"gopkg.in/yaml.v3": "yaml"}

		changed, err := fixImportAliasesInFile(path, pkgNames)
		if err != nil {
			t.Fatal(err)
		}
		if changed {
			got, _ := os.ReadFile(path)
			t.Errorf("expected no alias for gopkg.in/yaml.v3 with package yaml:\n%s", got)
		}
	})

	t.Run("mismatched", func(t *testing.T) {
		src := `package main

import "gopkg.in/yaml.v3"
`
		path := filepath.Join(t.TempDir(), "main.go")
		writeFile(t, path, src)

		pkgNames := map[string]string{"gopkg.in/yaml.v3": "myyaml"}

		changed, err := fixImportAliasesInFile(path, pkgNames)
		if err != nil {
			t.Fatal(err)
		}
		if !changed {
			t.Fatal("expected file to be changed")
		}
		got := string(mustReadFile(t, path))
		if !strings.Contains(got, `myyaml "gopkg.in/yaml.v3"`) {
			t.Errorf("expected alias myyaml:\n%s", got)
		}
	})
}

// ---------------------------------------------------------------------------
// collectGoFiles
// ---------------------------------------------------------------------------

// TestCollectGoFiles_Basic verifies that .go files in the pattern directory
// are collected and non-Go files are excluded.
func TestCollectGoFiles_Basic(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")
	writeFile(t, filepath.Join(root, "util.go"), "package main\n")
	writeFile(t, filepath.Join(root, "README.md"), "# readme\n")

	files, err := collectGoFiles(root, []string{"."})
	if err != nil {
		t.Fatalf("collectGoFiles: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 .go files, got %d: %v", len(files), files)
	}
}

// TestCollectGoFiles_Recursive verifies that subdirectories are walked.
func TestCollectGoFiles_Recursive(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")
	writeFile(t, filepath.Join(root, "sub", "sub.go"), "package sub\n")

	files, err := collectGoFiles(root, []string{"."})
	if err != nil {
		t.Fatalf("collectGoFiles: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 .go files, got %d: %v", len(files), files)
	}
}

// TestCollectGoFiles_SkipsVendor verifies that the vendor directory is skipped.
func TestCollectGoFiles_SkipsVendor(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")
	writeFile(t, filepath.Join(root, "vendor", "pkg", "pkg.go"), "package pkg\n")

	files, err := collectGoFiles(root, []string{"."})
	if err != nil {
		t.Fatalf("collectGoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 .go file (vendor excluded), got %d: %v", len(files), files)
	}
}

// TestCollectGoFiles_SkipsHiddenDirs verifies that hidden directories are
// skipped.
func TestCollectGoFiles_SkipsHiddenDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")
	writeFile(t, filepath.Join(root, ".git", "hook.go"), "package hook\n")

	files, err := collectGoFiles(root, []string{"."})
	if err != nil {
		t.Fatalf("collectGoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 .go file (.git excluded), got %d: %v", len(files), files)
	}
}

// TestCollectGoFiles_SkipsNonLocalPatterns verifies that import paths not
// starting with "." are silently ignored.
func TestCollectGoFiles_SkipsNonLocalPatterns(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")

	files, err := collectGoFiles(root, []string{"github.com/some/external"})
	if err != nil {
		t.Fatalf("collectGoFiles: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for non-local pattern, got %d: %v", len(files), files)
	}
}

// TestCollectGoFiles_SubPatterns verifies that a "./sub" pattern collects only
// files within that subdirectory.
func TestCollectGoFiles_SubPatterns(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")
	writeFile(t, filepath.Join(root, "sub", "sub.go"), "package sub\n")

	files, err := collectGoFiles(root, []string{"./sub"})
	if err != nil {
		t.Fatalf("collectGoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 .go file, got %d: %v", len(files), files)
	}
	if !strings.HasSuffix(files[0], "sub.go") {
		t.Errorf("expected sub.go, got %q", files[0])
	}
}

// TestCollectGoFiles_NoDuplicates verifies that the same file is not collected
// twice when patterns overlap.
func TestCollectGoFiles_NoDuplicates(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")

	files, err := collectGoFiles(root, []string{".", "."})
	if err != nil {
		t.Fatalf("collectGoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 .go file (no duplicates), got %d: %v", len(files), files)
	}
}

// TestCollectGoFiles_Empty verifies that an empty pattern list produces no
// files.
func TestCollectGoFiles_Empty(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "main.go"), "package main\n")

	files, err := collectGoFiles(root, nil)
	if err != nil {
		t.Fatalf("collectGoFiles: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for nil patterns, got %d", len(files))
	}
}

// ---------------------------------------------------------------------------
// runFixImportAliasesFrom (integration, uses real go list)
// ---------------------------------------------------------------------------

// TestRunFixImportAliasesFrom_FixesAlias is a lightweight integration test
// that verifies runFixImportAliasesFrom rewrites a Go file when the import's
// package name (from go list) differs from the last path component.
//
// To avoid an external network dependency, we reuse the known stdlib package
// "os/exec" whose package name is "exec" – it matches its last component, so
// no rewrite is expected. The real "needs alias" case is exercised by
// fixImportAliasesInFile tests above using a mock pkgNames map.
func TestRunFixImportAliasesFrom_NoopForStdlib(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/aliascheck\n\ngo 1.21\n")
	src := `package main

import "os/exec"

func main() { _ = exec.Command("ls") }
`
	goFile := filepath.Join(root, "main.go")
	writeFile(t, goFile, src)

	if err := runFixImportAliasesFrom(root, []string{"."}); err != nil {
		t.Fatalf("runFixImportAliasesFrom: %v", err)
	}

	// File should be unchanged: "exec" == importPathLastComponent("os/exec").
	got := string(mustReadFile(t, goFile))
	if got != src {
		t.Errorf("expected file unchanged, got:\n%s", got)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return b
}
