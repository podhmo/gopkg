package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestADRWorkflowDefinesRequiredFieldsAndStatusEnum(t *testing.T) {
	t.Parallel()

	paths := []string{
		filepath.Join("docs", "adr-workflow.md"),
		filepath.Join("docs", "standards", "adr-workflow.md"),
	}

	required := []string{
		"### Why these fields are mandatory",
		"### Status enum (closed set)",
		"- `Proposed`",
		"- `Accepted`",
		"- `Superseded by [ADR-XXX]`",
		"- `Deprecated`",
		"- Date: YYYY-MM-DD",
		"- Status:",
		"## Context",
		"## Decision",
		"## Consequences",
		"The `Status` line above shows alternatives; each ADR must choose exactly one value from the enum.",
	}

	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", path, err)
			}
			s := string(b)
			for _, item := range required {
				if !strings.Contains(s, item) {
					t.Fatalf("%s must contain %q", path, item)
				}
			}
		})
	}
}

func TestSubcommandADRsDescribeFixedNonCustomizableBehavior(t *testing.T) {
	t.Parallel()

	paths := []string{
		filepath.Join("docs", "adr", "001-init-subcommand.md"),
		filepath.Join("docs", "adr", "002-install-subcommand.md"),
		filepath.Join("docs", "adr", "003-upgrade-subcommand.md"),
		filepath.Join("docs", "adr", "004-format-subcommand.md"),
		filepath.Join("docs", "adr", "005-lint-subcommand.md"),
		filepath.Join("docs", "adr", "006-build-subcommand.md"),
		filepath.Join("docs", "adr", "007-run-subcommand.md"),
		filepath.Join("docs", "adr", "008-resolve-subcommand.md"),
	}

	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", path, err)
			}
			s := string(b)
			for _, marker := range []string{
				"Fixed and non-customizable in this command:",
				"Variable input:",
			} {
				if !strings.Contains(s, marker) {
					t.Fatalf("%s must contain %q", path, marker)
				}
			}
		})
	}

	formatADR, err := os.ReadFile(filepath.Join("docs", "adr", "004-format-subcommand.md"))
	if err != nil {
		t.Fatalf("reading format ADR: %v", err)
	}
	if !strings.Contains(string(formatADR), "not `gofumpt`") {
		t.Fatalf("format ADR must explicitly say goimports is fixed and gofumpt is not selected")
	}
	if !strings.Contains(string(formatADR), "not `gofmt`") {
		t.Fatalf("format ADR must explicitly say goimports is preferred over gofmt")
	}
	if !strings.Contains(string(formatADR), "automatically adds missing imports") {
		t.Fatalf("format ADR must explicitly say goimports is preferred because it auto-adds missing imports")
	}

	buildADR, err := os.ReadFile(filepath.Join("docs", "adr", "006-build-subcommand.md"))
	if err != nil {
		t.Fatalf("reading build ADR: %v", err)
	}
	if !strings.Contains(string(buildADR), "<moduleRoot>/.local/gobin") {
		t.Fatalf("build ADR must explicitly mention fixed .local/gobin install location")
	}
}
