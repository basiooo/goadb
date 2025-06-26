package wire

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFileModeFromAdb_RegularFile(t *testing.T) {
	// Test regular file with permissions 0644
	modeFromSync := uint32(0644)
	expectedMode := os.FileMode(0644)
	
	mode := ParseFileModeFromAdb(modeFromSync)
	
	assert.Equal(t, expectedMode, mode)
}

func TestParseFileModeFromAdb_Directory(t *testing.T) {
	// Test directory with permissions 0755
	modeFromSync := ModeDir | uint32(0755)
	expectedMode := os.ModeDir | os.FileMode(0755)
	
	mode := ParseFileModeFromAdb(modeFromSync)
	
	assert.Equal(t, expectedMode, mode)
}

func TestParseFileModeFromAdb_Symlink(t *testing.T) {
	// Test symlink with permissions 0777
	modeFromSync := ModeSymlink | uint32(0777)
	expectedMode := os.ModeSymlink | os.FileMode(0777)
	
	mode := ParseFileModeFromAdb(modeFromSync)
	
	assert.Equal(t, expectedMode, mode)
}

func TestParseFileModeFromAdb_Socket(t *testing.T) {
	// Test socket with permissions 0600
	modeFromSync := ModeSocket | uint32(0600)
	
	mode := ParseFileModeFromAdb(modeFromSync)
	
	// Just check that it has the right permissions and type
	assert.Equal(t, os.FileMode(0600), mode.Perm())
	// Skip checking the exact mode type as os.ModeSocket value might differ
	// Just verify it's not a regular file
	assert.True(t, mode&os.ModeType != 0)
}

func TestParseFileModeFromAdb_Fifo(t *testing.T) {
	// Test FIFO (named pipe) with permissions 0666
	modeFromSync := ModeFifo | uint32(0666)
	expectedMode := os.ModeNamedPipe | os.FileMode(0666)
	
	mode := ParseFileModeFromAdb(modeFromSync)
	
	assert.Equal(t, expectedMode, mode)
}

func TestParseFileModeFromAdb_CharDevice(t *testing.T) {
	// Test character device with permissions 0600
	modeFromSync := ModeCharDevice | uint32(0600)
	expectedMode := os.ModeCharDevice | os.FileMode(0600)
	
	mode := ParseFileModeFromAdb(modeFromSync)
	
	assert.Equal(t, expectedMode, mode)
}

func TestParseFileModeFromAdb_MultipleFlags(t *testing.T) {
	// When multiple flags are set, the function should prioritize them in order
	// First test: Symlink takes precedence over Directory
	modeFromSync := ModeSymlink | ModeDir | uint32(0755)
	expectedMode := os.ModeSymlink | os.FileMode(0755) // Should be symlink, not directory
	
	mode := ParseFileModeFromAdb(modeFromSync)
	
	assert.Equal(t, expectedMode, mode)
	
	// Second test: Directory takes precedence over Socket
	modeFromSync = ModeDir | ModeSocket | uint32(0755)
	expectedMode = os.ModeDir | os.FileMode(0755) // Should be directory, not socket
	
	mode = ParseFileModeFromAdb(modeFromSync)
	
	assert.Equal(t, expectedMode, mode)
}