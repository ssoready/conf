package conf_test

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ssoready/conf"
)

func ExampleLoad() {
	// utility function to restore os.Environ and os.Args, not needed outside of tests
	restore := restoreGlobalState()
	defer restore()

	// set up global state -- only needed because this is a unit test
	flag.CommandLine = flag.NewFlagSet("example", flag.PanicOnError)
	os.Args = []string{"example", "-Username=xxx"}
	os.Setenv("EXAMPLE_PASSWORD", "yyy")

	// this is the code you would actually write in the real world
	cfg := struct {
		Username string        `conf:"Username,noredact"`
		Password string        `conf:"Password"`
		Timeout  time.Duration `conf:"Timeout,noredact"`
	}{
		Timeout: time.Minute,
	}

	conf.Load(&cfg)
	fmt.Println(cfg) // Output: {xxx yyy 1m0s}
}

func ExampleLoad_advanced() {
	restore := restoreGlobalState()
	defer restore()

	// override flag.CommandLine to not panic, and write usage info to stdout,
	// that way we can see its output in an example function like this one
	flag.CommandLine = flag.NewFlagSet("example", flag.ContinueOnError)
	flag.CommandLine.SetOutput(os.Stdout)
	os.Args = []string{"example", "-h"}

	type subConfig struct {
		FastMode bool          `conf:"go-fast,noredact" usage:"do things faster"`
		Timeout  time.Duration `conf:"Timeout,noredact"`
	}

	cfg := struct {
		ClientConfig subConfig `conf:"ClientConfig,noredact"`
		ServerConfig subConfig `conf:"the-server-config,noredact"`
		Username     string    `conf:"my-user-name,noredact"`
	}{
		ClientConfig: subConfig{FastMode: true},
		ServerConfig: subConfig{Timeout: time.Hour},
		Username:     "jdoe",
	}

	conf.Load(&cfg)
	// Output:
	// Usage of example:
	//   -ClientConfig-Timeout duration
	//     	(env var EXAMPLE_CLIENT_CONFIG_TIMEOUT)
	//   -ClientConfig-go-fast
	//     	do things faster (env var EXAMPLE_CLIENT_CONFIG_GO_FAST) (default true)
	//   -my-user-name string
	//     	(env var EXAMPLE_MY_USER_NAME) (default "jdoe")
	//   -the-server-config-Timeout duration
	//     	(env var EXAMPLE_THE_SERVER_CONFIG_TIMEOUT) (default 1h0m0s)
	//   -the-server-config-go-fast
	//     	do things faster (env var EXAMPLE_THE_SERVER_CONFIG_GO_FAST)
}

func TestLoad(t *testing.T) {
	type subConfig struct {
		String string `conf:"String"`
	}

	type config struct {
		Bool      bool          `conf:"Bool"`
		Duration  time.Duration `conf:"Duration"`
		Float64   float64       `conf:"Float64"`
		Int       int           `conf:"Int" usage:"hello world"`
		Int64     int64         `conf:"Int64"`
		String    string        `conf:"String"`
		Uint      uint          `conf:"Uint"`
		Uint64    uint64        `conf:"Uint64"`
		SubConfig subConfig     `conf:"SubConfig"`

		StringCustomName    string    `conf:"string-custom-name-xxx"`
		SubConfigCustomName subConfig `conf:"sub-config-custom-name-xxx"`

		SkipMe string

		unexported string `conf:"unexported"`
	}

	t.Run("flags added", func(t *testing.T) {
		in := config{
			Int: 3,
			SubConfig: subConfig{
				String: "a",
			},
		}

		wantFlags := []flag.Flag{
			{Name: "Bool", Usage: "(env var CMD_BOOL)", DefValue: "false"},
			{Name: "Duration", Usage: "(env var CMD_DURATION)", DefValue: "0s"},
			{Name: "Float64", Usage: "(env var CMD_FLOAT64)", DefValue: "0"},
			{Name: "Int", Usage: "hello world (env var CMD_INT)", DefValue: "3"},
			{Name: "Int64", Usage: "(env var CMD_INT64)", DefValue: "0"},
			{Name: "String", Usage: "(env var CMD_STRING)", DefValue: ""},
			{Name: "Uint", Usage: "(env var CMD_UINT)", DefValue: "0"},
			{Name: "Uint64", Usage: "(env var CMD_UINT64)", DefValue: "0"},
			{Name: "SubConfig-String", Usage: "(env var CMD_SUB_CONFIG_STRING)", DefValue: "a"},
			{Name: "string-custom-name-xxx", Usage: "(env var CMD_STRING_CUSTOM_NAME_XXX)", DefValue: ""},
			{Name: "sub-config-custom-name-xxx-String", Usage: "(env var CMD_SUB_CONFIG_CUSTOM_NAME_XXX_STRING)", DefValue: ""},
		}

		wantFlagSet := map[flag.Flag]struct{}{}
		for _, f := range wantFlags {
			wantFlagSet[f] = struct{}{}
		}

		fs := flag.NewFlagSet("cmd", flag.ContinueOnError)
		restore := setGlobalState(nil, fs, []string{"cmd"})
		defer restore()

		conf.Load(&in)

		gotFlagSet := map[flag.Flag]struct{}{}
		fs.VisitAll(func(f *flag.Flag) {
			f.Value = nil // the "value loading" tests make sure this points to the right thing
			gotFlagSet[*f] = struct{}{}
		})

		if !reflect.DeepEqual(wantFlagSet, gotFlagSet) {
			t.Logf("len(want): %v, len(got): %v", len(wantFlagSet), len(gotFlagSet))
			t.Fatalf("want != got\nwant: %#v\ngot:  %#v", wantFlagSet, gotFlagSet)
		}
	})

	t.Run("value loading", func(t *testing.T) {
		testCases := []struct {
			Name string
			Args []string
			Env  map[string]string
			In   config
			Out  config
			Err  string
		}{
			{
				Name: "all empty",
				Args: []string{},
				Env:  map[string]string{},
				In:   config{},
				Out:  config{},
			},
			{
				Name: "set all from args",
				Args: []string{"-Bool", "-Duration=1m", "-Float64=1.0", "-Int=1", "-Int64=1", "-String=a", "-Uint=1", "-Uint64=1", "-SubConfig-String=a", "-string-custom-name-xxx=a", "-sub-config-custom-name-xxx-String=a"},
				Env:  map[string]string{},
				In:   config{},
				Out: config{
					Bool:                true,
					Duration:            time.Minute,
					Float64:             1.0,
					Int:                 1,
					Int64:               1,
					String:              "a",
					Uint:                1,
					Uint64:              1,
					SubConfig:           subConfig{String: "a"},
					StringCustomName:    "a",
					SubConfigCustomName: subConfig{String: "a"},
				},
			},
			{
				Name: "set all from env",
				Args: []string{},
				Env: map[string]string{
					"CMD_BOOL":                              "1",
					"CMD_DURATION":                          "1m",
					"CMD_FLOAT64":                           "1.0",
					"CMD_INT":                               "1",
					"CMD_INT64":                             "1",
					"CMD_STRING":                            "a",
					"CMD_UINT":                              "1",
					"CMD_UINT64":                            "1",
					"CMD_SUB_CONFIG_STRING":                 "a",
					"CMD_STRING_CUSTOM_NAME_XXX":            "a",
					"CMD_SUB_CONFIG_CUSTOM_NAME_XXX_STRING": "a",
				},
				In: config{},
				Out: config{
					Bool:                true,
					Duration:            time.Minute,
					Float64:             1.0,
					Int:                 1,
					Int64:               1,
					String:              "a",
					Uint:                1,
					Uint64:              1,
					SubConfig:           subConfig{String: "a"},
					StringCustomName:    "a",
					SubConfigCustomName: subConfig{String: "a"},
				},
			},
			{
				Name: "set from env, override by args",
				Args: []string{"-String=x"},
				Env: map[string]string{
					"CMD_STRING": "y",
				},
				In:  config{},
				Out: config{String: "x"},
			},
			{
				Name: "override with empty value from env",
				Env: map[string]string{
					"CMD_STRING": "",
				},
				In:  config{String: "a"},
				Out: config{},
			},
			{
				Name: "override with empty value from args",
				Args: []string{"-String="},
				In:   config{String: "a"},
				Out:  config{},
			},
			{
				Name: "return wrapped error from flag.Value.Set",
				Args: []string{},
				Env: map[string]string{
					"CMD_INT": "notint",
				},
				In:  config{},
				Out: config{},
				Err: `invalid value "notint" for env var CMD_INT: parse error`,
			},
			{
				Name: "return error from flag.FlagSet.Parse",
				Args: []string{"-Int=notint"},
				Env:  map[string]string{},
				In:   config{},
				Out:  config{},
				Err:  `invalid value "notint" for flag -Int: parse error`,
			},
		}

		for _, tt := range testCases {
			t.Run(tt.Name, func(t *testing.T) {
				defer func() {
					r := recover()

					errMsg := ""
					if err, ok := r.(error); ok {
						errMsg = err.Error()
					}

					if errMsg != tt.Err {
						t.Fatalf("failed to panic with expected error: %v", r)
					}
				}()

				fs := flag.NewFlagSet("some/path/to/cmd", flag.PanicOnError)
				restore := setGlobalState(tt.Env, fs, append([]string{fs.Name()}, tt.Args...))
				defer restore()

				conf.Load(&tt.In)

				if !reflect.DeepEqual(tt.Out, tt.In) {
					t.Fatalf("config: want != got, want: %#v, got: %#v", tt.Out, tt.In)
				}
			})
		}
	})
}

func setGlobalState(newEnv map[string]string, newFS *flag.FlagSet, newArgs []string) func() {
	restore := restoreGlobalState()

	os.Clearenv()
	for k, v := range newEnv {
		if err := os.Setenv(k, v); err != nil {
			panic(err)
		}
	}

	flag.CommandLine = newFS
	os.Args = newArgs

	return restore
}

func restoreGlobalState() func() {
	env := os.Environ()
	fs := flag.CommandLine
	args := os.Args

	return func() {
		for _, v := range env {
			parts := strings.SplitN(v, "=", 2)
			if err := os.Setenv(parts[0], parts[1]); err != nil {
				panic(err)
			}
		}

		flag.CommandLine = fs
		os.Args = args
	}
}
