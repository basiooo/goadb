package wire

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	adbErrors "github.com/basiooo/goadb/internal/errors"
	"github.com/stretchr/testify/assert"
)

// mockSyncConnScanner is a mock implementation of SyncScanner for testing
type mockSyncConnScanner struct {
	closeFunc func() error
}

func (m *mockSyncConnScanner) ReadStatus(expectedStatus string) (string, error) {
	return "", nil
}

func (m *mockSyncConnScanner) ReadInt32() (int32, error) {
	return 0, nil
}

func (m *mockSyncConnScanner) ReadFileMode() (os.FileMode, error) {
	return 0, nil
}

func (m *mockSyncConnScanner) ReadTime() (time.Time, error) {
	return time.Time{}, nil
}

func (m *mockSyncConnScanner) ReadString() (string, error) {
	return "", nil
}

func (m *mockSyncConnScanner) ReadBytes() (io.Reader, error) {
	return bytes.NewReader(nil), nil
}

func (m *mockSyncConnScanner) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// mockSyncConnSender is a mock implementation of SyncSender for testing
type mockSyncConnSender struct {
	closeFunc func() error
}

func (m *mockSyncConnSender) SendOctetString(s string) error {
	return nil
}

func (m *mockSyncConnSender) SendInt32(n int32) error {
	return nil
}

func (m *mockSyncConnSender) SendFileMode(mode os.FileMode) error {
	return nil
}

func (m *mockSyncConnSender) SendTime(t time.Time) error {
	return nil
}

func (m *mockSyncConnSender) SendBytes(data []byte) error {
	return nil
}

func (m *mockSyncConnSender) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestSyncConn_Close_NoErrors(t *testing.T) {
	// Create mocks
	scanner := &mockSyncConnScanner{
		closeFunc: func() error {
			return nil
		},
	}
	
	sender := &mockSyncConnSender{
		closeFunc: func() error {
			return nil
		},
	}
	
	syncConn := &SyncConn{scanner, sender}
	
	// Close should not return error
	err := syncConn.Close()
	assert.NoError(t, err)
}

func TestSyncConn_Close_ScannerError(t *testing.T) {
	// Create mocks
	scannerErr := errors.New("scanner close error")
	
	scanner := &mockSyncConnScanner{
		closeFunc: func() error {
			return scannerErr
		},
	}
	
	sender := &mockSyncConnSender{
		closeFunc: func() error {
			return nil
		},
	}
	
	syncConn := &SyncConn{scanner, sender}
	
	// Close should return error
	err := syncConn.Close()
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncConn_Close_SenderError(t *testing.T) {
	// Create mocks
	senderErr := errors.New("sender close error")
	
	scanner := &mockSyncConnScanner{
		closeFunc: func() error {
			return nil
		},
	}
	
	sender := &mockSyncConnSender{
		closeFunc: func() error {
			return senderErr
		},
	}
	
	syncConn := &SyncConn{scanner, sender}
	
	// Close should return error
	err := syncConn.Close()
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncConn_Close_BothErrors(t *testing.T) {
	// Create mocks
	scannerErr := errors.New("scanner close error")
	senderErr := errors.New("sender close error")
	
	scanner := &mockSyncConnScanner{
		closeFunc: func() error {
			return scannerErr
		},
	}
	
	sender := &mockSyncConnSender{
		closeFunc: func() error {
			return senderErr
		},
	}
	
	syncConn := &SyncConn{scanner, sender}
	
	// Close should return error
	err := syncConn.Close()
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}