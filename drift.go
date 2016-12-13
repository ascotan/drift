package drift

import (
  "io/ioutil"
)

func ReadMigrationFile(path string, fs filesystem) (string, error){
  f, err := fs.Open(path)
  if err != nil {
    return "", err
  }
  defer f.Close()
  out, err := ioutil.ReadAll(f)
  if err != nil {
    return "", err
  }
  return string(out), nil
}
