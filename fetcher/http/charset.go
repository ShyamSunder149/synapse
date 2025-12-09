package http

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/net/html/charset"
)

const defaultPeekSize = 1024

type CharsetMetadata struct {
	MIMEType string
	Charset  string
}

// Automatically detects the charset and converts to UTF-8 if needed.
// Returns the original reader unchanged if the content is non-textual.
func newUTF8WithFallbackReader(resp *http.Response, defaultEncoding string) (io.ReadCloser, error) {

	metadata, err := detectCharset(resp, defaultEncoding)
	if err != nil {
		return nil, err
	}

	if metadata == nil {
		return resp.Body, nil
	}

	utf8Reader, err := newCharsetReader(resp.Body, metadata.MIMEType, metadata.Charset)
	if err != nil {
		return nil, err
	}

	return &readCloser{
		Reader: utf8Reader,
		closer: resp.Body,
	}, nil
}

// SAFETY: Ensure that the `defaultEncoding` is valid at the caller side, if provided.
func detectCharset(resp *http.Response, defaultEncoding string) (*CharsetMetadata, error) {

	detectCharsetFromBody := func(resp *http.Response) (string, error) {
		buf := make([]byte, defaultPeekSize)
		n, err := io.ReadFull(resp.Body, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return "", err
		}

		buf = buf[:n]
		detector := chardet.NewTextDetector()
		results, err := detector.DetectAll(buf)
		if err != nil {
			return "", err
		}

		// Restore the body
		resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(buf), resp.Body))

		if len(results) == 0 {
			return "", nil
		}

		return results[0].Charset, nil
	}

	isTextualContent := func(mimeType string) bool {
		switch {
		case strings.HasPrefix(mimeType, "image/"),
			strings.HasPrefix(mimeType, "video/"),
			strings.HasPrefix(mimeType, "audio/"),
			strings.HasPrefix(mimeType, "font/"):
			return false
		default:
			return true
		}
	}

	if resp.ContentLength == 0 {
		return nil, errors.New("empty response body")
	}

	// 1. Try detecting from Content-Type header
	contentType := resp.Header.Get("Content-Type")
	mimeType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// mimeType = "text/plain"
		return nil, err
	}
	detectedCharset := strings.ToLower(params["charset"])

	if !isTextualContent(mimeType) {
		return nil, nil
	}

	if defaultEncoding != "" {
		detectedCharset = defaultEncoding
	}

	// 2. Determine charset from body
	if detectedCharset == "" {
		detectedCharset, err = detectCharsetFromBody(resp)
		if err != nil {
			return nil, err
		}
	}

	return &CharsetMetadata{
		MIMEType: "text/plain",
		Charset:  detectedCharset,
	}, nil
}

// Converts from the given charset to UTF-8.
func newCharsetReader(r io.Reader, mimeType, charsetName string) (io.Reader, error) {
	if strings.EqualFold(charsetName, "utf-8") || charsetName == "" {
		return r, nil
	}

	contentType := mimeType + "; charset=" + charsetName
	return charset.NewReader(r, contentType)
}
