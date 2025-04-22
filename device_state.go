package adb

import "github.com/basiooo/goadb/internal/errors"

// DeviceState represents one of the 3 possible states adb will report devices.
// A device can be communicated with when it's in StateOnline.
// A USB device will make the following state transitions:
//
//	Plugged in: StateDisconnected->StateOffline->StateOnline
//	Unplugged:  StateOnline->StateDisconnected
//
//go:generate stringer -type=DeviceState
type DeviceState int8

const (
	StateInvalid DeviceState = iota
	StateUnauthorized
	StateDisconnected
	StateOffline
	StateOnline
	StateAuthorizing
	StateRecovery
)

func (s DeviceState) String() string {
	switch DeviceState(s) {
	case StateDisconnected:
		return "Disconnected"
	case StateOffline:
		return "Offline"
	case StateOnline:
		return "Online"
	case StateUnauthorized:
		return "Unauthorized"
	case StateAuthorizing:
		return "Authorizing"
	case StateRecovery:
		return "Recovery"
	case StateInvalid:
		return "Invalid"
	default:
		return "Unknown"
	}
}

var deviceStateStrings = map[string]DeviceState{
	"":             StateDisconnected,
	"offline":      StateOffline,
	"device":       StateOnline,
	"unauthorized": StateUnauthorized,
	"authorizing":  StateAuthorizing,
	"recovery":     StateRecovery,
}

func parseDeviceState(str string) (DeviceState, error) {
	state, ok := deviceStateStrings[str]
	if !ok {
		return StateInvalid, errors.Errorf(errors.ParseError, "invalid device state: %q", state)
	}
	return state, nil
}
