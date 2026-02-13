package client

type logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

type KafkaLogger struct {
	Logger        logger
	IsErrorLogger bool
}

func (l *KafkaLogger) Printf(format string, args ...any) {
	if l.IsErrorLogger {
		l.Logger.Errorf(format, args...)
	} else {
		l.Logger.Infof(format, args...)
	}
}
