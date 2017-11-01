package cwl

import (
	"fmt"
	"net/url"
	"path/filepath"
)

var DOCKER_INPUT_DIR string = "/var/spool/cwl"
var DOCKER_WORK_DIR string = "/var/run/cwl"
var DOCKER_LOG_DIR string = "/var/log/cwl"

type MappedInput struct {
	StoragePath string
	MappedPath  string
}

type FileMapper interface {
	Input2Storage(basePath string, path string) string
	Storage2Volume(path string) string
	Volume2Storage(path string) string
}

type URLDockerMapper struct {
	StorageBase string
}

func (self URLDockerMapper) Input2Storage(basePath string, path string) string {
	u, err := url.Parse(path)
	if err == nil {
		if u.Scheme != "file" {
			return path
		}
	}
	fmt.Printf("Found %s", path)
	abs, _ := filepath.Abs(filepath.Join(basePath, path))
	return abs
}

func (self URLDockerMapper) Storage2Volume(path string) string {
	return filepath.Join(DOCKER_INPUT_DIR, filepath.Base(path))
}

func (self URLDockerMapper) Volume2Storage(path string) string {
	u, _ := url.Parse(self.StorageBase)
	u.Path = filepath.Join(u.Path, "output", filepath.Base(path))

	return u.String()
}
