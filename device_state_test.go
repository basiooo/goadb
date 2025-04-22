package adb

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDeviceState(t *testing.T) {
	for _, test := range []struct {
		String    string
		WantState DeviceState
		WantName  string
		WantError error // Compared by Error() message.
	}{
		{"", StateDisconnected, "Disconnected", nil},
		{"offline", StateOffline, "Offline", nil},
		{"device", StateOnline, "Online", nil},
		{"unauthorized", StateUnauthorized, "Unauthorized", nil},
		{"bad", StateInvalid, "Invalid", errors.New(`ParseError: invalid device state: "Invalid"`)},
	} {
		state, err := parseDeviceState(test.String)
		if test.WantError == nil {
			assert.NoError(t, err)
		} else {
			assert.EqualError(t, err, test.WantError.Error())
		}
		assert.Equal(t, test.WantState, state)
		assert.Equal(t, test.WantName, state.String())
	}
}
