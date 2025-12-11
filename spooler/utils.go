package spooler

import (
	"io/fs"
	"os"
)

const filePermissions fs.FileMode = 0770

func mkdirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, filePermissions); err != nil {
			return err
		}
	}
	return nil
}
