package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/engine"
)

func TestRun(t *testing.T) {
	for _, mode := range []engine.ParseMode{engine.ParseModeAST, engine.ParseModeTypes} {
		t.Run(string(mode), func(t *testing.T) {
			t.Run("succeeds", func(t *testing.T) {
				output := filepath.Join(t.TempDir(), "output.md")
				err := run(runOptions{
					rootDir:       "testdata",
					output:        output,
					envPrefix:     "PREFIX",
					extraWarn:     true,
					failOnWarning: false,
					parseMode:     mode,
				})
				require.NoError(t, err)

				got, err := os.ReadFile(output)
				require.NoError(t, err)

				expected, err := os.ReadFile("testdata/expected_output.md")
				require.NoError(t, err, "reading golden file")
				require.Equal(t, string(expected), string(got), "run output does not match golden file")
			})

			t.Run("fail_on_warning", func(t *testing.T) {
				output := filepath.Join(t.TempDir(), "output.md")
				err := run(runOptions{
					rootDir:       "testdata",
					output:        output,
					envPrefix:     "PREFIX",
					extraWarn:     true,
					failOnWarning: true,
					parseMode:     mode,
				})
				require.Error(t, err)
				require.Contains(t, err.Error(), "found")
				require.Contains(t, err.Error(), "warning")

				got, readErr := os.ReadFile(output)
				require.NoError(t, readErr)
				expected, readExpectedErr := os.ReadFile("testdata/expected_output.md")
				require.NoError(t, readExpectedErr)
				require.Equal(t, string(expected), string(got), "run output should still be generated")
			})

			t.Run("fail_on_warning_with_clean_input", func(t *testing.T) {
				rootDir := t.TempDir()
				src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	//cdoc:group General
	//cdoc:desc d
	conf.GetStringVar("value", "some.key")
}`
				err := os.WriteFile(filepath.Join(rootDir, "example.go"), []byte(src), 0o644)
				require.NoError(t, err)

				output := filepath.Join(t.TempDir(), "output.md")
				err = run(runOptions{
					rootDir:       rootDir,
					output:        output,
					envPrefix:     "PREFIX",
					extraWarn:     true,
					failOnWarning: true,
					parseMode:     mode,
				})
				require.NoError(t, err)

				got, readErr := os.ReadFile(output)
				require.NoError(t, readErr)
				require.Contains(t, string(got), "`some.key`")
			})
		})
	}
}
