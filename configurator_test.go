package configuration

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfigurator(t *testing.T) {
	// setting command line flag
	os.Args = []string{"smth", "-name=flag_value"}

	// setting env variable
	removeEnvKey, err := setEnv("AGE_ENV", "45")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer removeEnvKey()

	// defining a struct
	cfg := struct {
		Name     string `flag:"name"`
		LastName string `default:"defaultLastName"`
		Age      byte   `env:"AGE_ENV"`
		BoolPtr  *bool  `default:"false"`

		ObjPtr *struct {
			F32       float32       `default:"32"`
			StrPtr    *string       `default:"str_ptr_test"`
			HundredMS time.Duration `default:"100ms"`
		}

		Obj struct {
			IntPtr   *int16   `default:"123"`
			NameYML  int      `default:"24"`
			StrSlice []string `default:"one;two"`
			IntSlice []int64  `default:"3; 4"`
		}
	}{}

	fileProvider, err := NewFileProvider("./testdata/input.toml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configurator, err := New(&cfg,
		// order of execution will be kept:
		NewFlagProvider(&cfg), // 1st
		NewEnvProvider(),      // 2nd
		fileProvider,          // 3rd
		NewDefaultProvider(),  // 4th
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configurator.InitValues()

	assert.Equal(t, "flag_value", cfg.Name)
	assert.Equal(t, "defaultLastName", cfg.LastName)
	assert.Equal(t, byte(45), cfg.Age)
	assert.NotNil(t, cfg.BoolPtr)
	assert.Equal(t, false, *cfg.BoolPtr)

	assert.NotNil(t, cfg.ObjPtr)
	assert.Equal(t, float32(32), cfg.ObjPtr.F32)
	assert.NotNil(t, cfg.ObjPtr.StrPtr)
	assert.Equal(t, "str_ptr_test", *cfg.ObjPtr.StrPtr)

	assert.NotNil(t, cfg.Obj.IntPtr)
	assert.Equal(t, int16(123), *cfg.Obj.IntPtr)
	assert.Equal(t, int(42), cfg.Obj.NameYML)
	assert.Equal(t, []string{"one", "two"}, cfg.Obj.StrSlice)
	assert.Equal(t, []int64{3, 4}, cfg.Obj.IntSlice)
	assert.Equal(t, time.Millisecond*100, cfg.ObjPtr.HundredMS)
}

func TestConfigurator_Errors(t *testing.T) {
	tests := map[string]struct {
		input     interface{}
		providers []Provider
	}{
		"empty providers": {
			input:     &struct{}{},
			providers: []Provider{},
		},
		"non-pointer": {
			input: struct{}{},
			providers: []Provider{
				NewDefaultProvider(),
			},
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			_, err := New(test.input, test.providers...)
			if err == nil {
				t.Fatal("expected error but got nil")
			}
		})
	}
}

func TestEmbeddedFlags(t *testing.T) {
	type (
		Client struct {
			ServerAddress string `flag:"addr|127.0.0.1:443|server address"`
		}
		Config struct {
			Client *Client
		}
	)
	os.Args = []string{"smth", "-addr=addr_value"}

	var cfg Config
	c, err := New(&cfg, NewFlagProvider(&cfg))
	if err != nil {
		t.Fatal("unexpected err: ", err)
	}

	c.InitValues()

	assert.NotNil(t, cfg.Client)
	assert.Equal(t, cfg.Client.ServerAddress, "addr_value")
}

func TestSetLogger(t *testing.T) {
	var (
		cfg = struct {
			Name string `default:"test_name"`
		}{}
		logs  []string // collects log output into slice
		logFn = func(format string, v ...interface{}) {
			logs = append(logs, fmt.Sprintf(format, v...))
		}
		expectedLogs = []string{
			"configurator: current path: [Name]",
			"configurator: envProvider: key is empty",
		}
	)

	c, err := New(
		&cfg,
		NewEnvProvider(),
		NewDefaultProvider(),
	)
	if err != nil {
		t.Fatal("unexpected err: ", err)
	}

	c.SetLogger(logFn)
	c.EnableLogging(true)
	c.InitValues()

	assert.Equal(t, cfg.Name, "test_name")
	assert.Equal(t, expectedLogs, logs)
}

func TestFallBackToDefault(t *testing.T) {
	// defining a struct
	cfg := struct {
		NameFlag string `flag:"name_flag||Some description"   default:"default_val"`
	}{}

	configurator, err := New(&cfg,
		NewFlagProvider(&cfg),
		NewDefaultProvider(),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	configurator.EnableLogging(true)
	configurator.InitValues()

	assert.Equal(t, "default_val", cfg.NameFlag)
}
