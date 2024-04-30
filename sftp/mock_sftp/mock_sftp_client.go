// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rudderlabs/rudder-go-kit/sftp (interfaces: Client)

// Package mock_sftp is a generated GoMock package.
package mock_sftp

import (
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// MkdirAll mocks base method.
func (m *MockClient) MkdirAll(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MkdirAll", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// MkdirAll indicates an expected call of MkdirAll.
func (mr *MockClientMockRecorder) MkdirAll(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MkdirAll", reflect.TypeOf((*MockClient)(nil).MkdirAll), arg0)
}

// OpenFile mocks base method.
func (m *MockClient) OpenFile(arg0 string, arg1 int) (io.ReadWriteCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "OpenFile", arg0, arg1)
	ret0, _ := ret[0].(io.ReadWriteCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// OpenFile indicates an expected call of OpenFile.
func (mr *MockClientMockRecorder) OpenFile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "OpenFile", reflect.TypeOf((*MockClient)(nil).OpenFile), arg0, arg1)
}

// Remove mocks base method.
func (m *MockClient) Remove(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Remove", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Remove indicates an expected call of Remove.
func (mr *MockClientMockRecorder) Remove(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Remove", reflect.TypeOf((*MockClient)(nil).Remove), arg0)
}
