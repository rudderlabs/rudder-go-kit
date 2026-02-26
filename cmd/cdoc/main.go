package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	rootDir := flag.String("root", ".", "Project root directory")
	output := flag.String("output", "", "Output file path (default: stdout)")
	prefix := flag.String("prefix", "RSERVER", "Environment variable prefix")
	warn := flag.Bool("warn", false, "Print warnings for missing descriptions/groups to stderr")
	flag.Parse()

	if err := run(*rootDir, *output, *prefix, *warn); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
