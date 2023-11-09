package logger_test

import (
	"bufio"
	"bytes"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/zenizh/go-capturer"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

func Test_Print_All_Levels(t *testing.T) {
	fileName := t.TempDir() + "out.log"
	f, err := os.Create(fileName)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	c := config.New()
	c.Set("LOG_LEVEL", "EVENT")
	c.Set("Logger.enableConsole", false)
	c.Set("Logger.enableFile", true)
	c.Set("Logger.enableFileNameInLog", false)
	c.Set("Logger.logFileLocation", fileName)
	loggerFactory := logger.NewFactory(c, constantClockOpt)

	rootLogger := loggerFactory.NewLogger()
	require.True(t, rootLogger.IsDebugLevel())

	scanner := bufio.NewScanner(f)

	rootLogger.Debug("hello ", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	DEBUG	hello world", scanner.Text())

	rootLogger.Info("hello ", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	INFO	hello world", scanner.Text())

	rootLogger.Warn("hello ", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	WARN	hello world", scanner.Text())

	rootLogger.Error("hello ", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	ERROR	hello world", scanner.Text())

	rootLogger.Fatal("hello ", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	ERROR	hello world", scanner.Text())
}

func Test_Printf_All_Levels(t *testing.T) {
	fileName := t.TempDir() + "out.log"
	f, err := os.Create(fileName)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	c := config.New()
	c.Set("LOG_LEVEL", "EVENT")
	c.Set("Logger.enableConsole", false)
	c.Set("Logger.enableFile", true)
	c.Set("Logger.enableFileNameInLog", false)
	c.Set("Logger.logFileLocation", fileName)
	loggerFactory := logger.NewFactory(c, constantClockOpt)

	rootLogger := loggerFactory.NewLogger()
	require.True(t, rootLogger.IsDebugLevel())

	scanner := bufio.NewScanner(f)

	rootLogger.Debugf("hello %s", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	DEBUG	hello world", scanner.Text())

	rootLogger.Infof("hello %s", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	INFO	hello world", scanner.Text())

	rootLogger.Warnf("hello %s", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	WARN	hello world", scanner.Text())

	rootLogger.Errorf("hello %s", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	ERROR	hello world", scanner.Text())

	rootLogger.Fatalf("hello %s", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	ERROR	hello world", scanner.Text())
}

func Test_Printw_All_Levels(t *testing.T) {
	fileName := t.TempDir() + "out.log"
	f, err := os.Create(fileName)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	c := config.New()
	c.Set("LOG_LEVEL", "EVENT")
	c.Set("Logger.enableConsole", false)
	c.Set("Logger.enableFile", true)
	c.Set("Logger.enableFileNameInLog", false)
	c.Set("Logger.enableLoggerNameInLog", false)
	c.Set("Logger.logFileLocation", fileName)
	loggerFactory := logger.NewFactory(c, constantClockOpt)

	rootLogger := loggerFactory.NewLogger()
	require.True(t, rootLogger.IsDebugLevel())

	scanner := bufio.NewScanner(f)

	rootLogger.Debugw("hello world", "key", "value")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	DEBUG	hello world	{"key": "value"}`, scanner.Text())

	rootLogger.Infow("hello world", "key", "value")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	INFO	hello world	{"key": "value"}`, scanner.Text())

	rootLogger.Warnw("hello world", "key", "value")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	WARN	hello world	{"key": "value"}`, scanner.Text())

	rootLogger.Errorw("hello world", "key", "value")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	ERROR	hello world	{"key": "value"}`, scanner.Text())

	rootLogger.Fatalw("hello world", "key", "value")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	ERROR	hello world	{"key": "value"}`, scanner.Text())
}

func Test_Logger_With_Context(t *testing.T) {
	fileName := t.TempDir() + "out.log"
	f, err := os.Create(fileName)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	c := config.New()
	c.Set("LOG_LEVEL", "INFO")
	c.Set("Logger.enableConsole", false)
	c.Set("Logger.enableFile", true)
	c.Set("Logger.enableFileNameInLog", false)
	c.Set("Logger.enableLoggerNameInLog", false)
	c.Set("Logger.logFileLocation", fileName)
	loggerFactory := logger.NewFactory(c, constantClockOpt)
	rootLogger := loggerFactory.NewLogger()
	ctxLogger := rootLogger.With("key", "value")

	scanner := bufio.NewScanner(f)

	rootLogger.Info("hello world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	INFO	hello world`, scanner.Text())
	ctxLogger.Info("hello world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	INFO	hello world	{"key": "value"}`, scanner.Text())

	rootLogger.Infof("hello %s", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	INFO	hello world`, scanner.Text())
	ctxLogger.Infof("hello %s", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	INFO	hello world	{"key": "value"}`, scanner.Text())

	rootLogger.Infow("hello world", "key1", "value1")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	INFO	hello world	{"key1": "value1"}`, scanner.Text())
	ctxLogger.Infow("hello world", "key1", "value1")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `2077-01-23T10:15:13.000Z	INFO	hello world	{"key": "value", "key1": "value1"}`, scanner.Text())
}

func Test_Logger_Deep_Hierarchy(t *testing.T) {
	fileName := t.TempDir() + "out.log"
	f, err := os.Create(fileName)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	c := config.New()
	c.Set("LOG_LEVEL", "INFO")
	c.Set("Logger.enableConsole", false)
	c.Set("Logger.enableFile", true)
	c.Set("Logger.enableFileNameInLog", false)
	c.Set("Logger.logFileLocation", fileName)
	loggerFactory := logger.NewFactory(c, constantClockOpt)
	rootLogger := loggerFactory.NewLogger()
	lvl1Logger := rootLogger.Child("logger1")
	lvl2Logger := lvl1Logger.Child("logger2")
	lvl3Logger := lvl2Logger.Child("logger3")

	rootLogger.Info("hello world 0")
	scanner := bufio.NewScanner(f)
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	INFO	hello world 0", scanner.Text())

	lvl1Logger.Info("hello world 1")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	INFO	logger1	hello world 1", scanner.Text())

	lvl2Logger.Info("hello world 2")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	INFO	logger1.logger2	hello world 2", scanner.Text())

	lvl3Logger.Info("hello world 3")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, "2077-01-23T10:15:13.000Z	INFO	logger1.logger2.logger3	hello world 3", scanner.Text())
}

func Test_Logger_Json_Output(t *testing.T) {
	fileName := t.TempDir() + "out.log"
	f, err := os.Create(fileName)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	c := config.New()
	c.Set("LOG_LEVEL", "INFO")
	c.Set("Logger.enableConsole", false)
	c.Set("Logger.enableFile", true)
	c.Set("Logger.enableFileNameInLog", false)
	c.Set("Logger.logFileLocation", fileName)
	c.Set("Logger.fileJsonFormat", true)
	loggerFactory := logger.NewFactory(c, constantClockOpt)
	rootLogger := loggerFactory.NewLogger().Child("mylogger")
	ctxLogger := rootLogger.With("key", "value")

	scanner := bufio.NewScanner(f)

	rootLogger.Info("hello world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `{"level":"INFO","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world"}`, scanner.Text())
	ctxLogger.Info("hello world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `{"level":"INFO","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key":"value"}`, scanner.Text())

	rootLogger.Infof("hello %s", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `{"level":"INFO","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world"}`, scanner.Text())
	ctxLogger.Infof("hello %s", "world")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `{"level":"INFO","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key":"value"}`, scanner.Text())

	rootLogger.Infow("hello world", "key1", "value1")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `{"level":"INFO","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key1":"value1"}`, scanner.Text())
	ctxLogger.Infow("hello world", "key1", "value1")
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.Equal(t, `{"level":"INFO","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key":"value","key1":"value1"}`, scanner.Text())
}

func Test_Logger_NonSugared(t *testing.T) {
	fileName := t.TempDir() + "out.log"
	f, err := os.Create(fileName)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	c := config.New()
	c.Set("LOG_LEVEL", "EVENT")
	c.Set("Logger.enableConsole", false)
	c.Set("Logger.enableFile", true)
	c.Set("Logger.enableFileNameInLog", false)
	c.Set("Logger.enableStackTrace", false)
	c.Set("Logger.logFileLocation", fileName)
	c.Set("Logger.fileJsonFormat", true)
	loggerFactory := logger.NewFactory(c, constantClockOpt)
	rootLogger := loggerFactory.NewLogger().Child("mylogger")
	ctxLogger := rootLogger.Withn(logger.NewBoolField("foo", true))

	scanner := bufio.NewScanner(f)

	rootLogger.Debugn("hello world", logger.NewStringField("key1", "value1"))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"DEBUG","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key1":"value1"}`, scanner.Text())
	ctxLogger.Debugn("hello world", logger.NewIntField("key2", 2))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"DEBUG","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key2":2,"foo":true}`, scanner.Text())

	rootLogger.Infon("hello world", logger.NewStringField("key1", "value1"))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"INFO","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key1":"value1"}`, scanner.Text())
	ctxLogger.Infon("hello world", logger.NewIntField("key2", 2))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"INFO","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key2":2,"foo":true}`, scanner.Text())

	rootLogger.Warnn("hello world", logger.NewStringField("key1", "value1"))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"WARN","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key1":"value1"}`, scanner.Text())
	ctxLogger.Warnn("hello world", logger.NewIntField("key2", 2))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"WARN","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key2":2,"foo":true}`, scanner.Text())

	rootLogger.Errorn("hello world", logger.NewStringField("key1", "value1"))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"ERROR","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key1":"value1"}`, scanner.Text())
	ctxLogger.Errorn("hello world", logger.NewIntField("key2", 2))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"ERROR","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key2":2,"foo":true}`, scanner.Text())

	rootLogger.Fataln("hello world", logger.NewStringField("key1", "value1"))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"ERROR","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key1":"value1"}`, scanner.Text())
	require.True(t, scanner.Scan(), "it should print a stacktrace statement")
	require.Contains(t, scanner.Text(), `"level":"ERROR"`)
	require.Contains(t, scanner.Text(), `"ts":"2077-01-23T10:15:13.000Z"`)
	require.Contains(t, scanner.Text(), `"logger":"mylogger"`)
	require.Contains(t, scanner.Text(), `"key1":"value1"`)
	require.Contains(t, scanner.Text(), `"msg":"goroutine`)

	ctxLogger.Fataln("hello world", logger.NewIntField("key2", 2))
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{"level":"ERROR","ts":"2077-01-23T10:15:13.000Z","logger":"mylogger","msg":"hello world","key2":2,"foo":true}`, scanner.Text())
	require.True(t, scanner.Scan(), "it should print a stacktrace statement")
	require.Contains(t, scanner.Text(), `"level":"ERROR"`)
	require.Contains(t, scanner.Text(), `"ts":"2077-01-23T10:15:13.000Z"`)
	require.Contains(t, scanner.Text(), `"logger":"mylogger"`)
	require.Contains(t, scanner.Text(), `"key2":2`)
	require.Contains(t, scanner.Text(), `"foo":true`)
	require.Contains(t, scanner.Text(), `"msg":"goroutine`)

	rootLogger.Debugn("using all fields",
		logger.NewField("foo", "any value"),
		logger.NewStringField("myString", "hello"),
		logger.NewBoolField("myBool", true),
		logger.NewFloatField("myFloat", 1.1),
		logger.NewTimeField("myTime", time.Date(2077, 1, 23, 10, 15, 13, 0, time.UTC)),
		logger.NewDurationField("myDuration", time.Second*2+time.Millisecond*321),
		logger.NewErrorField(errors.New("a bad error")),
	)
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{
		"level":"DEBUG",
		"ts":"2077-01-23T10:15:13.000Z",
		"logger":"mylogger",
		"msg":"using all fields",
		"foo":"any value",
		"myString":"hello",
		"myBool":true,
		"myFloat":1.1,
		"myTime":"2077-01-23T10:15:13.000Z",
		"myDuration":2.321,
		"error":"a bad error"
	}`, scanner.Text())
}

func Test_Logger_Expand(t *testing.T) {
	fileName := t.TempDir() + "out.log"
	f, err := os.Create(fileName)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()

	c := config.New()
	c.Set("LOG_LEVEL", "EVENT")
	c.Set("Logger.enableConsole", false)
	c.Set("Logger.enableFile", true)
	c.Set("Logger.enableFileNameInLog", false)
	c.Set("Logger.enableStackTrace", false)
	c.Set("Logger.logFileLocation", fileName)
	c.Set("Logger.fileJsonFormat", true)
	loggerFactory := logger.NewFactory(c, constantClockOpt)
	rootLogger := loggerFactory.NewLogger().Child("mylogger")

	scanner := bufio.NewScanner(f)

	rootLogger.Debugw("using expand",
		logger.Expand(
			logger.NewField("foo", "any value"),
			logger.NewStringField("myString", "hello"),
			logger.NewBoolField("myBool", true),
			logger.NewFloatField("myFloat", 1.1),
			logger.NewTimeField("myTime", time.Date(2077, 1, 23, 10, 15, 13, 0, time.UTC)),
			logger.NewDurationField("myDuration", time.Second*2+time.Millisecond*321),
			logger.NewErrorField(errors.New("a bad error")),
		)...,
	)
	require.True(t, scanner.Scan(), "it should print a log statement")
	require.JSONEq(t, `{
		"level":"DEBUG",
		"ts":"2077-01-23T10:15:13.000Z",
		"logger":"mylogger",
		"msg":"using expand",
		"foo":"any value",
		"myString":"hello",
		"myBool":true,
		"myFloat":1.1,
		"myTime":"2077-01-23T10:15:13.000Z",
		"myDuration":2.321,
		"error":"a bad error"
	}`, scanner.Text())
}

func Test_LogRequest(t *testing.T) {
	json := `{"key":"value"}`
	request, err := http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte(json)))
	require.NoError(t, err)
	c := config.New()
	c.Set("LOG_LEVEL", "EVENT")
	c.Set("Logger.enableTimestamp", false)
	c.Set("Logger.enableFileNameInLog", false)
	c.Set("Logger.enableLoggerNameInLog", false)
	stdout := capturer.CaptureStdout(func() {
		loggerFactory := logger.NewFactory(c, constantClockOpt)
		l := loggerFactory.NewLogger()
		l.LogRequest(request)
		loggerFactory.Sync()
	})
	require.Equal(t, "DEBUG\tRequest Body\t{\"body\": \"{\\\"key\\\":\\\"value\\\"}\"}\n", stdout)
}

type constantClock time.Time

func (c constantClock) Now() time.Time { return time.Time(c) }
func (constantClock) NewTicker(_ time.Duration) *time.Ticker {
	return &time.Ticker{}
}

var (
	date             = time.Date(2077, 1, 23, 10, 15, 13, 0o00, time.UTC)
	constantClockOpt = logger.WithClock(constantClock(date))
)
