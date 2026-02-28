package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
)

// main parses CLI flags and runs cdoc.
func main() {
	rootDir := flag.String("root", ".", "Project root directory")
	output := flag.String("output", "", "Output file path (default: stdout)")
	prefix := flag.String("prefix", "RSERVER", "Environment variable prefix")
	extraWarn := flag.Bool("extrawarn", false, "Include extra warnings for missing descriptions/groups")
	failOnWarning := flag.Bool("fail-on-warning", false, "Exit with non-zero status if any warnings are emitted")
	parseMode := flag.String("parse-mode", string(engine.DefaultParseMode), "Parse mode: ast or types")
	flag.Parse()

	if err := run(runOptions{
		rootDir:       *rootDir,
		output:        *output,
		envPrefix:     *prefix,
		extraWarn:     *extraWarn,
		failOnWarning: *failOnWarning,
		parseMode:     engine.ParseMode(*parseMode),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type runOptions struct {
	rootDir       string
	output        string
	envPrefix     string
	extraWarn     bool
	failOnWarning bool
	parseMode     engine.ParseMode
}

// run executes cdoc with explicit options.
func run(opts runOptions) error {
	policy := model.DefaultWarningPolicy()
	if opts.failOnWarning {
		policy = model.StrictWarningPolicy()
	}
	return engine.Run(engine.RunOptions{
		RootDir:       opts.rootDir,
		Output:        opts.output,
		EnvPrefix:     opts.envPrefix,
		ExtraWarnings: opts.extraWarn,
		ParseMode:     opts.parseMode,
		Policy:        &policy,
	})
}
