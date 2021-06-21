package conf_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ucarion/conf"
)

func ExampleRedact() {
	cfg := struct {
		Username      string `conf:"username,noredact"`
		Password      string `conf:"password"`
		InnocentBytes []byte `conf:"innocent-bytes,noredact"`
		SecretBytes   []byte `conf:"secret-bytes"`
	}{
		Username:      "jdoe",
		Password:      "iloveyou",
		InnocentBytes: []byte{1, 2, 3},
		SecretBytes:   []byte{4, 5, 6},
	}

	fmt.Println(conf.Redact(cfg)) // returns a deep copy with redacted fields set to zero
	fmt.Println(cfg)              // original copy unmodified

	// Output:
	// {jdoe  [1 2 3] []}
	// {jdoe iloveyou [1 2 3] [4 5 6]}
}

func TestRedact(t *testing.T) {
	type subConfig struct {
		S1 string `conf:""`
		S2 string `conf:""`
		S3 string `conf:",noredact"`
		S4 string `conf:",noredact"`
		S5 string `conf:",noredact"`
	}

	type config struct {
		S1 string `conf:""`
		S2 string `conf:""`
		S3 string `conf:",noredact"`
		S4 string `conf:",noredact"`
		S5 string `conf:",noredact"`

		Bool      bool
		Int       int
		Array     [64]byte
		Chan      chan string
		Func      func()
		Interface interface{}
		Map       map[bool]bool

		StringPtr *string
		ChanPtr   *chan string

		SubConfig       subConfig `conf:",noredact"`
		SubConfigRedact subConfig

		unexported string
	}

	s := "a"
	c := make(chan string)
	got := conf.Redact(config{
		S1:        "a",
		S2:        "a",
		S3:        "a",
		S4:        "a",
		S5:        "a",
		Bool:      true,
		Int:       1,
		Array:     [64]byte{1},
		Chan:      make(chan string),
		Func:      func() {},
		Interface: 1,
		Map:       map[bool]bool{true: true},
		StringPtr: &s,
		ChanPtr:   &c,
		SubConfig: subConfig{
			S1: "a",
			S2: "a",
			S3: "a",
			S4: "a",
			S5: "a",
		},
		SubConfigRedact: subConfig{
			S1: "a",
			S2: "a",
			S3: "a",
			S4: "a",
			S5: "a",
		},
	})

	want := config{
		S1: "",
		S2: "",
		S3: "a",
		S4: "a",
		S5: "a",
		SubConfig: subConfig{
			S1: "",
			S2: "",
			S3: "a",
			S4: "a",
			S5: "a",
		},
	}

	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(config{})); diff != "" {
		t.Fatalf("Redact(%v) != %v (+want -got)\n%s ", want, got, diff)
	}
}

func TestRedact_Panic_Invalid_Kind(t *testing.T) {
	defer func() {
		r := recover()
		if r.(error).Error() != "conf: Redact called on ptr (only structs are acceptable)" {
			t.Fatalf("failed to panic with expected error: %v", r)
		}
	}()

	conf.Redact(&struct{}{})
}
