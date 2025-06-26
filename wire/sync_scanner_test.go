package wire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"syscall"
	"testing"
	"time"

	adbErrors "github.com/basiooo/goadb/internal/errors"
	"github.com/stretchr/testify/assert"
)

// mockReader implements io.Reader and io.Closer for testing
type mockSyncScannerReader struct {
	readFunc  func(p []byte) (n int, err error)
	closeFunc func() error
}

func (m *mockSyncScannerReader) Read(p []byte) (n int, err error) {
	if m.readFunc != nil {
		return m.readFunc(p)
	}
	return 0, io.EOF
}

func (m *mockSyncScannerReader) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestNewSyncScanner(t *testing.T) {
	reader := &mockSyncScannerReader{}
	scanner := NewSyncScanner(reader)

	// Verify scanner is created with the reader
	assert.NotNil(t, scanner)

	// Verify it's the correct type
	_, ok := scanner.(*realSyncScanner)
	assert.True(t, ok)
}

func TestSyncScanner_ReadStatus_Success(t *testing.T) {
	// Create a reader that returns a successful status
	data := []byte{0x4f, 0x4b, 0x41, 0x59} // "OKAY" in little-endian
	reader := bytes.NewReader(data)

	scanner := NewSyncScanner(reader)

	// Read status
	status, err := scanner.ReadStatus(StatusSuccess)

	assert.NoError(t, err)
	assert.Equal(t, StatusSuccess, status)
}

func TestSyncScanner_ReadStatus_Failure(t *testing.T) {
	// Create a reader that returns a failure status followed by an error message
	data := []byte{0x46, 0x41, 0x49, 0x4c} // "FAIL" in little-endian

	// Add error message length (4 bytes) and message
	errMsg := "error message"
	msgLen := int32(len(errMsg))

	buf := bytes.NewBuffer(data)
	err := binary.Write(buf, binary.LittleEndian, msgLen)
	assert.NoError(t, err)
	buf.WriteString(errMsg)

	reader := bytes.NewReader(buf.Bytes())
	scanner := NewSyncScanner(reader)

	// Read status
	status, err := scanner.ReadStatus(StatusSuccess)

	assert.Error(t, err)
	assert.Equal(t, StatusNone, status)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.AdbError))
}

func TestSyncScanner_ReadInt32_Success(t *testing.T) {
	// Create a reader that returns an int32
	expectedValue := int32(12345)
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, expectedValue)
	assert.NoError(t, err)

	reader := bytes.NewReader(buf.Bytes())
	scanner := NewSyncScanner(reader)

	// Read int32
	value, err := scanner.ReadInt32()

	assert.NoError(t, err)
	assert.Equal(t, expectedValue, value)
}

func TestSyncScanner_ReadInt32_Error(t *testing.T) {
	// Create a reader that returns an error
	reader := &mockSyncScannerReader{
		readFunc: func(p []byte) (n int, err error) {
			return 0, errors.New("read error")
		},
	}

	scanner := NewSyncScanner(reader)

	// Read int32
	_, err := scanner.ReadInt32()

	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncScanner_ReadFileMode_Success(t *testing.T) {
	// Create a reader that returns a file mode
	expectedMode := os.FileMode(0644)
	expectedValue := uint32(0644 | syscall.S_IFREG) // Regular file with 0644 permissions

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, expectedValue)
	assert.NoError(t, err)

	reader := bytes.NewReader(buf.Bytes())
	scanner := NewSyncScanner(reader)

	// Read file mode
	mode, err := scanner.ReadFileMode()

	assert.NoError(t, err)
	assert.Equal(t, expectedMode, mode&os.ModePerm)
}

func TestSyncScanner_ReadFileMode_Error(t *testing.T) {
	// Create a reader that returns an error
	reader := &mockSyncScannerReader{
		readFunc: func(p []byte) (n int, err error) {
			return 0, errors.New("read error")
		},
	}

	scanner := NewSyncScanner(reader)

	// Read file mode
	_, err := scanner.ReadFileMode()

	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncScanner_ReadTime_Success(t *testing.T) {
	// Create a reader that returns a time
	expectedSeconds := int32(1609459200) // 2021-01-01 00:00:00 UTC
	expectedTime := time.Unix(int64(expectedSeconds), 0).UTC()

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, expectedSeconds)
	assert.NoError(t, err)

	reader := bytes.NewReader(buf.Bytes())
	scanner := NewSyncScanner(reader)

	// Read time
	tm, err := scanner.ReadTime()

	assert.NoError(t, err)
	assert.Equal(t, expectedTime, tm)
}

func TestSyncScanner_ReadTime_Error(t *testing.T) {
	// Create a reader that returns an error
	reader := &mockSyncScannerReader{
		readFunc: func(p []byte) (n int, err error) {
			return 0, errors.New("read error")
		},
	}

	scanner := NewSyncScanner(reader)

	// Read time
	_, err := scanner.ReadTime()

	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncScanner_ReadString_Success(t *testing.T) {
	// Create a reader that returns a string
	expectedString := "test string"
	expectedLength := int32(len(expectedString))

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, expectedLength)
	assert.NoError(t, err)
	buf.WriteString(expectedString)

	reader := bytes.NewReader(buf.Bytes())
	scanner := NewSyncScanner(reader)

	// Read string
	str, err := scanner.ReadString()

	assert.NoError(t, err)
	assert.Equal(t, expectedString, str)
}

func TestSyncScanner_ReadString_LengthError(t *testing.T) {
	// Create a reader that returns an error when reading length
	reader := &mockSyncScannerReader{
		readFunc: func(p []byte) (n int, err error) {
			return 0, errors.New("read error")
		},
	}

	scanner := NewSyncScanner(reader)

	// Read string
	_, err := scanner.ReadString()

	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncScanner_ReadString_ContentError(t *testing.T) {
	// Create a reader that returns a length but then an error when reading content
	expectedLength := int32(10)

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, expectedLength)
	assert.NoError(t, err)

	reader := &mockSyncScannerReader{
		readFunc: func(p []byte) (n int, err error) {
			// First read the length successfully
			if len(p) == 4 {
				copy(p, buf.Bytes())
				return 4, nil
			}
			// Then fail on content read
			return 0, errors.New("read error")
		},
	}

	scanner := NewSyncScanner(reader)

	// Read string
	_, err = scanner.ReadString()

	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncScanner_ReadString_IncompleteContent(t *testing.T) {
	// Create a reader that returns a length but then incomplete content
	expectedLength := int32(10)
	incompleteContent := "test" // Only 4 bytes when 10 are expected

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, expectedLength)
	assert.NoError(t, err)
	buf.WriteString(incompleteContent)

	reader := bytes.NewReader(buf.Bytes())
	scanner := NewSyncScanner(reader)

	// Read string
	_, err = scanner.ReadString()

	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.ConnectionResetError))
}

func TestSyncScanner_ReadBytes_Success(t *testing.T) {
	// Create a reader that returns bytes
	expectedContent := []byte("test content")
	expectedLength := int32(len(expectedContent))

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, expectedLength)
	assert.NoError(t, err)
	buf.Write(expectedContent)

	reader := bytes.NewReader(buf.Bytes())
	scanner := NewSyncScanner(reader)

	// Read bytes
	bytesReader, err := scanner.ReadBytes()

	assert.NoError(t, err)
	assert.NotNil(t, bytesReader)

	// Read the content from the returned reader
	content, err := io.ReadAll(bytesReader)
	assert.NoError(t, err)
	assert.Equal(t, expectedContent, content)
}

func TestSyncScanner_ReadBytes_Error(t *testing.T) {
	// Create a reader that returns an error when reading length
	reader := &mockSyncScannerReader{
		readFunc: func(p []byte) (n int, err error) {
			return 0, errors.New("read error")
		},
	}

	scanner := NewSyncScanner(reader)

	// Read bytes
	_, err := scanner.ReadBytes()

	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncScanner_Close_Success(t *testing.T) {
	// Create a reader with a Close method
	reader := &mockSyncScannerReader{
		closeFunc: func() error {
			return nil
		},
	}

	scanner := NewSyncScanner(reader)

	// Close scanner
	err := scanner.Close()

	assert.NoError(t, err)
}

func TestSyncScanner_Close_Error(t *testing.T) {
	// Create a reader with a Close method that returns an error
	expectedErr := errors.New("close error")
	reader := &mockSyncScannerReader{
		closeFunc: func() error {
			return expectedErr
		},
	}

	scanner := NewSyncScanner(reader)

	// Close scanner
	err := scanner.Close()

	assert.Error(t, err)
	assert.True(t, adbErrors.HasErrCode(err, adbErrors.NetworkError))
}

func TestSyncScanner_Close_NotCloser(t *testing.T) {
	// Create a reader that doesn't implement io.Closer
	reader := bytes.NewReader([]byte{})

	scanner := NewSyncScanner(reader)

	// Close scanner
	err := scanner.Close()

	assert.NoError(t, err)
}
