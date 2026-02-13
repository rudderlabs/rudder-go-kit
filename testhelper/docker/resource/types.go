package resource

type Logger interface {
	Log(...any)
	Logf(string, ...any)
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
func (*NOPLogger) Log(...any) {}
