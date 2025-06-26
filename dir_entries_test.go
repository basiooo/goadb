package adb

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockSyncScanner is a mock implementation of wire.SyncScanner for testing
type mockDirEntriesSyncScanner struct {
	statusResponses []string
	statusIndex     int
	modeResponses   []os.FileMode
	modeIndex       int
	int32Responses  []int32
	int32Index      int
	timeResponses   []time.Time
	timeIndex       int
	stringResponses []string
	stringIndex     int
	bytesResponses  [][]byte
	bytesIndex      int
	closeFunc       func() error
	err             error
}

func (m *mockDirEntriesSyncScanner) ReadStatus(expectedStatus string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.statusIndex >= len(m.statusResponses) {
		return "", io.EOF
	}
	status := m.statusResponses[m.statusIndex]
	m.statusIndex++
	return status, nil
}

func (m *mockDirEntriesSyncScanner) ReadInt32() (int32, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.int32Index >= len(m.int32Responses) {
		return 0, io.EOF
	}
	val := m.int32Responses[m.int32Index]
	m.int32Index++
	return val, nil
}

func (m *mockDirEntriesSyncScanner) ReadFileMode() (os.FileMode, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.modeIndex >= len(m.modeResponses) {
		return 0, io.EOF
	}
	mode := m.modeResponses[m.modeIndex]
	m.modeIndex++
	return mode, nil
}

func (m *mockDirEntriesSyncScanner) ReadTime() (time.Time, error) {
	if m.err != nil {
		return time.Time{}, m.err
	}
	if m.timeIndex >= len(m.timeResponses) {
		return time.Time{}, io.EOF
	}
	t := m.timeResponses[m.timeIndex]
	m.timeIndex++
	return t, nil
}

func (m *mockDirEntriesSyncScanner) ReadString() (string, error) {
	if m.err != nil {
		return "", m.err
	}
	if m.stringIndex >= len(m.stringResponses) {
		return "", io.EOF
	}
	s := m.stringResponses[m.stringIndex]
	m.stringIndex++
	return s, nil
}

func (m *mockDirEntriesSyncScanner) ReadBytes() (io.Reader, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.bytesIndex >= len(m.bytesResponses) {
		return nil, io.EOF
	}
	b := m.bytesResponses[m.bytesIndex]
	m.bytesIndex++
	return bytes.NewReader(b), nil
}

func (m *mockDirEntriesSyncScanner) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestDirEntries_ReadAll(t *testing.T) {
	// Create a mock scanner that returns two directory entries followed by DONE
	now := time.Now()
	mock := &mockDirEntriesSyncScanner{
		statusResponses: []string{"DENT", "DENT", "DONE"},
		modeResponses:   []os.FileMode{0644, 0755},
		int32Responses:  []int32{1024, 2048},
		timeResponses:   []time.Time{now, now.Add(time.Hour)},
		stringResponses: []string{"file1.txt", "file2.exe"},
		closeFunc:       func() error { return nil },
	}

	entries := &DirEntries{scanner: mock}

	// Read all entries
	result, err := entries.ReadAll()

	// Verify results
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	// Check first entry
	assert.Equal(t, "file1.txt", result[0].Name)
	assert.Equal(t, os.FileMode(0644), result[0].Mode)
	assert.Equal(t, int32(1024), result[0].Size)
	assert.Equal(t, now, result[0].ModifiedAt)

	// Check second entry
	assert.Equal(t, "file2.exe", result[1].Name)
	assert.Equal(t, os.FileMode(0755), result[1].Mode)
	assert.Equal(t, int32(2048), result[1].Size)
	assert.Equal(t, now.Add(time.Hour), result[1].ModifiedAt)
}

func TestDirEntries_Next_Entry_Err(t *testing.T) {
	// Create a mock scanner that returns one entry, then DONE
	now := time.Now()
	mock := &mockDirEntriesSyncScanner{
		statusResponses: []string{"DENT", "DONE"},
		modeResponses:   []os.FileMode{0644},
		int32Responses:  []int32{1024},
		timeResponses:   []time.Time{now},
		stringResponses: []string{"file1.txt"},
		closeFunc:       func() error { return nil },
	}

	entries := &DirEntries{scanner: mock}

	// First call to Next should return true
	assert.True(t, entries.Next())
	// Entry should be available
	entry := entries.Entry()
	assert.NotNil(t, entry)
	assert.Equal(t, "file1.txt", entry.Name)
	// No error yet
	assert.NoError(t, entries.Err())

	// Second call to Next should return false (DONE)
	assert.False(t, entries.Next())
	// No error
	assert.NoError(t, entries.Err())
}

func TestDirEntries_Next_Error(t *testing.T) {
	// Create a mock scanner that returns an error
	mockErr := io.ErrUnexpectedEOF
	mock := &mockDirEntriesSyncScanner{
		err: mockErr,
	}

	entries := &DirEntries{scanner: mock}

	// Next should return false due to error
	assert.False(t, entries.Next())
	// Error should be available
	assert.Equal(t, mockErr, entries.Err())
}

func TestDirEntries_Next_InvalidStatus(t *testing.T) {
	// Create a mock scanner that returns an invalid status
	mock := &mockDirEntriesSyncScanner{
		statusResponses: []string{"INVALID"},
	}

	entries := &DirEntries{scanner: mock}

	// Next should return false due to invalid status
	assert.False(t, entries.Next())
	// Error should be available
	assert.Error(t, entries.Err())
	assert.Contains(t, entries.Err().Error(), "expected dir entry ID 'DENT', but got 'INVALID'")
}

func TestDirEntries_Close(t *testing.T) {
	// Create a mock scanner with a close function that counts calls
	closeCount := 0
	mock := &mockDirEntriesSyncScanner{
		closeFunc: func() error {
			closeCount++
			return nil
		},
	}

	entries := &DirEntries{scanner: mock}

	// Close should call the underlying scanner's Close
	err := entries.Close()
	assert.NoError(t, err)
	assert.Equal(t, 1, closeCount)
}
