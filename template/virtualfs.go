package main

import (
	"io/fs"
	"os"
)

// virtualFS encapsulates the stdin to enable text/template lib to receive it as input.
// Consequently, delegating file read handling to the text/template lib
type virtualFS struct{}

var _ fs.FS = (*virtualFS)(nil)

// Open always returns stdin
func (v virtualFS) Open(_ string) (fs.File, error) {
	return virtualFile{os.Stdin}, nil
}

type virtualFile struct{ fs.File }

var _ fs.File = (*virtualFile)(nil)

func (v virtualFile) Stat() (fs.FileInfo, error) {
	return v.File.Stat()
}

func (v virtualFile) Read(bytes []byte) (int, error) {
	return v.File.Read(bytes)
}

// Close avoid closing the stdin since the text/template lib will try to close the stdin.
// stdin cannot be closed since it can only be opened once per execution
func (v virtualFile) Close() error {
	return nil
}
