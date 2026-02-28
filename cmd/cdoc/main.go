package main

import (
	"flag"
	"fmt"
	"os"
)

// main parses CLI flags and runs cdoc.
func main() {
	rootDir := flag.String("root", ".", "Project root directory")
	output := flag.String("output", "", "Output file path (default: stdout)")
	prefix := flag.String("prefix", "RSERVER", "Environment variable prefix")
	extraWarn := flag.Bool("extrawarn", false, "Include extra warnings for missing descriptions/groups")
	failOnWarning := flag.Bool("fail-on-warning", false, "Exit with non-zero status if any warnings are emitted")
	parseMode := flag.String("parse-mode", "types", "Parse mode: ast or types")
	flag.Parse()

	if err := runWithMode(*rootDir, *output, *prefix, *extraWarn, *failOnWarning, *parseMode); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
