package resource

type Logger interface {
	Log(...interface{})
}

type Cleaner interface {
	Cleanup(func())
	Logger
}

type NOPLogger struct{}

// Log for the NOP Logger does nothing.
func (*NOPLogger) Log(...interface{}) {}
