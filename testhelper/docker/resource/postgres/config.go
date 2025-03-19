package postgres

import "github.com/ory/dockertest/v3/docker"

type Opt func(*Config)

func WithTag(tag string) Opt {
	return func(c *Config) {
		c.Tag = tag
	}
}

func WithOptions(options ...string) Opt {
	return func(c *Config) {
		c.Options = options
	}
}

func WithShmSize(shmSize int64) Opt {
	return func(c *Config) {
		c.ShmSize = shmSize
	}
}

func WithMemory(memory int64) Opt {
	return func(c *Config) {
		c.Memory = memory
	}
}

func WithOOMKillDisable(disable bool) Opt {
	return func(c *Config) {
		c.OOMKillDisable = disable
	}
}

func WithPrintLogsOnError(printLogsOnError bool) Opt {
	return func(c *Config) {
		c.PrintLogsOnError = printLogsOnError
	}
}

func WithNetwork(network *docker.Network) Opt {
	return func(c *Config) {
		c.NetworkID = network.ID
	}
}

func WithBindIP(bindIP string) Opt {
	return func(c *Config) {
		c.BindIP = bindIP
	}
}

type Config struct {
	Tag              string
	Options          []string
	ShmSize          int64
	Memory           int64
	OOMKillDisable   bool
	PrintLogsOnError bool
	NetworkID        string
	BindIP           string
}
