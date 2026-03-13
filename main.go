package main

import (
	"flag"
	"fmt"
	"os"
)

const usage = `gopkg – selfish Go package manager

Usage:
  gopkg <command> [flags]

Commands:
  install   Run go mod tidy (and optionally install dev tools)
  upgrade   Run go get -u ./... (and optionally upgrade dev tools)
  format    Run goimports -w ./... (and optionally go fix ./...)
  lint      Run go vet ./...

Run 'gopkg <command> -help' for per-command flags.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var err error
	switch args[0] {
	case "install":
		err = cmdInstall(args[1:])
	case "upgrade":
		err = cmdUpgrade(args[1:])
	case "format":
		err = cmdFormat(args[1:])
	case "lint":
		err = cmdLint(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		flag.Usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func cmdInstall(args []string) error {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	dev := fs.Bool("dev", false, "also install tools from go.mod tool directives")
	fs.Parse(args) //nolint:errcheck // ExitOnError
	return runInstall(*dev)
}

func cmdUpgrade(args []string) error {
	fs := flag.NewFlagSet("upgrade", flag.ExitOnError)
	dev := fs.Bool("dev", false, "also upgrade tools from go.mod tool directives")
	fs.Parse(args) //nolint:errcheck // ExitOnError
	return runUpgrade(*dev)
}

func cmdFormat(args []string) error {
	fs := flag.NewFlagSet("format", flag.ExitOnError)
	fix := fs.Bool("fix", false, "run go fix ./... before goimports")
	fs.Parse(args) //nolint:errcheck // ExitOnError
	return runFormat(*fix)
}

func cmdLint(args []string) error {
	fs := flag.NewFlagSet("lint", flag.ExitOnError)
	fs.Parse(args) //nolint:errcheck // ExitOnError
	return runLint()
}
