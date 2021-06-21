package conf

import (
	"fmt"
	"reflect"
)

// Redact returns a copy of v with all fields zeroed out unless they are marked
// as "noredact". Redact does not modify v.
//
// A field in v will be copied over as-is only if it is exported and has a
// "conf" tag of the form:
//
//	Foo string `conf:"foo,noredact"`
//
// This is the same syntax Load supports. If ",noredact" is not present, then
// Redact will not copy the field over. Unexported fields will always be zero.
//
// Redact will recursively apply to struct-valued fields in v only if that field
// is marked as "noredact". Otherwise, the entire field is left as zero.
func Redact[T any](v T) T {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		panic(fmt.Errorf("conf: Redact called on %v (only structs are acceptable)", val.Kind()))
	}

	return redact(val).Interface().(T)
}

func redact(v reflect.Value) reflect.Value {
	t := v.Type()

	if t.Kind() != reflect.Struct {
		return v
	}

	out := reflect.New(t)
	for i := 0; i < t.NumField(); i++ {
		_, noredact := parseConfTag(t.Field(i).Tag.Get(tagConf))
		if !noredact || !out.Elem().Field(i).CanSet() {
			continue
		}

		out.Elem().Field(i).Set(redact(v.Field(i)))
	}

	return out.Elem()
}
