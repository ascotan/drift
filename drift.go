// This package provides sql migration capabilities
package drift

import (
  "log"
  "io/ioutil"
)

type revision struct {
  data  []byte
  path  string
}

type changeset struct {
  headers string
  sql     string
}

// Reads a file from a path and parses the file into a revision struct
// the filesystem argument represents a generic filesystem
// OSFileSystem is an implemention of the local filesystem
// this can be used by ReadRevision('mypath', OSFileSystem{})
func ReadRevision(path string, fs filesystem) (*revision, error){
  f, err := fs.Open(path)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  out, err := ioutil.ReadAll(f)
  if err != nil {
    return nil, err
  }
  return &revision{out, path}, nil
}

// parses the changesets out of a revision file
func ParseChangesets(rev *revision) ([]changeset, error){
  s := NewScanner(rev.data)
  for s.HasMoreTokens() {
    tok, err := s.NextToken()
    if err != nil {
      log.Println(err)
    } else {
      log.Println(string(tok.runes), " : ", s.offset)
    }
  }
  return nil, nil
}
