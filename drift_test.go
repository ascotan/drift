package drift

import (
  "testing"
  "os"
  "strings"
  "errors"
  "fmt"
  "time"
  "path/filepath"
)

// ----------------------------------------------------------------------------
// filesystem mock
// ----------------------------------------------------------------------------
// mock file
type mockFile struct {
  path  string
  data  *strings.Reader
  info  *mockFileInfo
}
func newMockFile(data string, path string, mode os.FileMode) *mockFile {
    return &mockFile{
      path,
      strings.NewReader(data),
      &mockFileInfo {
        name:    filepath.Base(path),
        size:    int64(len([]byte(data))),
        mode:    mode,
        modtime: time.Now(),
        isdir:   false,
        sys:     nil,
      },
    }
}
func (m *mockFile) Path() (string) { return m.path }
func (m *mockFile) Close() error { return nil }
func (m *mockFile) Read(p []byte) (n int, err error) { return m.data.Read(p) }
func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
  return m.data.ReadAt(p, off)
}
func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
  return m.data.Seek(offset, whence)
}
func (m *mockFile) Stat() (os.FileInfo, error) { return m.info, nil }

// mock file properties
type mockFileInfo struct {
  name    string
  size    int64
  mode    os.FileMode
  modtime time.Time
  isdir   bool
  sys     interface{}
}
func (m *mockFileInfo) Name() string { return m.name }
func (m *mockFileInfo) Size() int64{ return m.size }
func (m *mockFileInfo) Mode() os.FileMode { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modtime }
func (m *mockFileInfo) IsDir() bool { return m.isdir }
func (m *mockFileInfo) Sys() interface{} { return m.sys }

// mock filesystem
type mockFS struct{ files  map[string]file }
func newMockFS(files ...*mockFile) *mockFS {
  m := make(map[string]file)
  for _, f := range files {
    m[f.Path()] = f
  }
  return &mockFS{m}
}
func (m *mockFS) Open(name string) (file, error) {
  val, exists := m.files[name]
  if !exists {
    return nil, errors.New(fmt.Sprintf("%s: no such file or directory", name))
  }
  return val, nil
}
func (m *mockFS) Stat(name string) (os.FileInfo, error) {
  val, exists := m.files[name]
  if !exists {
    return nil, errors.New(fmt.Sprintf("%s: unable to stat file", name))
  }
  info, err := val.Stat()
  if err != nil {
    return nil, err
  }
  return info, nil
}
// ---------------------------------------------------------------------------
// tests
// ----------------------------------------------------------------------------

func TestReadFile(t *testing.T) {
  changeset1 := `--- +changeset id:hello kitty author:jgilbert dbms:ql runalways:true, runonchange:true, failonerror:true
  --- + preconditions dbms:ql tableexists:tablename colexists:colname fkexists:fkname indexexists:indexname
  --- + precondition-sql-check expectedResult:0 select count(*) from mytable
  --- + precondition-sql-check expectedResult:0 select count(*) from mytable
  --- + precondition-sql-check expectedResult:0 select count(*) from mytable
  --- + rollback DROP TABLE xxx;
  CREATE TABLE xxx;`

  fs := newMockFS(newMockFile(changeset1, "/tmp/file", 0644))
  data, err := ReadMigrationFile("/tmp/file", fs)
  if err != nil {
    t.Error(err)
  }
  if data != changeset1 {
    t.Error("Data returned is different than expected")
  }
}
