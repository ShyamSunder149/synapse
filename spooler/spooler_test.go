package spooler

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/ritvikos/synapse/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Write more robust tests once spooler's internal implementation is completed, till then, basic ones.

func generateTestData(t *testing.T) string {
	t.Helper()

	paragraphCount := int(rand.Int64N(15))
	sentenceCount := int(rand.Int64N(20))
	wordCount := int(rand.Int64N(100))
	separation := " "

	return gofakeit.LoremIpsumParagraph(paragraphCount, sentenceCount, wordCount, separation)
}

func generateTestChunks(t *testing.T, numChunks int) [][]byte {
	t.Helper()

	chunks := make([][]byte, numChunks)
	for i := range numChunks {
		chunks[i] = []byte(generateTestData(t))
	}
	return chunks
}

func writeChunks(t *testing.T, spooler *Spooler, chunks [][]byte) int {
	t.Helper()

	totalSize := 0
	for _, chunk := range chunks {
		err := spooler.WriteChunk(chunk)
		require.NoError(t, err, "failed to write chunk")
		totalSize += len(chunk)
	}
	return totalSize
}

func TestSpoolerSync(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err, "failed to get working directory")

	baseDir := filepath.Join(wd, "test_spooler")

	config := SpoolerConfig{
		BatchConfig: BatchConfig{
			BaseDir:      baseDir,
			MaxBatchSize: 2 * 1024 * 1024,
			Processor: BatchProcessor{
				Async:        false,
				DeleteSource: true,
				Hooks:        nil,
			},
		},
		FileWriterConfig: FileWriterConfig{
			MaxFileSize: 5 * 1024 * 1024,
			logger:      logger.New(),
		},
	}

	spooler, err := NewSpooler(config)
	require.NoError(t, err, "failed to create spooler")

	t.Cleanup(func() {
		time.Sleep(10 * time.Second)
		os.RemoveAll(baseDir)
	})

	t.Run("write single file", func(t *testing.T) {
		ctx := t.Context()

		err := spooler.NewWriter(ctx, "testfile1.txt")
		require.NoError(t, err, "failed to create new writer")

		chunks := generateTestChunks(t, 3)
		totalSize := writeChunks(t, spooler, chunks)

		err = spooler.Commit()
		require.NoError(t, err, "failed to commit")

		assert.Equal(t, totalSize, spooler.BatchSize(), "batch size mismatch")
	})

	t.Run("write multiple files", func(t *testing.T) {
		ctx := t.Context()
		initialBatchSize := spooler.BatchSize()

		err := spooler.NewWriter(ctx, "testfile2.txt")
		require.NoError(t, err, "failed to create new writer")

		chunks := generateTestChunks(t, 5)
		totalSize := writeChunks(t, spooler, chunks)

		err = spooler.Commit()
		require.NoError(t, err, "failed to commit")

		expectedBatchSize := initialBatchSize + totalSize
		assert.Equal(t, expectedBatchSize, spooler.BatchSize(), "batch size mismatch after second file")
	})
}
