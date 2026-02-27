package main

import (
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine/model"
)

func run(rootDir, output, envPrefix string, extraWarn, failOnWarning bool) error {
	policy := model.DefaultWarningPolicy()
	if failOnWarning {
		policy = model.StrictWarningPolicy()
	}
	return engine.Run(engine.RunOptions{
		RootDir:       rootDir,
		Output:        output,
		EnvPrefix:     envPrefix,
		ExtraWarnings: extraWarn,
		Policy:        &policy,
	})
}
