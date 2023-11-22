package logger

import "net/http"

var NOP Logger = nop{}

type nop struct{}

func (nop) Debug(_ ...any)              {}
func (nop) Info(_ ...any)               {}
func (nop) Warn(_ ...any)               {}
func (nop) Error(_ ...any)              {}
func (nop) Fatal(_ ...any)              {}
func (nop) Debugf(_ string, _ ...any)   {}
func (nop) Infof(_ string, _ ...any)    {}
func (nop) Warnf(_ string, _ ...any)    {}
func (nop) Errorf(_ string, _ ...any)   {}
func (nop) Fatalf(_ string, _ ...any)   {}
func (nop) Debugw(_ string, _ ...any)   {}
func (nop) Infow(_ string, _ ...any)    {}
func (nop) Warnw(_ string, _ ...any)    {}
func (nop) Errorw(_ string, _ ...any)   {}
func (nop) Fatalw(_ string, _ ...any)   {}
func (nop) Debugn(_ string, _ ...Field) {}
func (nop) Infon(_ string, _ ...Field)  {}
func (nop) Warnn(_ string, _ ...Field)  {}
func (nop) Errorn(_ string, _ ...Field) {}
func (nop) Fataln(_ string, _ ...Field) {}
func (nop) LogRequest(_ *http.Request)  {}
func (nop) With(_ ...any) Logger        { return NOP }
func (nop) Withn(_ ...Field) Logger     { return NOP }
func (nop) Child(_ string) Logger       { return NOP }
func (nop) IsDebugLevel() bool          { return false }
