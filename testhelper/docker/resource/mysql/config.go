package mysql

type Opt func(*Config)

func WithTag(tag string) Opt {
	return func(c *Config) {
		c.Tag = tag
	}
}

func WithShmSize(shmSize int64) Opt {
	return func(c *Config) {
		c.ShmSize = shmSize
	}
}

type Config struct {
	Tag     string
	ShmSize int64
}
