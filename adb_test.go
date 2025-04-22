package adb

import (
	"testing"

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
