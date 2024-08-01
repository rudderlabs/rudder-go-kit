package scylla

type Option func(*config)

func WithTag(tag string) Option {
	return func(c *config) {
		c.tag = tag
	}
}

func WithKeyspace(keyspace string) Option {
	return func(c *config) {
		c.keyspace = keyspace
	}
}

type config struct {
	tag      string
	keyspace string
}
