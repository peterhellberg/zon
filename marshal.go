package zon

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Marshal(v interface{}) ([]byte, error) {
	var b bytes.Buffer
	if err := enc(reflect.ValueOf(v), &b); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func enc(v reflect.Value, b *bytes.Buffer) error {
	w, wb := b.WriteString, b.WriteByte

	if !v.IsValid() {
		w("null")
		return nil
	}

	switch v.Kind() {
	case reflect.Bool:
		w(strconv.FormatBool(v.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		w(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		w(strconv.FormatFloat(v.Float(), 'g', -1, 64))
	case reflect.String:
		wb('"')
		w(v.String())
		wb('"')
	case reflect.Slice, reflect.Array:
		wb('[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				w(", ")
			}
			if err := enc(v.Index(i), b); err != nil {
				return err
			}
		}
		wb(']')
	case reflect.Map, reflect.Struct:
		w(".{") // <-- leading dot before opening brace
		first := true
		if v.Kind() == reflect.Map {
			for _, k := range v.MapKeys() {
				if !first {
					w(", ")
				}
				first = false
				if k.Kind() == reflect.String {
					s := k.String()
					if !strings.HasPrefix(s, ".") {
						wb('.')
					}
					w(s)
				} else if err := enc(k, b); err != nil {
					return err
				}
				w(" = ")
				if err := enc(v.MapIndex(k), b); err != nil {
					return err
				}
			}
		} else {
			for i := 0; i < v.NumField(); i++ {
				f := v.Type().Field(i)
				if f.PkgPath != "" {
					continue
				}
				if !first {
					w(", ")
				}
				first = false
				name := f.Tag.Get("zon")
				if name == "" {
					name = f.Name
				}
				wb('.')
				w(name)
				w(" = ")
				if err := enc(v.Field(i), b); err != nil {
					return err
				}
			}
		}
		wb('}')
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			w("null")
		} else {
			return enc(v.Elem(), b)
		}
	default:
		return fmt.Errorf("zon: unsupported type %s", v.Type())
	}

	return nil
}
