package adb

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsExecutable(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "executable_test")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			log.Printf("[TestIsExecutable] error removing temporary directory: %s", err)
		}
	}()

	// Create a non-executable file
	nonExecPath := filepath.Join(tmpDir, "non_executable.txt")
	err = os.WriteFile(nonExecPath, []byte("test content"), 0644)
	require.NoError(t, err)

	// Create an executable file
	execPath := filepath.Join(tmpDir, "executable.sh")
	err = os.WriteFile(execPath, []byte("#!/bin/sh\necho test"), 0755)
	require.NoError(t, err)

	// Test non-executable file
	err = isExecutable(nonExecPath)
	assert.Error(t, err, "Non-executable file should return an error")

	// Test executable file
	err = isExecutable(execPath)
	assert.NoError(t, err, "Executable file should not return an error")

	// Test non-existent file
	err = isExecutable(filepath.Join(tmpDir, "does_not_exist"))
	assert.Error(t, err, "Non-existent file should return an error")
}
