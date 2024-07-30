package scylla

type Option func(*config)

func WithTag(tag string) Option {
	return func(c *config) {
		c.tag = tag
	}
}

type config struct {
	tag string
}
