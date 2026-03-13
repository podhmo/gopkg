package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const echoArgsMain = `package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println(strings.Join(os.Args[1:], " "))
}
`

// TestSplitAtDashDash verifies the "--" splitting logic used by cmdRun.
func TestSplitAtDashDash(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		before []string
		after  []string
	}{
		{
			name:   "no separator",
			args:   []string{"./foo.go", "arg1"},
			before: []string{"./foo.go", "arg1"},
			after:  nil,
		},
		{
			name:   "separator at start",
			args:   []string{"--", "./bar.go"},
			before: []string{},
			after:  []string{"./bar.go"},
		},
		{
			name:   "separator in middle",
			args:   []string{"./foo.go", "--", "./bar.go"},
			before: []string{"./foo.go"},
			after:  []string{"./bar.go"},
		},
		{
			name:   "multiple args after separator",
			args:   []string{"-v", "./foo.go", "--", "arg1", "arg2"},
			before: []string{"-v", "./foo.go"},
			after:  []string{"arg1", "arg2"},
		},
		{
			name:   "empty args",
			args:   nil,
			before: nil,
			after:  nil,
		},
		{
			name:   "only separator",
			args:   []string{"--"},
			before: []string{},
			after:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before, after := splitAtDashDash(tt.args)
			if len(before) != len(tt.before) {
				t.Fatalf("before: got %v, want %v", before, tt.before)
			}
			for i := range before {
				if before[i] != tt.before[i] {
					t.Errorf("before[%d]: got %q, want %q", i, before[i], tt.before[i])
				}
			}
			if len(after) != len(tt.after) {
				t.Fatalf("after: got %v, want %v", after, tt.after)
			}
			for i := range after {
				if after[i] != tt.after[i] {
					t.Errorf("after[%d]: got %q, want %q", i, after[i], tt.after[i])
				}
			}
		})
	}
}

// TestRunRunFrom_NoArgs verifies that runRunFrom builds and runs the binary.
func TestRunRunFrom_NoArgs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testrun\n\ngo 1.21\n")
	writeFile(t, filepath.Join(root, "main.go"), simpleMain)

	if err := runRunFrom(root, root, false, nil, nil); err != nil {
		t.Fatalf("runRunFrom: %v", err)
	}
}

// TestRunRunFrom_WithRunArgs verifies that args after "--" are forwarded to the
// binary.
func TestRunRunFrom_WithRunArgs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testrun\n\ngo 1.21\n")
	writeFile(t, filepath.Join(root, "main.go"), echoArgsMain)

	// Capture stdout.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	runErr := runRunFrom(root, root, false, nil, []string{"hello", "world"})

	w.Close()
	os.Stdout = origStdout
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading stdout: %v", err)
	}

	if runErr != nil {
		t.Fatalf("runRunFrom: %v", runErr)
	}

	output := buf.String()
	if !strings.Contains(output, "hello") || !strings.Contains(output, "world") {
		t.Errorf("expected 'hello world' in output, got: %q", output)
	}
}

// TestRunRunFrom_VerboseFlagPassedToGo verifies that -v is forwarded to the
// underlying go build command.
func TestRunRunFrom_VerboseFlagPassedToGo(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testrun\n\ngo 1.21\n")
	writeFile(t, filepath.Join(root, "main.go"), simpleMain)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	runRunFrom(root, root, true, nil, nil) //nolint:errcheck

	w.Close()
	os.Stdout = origStdout
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading stdout: %v", err)
	}

	if !strings.Contains(buf.String(), " -v") {
		t.Errorf("expected -v flag in logged command, got: %q", buf.String())
	}
}

// TestRunRunFrom_ExplicitPackage verifies that an explicit package path is
// forwarded to go build.
func TestRunRunFrom_ExplicitPackage(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/testrun\n\ngo 1.21\n")
	subDir := filepath.Join(root, "cmd", "hello")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(subDir, "main.go"), simpleMain)

	if err := runRunFrom(root, root, false, []string{"./cmd/hello"}, nil); err != nil {
		t.Fatalf("runRunFrom with explicit package: %v", err)
	}
}
