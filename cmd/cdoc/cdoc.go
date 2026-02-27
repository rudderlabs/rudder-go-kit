package main

import (
	cdoc "github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/cdoc"
	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/cdoc/model"
)

func run(rootDir, output, envPrefix string, extraWarn, failOnWarning bool) error {
	policy := model.DefaultWarningPolicy()
	if failOnWarning {
		policy = model.StrictWarningPolicy()
	}
	return cdoc.Run(cdoc.RunOptions{
		RootDir:       rootDir,
		Output:        output,
		EnvPrefix:     envPrefix,
		ExtraWarnings: extraWarn,
		Policy:        &policy,
	})
}
