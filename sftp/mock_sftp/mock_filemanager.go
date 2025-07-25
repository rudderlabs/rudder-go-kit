// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/rudderlabs/rudder-go-kit/sftp (interfaces: FileManager)
//
// Generated by this command:
//
//	mockgen -destination=mock_sftp/mock_filemanager.go -package mock_sftp github.com/rudderlabs/rudder-go-kit/sftp FileManager
//

// Package mock_sftp is a generated GoMock package.
package mock_sftp

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockFileManager is a mock of FileManager interface.
type MockFileManager struct {
	ctrl     *gomock.Controller
	recorder *MockFileManagerMockRecorder
	isgomock struct{}
}

// MockFileManagerMockRecorder is the mock recorder for MockFileManager.
type MockFileManagerMockRecorder struct {
	mock *MockFileManager
}

// NewMockFileManager creates a new mock instance.
func NewMockFileManager(ctrl *gomock.Controller) *MockFileManager {
	mock := &MockFileManager{ctrl: ctrl}
	mock.recorder = &MockFileManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockFileManager) EXPECT() *MockFileManagerMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockFileManager) Delete(remoteFilePath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", remoteFilePath)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockFileManagerMockRecorder) Delete(remoteFilePath any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockFileManager)(nil).Delete), remoteFilePath)
}

// Download mocks base method.
func (m *MockFileManager) Download(remoteFilePath, localDir string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Download", remoteFilePath, localDir)
	ret0, _ := ret[0].(error)
	return ret0
}

// Download indicates an expected call of Download.
func (mr *MockFileManagerMockRecorder) Download(remoteFilePath, localDir any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Download", reflect.TypeOf((*MockFileManager)(nil).Download), remoteFilePath, localDir)
}

// Upload mocks base method.
func (m *MockFileManager) Upload(localFilePath, remoteDir string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upload", localFilePath, remoteDir)
	ret0, _ := ret[0].(error)
	return ret0
}

// Upload indicates an expected call of Upload.
func (mr *MockFileManagerMockRecorder) Upload(localFilePath, remoteDir any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upload", reflect.TypeOf((*MockFileManager)(nil).Upload), localFilePath, remoteDir)
}
