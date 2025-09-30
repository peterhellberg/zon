package zon

import (
	"fmt"
	"reflect"
)

// Unmarshal parses the data into the value pointed to by out.
// out must be a non-nil pointer.
func Unmarshal(data []byte, out interface{}) error {
	v := reflect.ValueOf(out)

	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("zon: out must be a non-nil pointer")
	}

	return safeParseValue(&parser{data: data}, v)
}

// safeParseValue executes parseValue and converts any panic into an error.
func safeParseValue(p *parser, v reflect.Value) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("zon: panic during unmarshal: %v", r)
		}
	}()

	return p.parseValue(v)
}
