package logger

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rudderlabs/rudder-go-kit/config"
)

// BenchmarkKit results:
// non-sugared:	    496.6 ns/op
// sugared:         723.5 ns/op
func BenchmarkKit(b *testing.B) {
	c := config.New()
	c.Set("LOG_LEVEL", "DEBUG")
	c.Set("Logger.consoleJsonFormat", "json")
	c.Set("Logger.discardConsole", true)
	c.Set("Logger.enableFileNameInLog", true)

	b.Run("non-sugared", func(b *testing.B) {
		f := NewFactory(c)
		l := f.NewLogger()
		defer f.Sync()

		fields := []Field{
			NewStringField("key1", "111"), NewStringField("key2", "222"),
			NewStringField("key3", "333"), NewStringField("key4", "444"),
			NewStringField("key5", "555"), NewStringField("key6", "666"),
			NewStringField("key7", "777"), NewStringField("key8", "888"),
			NewStringField("key9", "999"), NewStringField("key10", "101010"),
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Debugn("test", fields...)
			}
		})
	})

	b.Run("sugared", func(b *testing.B) {
		f := NewFactory(c)
		l := f.NewLogger()
		defer f.Sync()

		fields := []any{
			"key1", "111", "key2", "222",
			"key3", "333", "key4", "444",
			"key5", "555", "key6", "666",
			"key7", "777", "key8", "888",
			"key9", "999", "key10", "101010",
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Debugw("test", fields...)
			}
		})
	})
}

// BenchmarkZap results:
// Zap:         60.96 ns/op
// Zap.Sugar:   94.57 ns/op
func BenchmarkZap(b *testing.B) {
	newZap := func(lvl zapcore.Level) *zap.Logger {
		ec := zap.NewProductionEncoderConfig()
		ec.EncodeDuration = zapcore.NanosDurationEncoder
		ec.EncodeTime = zapcore.EpochNanosTimeEncoder
		enc := zapcore.NewJSONEncoder(ec)
		return zap.New(zapcore.NewCore(
			enc,
			&discarder{},
			lvl,
		))
	}
	b.Run("Zap", func(b *testing.B) {
		logger := newZap(zapcore.DebugLevel)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Debug("message", zap.String("key1", "111"), zap.String("key2", "222"))
			}
		})
	})
	b.Run("Zap.Sugar", func(b *testing.B) {
		logger := newZap(zapcore.DebugLevel).Sugar()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Debugw("message", "key1", "111", "key2", "222")
			}
		})
	})
}
