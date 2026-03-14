package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestADRDocumentsForSubcommandsExistAndHaveRequiredMetadata(t *testing.T) {
	t.Parallel()

	files := []string{
		"001-init-subcommand.md",
		"002-install-subcommand.md",
		"003-upgrade-subcommand.md",
		"004-format-subcommand.md",
		"005-lint-subcommand.md",
		"006-build-subcommand.md",
		"007-run-subcommand.md",
		"008-resolve-subcommand.md",
		"009-adr-workflow-standard.md",
	}

	for _, name := range files {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join("docs", "adr", name)
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ERROR: missing ADR file %s\n  WHY: Each subcommand and ADR workflow requires a dedicated ADR document.\n  FIX: Create %s following docs/adr-workflow.md", path, path)
			}
			s := string(b)

			for _, required := range []string{
				"- Date: 2026-03-14",
				"- Status: Accepted",
				"## Context",
				"## Decision",
				"## Consequences",
			} {
				if !strings.Contains(s, required) {
					t.Fatalf("ERROR: %s is missing required section/metadata %q\n  WHY: ADR format is standardized by docs/adr-workflow.md.\n  FIX: Add %q to %s", path, required, required, path)
				}
			}
		})
	}
}

func TestADRWorkflowDocsAndAgentRouting(t *testing.T) {
	t.Parallel()

	workflowPaths := []string{
		filepath.Join("docs", "adr-workflow.md"),
		filepath.Join("docs", "standards", "adr-workflow.md"),
	}

	for _, path := range workflowPaths {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("missing workflow standard document %s: %v", path, err)
			}
			s := string(b)
			for _, required := range []string{
				"# ADR Workflow Standard",
				"## 1. Immutability Principle",
				"## 2. ADR Structure",
				"## 3. The Archgate Pattern",
			} {
				if !strings.Contains(s, required) {
					t.Fatalf("%s must contain %q", path, required)
				}
			}
		})
	}

	agents, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	agentText := string(agents)
	for _, required := range []string{
		"## 📍 Pointers & Routing",
		"docs/standards/adr-workflow.md",
		"## 🚫 Strict Prohibitions (NEVER DO THESE)",
		"DO NOT overwrite/delete existing ADRs",
		"DO NOT invent architectural rules",
	} {
		if !strings.Contains(agentText, required) {
			t.Fatalf("AGENTS.md must contain %q", required)
		}
	}
}
