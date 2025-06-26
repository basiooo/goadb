package adb

import (
	"context"
	"testing"
	"time"

	"github.com/basiooo/goadb/internal/errors"
	"github.com/basiooo/goadb/wire"
	"github.com/stretchr/testify/assert"
)

// MockScanner implements wire.Scanner for testing
type MockScanner struct {
	Messages [][]byte
	Errs     []error
	index    int
}

func (s *MockScanner) ReadMessage() ([]byte, error) {
	if s.index < len(s.Errs) && s.Errs[s.index] != nil {
		err := s.Errs[s.index]
		s.index++
		return nil, err
	}
	
	if s.index < len(s.Messages) {
		msg := s.Messages[s.index]
		s.index++
		return msg, nil
	}
	
	// Return EOF instead of blocking forever
	return nil, errors.Errorf(errors.ConnectionResetError, "end of messages")
}

func (s *MockScanner) ReadStatus(cmd string) (string, error) {
	return "OKAY", nil
}

func (s *MockScanner) Close() error {
	return nil
}

func (s *MockScanner) ReadUntilEof() ([]byte, error) {
	return nil, nil
}

func (s *MockScanner) NewSyncScanner() wire.SyncScanner {
	return nil
}

func TestParseDeviceStatesSingle(t *testing.T) {
	states, err := parseDeviceStates(`192.168.56.101:5555	offline
`)

	assert.NoError(t, err)
	assert.Len(t, states, 1)
	assert.Equal(t, StateOffline, states["192.168.56.101:5555"])
}

func TestParseDeviceStatesMultiple(t *testing.T) {
	states, err := parseDeviceStates(`192.168.56.101:5555	offline
0x0x0x0x	device
`)

	assert.NoError(t, err)
	assert.Len(t, states, 2)
	assert.Equal(t, StateOffline, states["192.168.56.101:5555"])
	assert.Equal(t, StateOnline, states["0x0x0x0x"])
}

func TestParseDeviceStatesMalformed(t *testing.T) {
	_, err := parseDeviceStates(`192.168.56.101:5555	offline
0x0x0x0x
`)

	assert.True(t, HasErrCode(err, ParseError))
	assert.Equal(t, "invalid device state line 1: 0x0x0x0x", err.(*errors.Err).Message)
}

func TestCalculateStateDiffsUnchangedEmpty(t *testing.T) {
	oldStates := map[string]DeviceState{}
	newStates := map[string]DeviceState{}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Empty(t, diffs)
}

func TestCalculateStateDiffsUnchangedNonEmpty(t *testing.T) {
	oldStates := map[string]DeviceState{
		"1": StateOnline,
		"2": StateOnline,
	}
	newStates := map[string]DeviceState{
		"1": StateOnline,
		"2": StateOnline,
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assert.Empty(t, diffs)
}

func TestCalculateStateDiffsOneAdded(t *testing.T) {
	oldStates := map[string]DeviceState{}
	newStates := map[string]DeviceState{
		"serial": StateOffline,
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assertContainsOnly(t, []DeviceStateChangedEvent{
		{"serial", StateDisconnected, StateOffline},
	}, diffs)
}

func TestCalculateStateDiffsOneRemoved(t *testing.T) {
	oldStates := map[string]DeviceState{
		"serial": StateOffline,
	}
	newStates := map[string]DeviceState{}

	diffs := calculateStateDiffs(oldStates, newStates)

	assertContainsOnly(t, []DeviceStateChangedEvent{
		{"serial", StateOffline, StateDisconnected},
	}, diffs)
}

func TestCalculateStateDiffsOneAddedOneUnchanged(t *testing.T) {
	oldStates := map[string]DeviceState{
		"1": StateOnline,
	}
	newStates := map[string]DeviceState{
		"1": StateOnline,
		"2": StateOffline,
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assertContainsOnly(t, []DeviceStateChangedEvent{
		{"2", StateDisconnected, StateOffline},
	}, diffs)
}

func TestCalculateStateDiffsOneRemovedOneUnchanged(t *testing.T) {
	oldStates := map[string]DeviceState{
		"1": StateOffline,
		"2": StateOnline,
	}
	newStates := map[string]DeviceState{
		"2": StateOnline,
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assertContainsOnly(t, []DeviceStateChangedEvent{
		{"1", StateOffline, StateDisconnected},
	}, diffs)
}

func TestCalculateStateDiffsOneAddedOneRemoved(t *testing.T) {
	oldStates := map[string]DeviceState{
		"1": StateOffline,
	}
	newStates := map[string]DeviceState{
		"2": StateOffline,
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assertContainsOnly(t, []DeviceStateChangedEvent{
		{"1", StateOffline, StateDisconnected},
		{"2", StateDisconnected, StateOffline},
	}, diffs)
}

func TestCalculateStateDiffsOneChangedOneUnchanged(t *testing.T) {
	oldStates := map[string]DeviceState{
		"1": StateOffline,
		"2": StateOnline,
	}
	newStates := map[string]DeviceState{
		"1": StateOnline,
		"2": StateOnline,
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assertContainsOnly(t, []DeviceStateChangedEvent{
		{"1", StateOffline, StateOnline},
	}, diffs)
}

func TestCalculateStateDiffsMultipleChanged(t *testing.T) {
	oldStates := map[string]DeviceState{
		"1": StateOffline,
		"2": StateOnline,
	}
	newStates := map[string]DeviceState{
		"1": StateOnline,
		"2": StateOffline,
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assertContainsOnly(t, []DeviceStateChangedEvent{
		{"1", StateOffline, StateOnline},
		{"2", StateOnline, StateOffline},
	}, diffs)
}

func TestCalculateStateDiffsOneAddedOneRemovedOneChanged(t *testing.T) {
	oldStates := map[string]DeviceState{
		"1": StateOffline,
		"2": StateOffline,
	}
	newStates := map[string]DeviceState{
		"1": StateOnline,
		"3": StateOffline,
	}

	diffs := calculateStateDiffs(oldStates, newStates)

	assertContainsOnly(t, []DeviceStateChangedEvent{
		{"1", StateOffline, StateOnline},
		{"2", StateOffline, StateDisconnected},
		{"3", StateDisconnected, StateOffline},
	}, diffs)
}

func TestCameOnline(t *testing.T) {
	assert.True(t, DeviceStateChangedEvent{"", StateDisconnected, StateOnline}.CameOnline())
	assert.True(t, DeviceStateChangedEvent{"", StateOffline, StateOnline}.CameOnline())
	assert.False(t, DeviceStateChangedEvent{"", StateOnline, StateOffline}.CameOnline())
	assert.False(t, DeviceStateChangedEvent{"", StateOnline, StateDisconnected}.CameOnline())
	assert.False(t, DeviceStateChangedEvent{"", StateOffline, StateDisconnected}.CameOnline())
}

func TestWentOffline(t *testing.T) {
	assert.True(t, DeviceStateChangedEvent{"", StateOnline, StateDisconnected}.WentOffline())
	assert.True(t, DeviceStateChangedEvent{"", StateOnline, StateOffline}.WentOffline())
	assert.False(t, DeviceStateChangedEvent{"", StateOffline, StateOnline}.WentOffline())
	assert.False(t, DeviceStateChangedEvent{"", StateDisconnected, StateOnline}.WentOffline())
	assert.False(t, DeviceStateChangedEvent{"", StateOffline, StateDisconnected}.WentOffline())
}

func TestPublishDevicesRestartsServer(t *testing.T) {
	ctx, ctxCancelFunc := context.WithTimeout(context.Background(), time.Second)
	server := &MockServer{
		Status: wire.StatusSuccess,
		Errs: []error{
			nil, nil, nil, // Successful dial.
			errors.Errorf(errors.ConnectionResetError, "failed first read"),
			errors.Errorf(errors.ServerNotAvailable, "failed redial"),
		},
	}
	watcher := deviceWatcherImpl{
		server:        server,
		eventChan:     make(chan DeviceStateChangedEvent),
		ctx:           ctx,
		ctxCancelFunc: ctxCancelFunc,
	}
	publishDevices(&watcher)

	assert.Empty(t, server.Errs)
	assert.Contains(t, server.Requests, "host:track-devices")
	assert.Subset(t, server.Trace, []string{"Dial", "SendMessage", "ReadStatus", "ReadMessage", "Start", "Dial"})
}

func TestConnectToTrackDevices_Success(t *testing.T) {
	server := &MockServer{
		Status: wire.StatusSuccess,
	}

	scanner, err := connectToTrackDevices(server)
	assert.NoError(t, err)
	assert.NotNil(t, scanner)
	assert.Equal(t, "host:track-devices", server.Requests[0])
}

func TestConnectToTrackDevices_DialError(t *testing.T) {
	server := &MockServer{
		Errs: []error{errors.Errorf(errors.ServerNotAvailable, "server not available")},
	}

	scanner, err := connectToTrackDevices(server)
	assert.Error(t, err)
	assert.Nil(t, scanner)
	assert.True(t, errors.HasErrCode(err, errors.ServerNotAvailable))
}

func TestConnectToTrackDevices_SendError(t *testing.T) {
	server := &MockServer{
		Status: wire.StatusSuccess,
		Errs: []error{
			nil, // Successful dial
			errors.Errorf(errors.ConnectionResetError, "connection reset"),
		},
	}

	scanner, err := connectToTrackDevices(server)
	assert.Error(t, err)
	assert.Nil(t, scanner)
	assert.True(t, errors.HasErrCode(err, errors.ConnectionResetError))
}

func TestDeviceWatcher_C(t *testing.T) {
	server := &MockServer{
		Status: wire.StatusSuccess,
	}
	watcher := newDeviceWatcher(server)

	channel := watcher.C()
	assert.NotNil(t, channel)
	// Clean up
	watcher.Shutdown()
}

func TestDeviceWatcher_Err(t *testing.T) {
	server := &MockServer{
		Status: wire.StatusSuccess,
	}
	watcher := newDeviceWatcher(server)

	err := watcher.Err()
	assert.Nil(t, err)

	// Manually set an error
	testErr := errors.Errorf(errors.ConnectionResetError, "test error")
	watcher.reportErr(testErr)

	err = watcher.Err()
	assert.NotNil(t, err)
	assert.True(t, errors.HasErrCode(err, errors.ConnectionResetError))
	// Clean upß
	watcher.Shutdown()
}

func TestDeviceWatcher_Shutdown(t *testing.T) {
	server := &MockServer{
		Status: wire.StatusSuccess,
	}
	watcher := newDeviceWatcher(server)

	// Get the channel before shutdown
	channel := watcher.C()

	// Shutdown should close the channel
	watcher.Shutdown()

	// Verify channel is closed by trying to receive from it
	_, ok := <-channel
	assert.False(t, ok, "Channel should be closed after Shutdown")
}

func TestPublishDevicesUntilError_Success(t *testing.T) {
	// Create a mock scanner that returns device states and then an error
	scanner := &MockScanner{
		Messages: [][]byte{[]byte("device1\tdevice\ndevice2\toffline\n")},
		Errs:     []error{nil, errors.Errorf(errors.ConnectionResetError, "end of messages")},
	}
	
	// Create a device watcher
	ctx, cancel := context.WithCancel(context.Background())
	watcher := &deviceWatcherImpl{
		ctx:       ctx,
		eventChan: make(chan DeviceStateChangedEvent, 10),
	}
	
	// Initial empty state
	lastKnownStates := make(map[string]DeviceState)
	
	// Run the function
	finished, err := publishDevicesUntilError(scanner, watcher, &lastKnownStates)
	
	// Clean up
	cancel()
	
	// Verify results
	assert.False(t, finished)
	assert.Error(t, err) // Now we expect an error
	assert.True(t, errors.HasErrCode(err, errors.ConnectionResetError))
	
	// Verify the states were updated
	assert.Equal(t, 2, len(lastKnownStates))
	assert.Equal(t, StateOnline, lastKnownStates["device1"])
	assert.Equal(t, StateOffline, lastKnownStates["device2"])
	
	// Verify events were published
	assert.Equal(t, 2, len(watcher.eventChan))
	
	// Read and verify events
	event1 := <-watcher.eventChan
	event2 := <-watcher.eventChan
	
	// Events could be in any order, so check both possibilities
	if event1.Serial == "device1" {
		assert.Equal(t, "device1", event1.Serial)
		assert.Equal(t, StateDisconnected, event1.OldState)
		assert.Equal(t, StateOnline, event1.NewState)
		
		assert.Equal(t, "device2", event2.Serial)
		assert.Equal(t, StateDisconnected, event2.OldState)
		assert.Equal(t, StateOffline, event2.NewState)
	} else {
		assert.Equal(t, "device2", event1.Serial)
		assert.Equal(t, StateDisconnected, event1.OldState)
		assert.Equal(t, StateOffline, event1.NewState)
		
		assert.Equal(t, "device1", event2.Serial)
		assert.Equal(t, StateDisconnected, event2.OldState)
		assert.Equal(t, StateOnline, event2.NewState)
	}
}

func TestPublishDevicesUntilError_ReadError(t *testing.T) {
	// Create a mock scanner that returns an error
	scanner := &MockScanner{
		Errs: []error{errors.Errorf(errors.ConnectionResetError, "connection reset")},
	}
	
	// Create a device watcher
	ctx, cancel := context.WithCancel(context.Background())
	watcher := &deviceWatcherImpl{
		ctx:       ctx,
		eventChan: make(chan DeviceStateChangedEvent, 10),
	}
	
	// Initial empty state
	lastKnownStates := make(map[string]DeviceState)
	
	// Run the function
	finished, err := publishDevicesUntilError(scanner, watcher, &lastKnownStates)
	
	// Clean up
	cancel()
	
	// Verify results
	assert.False(t, finished)
	assert.Error(t, err)
	assert.True(t, errors.HasErrCode(err, errors.ConnectionResetError))
}

func TestPublishDevicesUntilError_ContextCanceled(t *testing.T) {
	// Create a mock scanner
	scanner := &MockScanner{
		Messages: [][]byte{[]byte("device1\tdevice\n")},
	}
	
	// Create a device watcher with canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	watcher := &deviceWatcherImpl{
		ctx:       ctx,
		eventChan: make(chan DeviceStateChangedEvent, 10),
	}
	
	// Initial empty state
	lastKnownStates := make(map[string]DeviceState)
	
	// Run the function
	finished, err := publishDevicesUntilError(scanner, watcher, &lastKnownStates)
	
	// Verify results
	assert.True(t, finished)
	assert.NoError(t, err)
}

func assertContainsOnly(t *testing.T, expected, actual []DeviceStateChangedEvent) {
	assert.Len(t, actual, len(expected))
	for _, expectedEntry := range expected {
		assertContains(t, expectedEntry, actual)
	}
}

func assertContains(t *testing.T, expectedEntry DeviceStateChangedEvent, actual []DeviceStateChangedEvent) {
	for _, actualEntry := range actual {
		if expectedEntry == actualEntry {
			return
		}
	}
	assert.Fail(t, "expected to find %+v in %+v", expectedEntry, actual)
}
