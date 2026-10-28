package pulsar

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestContainerEnv(t *testing.T) {
	newConfig := func(opts ...Option) *config {
		c := &config{}
		for _, opt := range opts {
			opt(c)
		}
		return c
	}

	t.Run("test tuning by default", func(t *testing.T) {
		require.Equal(t, testTuningEnv, newConfig().containerEnv())
	})

	t.Run("without test tuning", func(t *testing.T) {
		require.Empty(t, newConfig(WithoutTestTuning()).containerEnv())
		require.Equal(t, []string{"FOO=bar"}, newConfig(WithoutTestTuning(), WithEnv("FOO=bar")).containerEnv())
	})

	t.Run("caller variables replace tuned ones and the rest are appended", func(t *testing.T) {
		env := newConfig(WithEnv("PULSAR_MEM=-Xms1g -Xmx1g", "FOO=bar")).containerEnv()
		require.Equal(t, []string{
			"PULSAR_MEM=-Xms1g -Xmx1g",
			testTuningEnv[1],
			"FOO=bar",
		}, env)
		require.Equal(t, "PULSAR_MEM=-Xms256m -Xmx512m -XX:MaxDirectMemorySize=512m", testTuningEnv[0],
			"the tuning defaults must not be modified")
	})

	t.Run("a later caller variable replaces an earlier one", func(t *testing.T) {
		env := newConfig(WithEnv("FOO=bar"), WithEnv("FOO=baz")).containerEnv()
		require.Equal(t, append(append([]string{}, testTuningEnv...), "FOO=baz"), env)
	})
}
