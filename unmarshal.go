package zon

import (
	"fmt"
	"reflect"
)

func Unmarshal(data []byte, out interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("zon: panic during unmarshal: %v", r)
		}
	}()
	p := &parser{data: data}
	v := reflect.ValueOf(out)
	if v.Kind() != reflect.Pointer || v.IsNil() {
		return fmt.Errorf("zon: out must be a non-nil pointer")
	}
	return p.parseValue(v)
}
