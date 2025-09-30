package zon

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeValue(reflect.ValueOf(v), &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeValue(v reflect.Value, buf *bytes.Buffer) error {
	if !v.IsValid() {
		buf.WriteString("null")
		return nil
	}

	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf.WriteString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		buf.WriteString(strconv.FormatFloat(v.Float(), 'g', -1, 64))
	case reflect.String:
		buf.WriteByte('"')
		buf.WriteString(v.String())
		buf.WriteByte('"')
	case reflect.Slice, reflect.Array:
		buf.WriteByte('[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				buf.WriteString(", ")
			}
			if err := encodeValue(v.Index(i), buf); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case reflect.Map:
		buf.WriteByte('{')
		keys := v.MapKeys()
		for i, k := range keys {
			if i > 0 {
				buf.WriteString(", ")
			}
			if k.Kind() == reflect.String {
				keyStr := k.String()
				if !strings.HasPrefix(keyStr, ".") {
					buf.WriteByte('.')
				}
				buf.WriteString(keyStr)
			} else {
				if err := encodeValue(k, buf); err != nil {
					return err
				}
			}
			buf.WriteString(" = ")
			if err := encodeValue(v.MapIndex(k), buf); err != nil {
				return err
			}
		}
		buf.WriteByte('}')

	case reflect.Struct:
		buf.WriteByte('{')
		t := v.Type()
		first := true
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			val := v.Field(i)
			if field.PkgPath != "" { // skip unexported
				continue
			}
			if !first {
				buf.WriteString(", ")
			}
			first = false
			name := field.Tag.Get("zon")
			if name == "" {
				name = field.Name
			}
			buf.WriteByte('.')
			buf.WriteString(name)
			buf.WriteString(" = ")
			if err := encodeValue(val, buf); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			buf.WriteString("null")
		} else {
			return encodeValue(v.Elem(), buf)
		}
	default:
		return fmt.Errorf("zon: unsupported type %s", v.Type())
	}
	return nil
}
