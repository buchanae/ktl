package cwl

import (
	"fmt"
	"github.com/ohsu-comp-bio/ktl/pbutil"
	"log"
	"net/url"
	"path/filepath"
	"strings"
)

var DOCKER_INPUT_DIR string = "/var/spool/cwl"
var DOCKER_WORK_DIR string = "/var/run/cwl"
var DOCKER_LOG_DIR string = "/var/log/cwl"

type MappedInput struct {
	StoragePath string
	MappedPath  string
	Type        string
}

type FileMapper interface {
	Input2Storage(path string) string
	Storage2Volume(path string) string
	Volume2Storage(path string) string
	StoragePath(path string) string
}

type URLDockerMapper struct {
	StorageBase string
}

func NewFileMapper(path string) FileMapper {
	u, err := url.Parse(path)
	if err == nil {
		if u.Scheme == "" {
			a, _ := filepath.Abs(path)
			return URLDockerMapper{fmt.Sprintf("file://%s", a)}
		}
		if u.Scheme != "file" {
			return URLDockerMapper{path}
		}
	}
	return URLDockerMapper{path}
}

func (self URLDockerMapper) Input2Storage(path string) string {
	//TODO: do object storeage movement
	u, err := url.Parse(path)
	if err == nil {
		if u.Scheme == "" {
			return fmt.Sprintf("file://%s", path)
		}
		if u.Scheme != "file" {
			return path
		}
	}
	return path
}

func (self URLDockerMapper) Storage2Volume(path string) string {
	return filepath.Join(DOCKER_INPUT_DIR, filepath.Base(path))
}

func (self URLDockerMapper) Volume2Storage(path string) string {
	u, _ := url.Parse(self.StorageBase)
	u.Path = filepath.Join(u.Path, "output", filepath.Base(path))
	return u.String()
}

func (self URLDockerMapper) StoragePath(path string) string {
	u, _ := url.Parse(self.StorageBase)
	u.Path = filepath.Join(u.Path, path)
	return u.String()
}

/*
	Take URL qualified input json and create dictionary with 'path' entries
	that can be used by Javascript interpteter and command line builder
*/
func SetInputVolumePath(input interface{}, mapper FileMapper) interface{} {
	if base, ok := input.(pbutil.JSONDict); ok {
		out := pbutil.JSONDict{}
		if class, ok := base["class"]; ok {
			if class == "File" {
				for k, v := range base {
					if k == "url" {
						out["path"] = mapper.Storage2Volume(v.(string))
						out["url"] = v.(string)
					} else {
						out[k] = v
					}
				}
			} else if class == "Directory" {
				for k, v := range base {
					if k == "url" {
						out["path"] = mapper.Storage2Volume(v.(string))
						out["url"] = v.(string)
					} else {
						out[k] = v
					}
				}
			} else {
				log.Printf("Unknown class type: %s", class)
			}
		} else {
			for k, v := range base {
				out[k] = SetInputVolumePath(v, mapper)
			}
		}
		return out
	} else if base, ok := input.([]interface{}); ok {
		out := []interface{}{}
		for _, i := range base {
			out = append(out, SetInputVolumePath(i, mapper))
		}
		return out
	}
	return input
}

func abspath(basePath, p string) string {
	if strings.HasPrefix(p, "/") {
		return p
	}
	a, _ := filepath.Abs(filepath.Join(basePath, p))
	return a
}

/*
	Convert user input JSON to full path qualified entries in the 'location' field
*/
func SetInputAbsPath(input interface{}, basePath string) interface{} {
	if base, ok := input.(pbutil.JSONDict); ok {
		out := pbutil.JSONDict{}
		if class, ok := base["class"]; ok {
			if class == "File" {
				for k, v := range base {
					if k == "path" {
						out["location"] = abspath(basePath, v.(string))
					} else if k == "location" {
						out["location"] = abspath(basePath, v.(string))
					} else {
						out[k] = v
					}
				}
			} else if class == "Directory" {
				for k, v := range base {
					if k == "path" {
						out["location"] = abspath(basePath, v.(string))
					} else if k == "location" {
						out["location"] = abspath(basePath, v.(string))
					} else {
						out[k] = v
					}
				}
			} else {
				log.Printf("Unknown class type: %s", class)
			}
		} else {
			for k, v := range base {
				out[k] = SetInputAbsPath(v, basePath)
			}
		}
		return out
	} else if base, ok := input.([]interface{}); ok {
		out := []interface{}{}
		for _, i := range base {
			out = append(out, SetInputAbsPath(i, basePath))
		}
		return out
	}
	return input
}

/*
	Convert qualified 'location' entries and convert them into URLs if needed
*/
func SetInputUrl(input interface{}, mapper FileMapper) interface{} {
	if base, ok := input.(pbutil.JSONDict); ok {
		out := pbutil.JSONDict{}
		if class, ok := base["class"]; ok {
			if class == "File" {
				for k, v := range base {
					if k == "location" {
						out["url"] = mapper.Input2Storage(v.(string))
					} else {
						out[k] = v
					}
				}
			} else if class == "Directory" {
				for k, v := range base {
					if k == "location" {
						out["url"] = mapper.Input2Storage(v.(string))
					} else {
						out[k] = v
					}
				}
			} else {
				log.Printf("Unknown class type: %s", class)
			}
		} else {
			for k, v := range base {
				out[k] = SetInputUrl(v, mapper)
			}
		}
		return out
	} else if base, ok := input.([]interface{}); ok {
		out := []interface{}{}
		for _, i := range base {
			out = append(out, SetInputUrl(i, mapper))
		}
		return out
	}
	return input
}

func GetFileInputs(input interface{}) []MappedInput {
	log.Printf("InputMapping: %s", input)
	out := []MappedInput{}
	if base, ok := input.(pbutil.JSONDict); ok {
		if class, ok := base["class"]; ok {
			if class == "File" {
				out = append(out, MappedInput{
					StoragePath: base["url"].(string),
					MappedPath:  base["path"].(string),
					Type:        "File",
				})
			} else if class == "Directory" {
				out = append(out, MappedInput{
					StoragePath: base["url"].(string),
					MappedPath:  base["path"].(string),
					Type:        "Directory",
				})
			}
		} else {
			for _, v := range base {
				o := GetFileInputs(v)
				out = append(out, o...)
			}
		}
		return out

	} else if base, ok := input.([]interface{}); ok {
		for _, i := range base {
			o := GetFileInputs(i)
			out = append(out, o...)
		}
	}
	return out
}
