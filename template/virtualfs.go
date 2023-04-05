package main

import (
	"io/fs"
	"os"
)

// virtualFS implements io/fs.FS to enable text/template.ParseFS to receive
// os.Stdin as input.
type virtualFS struct{}

var _ fs.FS = (*virtualFS)(nil)

// Open always returns os.Stdin.
func (_ virtualFS) Open(_ string) (fs.File, error) {
	return virtualFile{os.Stdin}, nil
}

// virtualFile is io/fs.File wrapper that overrides io/fs.File.Close.
type virtualFile struct{ fs.File }

// Close is NOP and always return nil to avoid closing the underlying file.
func (_ virtualFile) Close() error {
	return nil
}
