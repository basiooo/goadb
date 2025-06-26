package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	adbErrors "github.com/basiooo/goadb/internal/errors"
	"github.com/stretchr/testify/assert"
)

// mockWriter implements io.Writer and io.Closer for testing
type mockSyncSenderWriter struct {
	writeFunc func(p []byte) (n int, err error)
	closeFunc func() error
	written   []byte
}

func (m *mockSyncSenderWriter) Write(p []byte) (n int, err error) {
	if m.writeFunc != nil {
		return m.writeFunc(p)
	}
	m.written = append(m.written, p...)
	return len(p), nil
}

func (m *mockSyncSenderWriter) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestNewSyncSender(t *testing.T) {
	writer := &mockSyncSenderWriter{}
	sender := NewSyncSender(writer)
	
	// Verify sender is created with the writer
	assert.NotNil(t, sender)
	
	// Verify it's the correct type
	_, ok := sender.(*realSyncSender)
	assert.True(t, ok)
}

func TestSyncSender_SendOctetString_Success(t *testing.T) {
	writer := &mockSyncSenderWriter{}
	sender := NewSyncSender(writer)
	
	// Send a 4-byte string
	err := sender.SendOctetString("OKAY")
	
	assert.NoError(t, err)
	assert.Equal(t, []byte("OKAY"), writer.written)
}

func TestSyncSender_SendOctetString_InvalidLength(t *testing.T) {
	writer := &mockSyncSenderWriter{}
	sender := NewSyncSender(writer)
	
	// Send a string that's not 4 bytes
	err := sender.SendOctetString("TOO_LONG")
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "octet string must be exactly 4 bytes")
	assert.Empty(t, writer.written)
}

func TestSyncSender_SendOctetString_WriteError(t *testing.T) {
	expectedErr := errors.New("write error")
	writer := &mockSyncSenderWriter{
		writeFunc: func(p []byte) (n int, err error) {
			return 0, expectedErr
		},
	}
	sender := NewSyncSender(writer)
	
	// Send an octet string but get a write error
	err := sender.SendOctetString("ABCD")
	
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncSender_SendInt32_Success(t *testing.T) {
	writer := &mockSyncSenderWriter{}
	sender := NewSyncSender(writer)
	
	// Send an int32
	value := int32(12345)
	err := sender.SendInt32(value)
	
	assert.NoError(t, err)
	
	// Verify the correct bytes were written
	expectedBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(expectedBytes, uint32(value))
	assert.Equal(t, expectedBytes, writer.written)
}

func TestSyncSender_SendInt32_WriteError(t *testing.T) {
	expectedErr := errors.New("write error")
	writer := &mockSyncSenderWriter{
		writeFunc: func(p []byte) (n int, err error) {
			return 0, expectedErr
		},
	}
	sender := NewSyncSender(writer)
	
	// Send an int32 but get a write error
	err := sender.SendInt32(42)
	
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncSender_SendFileMode_Success(t *testing.T) {
	writer := &mockSyncSenderWriter{}
	sender := NewSyncSender(writer)
	
	// Send a file mode
	mode := os.FileMode(0644)
	err := sender.SendFileMode(mode)
	
	assert.NoError(t, err)
	
	// Verify the correct bytes were written
	expectedBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(expectedBytes, uint32(mode))
	assert.Equal(t, expectedBytes, writer.written)
}

func TestSyncSender_SendFileMode_WriteError(t *testing.T) {
	expectedErr := errors.New("write error")
	writer := &mockSyncSenderWriter{
		writeFunc: func(p []byte) (n int, err error) {
			return 0, expectedErr
		},
	}
	sender := NewSyncSender(writer)
	
	// Send a file mode but get a write error
	err := sender.SendFileMode(0644)
	
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncSender_SendTime_Success(t *testing.T) {
	writer := &mockSyncSenderWriter{}
	sender := NewSyncSender(writer)
	
	// Send a time
	tm := time.Unix(1609459200, 0) // 2021-01-01 00:00:00 UTC
	err := sender.SendTime(tm)
	
	assert.NoError(t, err)
	
	// Verify the correct bytes were written
	expectedBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(expectedBytes, uint32(tm.Unix()))
	assert.Equal(t, expectedBytes, writer.written)
}

func TestSyncSender_SendTime_WriteError(t *testing.T) {
	expectedErr := errors.New("write error")
	writer := &mockSyncSenderWriter{
		writeFunc: func(p []byte) (n int, err error) {
			return 0, expectedErr
		},
	}
	sender := NewSyncSender(writer)
	
	// Send a time but get a write error
	err := sender.SendTime(time.Now())
	
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncSender_SendBytes_Success(t *testing.T) {
	writer := &mockSyncSenderWriter{}
	sender := NewSyncSender(writer)
	
	// Send bytes
	data := []byte("test data")
	err := sender.SendBytes(data)
	
	assert.NoError(t, err)
	
	// Verify the correct bytes were written
	// First 4 bytes should be the length
	expectedLength := make([]byte, 4)
	binary.LittleEndian.PutUint32(expectedLength, uint32(len(data)))
	
	// Then the actual data
	expectedBytes := append(expectedLength, data...)
	assert.Equal(t, expectedBytes, writer.written)
}

func TestSyncSender_SendBytes_TooLarge(t *testing.T) {
	writer := &mockSyncSenderWriter{}
	sender := NewSyncSender(writer)
	
	// Create data larger than SyncMaxChunkSize
	data := make([]byte, SyncMaxChunkSize+1)
	
	// Send bytes
	err := sender.SendBytes(data)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data must be <=")
	assert.Empty(t, writer.written)
}

func TestSyncSender_SendBytes_LengthWriteError(t *testing.T) {
	expectedErr := errors.New("write error")
	writer := &mockSyncSenderWriter{
		writeFunc: func(p []byte) (n int, err error) {
			// Fail on the length write
			if len(p) == 4 {
				return 0, expectedErr
			}
			return len(p), nil
		},
	}
	sender := NewSyncSender(writer)
	
	// Send bytes but get a write error on the length
	err := sender.SendBytes([]byte("test data"))
	
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncSender_SendBytes_DataWriteError(t *testing.T) {
	expectedErr := errors.New("write error")
	writer := &mockSyncSenderWriter{
		writeFunc: func(p []byte) (int, error) {
			// First write (length) succeeds, second write (data) fails
			if len(p) == 4 {
				return len(p), nil
			}
			return 0, expectedErr
		},
	}
	sender := NewSyncSender(writer)
	
	// Send bytes but get a write error on the data
	err := sender.SendBytes([]byte("test data"))
	
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncSender_Close_Success(t *testing.T) {
	writer := &mockSyncSenderWriter{
		closeFunc: func() error {
			return nil
		},
	}
	sender := NewSyncSender(writer)
	
	// Close sender
	err := sender.Close()
	
	assert.NoError(t, err)
}

func TestSyncSender_Close_Error(t *testing.T) {
	expectedErr := errors.New("close error")
	writer := &mockSyncSenderWriter{
		closeFunc: func() error {
			return expectedErr
		},
	}
	sender := NewSyncSender(writer)
	
	// Close sender
	err := sender.Close()
	
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncSender_Close_NotCloser(t *testing.T) {
	// Create a writer that doesn't implement io.Closer
	writer := bytes.NewBuffer([]byte{})
	// Verify it doesn't implement io.Closer
	_, ok := interface{}(writer).(io.Closer)
	assert.False(t, ok)
	sender := NewSyncSender(writer)
	
	// Close sender
	err := sender.Close()
	
	assert.NoError(t, err)
}