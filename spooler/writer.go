package spooler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrWriteLimitReached = errors.New("write limit reached")

// Factory for creating file writers with specific configurations.
type writerFactory struct {
	maxFileSize int
}

// Create a new writer factory with the given configuration.
func newWriterFactory(config FileWriterConfig) (writerFactory, error) {
	if err := config.Validate(); err != nil {
		return writerFactory{}, fmt.Errorf("writer factory: invalid config: %w", err)
	}

	return writerFactory{
		maxFileSize: config.MaxFileSize,
	}, nil
}


// Writes data to a temporary file and commits it to the final location.
type fileWriter struct {
	// Temporary file handle
	tmpFile *os.File

	// Final commit path
	commitPath string

	// Maximum allowed file size
	maxFileSize int

	// Current size of the written data
	written int
}

// Create a new file writer that writes to a temporary file in `dir`.
// Upon commit, the file will be renamed to the specified `commitPath`.
func (f *writerFactory) NewFileWriter(dir, fileName string) (*fileWriter, error) {
	tmpFile, err := os.CreateTemp(dir, "*.tmp")
	if err != nil {
		return nil, err
	}

	return &fileWriter{
		tmpFile:     tmpFile,
		commitPath:  filepath.Join(dir, fileName),
		maxFileSize: f.maxFileSize,
	}, nil
}

// Append data to the temporary file.
func (w *fileWriter) Write(data []byte) error {
	written, err := w.tmpFile.Write(data)
	if err != nil {
		return err
	}

	postWriteSize := w.written + written
	if postWriteSize >= w.maxFileSize {
		return ErrWriteLimitReached
	}

	w.written = postWriteSize

	return nil
}

// Commit the write-op by renaming the temp file to the final path.
func (w *fileWriter) Commit() (int, error) {
	tmpPath := w.tmpFile.Name()
	written := w.written

	if err := os.Rename(tmpPath, w.commitPath); err != nil {
		return written, err
	}

	if err := w.tmpFile.Sync(); err != nil {
		return written, err
	}

	if err := w.tmpFile.Close(); err != nil {
		return written, err
	}

	return written, nil
}

// Abort the write-op by deleting the temporary file.
func (w *fileWriter) Abort() error {
	return os.Remove(w.tmpFile.Name())
}
