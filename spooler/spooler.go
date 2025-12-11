package spooler

import (
	"context"
	"errors"
)

// TODO:
// - Lightweight embedded database for book-keeping
// - Maintain log to track files at directory-level

// Manages local batches for remote storage or local persistence.
type Spooler struct {
	// Configuration
	config SpoolerConfig

	// Factory for creating file writers
	writer writerFactory

	// Manages batch directories and sizes
	batcher *batcher

	// Tracks current file writer
	currentWriter *fileWriter
}

// Create a new Spooler instance
func NewSpooler(config SpoolerConfig) (*Spooler, error) {
	batcher, err := newBatcher(config.BatchConfig)
	if err != nil {
		return nil, err
	}

	writerFactory, err := newWriterFactory(config.FileWriterConfig)
	if err != nil {
		return nil, err
	}

	return &Spooler{
		config:  config,
		writer:  writerFactory,
		batcher: batcher,
	}, nil
}

// Current batch size (in bytes)
func (s *Spooler) BatchSize() int {
	return int(s.batcher.Size())
}

// Create a new file writer
func (s *Spooler) NewWriter(ctx context.Context, fileName string) error {
	if s.currentWriter != nil {
		if err := s.Commit(); err != nil {
			return err
		}
	}

	writer, err := s.writer.NewFileWriter(s.batcher.CurrentDir(), fileName)
	if err != nil {
		return err
	}

	s.currentWriter = writer
	return nil
}

// Write a chunk of data to the current file writer
func (s *Spooler) WriteChunk(data []byte) error {
	var err error

	if err = s.currentWriter.Write(data); err != nil {
		if errors.Is(err, ErrWriteLimitReached) {
			if err = s.currentWriter.Abort(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Commit the current file writer
func (s *Spooler) Commit() error {
	var err error

	totalWritten, err := s.currentWriter.Commit()
	if err != nil {
		return err
	}
	s.currentWriter = nil

	s.batcher.AddBytes(totalWritten)
	if err = s.batcher.Rotate(); err != nil {
		return err
	}

	return nil
}
