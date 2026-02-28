package main

import (
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
)

// run executes cdoc with the default parse mode.
func run(rootDir, output, envPrefix string, extraWarn, failOnWarning bool) error {
	return runWithMode(rootDir, output, envPrefix, extraWarn, failOnWarning, "")
}

// runWithMode executes cdoc with an explicit parse mode.
func runWithMode(rootDir, output, envPrefix string, extraWarn, failOnWarning bool, parseMode string) error {
	policy := model.DefaultWarningPolicy()
	if failOnWarning {
		policy = model.StrictWarningPolicy()
	}
	return engine.Run(engine.RunOptions{
		RootDir:       rootDir,
		Output:        output,
		EnvPrefix:     envPrefix,
		ExtraWarnings: extraWarn,
		ParseMode:     engine.ParseMode(parseMode),
		Policy:        &policy,
	})
}
