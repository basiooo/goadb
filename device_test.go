package adb

import (
	"context"
	"testing"

	"github.com/basiooo/goadb/internal/errors"
	"github.com/basiooo/goadb/wire"
	"github.com/stretchr/testify/assert"
)

func code(err error) errors.ErrCode {
	if err, ok := err.(*errors.Err); ok {
		return err.Code
	}
	return errors.ErrCode(255) // Invalid code
}

func message(err error) string {
	if err, ok := err.(*errors.Err); ok {
		return err.Message
	}
	return err.Error()
}

func TestGetAttribute(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"value"},
	}
	client := (&Adb{s}).Device(DeviceWithSerial("serial"))

	v, err := client.getAttribute("attr")
	assert.Equal(t, "host-serial:serial:attr", s.Requests[0])
	assert.NoError(t, err)
	assert.Equal(t, "value", v)
}

func TestGetDeviceInfo(t *testing.T) {
	deviceLister := func() ([]*DeviceInfo, error) {
		return []*DeviceInfo{
			{
				Serial:  "abc",
				Product: "Foo",
			},
			{
				Serial:  "def",
				Product: "Bar",
			},
		}, nil
	}

	client := newDeviceClientWithDeviceLister("abc", deviceLister)
	device, err := client.DeviceInfo()
	assert.NoError(t, err)
	assert.Equal(t, "Foo", device.Product)

	client = newDeviceClientWithDeviceLister("def", deviceLister)
	device, err = client.DeviceInfo()
	assert.NoError(t, err)
	assert.Equal(t, "Bar", device.Product)

	client = newDeviceClientWithDeviceLister("serial", deviceLister)
	device, err = client.DeviceInfo()
	assert.True(t, HasErrCode(err, DeviceNotFound))
	assert.EqualError(t, err.(*errors.Err).Cause,
		"DeviceNotFound: device list doesn't contain serial serial")
	assert.Nil(t, device)
}

func newDeviceClientWithDeviceLister(serial string, deviceLister func() ([]*DeviceInfo, error)) *Device {
	client := (&Adb{&MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{serial},
	}}).Device(DeviceWithSerial(serial))
	client.deviceListFunc = deviceLister
	return client
}

func TestRunCommandNoArgs(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"output"},
	}
	client := (&Adb{s}).Device(AnyDevice())

	v, err := client.RunCommand("cmd")
	assert.Equal(t, "host:transport-any", s.Requests[0])
	assert.Equal(t, "shell:cmd", s.Requests[1])
	assert.NoError(t, err)
	assert.Equal(t, "output", v)
}

func TestPrepareCommandLineNoArgs(t *testing.T) {
	result, err := prepareCommandLine("cmd")
	assert.NoError(t, err)
	assert.Equal(t, "cmd", result)
}

func TestPrepareCommandLineEmptyCommand(t *testing.T) {
	_, err := prepareCommandLine("")
	assert.Equal(t, errors.AssertionError, code(err))
	assert.Equal(t, "command cannot be empty", message(err))
}

func TestPrepareCommandLineBlankCommand(t *testing.T) {
	_, err := prepareCommandLine("  ")
	assert.Equal(t, errors.AssertionError, code(err))
	assert.Equal(t, "command cannot be empty", message(err))
}

func TestPrepareCommandLineCleanArgs(t *testing.T) {
	result, err := prepareCommandLine("cmd", "arg1", "arg2")
	assert.NoError(t, err)
	assert.Equal(t, "cmd arg1 arg2", result)
}

func TestPrepareCommandLineArgWithWhitespaceQuotes(t *testing.T) {
	result, err := prepareCommandLine("cmd", "arg with spaces")
	assert.NoError(t, err)
	assert.Equal(t, "cmd \"arg with spaces\"", result)
}

func TestPrepareCommandLineArgWithDoubleQuoteFails(t *testing.T) {
	_, err := prepareCommandLine("cmd", "quoted\"arg")
	assert.Equal(t, errors.ParseError, code(err))
	assert.Equal(t, "arg at index 0 contains an invalid double quote: quoted\"arg", message(err))
}

func TestRunCommandWithTimeoutSuccess(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"output"},
	}
	client := (&Adb{s}).Device(AnyDevice())

	v, err := client.RunCommandWithTimeout("cmd", 5)
	assert.Equal(t, "host:transport-any", s.Requests[0])
	assert.Equal(t, "shell:cmd", s.Requests[1])
	assert.NoError(t, err)
	assert.Equal(t, "output", v)
}

func TestRunCommandWithTimeoutTimesOut(t *testing.T) {
	s := &MockServer{
		Status: wire.StatusSuccess,
		Errs:   []error{nil, nil, context.DeadlineExceeded},
	}
	client := (&Adb{s}).Device(AnyDevice())

	timeoutErr := &errors.Err{
		Code:    errors.CommandTimeout,
		Message: "command timed out",
		Cause:   context.DeadlineExceeded,
	}

	s.Errs = []error{nil, nil, timeoutErr}

	_, err := client.RunCommandWithTimeout("cmd", 1)
	assert.Error(t, err)
	assert.True(t, errors.HasErrCode(err, errors.CommandTimeout))
}

func TestRunCommandWithTimeoutWithArgs(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"output with args"},
	}
	client := (&Adb{s}).Device(AnyDevice())

	v, err := client.RunCommandWithTimeout("cmd arg1 arg2", 5)
	assert.Equal(t, "host:transport-any", s.Requests[0])
	assert.Equal(t, "shell:cmd arg1 arg2", s.Requests[1])
	assert.NoError(t, err)
	assert.Equal(t, "output with args", v)
}

func TestRunCommandWithTimeoutPreparationError(t *testing.T) {
	s := &MockServer{
		Status: wire.StatusSuccess,
	}
	client := (&Adb{s}).Device(AnyDevice())

	_, err := client.RunCommandWithTimeout("", 5)
	assert.Error(t, err)
	assert.True(t, errors.HasErrCode(err, errors.AssertionError))
}

func TestRunCommandWithTimeoutDialError(t *testing.T) {
	s := &MockServer{
		Status: wire.StatusSuccess,
		Errs:   []error{errors.Errorf(errors.NetworkError, "dial error")},
	}
	client := (&Adb{s}).Device(AnyDevice())

	_, err := client.RunCommandWithTimeout("cmd", 5)
	assert.Error(t, err)
	assert.True(t, errors.HasErrCode(err, errors.NetworkError))
}

func TestRunCommandWithTimeoutSendError(t *testing.T) {
	s := &MockServer{
		Status: wire.StatusSuccess,
		Errs:   []error{nil, errors.Errorf(errors.NetworkError, "send error")},
	}
	client := (&Adb{s}).Device(AnyDevice())

	_, err := client.RunCommandWithTimeout("cmd", 5)
	assert.Error(t, err)
	assert.True(t, errors.HasErrCode(err, errors.NetworkError))
}

func TestRunCommandWithTimeoutStatusError(t *testing.T) {
	s := &MockServer{
		Status: wire.StatusFailure,
		Errs:   []error{nil, nil, errors.Errorf(errors.AdbError, "status error")},
	}
	client := (&Adb{s}).Device(AnyDevice())

	_, err := client.RunCommandWithTimeout("cmd", 5)
	assert.Error(t, err)
	assert.True(t, errors.HasErrCode(err, errors.AdbError))
}

func TestDevice_ForwardPort(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"OKAY"},
	}
	dev := (&Adb{s}).Device(DeviceWithSerial("serial"))
	err := dev.ForwardPort(12345)
	assert.NoError(t, err)
	assert.Contains(t, s.Requests[1], "host-serial:serial:forward:tcp:12345")
}

func TestDevice_ForwardAbstract(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"OKAY"},
	}
	dev := (&Adb{s}).Device(DeviceWithSerial("serial"))
	err := dev.ForwardAbstract(12345, "testname")
	assert.NoError(t, err)
	assert.Contains(t, s.Requests[1], "host-serial:serial:forward:tcp:12345;localabstract:testname")
}

func TestDevice_ForwardRemovePort(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"OKAY"},
	}
	dev := (&Adb{s}).Device(DeviceWithSerial("serial"))
	err := dev.ForwardRemovePort(12345)
	assert.NoError(t, err)
	assert.Contains(t, s.Requests[1], "host-serial:serial:killforward:tcp:12345")
}

func TestDevice_ForwardRemoveAll(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"OKAY"},
	}
	dev := (&Adb{s}).Device(DeviceWithSerial("serial"))
	err := dev.ForwardRemoveAll()
	assert.NoError(t, err)
	assert.Contains(t, s.Requests[1], "host-serial:serial:killforward-all")
}

func TestDevice_ForwardList(t *testing.T) {
	s := &MockServer{
		Status:   wire.StatusSuccess,
		Messages: []string{"abcd\nserial1 tcp:12345 tcp:54321\nserial2 tcp:2222 tcp:3333\n"},
	}
	dev := (&Adb{s}).Device(DeviceWithSerial("serial"))
	rules, err := dev.ForwardList()
	assert.NoError(t, err)
	assert.Len(t, rules, 2)
	assert.Equal(t, "serial1", rules[0].Serial)
	assert.Equal(t, "tcp:12345", rules[0].Local)
	assert.Equal(t, "tcp:54321", rules[0].Remote)
	assert.Equal(t, "serial2", rules[1].Serial)
}
