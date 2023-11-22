package logger

import (
	"time"

	"go.uber.org/zap"
)

type FieldType uint8

const (
	UnknownType FieldType = iota
	StringType
	IntType
	BoolType
	FloatType
	TimeType
	DurationType
	ErrorType
)

type Field struct {
	name      string
	fieldType FieldType

	unknown  any
	string   string
	int      int64
	bool     bool
	float    float64
	time     time.Time
	duration time.Duration
	error    error
}

func (f Field) Name() string { return f.name }
func (f Field) Value() any {
	switch f.fieldType {
	case StringType:
		return f.string
	case IntType:
		return f.int
	case BoolType:
		return f.bool
	case FloatType:
		return f.float
	case TimeType:
		return f.time
	case DurationType:
		return f.duration
	case ErrorType:
		return f.error
	default:
		return f.unknown
	}
}

func (f Field) toZap() zap.Field {
	switch f.fieldType {
	case StringType:
		return zap.String(f.name, f.string)
	case IntType:
		return zap.Int64(f.name, f.int)
	case BoolType:
		return zap.Bool(f.name, f.bool)
	case FloatType:
		return zap.Float64(f.name, f.float)
	case TimeType:
		return zap.Time(f.name, f.time)
	case DurationType:
		return zap.Duration(f.name, f.duration)
	case ErrorType:
		return zap.Error(f.error)
	default:
		return zap.Any(f.name, f.unknown)
	}
}

func NewField(key string, v any) Field {
	return Field{name: key, unknown: v}
}

func NewStringField(key, v string) Field {
	return Field{name: key, string: v, fieldType: StringType}
}

func NewIntField(key string, v int64) Field {
	return Field{name: key, int: v, fieldType: IntType}
}

func NewBoolField(key string, v bool) Field {
	return Field{name: key, bool: v, fieldType: BoolType}
}

func NewFloatField(key string, v float64) Field {
	return Field{name: key, float: v, fieldType: FloatType}
}

func NewTimeField(key string, v time.Time) Field {
	return Field{name: key, time: v, fieldType: TimeType}
}

func NewDurationField(key string, v time.Duration) Field {
	return Field{name: key, duration: v, fieldType: DurationType}
}

func NewErrorField(v error) Field {
	return Field{name: "error", error: v, fieldType: ErrorType}
}

// Expand is useful if you want to use the type Field with the sugared logger
// e.g. l.Infow("my message", logger.Expand(f1, f2, f3)...)
func Expand(fields ...Field) []any {
	result := make([]any, 0, len(fields)*2)
	for _, field := range fields {
		result = append(result, field.name, field.Value())
	}
	return result
}

func toZap(fields []Field) []zap.Field {
	result := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		result = append(result, field.toZap())
	}
	return result
}
