package rollingfile

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWrite verifies that the logger writes messages with the correct prefix
// and ensures the written content matches the expected output.
func TestWrite(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	logger, err := New(logPath,
		WithMaxBytes(1024),
		WithMaxBackups(5),
	)
	assert.NoError(t, err)

	message := "hello log\n"
	n, err := logger.Write([]byte(message))
	assert.NoError(t, err)
	assert.Equal(t, len(message), n)

	err = logger.Sync()
	assert.NoError(t, err)
	err = logger.Close()
	assert.NoError(t, err)

	contents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	assert.Equal(t, string(contents), message)
}

// TestLineExceedsMaxSize ensures that attempting to write a line larger than the maximum size
// results in an error and no data is written to the log file.
func TestLineExceedsMaxSize(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "exceed.log")
	logger, err := New(logPath,
		WithMaxBytes(100),
		WithMaxBackups(5),
	)
	assert.NoError(t, err)

	longMessage := strings.Repeat("x", 200) + "\n"
	n, err := logger.Write([]byte(longMessage))
	assert.Error(t, err)
	assert.Equal(t, 0, n)

	err = logger.Sync()
	assert.NoError(t, err)
	err = logger.Close()
	assert.NoError(t, err)

	contents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}
	assert.Equal(t, len(contents), 0)
}

// TestLogRotationOccurs verifies that log rotation occurs when the maximum file size is exceeded.
func TestLogRotationOccurs(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "rotate.log")
	logger, err := New(logPath,
		WithMaxBytes(100),
		WithMaxBackups(5),
	)
	assert.NoError(t, err)

	// Write enough data to exceed 100 bytes
	for i := 0; i < 10; i++ {
		msg := strings.Repeat("x", 15) + "\n"
		n, err := logger.Write([]byte(msg))
		assert.NoError(t, err)
		assert.Equal(t, len(msg), n)
	}

	err = logger.Sync()
	assert.NoError(t, err)
	err = logger.Close()
	assert.NoError(t, err)

	files, err := filepath.Glob(logPath + ".*")
	if err != nil {
		t.Fatalf("failed to list rotated files: %v", err)
	}

	if len(files) == 0 {
		t.Error("expected rotated log file, found none")
	}
}

// TestMaxBackupsIsEnforced ensures that the maximum number of backup files is enforced.
func TestMaxBackupsIsEnforced(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "maxbackup.log")
	fmt.Println("Log path:", logPath)
	logger, err := New(logPath,
		WithMaxBytes(100),
		WithMaxBackups(3),
	)
	assert.NoError(t, err)

	for i := 0; i < 5; i++ {
		msg := strings.Repeat("y", 80) + "\n"
		_, err := logger.Write([]byte(msg))
		assert.NoError(t, err)
	}
	err = logger.Close()
	assert.NoError(t, err)

	files, err := filepath.Glob(logPath + ".*")
	assert.NoError(t, err)

	assert.Equal(t, 3, len(files), "expected 3 backups, found %d", len(files))
}

// TestRotationLinesRetained ensures that all log lines are retained across rotated files.
func TestRotationLinesRetained(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "file.log")

	const lineCount = 50
	const lineSize = 15
	const rotationSize = 100

	logger, err := New(logPath,
		WithMaxBytes(rotationSize),
		WithMaxBackups(100),
	)
	assert.NoError(t, err)

	for i := 0; i < lineCount; i++ {
		msg := strings.Repeat("y", lineSize) + "\n"
		_, err := logger.Write([]byte(msg))
		assert.NoError(t, err)
	}

	files, err := filepath.Glob(logPath + "*")
	assert.NoError(t, err)

	nLogFiles := lineCount/(rotationSize/lineSize) + 1
	assert.Equal(t, nLogFiles, len(files), "expected %d logFiles, found %d", nLogFiles, len(files))

	total := 0

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		total += len(bytes.Split(data, []byte("\n"))) - 1
	}

	assert.Equal(t, lineCount, total, "expected %d lines in all files, got %d", lineCount, total)
}
