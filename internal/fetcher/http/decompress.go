package http

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

func DecompressResponse(resp *http.Response) error {
	if resp.Body == nil {
		return nil
	}

	encoding := strings.ToLower(strings.TrimSpace(resp.Header.Get(HeaderContentEncoding)))
	if encoding == "" || encoding == "identity" {
		return nil
	}

	var reader io.ReadCloser
	var err error

	switch encoding {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}

	case "br":
		brotliReader := brotli.NewReader(resp.Body)
		reader = &readCloser{
			Reader: brotliReader,
			closer: resp.Body,
		}

	case "zstd":
		zstdReader, err := zstd.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create zstd reader: %w", err)
		}
		reader = &zstdReadCloser{
			Decoder: zstdReader,
			body:    resp.Body,
		}

	case "deflate":
		reader = flate.NewReader(resp.Body)

	default:
		return fmt.Errorf("unsupported content encoding: %s", encoding)
	}

	resp.Body = reader
	resp.Header.Del(HeaderContentEncoding)
	resp.Uncompressed = true

	return nil
}

type zstdReadCloser struct {
	*zstd.Decoder
	body io.Closer
}

func (z *zstdReadCloser) Close() error {
	z.Decoder.Close()
	return z.body.Close()
}
