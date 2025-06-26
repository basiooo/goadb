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

// mockScanner is a mock implementation of Scanner for testing
type mockConnScanner struct {
	readStatusFunc    func(expectedStatus string) (string, error)
	readMessageFunc   func() ([]byte, error)
	readUntilEofFunc  func() ([]byte, error)
	newSyncScannerFunc func() SyncScanner
	closeFunc         func() error
}

func (m *mockConnScanner) ReadStatus(expectedStatus string) (string, error) {
	return m.readStatusFunc(expectedStatus)
}

func (m *mockConnScanner) ReadMessage() ([]byte, error) {
	return m.readMessageFunc()
}

func (m *mockConnScanner) ReadUntilEof() ([]byte, error) {
	return m.readUntilEofFunc()
}

func (m *mockConnScanner) NewSyncScanner() SyncScanner {
	return m.newSyncScannerFunc()
}

func (m *mockConnScanner) Close() error {
	return m.closeFunc()
}

// mockSender is a mock implementation of Sender for testing
type mockConnSender struct {
	sendMessageStringFunc func(msg string) error
	sendMessageFunc       func(msg []byte) error
	newSyncSenderFunc     func() SyncSender
	closeFunc             func() error
}

func (m *mockConnSender) SendMessageString(msg string) error {
	return m.sendMessageStringFunc(msg)
}

func (m *mockConnSender) SendMessage(msg []byte) error {
	return m.sendMessageFunc(msg)
}

func (m *mockConnSender) NewSyncSender() SyncSender {
	return m.newSyncSenderFunc()
}

func (m *mockConnSender) Close() error {
	return m.closeFunc()
}

// mockSyncScanner is a mock implementation of SyncScanner for testing
type mockConnSyncScanner struct{}

func (m *mockConnSyncScanner) ReadStatus(expectedStatus string) (string, error) {
	return "", nil
}

func (m *mockConnSyncScanner) ReadInt32() (int32, error) {
	return 0, nil
}

func (m *mockConnSyncScanner) ReadFileMode() (os.FileMode, error) {
	return 0, nil
}

func (m *mockConnSyncScanner) ReadTime() (time.Time, error) {
	return time.Time{}, nil
}

func (m *mockConnSyncScanner) ReadString() (string, error) {
	return "", nil
}

func (m *mockConnSyncScanner) ReadBytes() (io.Reader, error) {
	return bytes.NewReader(nil), nil
}

func (m *mockConnSyncScanner) Close() error {
	return nil
}

// mockSyncSender is a mock implementation of SyncSender for testing
type mockConnSyncSender struct{}

func (m *mockConnSyncSender) SendOctetString(s string) error {
	return nil
}

func (m *mockConnSyncSender) SendInt32(n int32) error {
	return nil
}

func (m *mockConnSyncSender) SendFileMode(mode os.FileMode) error {
	return nil
}

func (m *mockConnSyncSender) SendTime(t time.Time) error {
	return nil
}

func (m *mockConnSyncSender) SendBytes(data []byte) error {
	return nil
}

func (m *mockConnSyncSender) Close() error {
	return nil
}

func TestNewConn(t *testing.T) {
	scanner := &mockConnScanner{}
	sender := &mockConnSender{}
	
	conn := NewConn(scanner, sender)
	
	assert.Equal(t, scanner, conn.Scanner)
	assert.Equal(t, sender, conn.Sender)
}

func TestConn_NewSyncConn(t *testing.T) {
	// Create mocks
	syncScanner := &mockConnSyncScanner{}
	syncSender := &mockConnSyncSender{}
	
	scanner := &mockConnScanner{
		newSyncScannerFunc: func() SyncScanner {
			return syncScanner
		},
	}
	
	sender := &mockConnSender{
		newSyncSenderFunc: func() SyncSender {
			return syncSender
		},
	}
	
	conn := &Conn{scanner, sender}
	
	// Create sync conn
	syncConn := conn.NewSyncConn()
	
	assert.Equal(t, syncScanner, syncConn.SyncScanner)
	assert.Equal(t, syncSender, syncConn.SyncSender)
}

func TestConn_RoundTripSingleResponse_Success(t *testing.T) {
	// Create mocks
	scanner := &mockConnScanner{
		readStatusFunc: func(expectedStatus string) (string, error) {
			return StatusSuccess, nil
		},
		readMessageFunc: func() ([]byte, error) {
			return []byte("response data"), nil
		},
	}
	
	sender := &mockConnSender{
		sendMessageFunc: func(msg []byte) error {
			assert.Equal(t, []byte("request data"), msg)
			return nil
		},
	}
	
	conn := &Conn{scanner, sender}
	
	// Send request and get response
	resp, err := conn.RoundTripSingleResponse([]byte("request data"))
	
	assert.NoError(t, err)
	assert.Equal(t, []byte("response data"), resp)
}

func TestConn_RoundTripSingleResponse_SendError(t *testing.T) {
	// Create mocks
	sendErr := errors.New("send error")
	
	sender := &mockConnSender{
		sendMessageFunc: func(msg []byte) error {
			return sendErr
		},
	}
	
	conn := &Conn{nil, sender}
	
	// Send request and get error
	resp, err := conn.RoundTripSingleResponse([]byte("request data"))
	
	assert.Error(t, err)
	assert.Equal(t, sendErr, err)
	assert.Nil(t, resp)
}

func TestConn_RoundTripSingleResponse_StatusError(t *testing.T) {
	// Create mocks
	statusErr := errors.New("status error")
	
	scanner := &mockConnScanner{
		readStatusFunc: func(expectedStatus string) (string, error) {
			return "", statusErr
		},
	}
	
	sender := &mockConnSender{
		sendMessageFunc: func(msg []byte) error {
			return nil
		},
	}
	
	conn := &Conn{scanner, sender}
	
	// Send request and get error
	resp, err := conn.RoundTripSingleResponse([]byte("request data"))
	
	assert.Error(t, err)
	assert.Equal(t, statusErr, err)
	assert.Nil(t, resp)
}

func TestConn_Close_NoErrors(t *testing.T) {
	// Create mocks
	scanner := &mockConnScanner{
		closeFunc: func() error {
			return nil
		},
	}
	
	sender := &mockConnSender{
		closeFunc: func() error {
			return nil
		},
	}
	
	conn := &Conn{scanner, sender}
	
	// Close should not return error
	err := conn.Close()
	assert.NoError(t, err)
}

func TestConn_Close_WithErrors(t *testing.T) {
	// Create mocks
	scannerErr := errors.New("scanner close error")
	senderErr := errors.New("sender close error")
	
	scanner := &mockConnScanner{
		closeFunc: func() error {
			return scannerErr
		},
	}
	
	sender := &mockConnSender{
		closeFunc: func() error {
			return senderErr
		},
	}
	
	conn := &Conn{scanner, sender}
	
	// Close should return error
	err := conn.Close()
	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
	
	// Error should contain both underlying errors
	errObj, ok := err.(*adbErrors.Err)
	assert.True(t, ok)
	details, ok := errObj.Details.(struct {
		SenderErr  error
		ScannerErr error
	})
	assert.True(t, ok)
	assert.Equal(t, senderErr, details.SenderErr)
	assert.Equal(t, scannerErr, details.ScannerErr)
}