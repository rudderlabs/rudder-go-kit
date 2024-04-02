package mock_sftp

import (
	"io"
	"os"
)

type MockWriteCloser struct {
	WrittenBytes []byte   // Store written bytes for inspection in tests if needed
	File         *os.File // File to store written bytes
}

// Write writes data to the buffer.
func (m *MockWriteCloser) Write(p []byte) (n int, err error) {
	m.WrittenBytes = append(m.WrittenBytes, p...)
	if m.File != nil {
		_, err := m.File.Write(p)
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

func (m *MockWriteCloser) Close() error {
	if m.File != nil {
		err := m.File.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// MockReadCloser is a mock implementation of io.ReadCloser for testing purposes.
type MockReadCloser struct {
	ReadBytes []byte   // Store read bytes for inspection in tests if needed
	File      *os.File // File to read bytes from
}

// Read reads data from the buffer.
func (m *MockReadCloser) Read(p []byte) (n int, err error) {
	if m.File != nil {
		bytesRead, err := m.File.Read(p)
		if err != nil && err != io.EOF {
			return 0, err
		}
		m.ReadBytes = append(m.ReadBytes, p[:bytesRead]...)
		return bytesRead, err
	}
	return 0, io.EOF
}

// Close closes the file.
func (m *MockReadCloser) Close() error {
	if m.File != nil {
		err := m.File.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
