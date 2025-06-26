package wire

import (
	"errors"
	"testing"

	adbErrors "github.com/basiooo/goadb/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestAdbServerError(t *testing.T) {
	// Test with empty request
	err1 := adbServerError("", "some error")
	assert.Error(t, err1)
	assert.Contains(t, err1.Error(), "server error: some error")
	assert.True(t, adbErrors.HasErrCode(err1, adbErrors.AdbError))
	
	// Test with request
	err2 := adbServerError("test-request", "some error")
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "server error for test-request request: some error")
	assert.True(t, adbErrors.HasErrCode(err2, adbErrors.AdbError))
	
	// Test with device not found message
	err3 := adbServerError("test-request", "device not found")
	assert.Error(t, err3)
	assert.True(t, adbErrors.HasErrCode(err3, adbErrors.DeviceNotFound))
	
	// Test with device not found message with serial
	err4 := adbServerError("test-request", "device 'serial123' not found")
	assert.Error(t, err4)
	assert.True(t, adbErrors.HasErrCode(err4, adbErrors.DeviceNotFound))
}

func TestIsAdbServerErrorMatching(t *testing.T) {
	// Test with matching AdbError
	err1 := adbServerError("test-request", "some error")
	matched1 := IsAdbServerErrorMatching(err1, func(msg string) bool {
		return msg == "some error"
	})
	assert.True(t, matched1)
	
	// Test with non-matching AdbError
	matched2 := IsAdbServerErrorMatching(err1, func(msg string) bool {
		return msg == "different error"
	})
	assert.False(t, matched2)
	
	// Test with DeviceNotFound error (should return false since it's not AdbError)
	err3 := adbServerError("test-request", "device not found")
	matched3 := IsAdbServerErrorMatching(err3, func(msg string) bool {
		return msg == "device not found"
	})
	assert.False(t, matched3)
	
	// Test with non-Err error
	err4 := errors.New("some other error")
	matched4 := IsAdbServerErrorMatching(err4, func(msg string) bool {
		return true
	})
	assert.False(t, matched4)
}

func TestErrIncompleteMessage(t *testing.T) {
	err := errIncompleteMessage("test message", 5, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incomplete test message: read 5 bytes, expecting 10")
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.ConnectionResetError))
	
	// Check the details
	errObj, ok := err.(*adbErrors.Err)
	assert.True(t, ok)
	details, ok := errObj.Details.(struct {
		ActualReadBytes int
		ExpectedBytes   int
	})
	assert.True(t, ok)
	assert.Equal(t, 5, details.ActualReadBytes)
	assert.Equal(t, 10, details.ExpectedBytes)
}