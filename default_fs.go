package testfixtures

import (
	"io/fs"
	"os"
	"path/filepath"
)

type defaultFS struct{}

func (defaultFS) Open(name string) (fs.File, error) {
	return os.Open(filepath.FromSlash(name))
}
