package adb

import (
	"io"
	"log"
	"net"
	"runtime"

	"github.com/basiooo/goadb/internal/errors"
	"github.com/basiooo/goadb/wire"
)

// Variable to allow mocking in tests
var netDial = net.Dial

// Dialer knows how to create connections to an adb server.
type Dialer interface {
	Dial(address string) (*wire.Conn, error)
}

type tcpDialer struct{}

// Dial connects to the adb server on the host and port set on the netDialer.
// The zero-value will connect to the default, localhost:5037.
func (tcpDialer) Dial(address string) (*wire.Conn, error) {
	netConn, err := netDial("tcp", address)
	if err != nil {
		return nil, errors.WrapErrorf(err, errors.ServerNotAvailable, "error dialing %s", address)
	}

	// net.Conn can't be closed more than once, but wire.Conn will try to close both sender and scanner
	// so we need to wrap it to make it safe.
	safeConn := wire.MultiCloseable(netConn)

	// Prevent leaking the network connection, not sure if TCPConn does this itself.
	// Note that the network connection may still be in use after the conn isn't (scanners/senders
	// can give their underlying connections to other scanner/sender types), so we can't
	// set the finalizer on conn.
	runtime.SetFinalizer(safeConn, func(conn io.ReadWriteCloser) {
		if err := conn.Close(); err != nil {
			log.Printf("[Dialer] error closing connection: %s", err)
		}
	})

	return &wire.Conn{
		Scanner: wire.NewScanner(safeConn),
		Sender:  wire.NewSender(safeConn),
	}, nil
}
