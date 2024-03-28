// Code generated by MockGen. DO NOT EDIT.
// Source: sftp.go

// Package mock_sftp is a generated GoMock package.
package mock_sftp

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSFTPClient is a mock of SFTPClient interface.
type MockSFTPClient struct {
	ctrl     *gomock.Controller
	recorder *MockSFTPClientMockRecorder
}

// MockSFTPClientMockRecorder is the mock recorder for MockSFTPClient.
type MockSFTPClientMockRecorder struct {
	mock *MockSFTPClient
}

// NewMockSFTPClient creates a new mock instance.
func NewMockSFTPClient(ctrl *gomock.Controller) *MockSFTPClient {
	mock := &MockSFTPClient{ctrl: ctrl}
	mock.recorder = &MockSFTPClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSFTPClient) EXPECT() *MockSFTPClientMockRecorder {
	return m.recorder
}

// DeleteFile mocks base method.
func (m *MockSFTPClient) DeleteFile(remoteFilePath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteFile", remoteFilePath)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteFile indicates an expected call of DeleteFile.
func (mr *MockSFTPClientMockRecorder) DeleteFile(remoteFilePath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteFile", reflect.TypeOf((*MockSFTPClient)(nil).DeleteFile), remoteFilePath)
}

// DownloadFile mocks base method.
func (m *MockSFTPClient) DownloadFile(remoteFilePath, localDir string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadFile", remoteFilePath, localDir)
	ret0, _ := ret[0].(error)
	return ret0
}

// DownloadFile indicates an expected call of DownloadFile.
func (mr *MockSFTPClientMockRecorder) DownloadFile(remoteFilePath, localDir interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadFile", reflect.TypeOf((*MockSFTPClient)(nil).DownloadFile), remoteFilePath, localDir)
}

// UploadFile mocks base method.
func (m *MockSFTPClient) UploadFile(localFilePath, remoteDir string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadFile", localFilePath, remoteDir)
	ret0, _ := ret[0].(error)
	return ret0
}

// UploadFile indicates an expected call of UploadFile.
func (mr *MockSFTPClientMockRecorder) UploadFile(localFilePath, remoteDir interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadFile", reflect.TypeOf((*MockSFTPClient)(nil).UploadFile), localFilePath, remoteDir)
}
