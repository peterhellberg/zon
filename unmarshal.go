package zon

import (
	"fmt"
	"reflect"
)

// Unmarshal parses the data into the value pointed to by v.
// v must be a non-nil pointer.
func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("zon: v must be a non-nil pointer")
	}

	return safeParseValue(&parser{data: data}, rv)
}

// safeParseValue executes parseValue and converts any panic into an error.
func safeParseValue(p *parser, rv reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("zon: panic during unmarshal: %v", rv)
		}
	}()

	return p.parseValue(rv)
}
