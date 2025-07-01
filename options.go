package rollingfile

import (
	"fmt"
	"os"
	"time"
)

type Option func(*RollingFile)

// New creates a new RollingFile wrapping a file under the specified path.
// It opens or creates the file and applies functional options for configuration.
// The file is opened in append mode, and the file permissions are set to the same as the existing file if it exists.
// If the file does not exist, it is created with default permissions (0644).

func New(path string, options ...Option) (logger *RollingFile, err error) {
	mode := os.FileMode(0644)
	info, err := os.Stat(path)
	if err == nil {
		mode = info.Mode()
	}
	logger = &RollingFile{
		mode: mode,
		errorHandler: func(err error) {
			fmt.Fprintf(os.Stderr, "RollingFile error: %v\n", err)
		},
	}

	for _, o := range options {
		o(logger)
	}
	logger.file, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, logger.mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	stat, err := logger.file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat log file: %v", err)
	}
	logger.size = stat.Size()

	return logger, nil
}

// WithMaxBytes returns an option to set the maximum size in bytes before rotation.
func WithMaxBytes(maxBytes int64) Option {
	return func(w *RollingFile) {
		w.maxSize = maxBytes
	}
}

// WithErrorHandler returns an option to set a custom handler for errors occurring during the cleanup of backup files.
func WithErrorHandler(handler func(error)) Option {
	return func(w *RollingFile) {
		w.errorHandler = handler
	}
}

// WithMaxBackups returns an option to set the maximum number of backup files to retain.
func WithMaxBackups(maxBackups int) Option {
	return func(w *RollingFile) {
		w.maxBackups = maxBackups
	}
}

// WithMaxAge returns an option to set the maximum age of backup files before deletion.
func WithMaxAge(age time.Duration) Option {
	return func(w *RollingFile) {
		w.maxAge = age
	}
}

// WithMode returns an option to set the file mode for the log file on creation.
func WithMode(mode os.FileMode) Option {
	return func(w *RollingFile) {
		w.mode = mode
	}
}
