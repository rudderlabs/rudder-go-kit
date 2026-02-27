package cdoc

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/cmd/cdoc/internal/cdoc/model"
)

func TestParseProject_ParseWarningCode(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "bad.go"), []byte("package bad\nfunc broken("), 0o644))

	entries, warnings, err := ParseProject(root)
	require.NoError(t, err)
	require.Empty(t, entries)
	require.Len(t, warnings, 1)
	require.Equal(t, model.WarningCodeParseFailed, warnings[0].Code)
	require.Contains(t, warnings[0].Message, "failed to parse")
}

func TestGenerateWarnings_Codes(t *testing.T) {
	entries := []model.Entry{{PrimaryKey: "missing.all", File: "x.go", Line: 9}}

	warnings := GenerateWarnings(entries)
	require.Len(t, warnings, 2)
	require.Equal(t, model.WarningCodeMissingDescription, warnings[0].Code)
	require.Equal(t, model.WarningCodeMissingGroup, warnings[1].Code)
}

func TestRun_WarningPolicyControlsBehavior(t *testing.T) {
	root := t.TempDir()
	src := `package test
import "github.com/rudderlabs/rudder-go-kit/config"
func f(conf *config.Config) {
	conf.GetStringVar("value", "some.key")
}`
	require.NoError(t, os.WriteFile(filepath.Join(root, "example.go"), []byte(src), 0o644))

	t.Run("warn policy does not fail", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := Run(RunOptions{
			RootDir:       root,
			EnvPrefix:     "PREFIX",
			ExtraWarnings: true,
			Policy:        new(model.DefaultWarningPolicy()),
			Stdout:        &stdout,
			Stderr:        &stderr,
		})
		require.NoError(t, err)
		require.Contains(t, stdout.String(), "`some.key`")
		require.Contains(t, stderr.String(), "has no //cdoc:desc")
		require.Contains(t, stderr.String(), "has no //cdoc:group")
	})

	t.Run("strict policy fails", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := Run(RunOptions{
			RootDir:       root,
			EnvPrefix:     "PREFIX",
			ExtraWarnings: true,
			Policy:        new(model.StrictWarningPolicy()),
			Stdout:        &stdout,
			Stderr:        &stderr,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "found 2 warning(s)")
		require.Contains(t, stdout.String(), "`some.key`")
		require.NotEmpty(t, stderr.String())
	})

	t.Run("ignore policy with explicit overrides map suppresses warnings", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := Run(RunOptions{
			RootDir:       root,
			EnvPrefix:     "PREFIX",
			ExtraWarnings: true,
			Policy: new(model.WarningPolicy{
				DefaultSeverity: model.SeverityIgnore,
				Overrides:       map[model.WarningCode]model.WarningSeverity{},
			}),
			Stdout: &stdout,
			Stderr: &stderr,
		})
		require.NoError(t, err)
		require.Contains(t, stdout.String(), "`some.key`")
		require.Empty(t, stderr.String())
	})

	t.Run("nil policy falls back to default warn behavior", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := Run(RunOptions{
			RootDir:       root,
			EnvPrefix:     "PREFIX",
			ExtraWarnings: true,
			Stdout:        &stdout,
			Stderr:        &stderr,
		})
		require.NoError(t, err)
		require.Contains(t, stdout.String(), "`some.key`")
		require.Contains(t, stderr.String(), "has no //cdoc:desc")
	})

	t.Run("zero-value policy can explicitly ignore warnings", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		err := Run(RunOptions{
			RootDir:       root,
			EnvPrefix:     "PREFIX",
			ExtraWarnings: true,
			Policy:        new(model.WarningPolicy{}),
			Stdout:        &stdout,
			Stderr:        &stderr,
		})
		require.NoError(t, err)
		require.Contains(t, stdout.String(), "`some.key`")
		require.Empty(t, stderr.String())
	})
}
