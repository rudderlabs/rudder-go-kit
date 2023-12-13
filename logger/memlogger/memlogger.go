package memlogger

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/samber/lo"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

var _ logger.Logger = (*Store)(nil)

type Store struct {
	name string

	mu                  sync.Mutex // protects the fields below
	commonKeysAndValues []any
	commonFields        []logger.Field
	entries             []entry
	childStores         []*Store
}

type entry struct {
	level         string
	template      string
	fmtArgs       []any
	keysAndValues []any
	fields        []logger.Field
}

func New() *Store {
	return &Store{}
}

func (ms *Store) IsDebugLevel() bool       { return false }
func (ms *Store) LogRequest(*http.Request) {}

func (ms *Store) Debug(args ...any) {
	ms.log("DEBUG", "", args, nil, nil)
}

func (ms *Store) Info(args ...any) {
	ms.log("INFO", "", args, nil, nil)
}

func (ms *Store) Warn(args ...any) {
	ms.log("WARN", "", args, nil, nil)
}

func (ms *Store) Error(args ...any) {
	ms.log("ERROR", "", args, nil, nil)
}

func (ms *Store) Fatal(args ...any) {
	ms.log("FATAL", "", args, nil, nil)
}

func (ms *Store) Debugf(template string, args ...any) {
	ms.log("DEBUG", template, args, nil, nil)
}

func (ms *Store) Infof(template string, args ...any) {
	ms.log("INFO", template, args, nil, nil)
}

func (ms *Store) Warnf(template string, args ...any) {
	ms.log("WARN", template, args, nil, nil)
}

func (ms *Store) Errorf(template string, args ...any) {
	ms.log("ERROR", template, args, nil, nil)
}

func (ms *Store) Fatalf(template string, args ...any) {
	ms.log("FATAL", template, args, nil, nil)
}

func (ms *Store) Debugw(msg string, keysAndValues ...any) {
	ms.log("DEBUG", msg, nil, keysAndValues, nil)
}

func (ms *Store) Infow(msg string, keysAndValues ...any) {
	ms.log("INFO", msg, nil, keysAndValues, nil)
}

func (ms *Store) Warnw(msg string, keysAndValues ...any) {
	ms.log("WARN", msg, nil, keysAndValues, nil)
}

func (ms *Store) Errorw(msg string, keysAndValues ...any) {
	ms.log("ERROR", msg, nil, keysAndValues, nil)
}

func (ms *Store) Fatalw(msg string, keysAndValues ...any) {
	ms.log("FATAL", msg, nil, keysAndValues, nil)
}

func (ms *Store) Debugn(msg string, fields ...logger.Field) {
	ms.log("DEBUG", msg, nil, nil, fields)
}

func (ms *Store) Infon(msg string, fields ...logger.Field) {
	ms.log("INFO", msg, nil, nil, fields)
}

func (ms *Store) Warnn(msg string, fields ...logger.Field) {
	ms.log("WARN", msg, nil, nil, fields)
}

func (ms *Store) Errorn(msg string, fields ...logger.Field) {
	ms.log("ERROR", msg, nil, nil, fields)
}

func (ms *Store) Fataln(format string, fields ...logger.Field) {
	ms.log("FATAL", format, nil, nil, fields)
}

func (ms *Store) Child(s string) logger.Logger {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	child := New()
	child.commonKeysAndValues = append(child.commonKeysAndValues, ms.commonKeysAndValues...)
	child.commonFields = append(child.commonFields, ms.commonFields...)

	if ms.name == "" {
		child.name = s
	} else {
		child.name = strings.Join([]string{ms.name, s}, ".")
	}

	ms.childStores = append(ms.childStores, child)
	return child
}

func (ms *Store) With(args ...any) logger.Logger {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.commonKeysAndValues = append(ms.commonKeysAndValues, args...)
	return ms
}

func (ms *Store) Withn(args ...logger.Field) logger.Logger {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.commonFields = append(ms.commonFields, args...)
	return ms
}

func (ms *Store) log(level, template string, fmtArgs, keysAndValues []interface{}, fields []logger.Field) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.entries = append(ms.entries, entry{
		level:         level,
		template:      template,
		fmtArgs:       fmtArgs,
		keysAndValues: append(ms.commonKeysAndValues, keysAndValues...),
		fields:        append(ms.commonFields, fields...),
	})
}

type SearchOptions struct {
	Name          string
	Level         string
	Msg           string
	KeysAndValues []interface{}
	Fields        []logger.Field
}

func (ms *Store) Search(opts SearchOptions) int {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	filtered := len(lo.Filter(ms.entries, func(e entry, i int) bool {
		// check for name
		if opts.Name != "" && opts.Name != ms.name {
			return false
		}

		// check for level
		if opts.Level != "" && opts.Level != e.level {
			return false
		}

		// Check for message
		msg := message(e.template, e.fmtArgs)
		if opts.Msg != msg {
			return false
		}

		// check for keys and values
		for _, kv := range opts.KeysAndValues {
			if !lo.Contains(e.keysAndValues, kv) {
				return false
			}
		}

		// check for fields
		for _, f := range opts.Fields {
			if !lo.Contains(e.fields, f) {
				return false
			}
		}
		return true
	}))

	for _, child := range ms.childStores {
		filtered += child.Search(opts)
	}
	return filtered
}

func message(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}
