package configuration

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml"
)

// NewFileProvider creates new provider which read values from files (toml)
func NewFileProvider(fileName string) (fp fileProvider, err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return fp, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return fp, err
	}

	if err := toml.Unmarshal(b, &fp.fileData); err != nil {
		return fp, err
	}
	return
}

type fileProvider struct {
	fileData interface{}
}

func (fp fileProvider) Provide(field reflect.StructField, v reflect.Value, path ...string) error {
	valStr, ok := findValStrByPath(fp.fileData, path)
	if !ok {
		return fmt.Errorf("fileProvider: findValStrByPath returns empty value")
	}

	return SetField(field, v, valStr)
}

func findValStrByPath(i interface{}, path []string) (string, bool) {
	if len(path) == 0 {
		return "", false
	}
	firstInPath := strings.ToLower(path[0])

	currentFieldStr, ok := i.(map[string]interface{})
	if !ok {
		return "", false
	}

	for k, v := range currentFieldStr {
		currentFieldStr[strings.ToLower(k)] = v
	}

	if len(path) == 1 {
		val, ok := currentFieldStr[firstInPath]
		if !ok {
			return "", false
		}
		return fmt.Sprint(val), true
	}

	return findValStrByPath(currentFieldStr[firstInPath], path[1:])
}
