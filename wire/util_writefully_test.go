package wire

import (
	"errors"
	"testing"

	adbErrors "github.com/basiooo/goadb/internal/errors"
	"github.com/stretchr/testify/assert"
)

// mockWriter is a mock implementation of io.Writer for testing
type mockWriter struct {
	writeFunc func(p []byte) (n int, err error)
	callCount int
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.callCount++
	return m.writeFunc(p)
}

func TestWriteFully_Success(t *testing.T) {
	// Test successful write in one call
	mock := &mockWriter{
		writeFunc: func(p []byte) (int, error) {
			return len(p), nil
		},
	}
	
	err := writeFully(mock, []byte("test data"))
	assert.NoError(t, err)
	assert.Equal(t, 1, mock.callCount)
}

func TestWriteFully_PartialWrites(t *testing.T) {
	// Test successful write with multiple partial writes
	callNum := 0
	mock := &mockWriter{
		writeFunc: func(p []byte) (int, error) {
			callNum++
			switch callNum {
			case 1:
				// First call writes 2 bytes
				return 2, nil
			case 2:
				// Second call writes 3 bytes
				return 3, nil
			default:
				// Third call writes the rest
				return len(p), nil
			}
		},
	}
	
	err := writeFully(mock, []byte("test data"))
	assert.NoError(t, err)
	assert.Equal(t, 3, mock.callCount)
}

func TestWriteFully_Error(t *testing.T) {
	// Test write with error
	mockErr := errors.New("write error")
	mock := &mockWriter{
		writeFunc: func(p []byte) (int, error) {
			return 0, mockErr
		},
	}
	
	err := writeFully(mock, []byte("test data"))
	assert.Error(t, err)
	assert.Equal(t, 1, mock.callCount)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestWriteFully_PartialWriteThenError(t *testing.T) {
	// Test partial write followed by error
	mockErr := errors.New("write error")
	callNum := 0
	mock := &mockWriter{
		writeFunc: func(p []byte) (int, error) {
			callNum++
			switch callNum {
			case 1:
				// First call writes 2 bytes
				return 2, nil
			default:
				// Second call returns error
				return 0, mockErr
			}
		},
	}
	
	err := writeFully(mock, []byte("test data"))
	assert.Error(t, err)
	assert.Equal(t, 2, mock.callCount)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestWriteFully_EmptyData(t *testing.T) {
	// Test with empty data
	mock := &mockWriter{
		writeFunc: func(p []byte) (int, error) {
			return len(p), nil
		},
	}
	
	err := writeFully(mock, []byte{})
	assert.NoError(t, err)
	// Write should not be called for empty data
	assert.Equal(t, 0, mock.callCount)
}