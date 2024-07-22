// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rudderlabs/rudder-go-kit/logger (interfaces: Logger)
//
// Generated by this command:
//
//	mockgen -destination=mock_logger/mock_logger.go -package mock_logger github.com/rudderlabs/rudder-go-kit/logger Logger
//

// Package mock_logger is a generated GoMock package.
package mock_logger

import (
	http "net/http"
	reflect "reflect"

	logger "github.com/rudderlabs/rudder-go-kit/logger"
	gomock "go.uber.org/mock/gomock"
)

// MockLogger is a mock of Logger interface.
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger.
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance.
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// Child mocks base method.
func (m *MockLogger) Child(arg0 string) logger.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Child", arg0)
	ret0, _ := ret[0].(logger.Logger)
	return ret0
}

// Child indicates an expected call of Child.
func (mr *MockLoggerMockRecorder) Child(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Child", reflect.TypeOf((*MockLogger)(nil).Child), arg0)
}

// Debug mocks base method.
func (m *MockLogger) Debug(arg0 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debug", varargs...)
}

// Debug indicates an expected call of Debug.
func (mr *MockLoggerMockRecorder) Debug(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLogger)(nil).Debug), arg0...)
}

// Debugf mocks base method.
func (m *MockLogger) Debugf(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debugf", varargs...)
}

// Debugf indicates an expected call of Debugf.
func (mr *MockLoggerMockRecorder) Debugf(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debugf", reflect.TypeOf((*MockLogger)(nil).Debugf), varargs...)
}

// Debugn mocks base method.
func (m *MockLogger) Debugn(arg0 string, arg1 ...logger.Field) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debugn", varargs...)
}

// Debugn indicates an expected call of Debugn.
func (mr *MockLoggerMockRecorder) Debugn(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debugn", reflect.TypeOf((*MockLogger)(nil).Debugn), varargs...)
}

// Debugw mocks base method.
func (m *MockLogger) Debugw(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debugw", varargs...)
}

// Debugw indicates an expected call of Debugw.
func (mr *MockLoggerMockRecorder) Debugw(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debugw", reflect.TypeOf((*MockLogger)(nil).Debugw), varargs...)
}

// Error mocks base method.
func (m *MockLogger) Error(arg0 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Error", varargs...)
}

// Error indicates an expected call of Error.
func (mr *MockLoggerMockRecorder) Error(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLogger)(nil).Error), arg0...)
}

// Errorf mocks base method.
func (m *MockLogger) Errorf(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Errorf", varargs...)
}

// Errorf indicates an expected call of Errorf.
func (mr *MockLoggerMockRecorder) Errorf(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Errorf", reflect.TypeOf((*MockLogger)(nil).Errorf), varargs...)
}

// Errorn mocks base method.
func (m *MockLogger) Errorn(arg0 string, arg1 ...logger.Field) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Errorn", varargs...)
}

// Errorn indicates an expected call of Errorn.
func (mr *MockLoggerMockRecorder) Errorn(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Errorn", reflect.TypeOf((*MockLogger)(nil).Errorn), varargs...)
}

// Errorw mocks base method.
func (m *MockLogger) Errorw(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Errorw", varargs...)
}

// Errorw indicates an expected call of Errorw.
func (mr *MockLoggerMockRecorder) Errorw(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Errorw", reflect.TypeOf((*MockLogger)(nil).Errorw), varargs...)
}

// Fatal mocks base method.
func (m *MockLogger) Fatal(arg0 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Fatal", varargs...)
}

// Fatal indicates an expected call of Fatal.
func (mr *MockLoggerMockRecorder) Fatal(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fatal", reflect.TypeOf((*MockLogger)(nil).Fatal), arg0...)
}

// Fatalf mocks base method.
func (m *MockLogger) Fatalf(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Fatalf", varargs...)
}

// Fatalf indicates an expected call of Fatalf.
func (mr *MockLoggerMockRecorder) Fatalf(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fatalf", reflect.TypeOf((*MockLogger)(nil).Fatalf), varargs...)
}

// Fataln mocks base method.
func (m *MockLogger) Fataln(arg0 string, arg1 ...logger.Field) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Fataln", varargs...)
}

// Fataln indicates an expected call of Fataln.
func (mr *MockLoggerMockRecorder) Fataln(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fataln", reflect.TypeOf((*MockLogger)(nil).Fataln), varargs...)
}

// Fatalw mocks base method.
func (m *MockLogger) Fatalw(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Fatalw", varargs...)
}

// Fatalw indicates an expected call of Fatalw.
func (mr *MockLoggerMockRecorder) Fatalw(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Fatalw", reflect.TypeOf((*MockLogger)(nil).Fatalw), varargs...)
}

// Info mocks base method.
func (m *MockLogger) Info(arg0 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Info", varargs...)
}

// Info indicates an expected call of Info.
func (mr *MockLoggerMockRecorder) Info(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLogger)(nil).Info), arg0...)
}

// Infof mocks base method.
func (m *MockLogger) Infof(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Infof", varargs...)
}

// Infof indicates an expected call of Infof.
func (mr *MockLoggerMockRecorder) Infof(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Infof", reflect.TypeOf((*MockLogger)(nil).Infof), varargs...)
}

// Infon mocks base method.
func (m *MockLogger) Infon(arg0 string, arg1 ...logger.Field) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Infon", varargs...)
}

// Infon indicates an expected call of Infon.
func (mr *MockLoggerMockRecorder) Infon(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Infon", reflect.TypeOf((*MockLogger)(nil).Infon), varargs...)
}

// Infow mocks base method.
func (m *MockLogger) Infow(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Infow", varargs...)
}

// Infow indicates an expected call of Infow.
func (mr *MockLoggerMockRecorder) Infow(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Infow", reflect.TypeOf((*MockLogger)(nil).Infow), varargs...)
}

// IsDebugLevel mocks base method.
func (m *MockLogger) IsDebugLevel() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDebugLevel")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsDebugLevel indicates an expected call of IsDebugLevel.
func (mr *MockLoggerMockRecorder) IsDebugLevel() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDebugLevel", reflect.TypeOf((*MockLogger)(nil).IsDebugLevel))
}

// LogRequest mocks base method.
func (m *MockLogger) LogRequest(arg0 *http.Request) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "LogRequest", arg0)
}

// LogRequest indicates an expected call of LogRequest.
func (mr *MockLoggerMockRecorder) LogRequest(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogRequest", reflect.TypeOf((*MockLogger)(nil).LogRequest), arg0)
}

// Warn mocks base method.
func (m *MockLogger) Warn(arg0 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warn", varargs...)
}

// Warn indicates an expected call of Warn.
func (mr *MockLoggerMockRecorder) Warn(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLogger)(nil).Warn), arg0...)
}

// Warnf mocks base method.
func (m *MockLogger) Warnf(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warnf", varargs...)
}

// Warnf indicates an expected call of Warnf.
func (mr *MockLoggerMockRecorder) Warnf(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warnf", reflect.TypeOf((*MockLogger)(nil).Warnf), varargs...)
}

// Warnn mocks base method.
func (m *MockLogger) Warnn(arg0 string, arg1 ...logger.Field) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warnn", varargs...)
}

// Warnn indicates an expected call of Warnn.
func (mr *MockLoggerMockRecorder) Warnn(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warnn", reflect.TypeOf((*MockLogger)(nil).Warnn), varargs...)
}

// Warnw mocks base method.
func (m *MockLogger) Warnw(arg0 string, arg1 ...any) {
	m.ctrl.T.Helper()
	varargs := []any{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warnw", varargs...)
}

// Warnw indicates an expected call of Warnw.
func (mr *MockLoggerMockRecorder) Warnw(arg0 any, arg1 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warnw", reflect.TypeOf((*MockLogger)(nil).Warnw), varargs...)
}

// With mocks base method.
func (m *MockLogger) With(arg0 ...any) logger.Logger {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "With", varargs...)
	ret0, _ := ret[0].(logger.Logger)
	return ret0
}

// With indicates an expected call of With.
func (mr *MockLoggerMockRecorder) With(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "With", reflect.TypeOf((*MockLogger)(nil).With), arg0...)
}

// Withn mocks base method.
func (m *MockLogger) Withn(arg0 ...logger.Field) logger.Logger {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Withn", varargs...)
	ret0, _ := ret[0].(logger.Logger)
	return ret0
}

// Withn indicates an expected call of Withn.
func (mr *MockLoggerMockRecorder) Withn(arg0 ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Withn", reflect.TypeOf((*MockLogger)(nil).Withn), arg0...)
}
