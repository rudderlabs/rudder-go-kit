package resource

type Logger interface {
	Log(...interface{})
	Logf(string, ...interface{})
}

type FailIndicator interface {
	Failed() bool
}

type Cleaner interface {
	Cleanup(func())
	Logger
	FailIndicator
}

type NOPLogger struct{}

// Log for the NOP Logger does nothing.
func (*NOPLogger) Log(...interface{}) {}
