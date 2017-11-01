package cwl

import (
  "fmt"
  "path/filepath"
  "net/url"
)

var DOCKER_PREFIX string = "/var/spool/cwl"
type MappedInput struct {
  StoragePath string
  MappedPath string
}


type FileMapper interface {
  AdjustPath(basePath string, path string) string
}


type URLDockerMapper struct {}

func (self URLDockerMapper) AdjustPath(basePath string, path string) string {
  u, err := url.Parse(path)
  if err != nil {
      fmt.Printf("Found: %s\n", u.Path)
      return filepath.Join(DOCKER_PREFIX, filepath.Base(u.Path))
  }
  return filepath.Join(basePath, path)
}
