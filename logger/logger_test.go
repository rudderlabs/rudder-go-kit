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
		logger.NewIntField("myInt", 666),
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
		"myInt":666,
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
			logger.NewIntField("myInt", 666),
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
		"myInt":666,
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

func Test_NewIntSliceField(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		slice := []int{-1, 0, 1, 42}
		field := logger.NewIntSliceField("test_int", slice)
		require.Equal(t, "test_int", field.Name())
		require.Equal(t, "-1,0,1,42", field.Value())
	})

	t.Run("int8", func(t *testing.T) {
		slice := []int8{-128, -1, 0, 1, 127}
		field := logger.NewIntSliceField("test_int8", slice)
		require.Equal(t, "test_int8", field.Name())
		require.Equal(t, "-128,-1,0,1,127", field.Value())
	})

	t.Run("int16", func(t *testing.T) {
		slice := []int16{-32768, -1, 0, 1, 32767}
		field := logger.NewIntSliceField("test_int16", slice)
		require.Equal(t, "test_int16", field.Name())
		require.Equal(t, "-32768,-1,0,1,32767", field.Value())
	})

	t.Run("int32", func(t *testing.T) {
		slice := []int32{-2147483648, -1, 0, 1, 2147483647}
		field := logger.NewIntSliceField("test_int32", slice)
		require.Equal(t, "test_int32", field.Name())
		require.Equal(t, "-2147483648,-1,0,1,2147483647", field.Value())
	})

	t.Run("int64", func(t *testing.T) {
		slice := []int64{-9223372036854775808, -1, 0, 1, 9223372036854775807}
		field := logger.NewIntSliceField("test_int64", slice)
		require.Equal(t, "test_int64", field.Name())
		require.Equal(t, "-9223372036854775808,-1,0,1,9223372036854775807", field.Value())
	})

	t.Run("uint", func(t *testing.T) {
		slice := []uint{0, 1, 42, 4294967295}
		field := logger.NewIntSliceField("test_uint", slice)
		require.Equal(t, "test_uint", field.Name())
		require.Equal(t, "0,1,42,4294967295", field.Value())
	})

	t.Run("uint8", func(t *testing.T) {
		slice := []uint8{0, 1, 42, 255}
		field := logger.NewIntSliceField("test_uint8", slice)
		require.Equal(t, "test_uint8", field.Name())
		require.Equal(t, "0,1,42,255", field.Value())
	})

	t.Run("uint16", func(t *testing.T) {
		slice := []uint16{0, 1, 42, 65535}
		field := logger.NewIntSliceField("test_uint16", slice)
		require.Equal(t, "test_uint16", field.Name())
		require.Equal(t, "0,1,42,65535", field.Value())
	})

	t.Run("uint32", func(t *testing.T) {
		slice := []uint32{0, 1, 42, 4294967295}
		field := logger.NewIntSliceField("test_uint32", slice)
		require.Equal(t, "test_uint32", field.Name())
		require.Equal(t, "0,1,42,4294967295", field.Value())
	})

	t.Run("uint64", func(t *testing.T) {
		slice := []uint64{0, 1, 42, 18446744073709551615}
		field := logger.NewIntSliceField("test_uint64", slice)
		require.Equal(t, "test_uint64", field.Name())
		require.Equal(t, "0,1,42,18446744073709551615", field.Value())
	})

	t.Run("data_truncation", func(t *testing.T) {
		// Test with uint64 values that exceed int64 max to ensure no truncation
		largeUint64 := uint64(9223372036854775808) // int64 max + 1
		maxUint64 := uint64(18446744073709551615)  // uint64 max
		slice := []uint64{largeUint64, maxUint64}
		field := logger.NewIntSliceField("test_truncation", slice)
		require.Equal(t, "test_truncation", field.Name())
		require.Equal(t, "9223372036854775808,18446744073709551615", field.Value())

		// Verify that the values are correctly formatted without truncation
		// If truncated to int64, these would be negative values
		require.NotContains(t, field.Value(), "-")
	})

	t.Run("empty_slice", func(t *testing.T) {
		var slice []int
		field := logger.NewIntSliceField("test_empty", slice)
		require.Equal(t, "test_empty", field.Name())
		require.Equal(t, "", field.Value())
	})

	t.Run("single_element", func(t *testing.T) {
		slice := []int{42}
		field := logger.NewIntSliceField("test_single", slice)
		require.Equal(t, "test_single", field.Name())
		require.Equal(t, "42", field.Value())
	})
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
