package memlogger_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/logger/memlogger"
)

func TestLogger(t *testing.T) {
	t.Run("With messages", func(t *testing.T) {
		ml := memlogger.New()
		for i := 0; i < 3; i++ {
			ml.Debug("debug", " message")
		}
		for i := 0; i < 4; i++ {
			ml.Info("info", " message")
		}
		for i := 0; i < 5; i++ {
			ml.Warn("warn", " message")
		}
		for i := 0; i < 6; i++ {
			ml.Error("error", " message")
		}
		for i := 0; i < 7; i++ {
			ml.Fatal("fatal", " message")
		}

		require.Equal(t, 3, ml.Search(memlogger.SearchOptions{
			Level: "DEBUG",
			Msg:   "debug message",
		}))
		require.Equal(t, 0, ml.Search(memlogger.SearchOptions{
			Level: "DEBUG",
			Msg:   "some other debug message",
		}))
		require.Equal(t, 4, ml.Search(memlogger.SearchOptions{
			Level: "INFO",
			Msg:   "info message",
		}))
		require.Equal(t, 5, ml.Search(memlogger.SearchOptions{
			Level: "WARN",
			Msg:   "warn message",
		}))
		require.Equal(t, 6, ml.Search(memlogger.SearchOptions{
			Level: "ERROR",
			Msg:   "error message",
		}))
		require.Equal(t, 7, ml.Search(memlogger.SearchOptions{
			Level: "FATAL",
			Msg:   "fatal message",
		}))
	})
	t.Run("With formatted message", func(t *testing.T) {
		ml := memlogger.New()
		for i := 0; i < 3; i++ {
			ml.Debugf("debug %s %s", "formatted", "message")
		}
		for i := 0; i < 4; i++ {
			ml.Infof("info %s %s", "formatted", "message")
		}
		for i := 0; i < 5; i++ {
			ml.Warnf("warn %s %s", "formatted", "message")
		}
		for i := 0; i < 6; i++ {
			ml.Errorf("error %s %s", "formatted", "message")
		}
		for i := 0; i < 7; i++ {
			ml.Fatalf("fatal %s %s", "formatted", "message")
		}

		require.Equal(t, 3, ml.Search(memlogger.SearchOptions{
			Level: "DEBUG",
			Msg:   "debug formatted message",
		}))
		require.Equal(t, 0, ml.Search(memlogger.SearchOptions{
			Level: "DEBUG",
			Msg:   "some other debug formatted message",
		}))
		require.Equal(t, 4, ml.Search(memlogger.SearchOptions{
			Level: "INFO",
			Msg:   "info formatted message",
		}))
		require.Equal(t, 5, ml.Search(memlogger.SearchOptions{
			Level: "WARN",
			Msg:   "warn formatted message",
		}))
		require.Equal(t, 6, ml.Search(memlogger.SearchOptions{
			Level: "ERROR",
			Msg:   "error formatted message",
		}))
		require.Equal(t, 7, ml.Search(memlogger.SearchOptions{
			Level: "FATAL",
			Msg:   "fatal formatted message",
		}))
	})
	t.Run("With keys and values", func(t *testing.T) {
		ml := memlogger.New()
		for i := 0; i < 3; i++ {
			ml.Debugw("debug", "key", "value")
		}
		for i := 0; i < 4; i++ {
			ml.Infow("info", "key", "value")
		}
		for i := 0; i < 5; i++ {
			ml.Warnw("warn", "key", "value")
		}
		for i := 0; i < 6; i++ {
			ml.Errorw("error", "key", "value")
		}
		for i := 0; i < 7; i++ {
			ml.Fatalw("fatal", "key", "value")
		}

		require.Equal(t, 3, ml.Search(memlogger.SearchOptions{
			Level:         "DEBUG",
			Msg:           "debug",
			KeysAndValues: []any{"key", "value"},
		}))
		require.Equal(t, 0, ml.Search(memlogger.SearchOptions{
			Level:         "DEBUG",
			Msg:           "some other debug",
			KeysAndValues: []any{"key", "value"},
		}))
		require.Equal(t, 0, ml.Search(memlogger.SearchOptions{
			Level:         "DEBUG",
			Msg:           "debug",
			KeysAndValues: []any{"some other key", "value"},
		}))
		require.Equal(t, 0, ml.Search(memlogger.SearchOptions{
			Level:         "DEBUG",
			Msg:           "some other debug",
			KeysAndValues: []any{"key", "some other value"},
		}))
		require.Equal(t, 4, ml.Search(memlogger.SearchOptions{
			Level:         "INFO",
			Msg:           "info",
			KeysAndValues: []any{"key", "value"},
		}))
		require.Equal(t, 5, ml.Search(memlogger.SearchOptions{
			Level:         "WARN",
			Msg:           "warn",
			KeysAndValues: []any{"key", "value"},
		}))
		require.Equal(t, 6, ml.Search(memlogger.SearchOptions{
			Level:         "ERROR",
			Msg:           "error",
			KeysAndValues: []any{"key", "value"},
		}))
		require.Equal(t, 7, ml.Search(memlogger.SearchOptions{
			Level:         "FATAL",
			Msg:           "fatal",
			KeysAndValues: []any{"key", "value"},
		}))
	})
	t.Run("With fields", func(t *testing.T) {
		ml := memlogger.New()
		for i := 0; i < 3; i++ {
			ml.Debugn("debug", logger.NewStringField("message", "message"))
		}
		for i := 0; i < 4; i++ {
			ml.Infon("info", logger.NewStringField("message", "message"))
		}
		for i := 0; i < 5; i++ {
			ml.Warnn("warn", logger.NewStringField("message", "message"))
		}
		for i := 0; i < 6; i++ {
			ml.Errorn("error", logger.NewStringField("message", "message"))
		}
		for i := 0; i < 7; i++ {
			ml.Fataln("fatal", logger.NewStringField("message", "message"))
		}

		require.Equal(t, 3, ml.Search(memlogger.SearchOptions{
			Level:  "DEBUG",
			Msg:    "debug",
			Fields: []logger.Field{logger.NewStringField("message", "message")},
		}))
		require.Equal(t, 0, ml.Search(memlogger.SearchOptions{
			Level:  "DEBUG",
			Msg:    "some other debug",
			Fields: []logger.Field{logger.NewStringField("message", "message")},
		}))
		require.Equal(t, 0, ml.Search(memlogger.SearchOptions{
			Level:  "DEBUG",
			Msg:    "debug",
			Fields: []logger.Field{logger.NewStringField("some other message", "message")},
		}))
		require.Equal(t, 0, ml.Search(memlogger.SearchOptions{
			Level:  "DEBUG",
			Msg:    "debug",
			Fields: []logger.Field{logger.NewStringField("message", "some other message")},
		}))
		require.Equal(t, 4, ml.Search(memlogger.SearchOptions{
			Level:  "INFO",
			Msg:    "info",
			Fields: []logger.Field{logger.NewStringField("message", "message")},
		}))
		require.Equal(t, 5, ml.Search(memlogger.SearchOptions{
			Level:  "WARN",
			Msg:    "warn",
			Fields: []logger.Field{logger.NewStringField("message", "message")},
		}))
		require.Equal(t, 6, ml.Search(memlogger.SearchOptions{
			Level:  "ERROR",
			Msg:    "error",
			Fields: []logger.Field{logger.NewStringField("message", "message")},
		}))
	})
	t.Run("With children's", func(t *testing.T) {
		ml := memlogger.New()
		child := ml.Child("child")
		grandChild := child.Child("grandchild")

		child.Debug("debug", " child ", "message")
		grandChild.Debug("debug", " grandchild ", "message")

		require.Equal(t, 1, ml.Search(memlogger.SearchOptions{
			Name:  "child",
			Level: "DEBUG",
			Msg:   "debug child message",
		}))
		require.Equal(t, 1, ml.Search(memlogger.SearchOptions{
			Name:  "child.grandchild",
			Level: "DEBUG",
			Msg:   "debug grandchild message",
		}))
	})
	t.Run("With WithArgs", func(t *testing.T) {
		ml := memlogger.New()
		mlW := ml.With("key 1", "value 1", "key 2", "value 2")
		mlW.Debug("debug", " message")
		require.Equal(t, 1, ml.Search(memlogger.SearchOptions{
			Level:         "DEBUG",
			Msg:           "debug message",
			KeysAndValues: []any{"key 1", "value 1", "key 2", "value 2"},
		}))

		child := mlW.Child("child")
		childW := child.With("key 3", "value 3", "key 4", "value 4")
		childW.Debug("debug", " child ", "message")
		require.Equal(t, 1, ml.Search(memlogger.SearchOptions{
			Name:          "child",
			Level:         "DEBUG",
			Msg:           "debug child message",
			KeysAndValues: []any{"key 1", "value 1", "key 2", "value 2", "key 3", "value 3", "key 4", "value 4"},
		}))
	})
	t.Run("With WithnArgs", func(t *testing.T) {
		ml := memlogger.New()
		mlW := ml.Withn(logger.NewStringField("key 1", "value 1"), logger.NewStringField("key 2", "value 2"))
		mlW.Debug("debug", " message")
		require.Equal(t, 1, ml.Search(memlogger.SearchOptions{
			Level:  "DEBUG",
			Msg:    "debug message",
			Fields: []logger.Field{logger.NewStringField("key 1", "value 1"), logger.NewStringField("key 2", "value 2")},
		}))

		child := mlW.Child("child")
		childW := child.Withn(logger.NewStringField("key 3", "value 3"), logger.NewStringField("key 4", "value 4"))
		childW.Debug("debug", " child ", "message")
		require.Equal(t, 1, ml.Search(memlogger.SearchOptions{
			Name:   "child",
			Level:  "DEBUG",
			Msg:    "debug child message",
			Fields: []logger.Field{logger.NewStringField("key 1", "value 1"), logger.NewStringField("key 2", "value 2"), logger.NewStringField("key 3", "value 3"), logger.NewStringField("key 4", "value 4")},
		}))
	})
}
