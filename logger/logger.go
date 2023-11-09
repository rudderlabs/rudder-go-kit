// Package logger
/*
Logger Interface Use instance of logger instead of exported functions

usage example

   import (
	   "errors"
	   "github.com/rudderlabs/rudder-go-kit/config"
	   "github.com/rudderlabs/rudder-go-kit/logger"
   )

   var c = config.New()
   var loggerFactory = logger.NewFactory(c)
   var log logger.Logger = loggerFactory.NewLogger()
   ...
   log.Error(...)

or if you want to use the default logger factory (not advised):

   var log logger.Logger = logger.NewLogger()
   ...
   log.Error(...)

*/
//go:generate mockgen -destination=mock_logger/mock_logger.go -package mock_logger github.com/rudderlabs/rudder-go-kit/logger Logger
package logger

import (
	"bytes"
	"io"
	"net/http"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

/*
Using levels(like Debug, Info etc.) in logging is a way to categorize logs based on their importance.
The idea is to have the option of running the application in different logging levels based on
how verbose we want the logging to be.
For example, using Debug level of logging, logs everything, and it might slow the application, so we run application
in DEBUG level for local development or when we want to look through the entire flow of events in detail.
We use 4 logging levels here Debug, Info, Warn and Error.
*/

type Logger interface {
	// IsDebugLevel Returns true is debug lvl is enabled
	IsDebugLevel() bool

	// Debug level logging. Most verbose logging level.
	Debug(args ...any)

	// Debugf does debug level logging similar to fmt.Printf. Most verbose logging level
	Debugf(format string, args ...any)

	// Debugw does debug level structured logging. Most verbose logging level
	Debugw(msg string, keysAndValues ...any)

	// Debugn does debug level non-sugared structured logging. Most verbose logging level
	Debugn(msg string, fields ...Field)

	// Info level logging. Use this to log the state of the application.
	// Don't use Logger.Info in the flow of individual events. Use Logger.Debug instead.
	Info(args ...any)

	// Infof does info level logging similar to fmt.Printf. Use this to log the state of the application.
	// Don't use Logger.Info in the flow of individual events. Use Logger.Debug instead.
	Infof(format string, args ...any)

	// Infow does info level structured logging. Use this to log the state of the application.
	// Don't use Logger.Info in the flow of individual events. Use Logger.Debug instead.
	Infow(msg string, keysAndValues ...any)

	// Infon does info level non-sugared structured logging. Use this to log the state of the application.
	// Don't use Logger.Info in the flow of individual events. Use Logger.Debug instead.
	Infon(msg string, fields ...Field)

	// Warn level logging. Use this to log warnings
	Warn(args ...any)

	// Warnf does warn level logging similar to fmt.Printf. Use this to log warnings
	Warnf(format string, args ...any)

	// Warnw does warn level structured logging. Use this to log warnings
	Warnw(msg string, keysAndValues ...any)

	// Warnn does warn level non-sugared structured logging. Use this to log warnings
	Warnn(msg string, fields ...Field)

	// Error level logging. Use this to log errors which don't immediately halt the application.
	Error(args ...any)

	// Errorf does error level logging similar to fmt.Printf. Use this to log errors which don't immediately halt the application.
	Errorf(format string, args ...any)

	// Errorw does error level structured logging.
	// Use this to log errors which don't immediately halt the application.
	Errorw(msg string, keysAndValues ...any)

	// Errorn does error level non-sugared structured logging.
	// Use this to log errors which don't immediately halt the application.
	Errorn(msg string, fields ...Field)

	// Fatal level logging. Use this to log errors which crash the application.
	Fatal(args ...any)

	// Fatalf does fatal level logging similar to fmt.Printf. Use this to log errors which crash the application.
	Fatalf(format string, args ...any)

	// Fatalw does fatal level structured logging.
	// Use this to log errors which crash the application.
	Fatalw(format string, keysAndValues ...any)

	// Fataln does fatal level non-sugared structured logging.
	// Use this to log errors which crash the application.
	Fataln(format string, fields ...Field)

	LogRequest(req *http.Request)

	// Child creates a child logger with the given name
	Child(s string) Logger

	// With adds the provided key value pairs to the logger context
	With(args ...any) Logger

	// Withn adds the provided key value pairs to the logger context
	Withn(args ...Field) Logger
}

type logger struct {
	logConfig  *factoryConfig
	name       string
	zap        *zap.Logger
	sugaredZap *zap.SugaredLogger
	parent     *logger
}

func (l *logger) Child(s string) Logger {
	if s == "" {
		return l
	}
	cp := *l
	cp.parent = l
	if l.name == "" {
		cp.name = s
	} else {
		cp.name = strings.Join([]string{l.name, s}, ".")
	}
	if l.logConfig.enableNameInLog {
		cp.sugaredZap = l.sugaredZap.Named(s)
		cp.zap = l.zap.Named(s)
	}
	return &cp
}

// With adds a variadic number of fields to the logging context.
// It accepts a mix of strongly-typed Field objects and loosely-typed key-value pairs.
// When processing pairs, the first element of the pair is used as the field key and the second as the field value.
func (l *logger) With(args ...any) Logger {
	cp := *l
	cp.sugaredZap = cp.sugaredZap.With(args...)
	for i := 0; i < len(args)-1; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			cp.Warnw("Logger.With called with non-string key",
				"parentName", l.parent.name, "name", l.name,
			)
			break
		}
		cp.zap = cp.zap.With(zap.Any(key, args[i+1]))
	}
	return &cp
}

// Withn adds a variadic number of fields to the logging context.
// It accepts a mix of strongly-typed Field objects and loosely-typed key-value pairs.
// When processing pairs, the first element of the pair is used as the field key and the second as the field value.
func (l *logger) Withn(args ...Field) Logger {
	cp := *l
	cp.zap = l.zap.With(toZap(args)...)
	return &cp
}

func (l *logger) getLoggingLevel() int {
	return l.logConfig.getOrSetLogLevel(l.name, l.parent.getLoggingLevel)
}

// IsDebugLevel Returns true is debug lvl is enabled
func (l *logger) IsDebugLevel() bool {
	return levelDebug >= l.getLoggingLevel()
}

// Debug level logging.
// Most verbose logging level.
func (l *logger) Debug(args ...any) {
	if levelDebug >= l.getLoggingLevel() {
		l.sugaredZap.Debug(args...)
	}
}

// Info level logging.
// Use this to log the state of the application. Don't use Logger.Info in the flow of individual events. Use Logger.Debug instead.
func (l *logger) Info(args ...any) {
	if levelInfo >= l.getLoggingLevel() {
		l.sugaredZap.Info(args...)
	}
}

// Warn level logging.
// Use this to log warnings
func (l *logger) Warn(args ...any) {
	if levelWarn >= l.getLoggingLevel() {
		l.sugaredZap.Warn(args...)
	}
}

// Error level logging.
// Use this to log errors which don't immediately halt the application.
func (l *logger) Error(args ...any) {
	if levelError >= l.getLoggingLevel() {
		l.sugaredZap.Error(args...)
	}
}

// Fatal level logging.
// Use this to log errors which crash the application.
func (l *logger) Fatal(args ...any) {
	l.sugaredZap.Error(args...)

	// If enableStackTrace is true, Zaplogger will take care of writing stacktrace to the file.
	// Else, we are force writing the stacktrace to the file.
	if !l.logConfig.enableStackTrace.Load() {
		byteArr := make([]byte, 2048)
		n := runtime.Stack(byteArr, false)
		stackTrace := string(byteArr[:n])
		l.sugaredZap.Error(stackTrace)
	}
	_ = l.sugaredZap.Sync()
}

// Debugf does debug level logging similar to fmt.Printf.
// Most verbose logging level
func (l *logger) Debugf(format string, args ...any) {
	if levelDebug >= l.getLoggingLevel() {
		l.sugaredZap.Debugf(format, args...)
	}
}

// Infof does info level logging similar to fmt.Printf.
// Use this to log the state of the application. Don't use Logger.Info in the flow of individual events. Use Logger.Debug instead.
func (l *logger) Infof(format string, args ...any) {
	if levelInfo >= l.getLoggingLevel() {
		l.sugaredZap.Infof(format, args...)
	}
}

// Warnf does warn level logging similar to fmt.Printf.
// Use this to log warnings
func (l *logger) Warnf(format string, args ...any) {
	if levelWarn >= l.getLoggingLevel() {
		l.sugaredZap.Warnf(format, args...)
	}
}

// Errorf does error level logging similar to fmt.Printf.
// Use this to log errors which don't immediately halt the application.
func (l *logger) Errorf(format string, args ...any) {
	if levelError >= l.getLoggingLevel() {
		l.sugaredZap.Errorf(format, args...)
	}
}

// Fatalf does fatal level logging similar to fmt.Printf.
// Use this to log errors which crash the application.
func (l *logger) Fatalf(format string, args ...any) {
	l.sugaredZap.Errorf(format, args...)

	// If enableStackTrace is true, Zaplogger will take care of writing stacktrace to the file.
	// Else, we are force writing the stacktrace to the file.
	if !l.logConfig.enableStackTrace.Load() {
		byteArr := make([]byte, 2048)
		n := runtime.Stack(byteArr, false)
		stackTrace := string(byteArr[:n])
		l.sugaredZap.Error(stackTrace)
	}
	_ = l.sugaredZap.Sync()
}

// Debugw does debug level structured logging.
// Most verbose logging level
func (l *logger) Debugw(msg string, keysAndValues ...any) {
	if levelDebug >= l.getLoggingLevel() {
		l.sugaredZap.Debugw(msg, keysAndValues...)
	}
}

// Infow does info level structured logging.
// Use this to log the state of the application. Don't use Logger.Info in the flow of individual events. Use Logger.Debug instead.
func (l *logger) Infow(msg string, keysAndValues ...any) {
	if levelInfo >= l.getLoggingLevel() {
		l.sugaredZap.Infow(msg, keysAndValues...)
	}
}

// Warnw does warn level structured logging.
// Use this to log warnings
func (l *logger) Warnw(msg string, keysAndValues ...any) {
	if levelWarn >= l.getLoggingLevel() {
		l.sugaredZap.Warnw(msg, keysAndValues...)
	}
}

// Errorw does error level structured logging.
// Use this to log errors which don't immediately halt the application.
func (l *logger) Errorw(msg string, keysAndValues ...any) {
	if levelError >= l.getLoggingLevel() {
		l.sugaredZap.Errorw(msg, keysAndValues...)
	}
}

// Fatalw does fatal level structured logging.
// Use this to log errors which crash the application.
func (l *logger) Fatalw(msg string, keysAndValues ...any) {
	l.sugaredZap.Errorw(msg, keysAndValues...)

	// If enableStackTrace is true, Zaplogger will take care of writing stacktrace to the file.
	// Else, we are force writing the stacktrace to the file.
	if !l.logConfig.enableStackTrace.Load() {
		byteArr := make([]byte, 2048)
		n := runtime.Stack(byteArr, false)
		stackTrace := string(byteArr[:n])
		l.sugaredZap.Error(stackTrace)
	}
	_ = l.sugaredZap.Sync()
}

// Debugn does debug level non-sugared structured logging.
func (l *logger) Debugn(msg string, fields ...Field) {
	if levelDebug >= l.getLoggingLevel() {
		l.zap.Debug(msg, toZap(fields)...)
	}
}

// Infon does info level non-sugared structured logging.
// Use this to log the state of the application.
// Don't use Logger.Info in the flow of individual events. Use Logger.Debug instead.
func (l *logger) Infon(msg string, fields ...Field) {
	if levelInfo >= l.getLoggingLevel() {
		l.zap.Info(msg, toZap(fields)...)
	}
}

// Warnn does warn level non-sugared structured logging.
// Use this to log warnings
func (l *logger) Warnn(msg string, fields ...Field) {
	if levelWarn >= l.getLoggingLevel() {
		l.zap.Warn(msg, toZap(fields)...)
	}
}

// Errorn does error level non-sugared structured logging.
// Use this to log errors which don't immediately halt the application.
func (l *logger) Errorn(msg string, fields ...Field) {
	if levelError >= l.getLoggingLevel() {
		l.zap.Error(msg, toZap(fields)...)
	}
}

// Fataln does fatal level non-sugared structured logging.
// Use this to log errors which crash the application.
func (l *logger) Fataln(msg string, fields ...Field) {
	zapFields := toZap(fields)
	l.zap.Error(msg, zapFields...)

	// If enableStackTrace is true, Zaplogger will take care of writing stacktrace to the file.
	// Else, we are force writing the stacktrace to the file.
	if !l.logConfig.enableStackTrace.Load() {
		byteArr := make([]byte, 2048)
		n := runtime.Stack(byteArr, false)
		stackTrace := string(byteArr[:n])
		l.zap.Error(stackTrace, zapFields...)
	}
	_ = l.zap.Sync()
}

// LogRequest reads and logs the request body and resets the body to original state.
func (l *logger) LogRequest(req *http.Request) {
	if levelEvent >= l.getLoggingLevel() {
		defer func() { _ = req.Body.Close() }()
		bodyBytes, _ := io.ReadAll(req.Body)
		bodyString := string(bodyBytes)
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		// print raw request body for debugging purposes
		l.zap.Debug("Request Body", zap.String("body", bodyString))
	}
}
