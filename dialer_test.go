package adb

import (
	"errors"
	"net"
	"testing"
	"time"

	adbErrors "github.com/basiooo/goadb/internal/errors"
	"github.com/stretchr/testify/assert"
)

// mockNetConn is used below, but we don't need a mockNetDialer since we're
// directly replacing the netDial function in the tests

// mockNetConn is a mock implementation of net.Conn for testing
type mockNetConn struct {
	ReadFunc  func(b []byte) (n int, err error)
	WriteFunc func(b []byte) (n int, err error)
	CloseFunc func() error
}

func (m *mockNetConn) Read(b []byte) (n int, err error)       { return m.ReadFunc(b) }
func (m *mockNetConn) Write(b []byte) (n int, err error)      { return m.WriteFunc(b) }
func (m *mockNetConn) Close() error                           { return m.CloseFunc() }
func (m *mockNetConn) LocalAddr() net.Addr                    { return nil }
func (m *mockNetConn) RemoteAddr() net.Addr                   { return nil }
func (m *mockNetConn) SetDeadline(t time.Time) error         { return nil }
func (m *mockNetConn) SetReadDeadline(t time.Time) error     { return nil }
func (m *mockNetConn) SetWriteDeadline(t time.Time) error    { return nil }

func TestTcpDialer_Dial_Success(t *testing.T) {
	// Create a mock dialer with our own implementation
	originalDial := netDial
	netDial = func(network, address string) (net.Conn, error) {
		assert.Equal(t, "tcp", network)
		assert.Equal(t, "localhost:5037", address)
		
		// Return a mock connection that does nothing
		return &mockNetConn{
			ReadFunc: func(b []byte) (int, error) { return 0, nil },
			WriteFunc: func(b []byte) (int, error) { return len(b), nil },
			CloseFunc: func() error { return nil },
		}, nil
	}
	// Restore the original function after the test
	defer func() { netDial = originalDial }()

	dialer := tcpDialer{}
	conn, err := dialer.Dial("localhost:5037")

	assert.NoError(t, err)
	assert.NotNil(t, conn)
	assert.NotNil(t, conn.Scanner)
	assert.NotNil(t, conn.Sender)
}

func TestTcpDialer_Dial_Error(t *testing.T) {
	// Create a mock dialer with our own implementation
	originalDial := netDial
	netDial = func(network, address string) (net.Conn, error) {
		return nil, errors.New("connection refused")
	}
	// Restore the original function after the test
	defer func() { netDial = originalDial }()

	dialer := tcpDialer{}
	conn, err := dialer.Dial("localhost:5037")

	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.ServerNotAvailable))
}