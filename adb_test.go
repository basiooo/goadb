package adb

import (
	"testing"

	"github.com/basiooo/goadb/internal/errors"
	"github.com/basiooo/goadb/wire"
	"github.com/stretchr/testify/assert"
)

func TestGetServerVersion(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"000a"},
	}
	client := &Adb{s}

	v, err := client.ServerVersion()
	assert.Equal(t, "host:version", s.Requests[0])
	assert.NoError(t, err)
	assert.Equal(t, 10, v)
}

func TestDisconnectAll(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{""},
	}
	client := &Adb{s}

	err := client.DisconnectAll()
	assert.Equal(t, "host:disconnect:", s.Requests[0])
	assert.NoError(t, err)
}

func TestDisconnect(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{""},
	}
	client := &Adb{s}

	err := client.Disconnect("123456")
	assert.Equal(t, "host:disconnect:123456", s.Requests[0])
	assert.NoError(t, err)
}

func TestKillServer(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{""},
	}
	client := &Adb{s}

	err := client.KillServer()
	assert.Equal(t, "host:kill", s.Requests[0])
	assert.NoError(t, err)
}

func TestListDeviceSerials(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"device1\tdevice\ndevice2\toffline\n"},
	}
	client := &Adb{s}

	serials, err := client.ListDeviceSerials()
	assert.Equal(t, "host:devices", s.Requests[0])
	assert.NoError(t, err)
	assert.Equal(t, []string{"device1", "device2"}, serials)
}

func TestListDevices(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"device1\tdevice\tproduct:p1 model:m1 device:d1\ndevice2\toffline\tproduct:p2 model:m2 device:d2\n"},
	}
	client := &Adb{s}

	devices, err := client.ListDevices()
	assert.Equal(t, "host:devices-l", s.Requests[0])
	assert.NoError(t, err)
	assert.Equal(t, 2, len(devices))
	assert.Equal(t, "device1", devices[0].Serial)
	assert.Equal(t, "p1", devices[0].Product)
	assert.Equal(t, "m1", devices[0].Model)
	assert.Equal(t, "d1", devices[0].DeviceInfo)
	assert.Equal(t, "device2", devices[1].Serial)
	assert.Equal(t, "p2", devices[1].Product)
	assert.Equal(t, "m2", devices[1].Model)
	assert.Equal(t, "d2", devices[1].DeviceInfo)
}

func TestConnect(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{""},
	}
	client := &Adb{s}

	err := client.Connect("192.168.1.100", 5555)
	assert.Equal(t, "host:connect:192.168.1.100:5555", s.Requests[0])
	assert.NoError(t, err)
}

func TestGetDeviceBySerial_Success(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"serial123"}, // Response for Serial() call
	}
	client := &Adb{s}

	device, err := client.GetDeviceBySerial("serial123")
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "host-serial:serial123:get-serialno", s.Requests[0])
}

func TestGetDeviceBySerial_DeviceNotFound(t *testing.T) {
	s := &MockServer{
		Errs: []error{errors.Errorf(errors.DeviceNotFound, "device not found")},
	}
	client := &Adb{s}

	device, err := client.GetDeviceBySerial("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, device)
	assert.True(t, errors.HasErrCode(err, errors.DeviceNotFound))
}

func TestParseServerVersion(t *testing.T) {
	client := &Adb{}
	
	// Test valid version
	version, err := client.parseServerVersion([]byte("001e"))
	assert.NoError(t, err)
	assert.Equal(t, 30, version)
	
	// Test another valid version
	version, err = client.parseServerVersion([]byte("0032"))
	assert.NoError(t, err)
	assert.Equal(t, 50, version)
	
	// Test invalid version
	_, err = client.parseServerVersion([]byte("xyz"))
	assert.Error(t, err)
	assert.True(t, errors.HasErrCode(err, errors.ParseError))
}
