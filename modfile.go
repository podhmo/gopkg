package main

import (
	"bufio"
	"os"
	"strings"
)

// readModuleName parses the module directive from a go.mod file at modPath
// and returns the module path. It returns an empty string when the directive
// is absent.
func readModuleName(modPath string) (string, error) {
	f, err := os.Open(modPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Strip inline comments.
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", scanner.Err()
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

		if strings.HasPrefix(line, "tool ") {
			rest := strings.TrimSpace(strings.TrimPrefix(line, "tool "))
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
