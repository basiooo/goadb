package wire

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockReadWriteCloser is a mock implementation of io.ReadWriteCloser for testing
type mockReadWriteCloser struct {
	readFunc  func(p []byte) (n int, err error)
	writeFunc func(p []byte) (n int, err error)
	closeFunc func() error
	closeCount int
}

func (m *mockReadWriteCloser) Read(p []byte) (n int, err error) {
	return m.readFunc(p)
}

func (m *mockReadWriteCloser) Write(p []byte) (n int, err error) {
	return m.writeFunc(p)
}

func (m *mockReadWriteCloser) Close() error {
	m.closeCount++
	return m.closeFunc()
}

func TestMultiCloseable(t *testing.T) {
	// Test that MultiCloseable can be closed multiple times
	// and only calls the underlying Close() once
	
	mockErr := errors.New("mock close error")
	mock := &mockReadWriteCloser{
		readFunc: func(p []byte) (int, error) { return 0, io.EOF },
		writeFunc: func(p []byte) (int, error) { return len(p), nil },
		closeFunc: func() error { return mockErr },
	}
	
	multiCloser := MultiCloseable(mock)
	
	// First close should call the underlying Close()
	err1 := multiCloser.Close()
	assert.Equal(t, mockErr, err1)
	assert.Equal(t, 1, mock.closeCount)
	
	// Second close should not call the underlying Close() again
	// but should return the same error
	err2 := multiCloser.Close()
	assert.Equal(t, mockErr, err2)
	assert.Equal(t, 1, mock.closeCount)
	
	// Third close should behave the same
	err3 := multiCloser.Close()
	assert.Equal(t, mockErr, err3)
	assert.Equal(t, 1, mock.closeCount)
}

func TestMultiCloseable_NoError(t *testing.T) {
	// Test with no error from Close()
	mock := &mockReadWriteCloser{
		readFunc: func(p []byte) (int, error) { return 0, io.EOF },
		writeFunc: func(p []byte) (int, error) { return len(p), nil },
		closeFunc: func() error { return nil },
	}
	
	multiCloser := MultiCloseable(mock)
	
	// First close should call the underlying Close()
	err1 := multiCloser.Close()
	assert.NoError(t, err1)
	assert.Equal(t, 1, mock.closeCount)
	
	// Second close should not call the underlying Close() again
	err2 := multiCloser.Close()
	assert.NoError(t, err2)
	assert.Equal(t, 1, mock.closeCount)
}

func TestMultiCloseable_ReadWrite(t *testing.T) {
	// Test that Read and Write are passed through to the underlying ReadWriteCloser
	mock := &mockReadWriteCloser{
		readFunc: func(p []byte) (int, error) { 
			if len(p) > 0 {
				p[0] = 'A'
			}
			return 1, nil 
		},
		writeFunc: func(p []byte) (int, error) { return len(p), nil },
		closeFunc: func() error { return nil },
	}
	
	multiCloser := MultiCloseable(mock)
	
	// Read should be passed through
	buf := make([]byte, 10)
	n, err := multiCloser.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, byte('A'), buf[0])
	
	// Write should be passed through
	n, err = multiCloser.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
}