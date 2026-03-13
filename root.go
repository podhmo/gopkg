// Package main implements gopkg, a selfish Go package manager.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// findProjectRoot walks up from the current working directory until it finds
// a directory containing go.mod.  If a .git directory is encountered before
// go.mod, the search stops and an error is returned.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getwd: %w", err)
	}
	return findProjectRootFrom(dir)
}

// findProjectRootFrom is the testable core of findProjectRoot.
func findProjectRootFrom(start string) (string, error) {
	dir := start
	for {
		// go.mod found – this is the project root.
		if exists(filepath.Join(dir, "go.mod")) {
			return dir, nil
		}

		// .git found but no go.mod – stop searching.
		if exists(filepath.Join(dir, ".git")) {
			return "", errors.New("found .git directory but no go.mod; is this a Go module?")
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the filesystem root without finding go.mod.
			return "", errors.New("no go.mod found in directory hierarchy")
		}
		dir = parent
	}
}

// exists reports whether path exists (any file/dir type).
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
