// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rudderlabs/rudder-go-kit/stats (interfaces: Stats,Measurement)

// Package mock_stats is a generated GoMock package.
package mock_stats

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	stats "github.com/rudderlabs/rudder-go-kit/stats"
)

// MockStats is a mock of Stats interface.
type MockStats struct {
	ctrl     *gomock.Controller
	recorder *MockStatsMockRecorder
}

// MockStatsMockRecorder is the mock recorder for MockStats.
type MockStatsMockRecorder struct {
	mock *MockStats
}

// NewMockStats creates a new mock instance.
func NewMockStats(ctrl *gomock.Controller) *MockStats {
	mock := &MockStats{ctrl: ctrl}
	mock.recorder = &MockStatsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStats) EXPECT() *MockStatsMockRecorder {
	return m.recorder
}

// NewSampledTaggedStat mocks base method.
func (m *MockStats) NewSampledTaggedStat(arg0, arg1 string, arg2 stats.Tags) stats.Measurement {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewSampledTaggedStat", arg0, arg1, arg2)
	ret0, _ := ret[0].(stats.Measurement)
	return ret0
}

// NewSampledTaggedStat indicates an expected call of NewSampledTaggedStat.
func (mr *MockStatsMockRecorder) NewSampledTaggedStat(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewSampledTaggedStat", reflect.TypeOf((*MockStats)(nil).NewSampledTaggedStat), arg0, arg1, arg2)
}

// NewStat mocks base method.
func (m *MockStats) NewStat(arg0, arg1 string) stats.Measurement {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewStat", arg0, arg1)
	ret0, _ := ret[0].(stats.Measurement)
	return ret0
}

// NewStat indicates an expected call of NewStat.
func (mr *MockStatsMockRecorder) NewStat(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewStat", reflect.TypeOf((*MockStats)(nil).NewStat), arg0, arg1)
}

// NewTaggedStat mocks base method.
func (m *MockStats) NewTaggedStat(arg0, arg1 string, arg2 stats.Tags) stats.Measurement {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewTaggedStat", arg0, arg1, arg2)
	ret0, _ := ret[0].(stats.Measurement)
	return ret0
}

// NewTaggedStat indicates an expected call of NewTaggedStat.
func (mr *MockStatsMockRecorder) NewTaggedStat(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewTaggedStat", reflect.TypeOf((*MockStats)(nil).NewTaggedStat), arg0, arg1, arg2)
}

// Start mocks base method.
func (m *MockStats) Start(arg0 context.Context, arg1 stats.GoRoutineFactory) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Start", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Start indicates an expected call of Start.
func (mr *MockStatsMockRecorder) Start(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Start", reflect.TypeOf((*MockStats)(nil).Start), arg0, arg1)
}

// Stop mocks base method.
func (m *MockStats) Stop() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Stop")
}

// Stop indicates an expected call of Stop.
func (mr *MockStatsMockRecorder) Stop() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockStats)(nil).Stop))
}

// MockMeasurement is a mock of Measurement interface.
type MockMeasurement struct {
	ctrl     *gomock.Controller
	recorder *MockMeasurementMockRecorder
}

// MockMeasurementMockRecorder is the mock recorder for MockMeasurement.
type MockMeasurementMockRecorder struct {
	mock *MockMeasurement
}

// NewMockMeasurement creates a new mock instance.
func NewMockMeasurement(ctrl *gomock.Controller) *MockMeasurement {
	mock := &MockMeasurement{ctrl: ctrl}
	mock.recorder = &MockMeasurementMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMeasurement) EXPECT() *MockMeasurementMockRecorder {
	return m.recorder
}

// Count mocks base method.
func (m *MockMeasurement) Count(arg0 int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Count", arg0)
}

// Count indicates an expected call of Count.
func (mr *MockMeasurementMockRecorder) Count(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Count", reflect.TypeOf((*MockMeasurement)(nil).Count), arg0)
}

// Gauge mocks base method.
func (m *MockMeasurement) Gauge(arg0 interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Gauge", arg0)
}

// Gauge indicates an expected call of Gauge.
func (mr *MockMeasurementMockRecorder) Gauge(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Gauge", reflect.TypeOf((*MockMeasurement)(nil).Gauge), arg0)
}

// Increment mocks base method.
func (m *MockMeasurement) Increment() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Increment")
}

// Increment indicates an expected call of Increment.
func (mr *MockMeasurementMockRecorder) Increment() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Increment", reflect.TypeOf((*MockMeasurement)(nil).Increment))
}

// Observe mocks base method.
func (m *MockMeasurement) Observe(arg0 float64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Observe", arg0)
}

// Observe indicates an expected call of Observe.
func (mr *MockMeasurementMockRecorder) Observe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Observe", reflect.TypeOf((*MockMeasurement)(nil).Observe), arg0)
}

// RecordDuration mocks base method.
func (m *MockMeasurement) RecordDuration() func() {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RecordDuration")
	ret0, _ := ret[0].(func())
	return ret0
}

// RecordDuration indicates an expected call of RecordDuration.
func (mr *MockMeasurementMockRecorder) RecordDuration() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordDuration", reflect.TypeOf((*MockMeasurement)(nil).RecordDuration))
}

// SendTiming mocks base method.
func (m *MockMeasurement) SendTiming(arg0 time.Duration) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SendTiming", arg0)
}

// SendTiming indicates an expected call of SendTiming.
func (mr *MockMeasurementMockRecorder) SendTiming(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendTiming", reflect.TypeOf((*MockMeasurement)(nil).SendTiming), arg0)
}

// Since mocks base method.
func (m *MockMeasurement) Since(arg0 time.Time) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Since", arg0)
}

// Since indicates an expected call of Since.
func (mr *MockMeasurementMockRecorder) Since(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Since", reflect.TypeOf((*MockMeasurement)(nil).Since), arg0)
}