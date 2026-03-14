package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// isVersionComponent reports whether s is a major-version path component: the
// letter "v" followed by a decimal integer ≥ 2 (e.g. "v2", "v3", "v10").
// "v0" and "v1" are not considered major-version components by convention.
func isVersionComponent(s string) bool {
	if len(s) < 2 || s[0] != 'v' {
		return false
	}
	for _, b := range []byte(s[1:]) {
		if b < '0' || b > '9' {
			return false
		}
	}
	// Single digit: must be ≥ 2.
	if len(s) == 2 {
		return s[1] >= '2'
	}
	// Multiple digits: no leading zeros and value is implicitly ≥ 2.
	return s[1] != '0'
}

// stripVersionSuffix removes a major-version suffix from importPath if one is
// present. Two suffix styles are recognised:
//
//   - "/vN" – a trailing slash-separated path component where N ≥ 2, used by
//     the Go module system (e.g. "github.com/foo/bar/v3" → "github.com/foo/bar").
//   - ".vN" – a dot-separated version extension on the last path component,
//     used by gopkg.in (e.g. "gopkg.in/yaml.v3" → "gopkg.in/yaml").
//
// If no recognised suffix is present, importPath is returned unchanged.
func stripVersionSuffix(importPath string) string {
	// "/vN" style: inspect the last slash-separated component.
	if idx := strings.LastIndex(importPath, "/"); idx >= 0 {
		if isVersionComponent(importPath[idx+1:]) {
			return importPath[:idx]
		}
	}
	// ".vN" style: inspect the last dot within the last path component.
	lastSlash := strings.LastIndex(importPath, "/")
	lastComp := importPath[lastSlash+1:] // safe when lastSlash == -1
	if dotIdx := strings.LastIndex(lastComp, "."); dotIdx >= 0 {
		if isVersionComponent(lastComp[dotIdx+1:]) {
			return importPath[:lastSlash+1+dotIdx]
		}
	}
	return importPath
}

// importPathLastComponent returns the last meaningful element of importPath
// after stripping any major-version suffix.
//
// Examples:
//
//	"github.com/foo/bar"    → "bar"
//	"github.com/foo/bar/v3" → "bar"
//	"gopkg.in/yaml.v3"      → "yaml"
//	"fmt"                    → "fmt"
func importPathLastComponent(importPath string) string {
	stripped := stripVersionSuffix(importPath)
	if idx := strings.LastIndex(stripped, "/"); idx >= 0 {
		return stripped[idx+1:]
	}
	return stripped
}

// needsImportAlias reports whether an import of importPath whose declared
// package name is packageName should carry an explicit alias in the import
// declaration.
//
// An alias is needed when packageName differs from the last non-version
// component of importPath. For example, "github.com/foo/go-bar" imported as
// package "bar" needs the alias "bar"; but "github.com/foo/bar/v3" imported
// as package "bar" does not (the last non-version component already matches).
func needsImportAlias(importPath, packageName string) bool {
	return importPathLastComponent(importPath) != packageName
}

// formatImportSpec returns the Go import spec string for importPath using its
// declared packageName. When an alias is required (i.e. the package name
// differs from the last non-version component of importPath), the alias is
// prepended:
//
//	packageName "importPath"
//
// Otherwise the plain quoted path is returned:
//
//	"importPath"
func formatImportSpec(importPath, packageName string) string {
	if needsImportAlias(importPath, packageName) {
		return packageName + ` "` + importPath + `"`
	}
	return `"` + importPath + `"`
}

// fixImportAliasesInFile rewrites the Go source file at path so that imports
// whose declared package name differs from the last non-version component of
// their import path are given an explicit alias. pkgNames maps import paths to
// their declared package names (as reported by go list).
//
// Files that require no changes are left untouched. The function returns true
// when the file was rewritten.
func fixImportAliasesInFile(path string, pkgNames map[string]string) (bool, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		// Not valid Go – skip silently.
		return false, nil
	}

	changed := false
	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Leave blank imports (_) and dot imports (.) alone; also skip any
		// import that already carries an explicit alias.
		if imp.Name != nil {
			continue
		}

		pkgName, ok := pkgNames[importPath]
		if !ok {
			continue // no information available for this import
		}

		if !needsImportAlias(importPath, pkgName) {
			continue // name already matches the path component
		}

		imp.Name = ast.NewIdent(pkgName)
		changed = true
	}

	if !changed {
		return false, nil
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return false, fmt.Errorf("formatting %s: %w", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// collectGoFiles returns the paths of all Go source files within the
// directories described by patterns (relative to root). Patterns that do not
// start with "." (i.e. non-local import paths) are ignored. Hidden directories
// and "vendor" directories are skipped.
func collectGoFiles(root string, patterns []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, pat := range patterns {
		// Only process local/relative patterns.
		if pat != "." && !strings.HasPrefix(pat, "./") && !strings.HasPrefix(pat, "../") {
			continue
		}

		dir := filepath.Join(root, pat)
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // skip unreadable entries
			}
			if d.IsDir() {
				name := d.Name()
				// Skip hidden directories (e.g. .git) and vendor.
				if name != "." && (name[0] == '.' || name == "vendor" || name == "testdata") {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".go") && !seen[path] {
				files = append(files, path)
				seen[path] = true
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// collectPackageNames runs "go list -json -deps ./..." in root and returns a
// map from import path to the package's declared name. If the command fails
// (e.g. the module is not yet valid), an empty map is returned so that alias
// fixing degrades gracefully.
func collectPackageNames(root string) (map[string]string, error) {
	cmd := exec.Command("go", "list", "-json", "-deps", "./...")
	cmd.Dir = root
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	// Discard stderr; failures are communicated via the exit code.
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return map[string]string{}, nil
	}

	names := make(map[string]string)
	dec := json.NewDecoder(&stdout)
	for dec.More() {
		var pkg struct {
			ImportPath string `json:"ImportPath"`
			Name       string `json:"Name"`
		}
		if err := dec.Decode(&pkg); err != nil {
			continue // skip malformed entries
		}
		if pkg.ImportPath != "" && pkg.Name != "" {
			names[pkg.ImportPath] = pkg.Name
		}
	}
	return names, nil
}

// runFixImportAliasesFrom adds explicit import aliases to Go source files
// within the directories described by patterns (relative to root). An alias is
// added whenever the declared package name differs from the last non-version
// component of the import path (e.g. package "bar" imported from
// "github.com/foo/go-bar" receives the alias "bar").
//
// Package name information is obtained by running "go list -json -deps ./...".
// If that command fails the function returns nil (best-effort operation).
func runFixImportAliasesFrom(root string, patterns []string) error {
	pkgNames, err := collectPackageNames(root)
	if err != nil {
		return err
	}
	if len(pkgNames) == 0 {
		return nil
	}

	files, err := collectGoFiles(root, patterns)
	if err != nil {
		return fmt.Errorf("collecting Go files: %w", err)
	}

	for _, f := range files {
		if _, err := fixImportAliasesInFile(f, pkgNames); err != nil {
			return err
		}
	}
	return nil
}
