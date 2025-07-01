package rolfilite

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type RollingFile struct {
	maxBackups   int
	maxSize      int64
	maxAge       time.Duration
	file         *os.File
	size         int64
	mode         os.FileMode
	errorHandler func(error)
	cleanupMutex sync.Mutex
}

func (l *RollingFile) Write(line []byte) (n int, err error) {
	if len(line) == 0 {
		return 0, nil
	}
	n = len(line)
	if int64(n) > l.maxSize && l.maxSize > 0 {
		return 0, fmt.Errorf("line exceeds max size")
	}

	if l.size+int64(n) >= l.maxSize && l.maxSize > 0 {
		if err = l.rotate(); err != nil {
			return 0, fmt.Errorf("failed to rotate log file: %w", err)
		}
	}

	n, err = l.file.Write(line)
	if err != nil {
		return n, err
	}
	l.size += int64(n)

	return n, nil
}

// rotate creates a timestamped backup of the current log file, truncates the original, and cleans up old backups.
func (l *RollingFile) rotate() error {
	// Close the current file before renaming
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("failed to close file before rotation: %w", err)
	}

	i := 0
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.%s.%d", l.file.Name(), timestamp, i)

	// Find a unique backup filename
	_, err := os.Stat(backupPath)
	for err == nil {
		i++
		backupPath = fmt.Sprintf("%s.%s.%d", l.file.Name(), timestamp, i)
		_, err = os.Stat(backupPath)
	}

	// Rename the current file to the backup name
	if err := os.Rename(l.file.Name(), backupPath); err != nil {
		return fmt.Errorf("failed to rename file for rotation: %w", err)
	}

	// Create a new file with the original name and same mode
	newFile, err := os.OpenFile(l.file.Name(), os.O_CREATE|os.O_RDWR|os.O_APPEND, l.mode)
	if err != nil {
		return fmt.Errorf("failed to create new log file after rotation: %w", err)
	}
	l.file = newFile
	l.size = 0

	go l.cleanupBackups()
	return nil
}

// cleanupBackups deletes oldest backup files to enforce the maxBackups limit.
func (l *RollingFile) cleanupBackups() {
	l.cleanupMutex.Lock()
	defer l.cleanupMutex.Unlock()
	matches, err := filepath.Glob(l.file.Name() + ".*")
	if err != nil {
		l.errorHandler(fmt.Errorf("failed to list backup files: %w", err))
		return
	}

	var backups []string
	for _, file := range matches {
		if strings.HasPrefix(file, l.file.Name()+".") && len(file) > len(l.file.Name())+1 {
			backups = append(backups, file)
		}
	}

	sort.Strings(backups)
	for i, file := range backups {
		expired, err := l.isOlderThanFilename(file)
		if err != nil {
			l.errorHandler(fmt.Errorf("failed to check backup file age: %w", err))
			continue
		}
		if (len(backups)-i > l.maxBackups && l.maxBackups > 0) || expired {
			err = os.Remove(file)
			if err != nil {
				l.errorHandler(fmt.Errorf("failed to remove backup file %q: %w", file, err))
			}
		}
	}
}

// isOlderThanFilename returns true if the embedded timestamp in fname
// (in the form ".log.YYYYMMDD-HHMMSS.") is before cutoff.
func (l *RollingFile) isOlderThanFilename(fname string) (bool, error) {
	if l.maxAge <= 0 {
		return false, nil
	}
	re := regexp.MustCompile(`\.log\.(\d{8}-\d{6})\.`)
	matches := re.FindStringSubmatch(fname)
	if len(matches) < 2 {
		return false, fmt.Errorf("no timestamp found in %q", fname)
	}

	ts, err := time.Parse("20060102-150405", matches[1])
	if err != nil {
		return false, fmt.Errorf("cannot parse timestamp %q: %w", matches[1], err)
	}
	cutoff := time.Now().Add(-l.maxAge)

	return ts.Before(cutoff), nil
}

// Close calls the Close function on the underlying file.
func (l *RollingFile) Close() error {
	return l.file.Close()
}

// Sync calls the Sync function on the underlying file.
func (l *RollingFile) Sync() error {
	return l.file.Sync()
}

// Name returns the name of the underlying file.
func (l *RollingFile) Name() string {
	return l.file.Name()
}
