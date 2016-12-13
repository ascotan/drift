package drift

import (
  "os"
  "io"
)

type filesystem interface {
  Open(name string) (file, error)
  Stat(name string) (os.FileInfo, error)
}

type file interface {
  io.Closer
  io.Reader
  io.ReaderAt
  io.Seeker
  Stat() (os.FileInfo, error)
}

type OSFileSystem struct{}
func (OSFileSystem) Open(name string) (file, error) {
  return os.Open(name)
}
func (OSFileSystem) Stat(name string) (os.FileInfo, error) {
  return os.Stat(name)
}
