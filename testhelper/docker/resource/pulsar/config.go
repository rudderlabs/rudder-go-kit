package pulsar

import (
	"slices"
	"strings"

	"github.com/ory/dockertest/v3/docker"
)

// testTuningEnv sizes the Pulsar JVM for short-lived tests. Pulsar's own defaults (conf/pulsar_env.sh) commit and
// pre-touch a 2 GiB ZGC heap at startup, so every container costs about 2.3 GiB before the first message.
var testTuningEnv = []string{
	"PULSAR_MEM=-Xms256m -Xmx512m -XX:MaxDirectMemorySize=512m",
	"PULSAR_GC=-XX:+UseG1GC -XX:+PerfDisableSharedMem",
}

// testTuningFlags skip the functions worker and the stream storage, which tests do not use. On a busy host they
// delay the creation of the public/default namespace, which the setup waits for, by tens of seconds.
var testTuningFlags = []string{"--no-functions-worker", "-nss"}

type Option func(*config)

func WithTag(tag string) Option {
	return func(c *config) {
		c.tag = tag
	}
}

func WithNetwork(network *docker.Network) Option {
	return func(c *config) {
		c.network = network
	}
}

// WithEnv adds environment variables to the Pulsar container, as KEY=value. A variable that the test tuning also
// sets, such as PULSAR_MEM or PULSAR_GC, replaces the tuned value.
func WithEnv(env ...string) Option {
	return func(c *config) {
		c.env = append(c.env, env...)
	}
}

// WithoutTestTuning keeps Pulsar's stock JVM settings and starts the functions worker and the stream storage, for
// tests that depend on them.
func WithoutTestTuning() Option {
	return func(c *config) {
		c.disableTestTuning = true
	}
}

type config struct {
	tag               string
	network           *docker.Network
	env               []string
	disableTestTuning bool
}

// containerEnv returns the test tuning followed by the caller's variables. A caller's variable replaces an earlier
// one with the same name.
func (c *config) containerEnv() []string {
	var env []string
	if !c.disableTestTuning {
		env = append(env, testTuningEnv...)
	}
	for _, kv := range c.env {
		name, _, _ := strings.Cut(kv, "=")
		i := slices.IndexFunc(env, func(e string) bool {
			n, _, _ := strings.Cut(e, "=")
			return n == name
		})
		if i >= 0 {
			env[i] = kv
			continue
		}
		env = append(env, kv)
	}
	return env
}

func (c *config) containerCmd() []string {
	cmd := []string{"bin/pulsar", "standalone"}
	if !c.disableTestTuning {
		cmd = append(cmd, testTuningFlags...)
	}
	return cmd
}
