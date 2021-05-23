package configuration

import (
	"testing"
	"time"

	"github.com/pelletier/go-toml"
)

type testStruct struct {
	Name   string
	Inside struct {
		Beta int
	}
	Timeout time.Duration
}

func TestFindValStrByPath(t *testing.T) {
	var testObjFromToml interface{}
	data, _ := toml.Marshal(testStruct{
		Name:   "test",
		Inside: struct{ Beta int }{Beta: 42},
	})
	_ = toml.Unmarshal(data, &testObjFromToml)

	tests := []struct {
		name         string
		input        interface{}
		path         []string
		expectedStr  string
		expectedBool bool
	}{
		{
			name:         "empty path",
			path:         nil,
			expectedStr:  "",
			expectedBool: false,
		},
		{
			name:         "at root level | Name | yaml",
			input:        testObjFromToml,
			path:         []string{"Name"},
			expectedStr:  "test",
			expectedBool: true,
		},
		{
			name:         "substructures | Inside.Beta | yaml",
			input:        testObjFromToml,
			path:         []string{"Inside", "Beta"},
			expectedStr:  "42",
			expectedBool: true,
		},
	}

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			gotStr, gotBool := findValStrByPath(testObjFromToml, test.path)
			if gotStr != test.expectedStr || gotBool != test.expectedBool {
				t.Fatalf("expected: [%q %v] but got [%q %v]", test.expectedStr, test.expectedBool, gotStr, gotBool)
			}
		})
	}
}
