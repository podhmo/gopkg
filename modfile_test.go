package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func TestReadToolDirectives_Empty(t *testing.T) {
	dir := t.TempDir()
	modPath := filepath.Join(dir, "go.mod")
	writeFile(t, modPath, `module example.com/mod

go 1.24
`)
	tools, err := readToolDirectives(modPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("expected no tools, got %v", tools)
	}
}

func TestReadToolDirectives_SingleLine(t *testing.T) {
	dir := t.TempDir()
	modPath := filepath.Join(dir, "go.mod")
	writeFile(t, modPath, `module example.com/mod

go 1.24

tool golang.org/x/tools/cmd/goimports
`)
	tools, err := readToolDirectives(modPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 1 || tools[0] != "golang.org/x/tools/cmd/goimports" {
		t.Errorf("unexpected tools: %v", tools)
	}
}

func TestReadToolDirectives_Block(t *testing.T) {
	dir := t.TempDir()
	modPath := filepath.Join(dir, "go.mod")
	writeFile(t, modPath, `module example.com/mod

go 1.24

tool (
	golang.org/x/tools/cmd/goimports
	github.com/some/tool
)
`)
	tools, err := readToolDirectives(modPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %v", tools)
	}
	if tools[0] != "golang.org/x/tools/cmd/goimports" {
		t.Errorf("tools[0] = %q", tools[0])
	}
	if tools[1] != "github.com/some/tool" {
		t.Errorf("tools[1] = %q", tools[1])
	}
}

func TestReadToolDirectives_InlineComment(t *testing.T) {
	dir := t.TempDir()
	modPath := filepath.Join(dir, "go.mod")
	writeFile(t, modPath, `module example.com/mod

go 1.24

tool (
	golang.org/x/tools/cmd/goimports // formatting
)
`)
	tools, err := readToolDirectives(modPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 1 || tools[0] != "golang.org/x/tools/cmd/goimports" {
		t.Errorf("unexpected tools: %v", tools)
	}
}

func TestReadToolDirectives_MultipleBlocks(t *testing.T) {
	dir := t.TempDir()
	modPath := filepath.Join(dir, "go.mod")
	writeFile(t, modPath, `module example.com/mod

go 1.24

tool golang.org/x/tools/cmd/goimports

tool (
	github.com/some/tool
	github.com/other/tool
)
`)
	tools, err := readToolDirectives(modPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %v", tools)
	}
}

func TestReadModuleName_Simple(t *testing.T) {
	dir := t.TempDir()
	modPath := filepath.Join(dir, "go.mod")
	writeFile(t, modPath, "module github.com/podhmo/gopkg\n\ngo 1.24\n")

	got, err := readModuleName(modPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "github.com/podhmo/gopkg" {
		t.Errorf("got %q, want %q", got, "github.com/podhmo/gopkg")
	}
}

func TestReadModuleName_Missing(t *testing.T) {
	dir := t.TempDir()
	modPath := filepath.Join(dir, "go.mod")
	writeFile(t, modPath, "go 1.24\n")

	_, err := readModuleName(modPath)
	if err == nil {
		t.Fatal("expected error when module directive is missing, got nil")
	}
}
