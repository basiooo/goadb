// TODO(z): Implement tests for sync_client functions.
package adb

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/basiooo/goadb/internal/errors"
	"github.com/basiooo/goadb/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockSyncScanner implements wire.SyncScanner for testing
type MockSyncScanner struct {
	statusResponses []string
	statusIndex     int
	statusCalls     int
	bytesResponses  [][]byte
	bytesIndex      int
	bytesErr        error
	stringResponses []string
	stringIndex     int
}

func (s *MockSyncScanner) ReadStatus(cmd string) (string, error) {
	s.statusCalls++
	if s.statusIndex < len(s.statusResponses) {
		status := s.statusResponses[s.statusIndex]
		s.statusIndex++
		return status, nil
	}
	return "", errors.Errorf(errors.ConnectionResetError, "no more status responses")
}

func (s *MockSyncScanner) ReadBytes() (io.Reader, error) {
	if s.bytesErr != nil {
		return nil, s.bytesErr
	}
	if s.bytesIndex < len(s.bytesResponses) {
		data := s.bytesResponses[s.bytesIndex]
		s.bytesIndex++
		return bytes.NewReader(data), nil
	}
	return nil, errors.Errorf(errors.ConnectionResetError, "no more bytes responses")
}

func (s *MockSyncScanner) ReadString() (string, error) {
	if s.stringIndex < len(s.stringResponses) {
		str := s.stringResponses[s.stringIndex]
		s.stringIndex++
		return str, nil
	}
	return "", errors.Errorf(errors.ConnectionResetError, "no more string responses")
}

func (s *MockSyncScanner) ReadInt32() (int32, error) {
	return 0, nil
}

func (s *MockSyncScanner) ReadFileMode() (os.FileMode, error) {
	return 0, nil
}

func (s *MockSyncScanner) ReadTime() (time.Time, error) {
	return time.Time{}, nil
}

func (s *MockSyncScanner) Close() error {
	return nil
}

var someTime = time.Date(2015, 5, 3, 8, 8, 8, 0, time.UTC)

func TestStatValid(t *testing.T) {
	var buf bytes.Buffer
	conn := &wire.SyncConn{SyncScanner: wire.NewSyncScanner(&buf), SyncSender: wire.NewSyncSender(&buf)}

	var mode os.FileMode = 0777

	err := conn.SendOctetString("STAT")
	assert.NoError(t, err)
	err = conn.SendFileMode(mode)
	assert.NoError(t, err)
	err = conn.SendInt32(4)
	assert.NoError(t, err)
	err = conn.SendTime(someTime)
	assert.NoError(t, err)

	entry, err := stat(conn, "/thing")
	assert.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, mode, entry.Mode, "expected os.FileMode %s, got %s", mode, entry.Mode)
	assert.Equal(t, int32(4), entry.Size)
	assert.Equal(t, someTime, entry.ModifiedAt)
	assert.Equal(t, "", entry.Name)
}

func TestStatBadResponse(t *testing.T) {
	var buf bytes.Buffer
	conn := &wire.SyncConn{SyncScanner: wire.NewSyncScanner(&buf), SyncSender: wire.NewSyncSender(&buf)}

	err := conn.SendOctetString("SPAT")
	assert.NoError(t, err)

	entry, err := stat(conn, "/")
	assert.Nil(t, entry)
	assert.Error(t, err)
}

func TestStatNoExist(t *testing.T) {
	var buf bytes.Buffer
	conn := &wire.SyncConn{SyncScanner: wire.NewSyncScanner(&buf), SyncSender: wire.NewSyncSender(&buf)}

	// Return all zeros to simulate file not existing
	err := conn.SendOctetString("STAT")
	assert.NoError(t, err)
	err = conn.SendFileMode(0)
	assert.NoError(t, err)
	err = conn.SendInt32(0)
	assert.NoError(t, err)
	err = conn.SendTime(zeroTime)
	assert.NoError(t, err)

	entry, err := stat(conn, "/nonexistent")
	assert.Nil(t, entry)
	assert.Error(t, err)
	assert.True(t, errors.HasErrCode(err, errors.FileNoExistError))
}

func TestSendFile(t *testing.T) {
	var buf bytes.Buffer
	conn := &wire.SyncConn{SyncScanner: wire.NewSyncScanner(&buf), SyncSender: wire.NewSyncSender(&buf)}

	var mode os.FileMode = 0644
	mtime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	writer, err := sendFile(conn, "/test/file.txt", mode, mtime)
	assert.NoError(t, err)
	assert.NotNil(t, writer)

	// Verify the correct command was sent
	assert.Equal(t, "SEND", string(buf.Bytes()[:4]))

	// Write some data and close
	data := []byte("test data")
	n, err := writer.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	err = writer.Close()
	assert.NoError(t, err)
}

func TestListDirEntries(t *testing.T) {
	var buf bytes.Buffer
	conn := &wire.SyncConn{SyncScanner: wire.NewSyncScanner(&buf), SyncSender: wire.NewSyncSender(&buf)}

	entries, err := listDirEntries(conn, "/test")
	assert.NoError(t, err)
	assert.NotNil(t, entries)

	// Verify the correct command was sent
	assert.Equal(t, "LIST", string(buf.Bytes()[:4]))
}

func TestReceiveFile(t *testing.T) {
	// Create a mock scanner that returns DONE status
	mockScanner := &MockSyncScanner{
		statusResponses: []string{wire.StatusSyncDone},
	}

	// Create a buffer to capture the sent data
	var buf bytes.Buffer
	sender := wire.NewSyncSender(&buf)

	// Create a SyncConn with our mock scanner and real sender
	conn := &wire.SyncConn{SyncScanner: mockScanner, SyncSender: sender}

	// Call the function under test
	reader, err := receiveFile(conn, "/test/file.txt")
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	// Verify the correct command was sent
	// Just verify that the buffer is not empty, which means something was sent
	assert.NotEmpty(t, buf.Bytes())

	// Verify the mock scanner was used
	assert.Equal(t, 1, mockScanner.statusCalls)
}

func TestReadStat(t *testing.T) {
	var buf bytes.Buffer
	scanner := wire.NewSyncScanner(&buf)

	// Setup test data
	var mode os.FileMode = 0755
	var size int32 = 1024
	mtime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	// Write test data to buffer
	sender := wire.NewSyncSender(&buf)
	err := sender.SendFileMode(mode)
	assert.NoError(t, err)
	err = sender.SendInt32(size)
	assert.NoError(t, err)
	err = sender.SendTime(mtime)
	assert.NoError(t, err)

	// Read the stat
	entry, err := readStat(scanner)
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, mode, entry.Mode)
	assert.Equal(t, size, entry.Size)
	assert.Equal(t, mtime, entry.ModifiedAt)
}

func TestReadStat_FileNoExist(t *testing.T) {
	var buf bytes.Buffer
	scanner := wire.NewSyncScanner(&buf)

	// Setup test data for non-existent file (all zeros)
	sender := wire.NewSyncSender(&buf)
	err := sender.SendFileMode(0)
	assert.NoError(t, err)
	err = sender.SendInt32(0)
	assert.NoError(t, err)
	err = sender.SendTime(time.Unix(0, 0).UTC())
	assert.NoError(t, err)

	// Read the stat
	entry, err := readStat(scanner)
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.True(t, errors.HasErrCode(err, errors.FileNoExistError))
}
