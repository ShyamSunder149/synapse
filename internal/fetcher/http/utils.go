package http

import "io"

type readCloser struct {
	io.Reader
	closer io.Closer
}

func (rc *readCloser) Close() error {
	if rc.closer != nil {
		return rc.closer.Close()
	}
	return nil
}
