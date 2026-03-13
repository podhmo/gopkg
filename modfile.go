package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// readModuleName reads the module declaration from the go.mod file at modPath
// and returns the module path (e.g. "github.com/podhmo/gopkg").
func readModuleName(modPath string) (string, error) {
	f, err := os.Open(modPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(after), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module directive not found in %s", modPath)
}

// readToolDirectives parses the tool directives from a go.mod file at modPath
// and returns the list of module paths declared with the tool directive.
//
// Supported formats:
//
//	tool golang.org/x/tools/cmd/goimports
//
//	tool (
//	    golang.org/x/tools/cmd/goimports
//	    github.com/some/other/tool
//	)
func readToolDirectives(modPath string) ([]string, error) {
	f, err := os.Open(modPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var tools []string
	inBlock := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Strip inline comments.
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		if inBlock {
			if line == ")" {
				inBlock = false
				continue
			}
			if line != "" {
				tools = append(tools, line)
			}
			continue
		}

		if after, ok := strings.CutPrefix(line, "tool "); ok {
			rest := strings.TrimSpace(after)
			if rest == "(" {
				inBlock = true
				continue
			}
			if rest != "" {
				tools = append(tools, rest)
			}
		}
	}

	return tools, scanner.Err()
}
