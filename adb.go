package adb

import (
	"fmt"
	"strconv"

	"github.com/basiooo/goadb/internal/errors"
	"github.com/basiooo/goadb/wire"
)

/*
Adb communicates with host services on the adb server.

Eg.
	client := adb.New()
	client.ListDevices()

See list of services at https://android.googlesource.com/platform/system/core/+/master/adb/SERVICES.TXT.
*/
// TODO(z): Finish implementing host services.
type Adb struct {
	Server server
}

// New creates a new Adb client that uses the default ServerConfig.
func New() (*Adb, error) {
	return NewWithConfig(ServerConfig{})
}

func NewWithConfig(config ServerConfig) (*Adb, error) {
	server, err := newServer(config)
	if err != nil {
		return nil, err
	}
	return &Adb{server}, nil
}

// Dial establishes a connection with the adb server.
func (c *Adb) Dial() (*wire.Conn, error) {
	return c.Server.Dial()
}

// Starts the adb server if it’s not running.
func (c *Adb) StartServer() error {
	return c.Server.Start()
}

func (c *Adb) Device(descriptor DeviceDescriptor) *Device {
	return &Device{
		server:         c.Server,
		descriptor:     descriptor,
		deviceListFunc: c.ListDevices,
	}
}

func (c *Adb) NewDeviceWatcher() *DeviceWatcher {
	return newDeviceWatcher(c.Server)
}

// ServerVersion asks the ADB server for its internal version number.
func (c *Adb) ServerVersion() (int, error) {
	resp, err := roundTripSingleResponse(c.Server, "host:version")
	if err != nil {
		return 0, wrapClientError(err, c, "GetServerVersion")
	}

	version, err := c.parseServerVersion(resp)
	if err != nil {
		return 0, wrapClientError(err, c, "GetServerVersion")
	}
	return version, nil
}

/*
KillServer tells the server to quit immediately.

Corresponds to the command:

	adb kill-server
*/
func (c *Adb) KillServer() error {
	conn, err := c.Server.Dial()
	if err != nil {
		return wrapClientError(err, c, "KillServer")
	}
	defer conn.Close()

	if err = wire.SendMessageString(conn, "host:kill"); err != nil {
		return wrapClientError(err, c, "KillServer")
	}

	return nil
}

/*
ListDeviceSerials returns the serial numbers of all attached devices.

Corresponds to the command:

	adb devices
*/
func (c *Adb) ListDeviceSerials() ([]string, error) {
	resp, err := roundTripSingleResponse(c.Server, "host:devices")
	if err != nil {
		return nil, wrapClientError(err, c, "ListDeviceSerials")
	}

	devices, err := parseDeviceList(string(resp), parseDeviceShort)
	if err != nil {
		return nil, wrapClientError(err, c, "ListDeviceSerials")
	}

	serials := make([]string, len(devices))
	for i, dev := range devices {
		serials[i] = dev.Serial
	}
	return serials, nil
}

/*
ListDevices returns the list of connected devices.

Corresponds to the command:

	adb devices -l
*/
func (c *Adb) ListDevices() ([]*DeviceInfo, error) {
	resp, err := roundTripSingleResponse(c.Server, "host:devices-l")
	if err != nil {
		return nil, wrapClientError(err, c, "ListDevices")
	}

	devices, err := parseDeviceList(string(resp), parseDeviceLong)
	if err != nil {
		return nil, wrapClientError(err, c, "ListDevices")
	}
	return devices, nil
}

/*
Connect connect to a device via TCP/IP

Corresponds to the command:

	adb connect
*/
func (c *Adb) Connect(host string, port int) error {
	_, err := roundTripSingleResponse(c.Server, fmt.Sprintf("host:connect:%s:%d", host, port))
	if err != nil {
		return wrapClientError(err, c, "Connect")
	}
	return nil
}

func (c *Adb) DisconnectAll() error {
	_, err := roundTripSingleResponse(c.Server, "host:disconnect:")
	if err != nil {
		return fmt.Errorf("disconnect: %w", err)
	}
	return nil
}

func (c *Adb) Disconnect(addr string) error {
	_, err := roundTripSingleResponse(c.Server, fmt.Sprintf("host:disconnect:%s", addr))
	if err != nil {
		return fmt.Errorf("disconnect: %w", err)
	}
	return nil
}

func (c *Adb) parseServerVersion(versionRaw []byte) (int, error) {
	versionStr := string(versionRaw)
	version, err := strconv.ParseInt(versionStr, 16, 32)
	if err != nil {
		return 0, errors.WrapErrorf(err, errors.ParseError,
			"error parsing server version: %s", versionStr)
	}
	return int(version), nil
}

// Get device by serial
func (c *Adb) GetDeviceBySerial(serial string) (*Device, error) {
	deviceDescriptor := DeviceWithSerial(serial)
	device := c.Device(deviceDescriptor)
	if _, err := device.Serial(); err != nil {
		return nil, errors.Errorf(errors.DeviceNotFound, "%s", fmt.Sprintf("Device with serial %s not found", serial))
	}
	return device, nil
}
