package spooler

import (
	"errors"
)

type SpoolerConfig struct {
	BatchConfig      BatchConfig
	FileWriterConfig FileWriterConfig
}

func (c SpoolerConfig) Validate() error {
	var err error

	if err = c.BatchConfig.Validate(); err != nil {
		return err
	}

	if err = c.FileWriterConfig.Validate(); err != nil {
		return err
	}

	return nil
}

type BatchConfig struct {
	// Root directory where batch directories are created
	BaseDir string

	// Maximum batch batchDirSize (per directory)
	MaxBatchSize int64

	// Handle batch processing
	Processor BatchProcessor
}

func (c BatchConfig) Validate() error {
	var err error

	if err = mkdirIfNotExists(c.BaseDir); err != nil {
		return errors.New("batch config: " + err.Error())
	}

	if c.MaxBatchSize <= 0 {
		return errors.New("batch config: max batch size must be greater than zero")
	}

	if err = c.Processor.Validate(); err != nil {
		return err
	}

	return nil
}

type BatchProcessor struct {
	// Perform batch processing asynchronously.
	//
	// Useful for long-running tasks like converting to WARC.
	Async bool

	// Delete source batch directory after processing
	//
	// Useful to save disk space after processing is complete.
	DeleteSource bool

	// Hooks for batch processing events
	Hooks *BatchHooks
}

func (c BatchProcessor) Validate() error {
	if c.Hooks == nil {
		c.Hooks = &BatchHooks{
			OnBatchReady: func(batchDir string, totalBytes int64) error { return nil },
			OnBatchError: func(batchDir string, err *error) {},
		}
	}
	return nil
}

// BatchHandler receives notifications when batches are ready for processing.
type BatchHooks struct {
	// It's called when a batch reaches its threshold and is ready
	// for processing. The batchDir contains all files in the batch.
	OnBatchReady func(batchDir string, totalBytes int64) error

	// OnBatchError is called when an error occurs during batch management.
	OnBatchError func(batchDir string, err *error)
}

type FileWriterConfig struct {
	MaxFileSize int
}

func (c FileWriterConfig) Validate() error {
	if c.MaxFileSize <= 0 {
		return errors.New("file writer config: max file size must be greater than zero")
	}

	return nil
}
