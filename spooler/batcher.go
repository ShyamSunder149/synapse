package spooler

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
)

// Manages batch directories and size.
type batcher struct {
	config BatchConfig

	currentDir  string
	currentSize atomic.Int64
}

// TODO: Resume batching logic
func newBatcher(config BatchConfig) (*batcher, error) {
	var err error

	if err = config.Validate(); err != nil {
		return nil, fmt.Errorf("batcher config: %w", err)
	}

	currentDir, err := newBatch(config.BaseDir)
	if err != nil {
		return nil, err
	}

	batcher := &batcher{
		config:     config,
		currentDir: currentDir,
	}
	batcher.currentSize.Store(0)

	return batcher, nil
}

// Current batch size (in bytes)
func (b *batcher) Size() int64 {
	return b.currentSize.Load()
}

// Current batch directory
func (b *batcher) CurrentDir() string {
	return b.currentDir
}

// Increment current batch size (in bytes)
func (b *batcher) AddBytes(delta int) {
	b.currentSize.Add(int64(delta))
}

// Rotate the batch if it exceeds the maximum size
func (b *batcher) Rotate() error {
	currentDir := b.currentDir
	currentSize := b.currentSize.Load()

	if currentSize <= b.config.MaxBatchSize {
		return nil
	}

	// TODO: Make this robust
	if b.config.Processor.Async {
		go b.processBatchWorker(currentDir, currentSize)
	} else {
		if err := b.config.Processor.Hooks.OnBatchReady(currentDir, currentSize); err != nil {
			return err
		}

		if b.config.Processor.DeleteSource {
			if err := os.RemoveAll(currentDir); err != nil {
				return fmt.Errorf("failed to delete batch directory: %w", err)
			}
		}
	}

	return b.reset()
}

// Create a new batch directory
func newBatch(baseDir string) (string, error) {
	// Generate a unique file ID
	createFileID := func() string {
		return strconv.Itoa(rand.Int())
	}

	batchDir := filepath.Join(baseDir, "batch-"+createFileID())
	if err := os.MkdirAll(batchDir, filePermissions); err != nil {
		return "", fmt.Errorf("failed to create batch directory: %w", err)
	}
	return batchDir, nil
}

// Reset the batcher to start a new batch
func (b *batcher) reset() error {
	currentDir, err := newBatch(b.config.BaseDir)
	if err != nil {
		return err
	}
	b.currentDir = currentDir
	b.currentSize.Store(0)
	return nil
}

func (b *batcher) processBatchWorker(dir string, size int64) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "batch worker: panic in async batch processing: %v\n", r)
		}
	}()

	if err := b.config.Processor.Hooks.OnBatchReady(dir, size); err != nil {
		b.config.Processor.Hooks.OnBatchError(dir, &err)
		return
	}

	if b.config.Processor.DeleteSource {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Fprintf(os.Stderr, "batch worker: failed to delete batch directory %s: %v\n", dir, err)
		}
	}
}
