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

	if err := marshal(reflect.ValueOf(v), &b); err != nil {
		return nil, err
	}

	_, err := b.WriteString("\n")

	return b.Bytes(), err
}

func marshal(v reflect.Value, b *bytes.Buffer) error {
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
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct:
		w(".{")
		first := true
		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			for i := 0; i < v.Len(); i++ {
				if !first {
					w(", ")
				}
				first = false
				if err := marshal(v.Index(i), b); err != nil {
					return err
				}
			}
		case reflect.Map:
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
				} else if err := marshal(k, b); err != nil {
					return err
				}
				w(" = ")
				if err := marshal(v.MapIndex(k), b); err != nil {
					return err
				}
			}
		case reflect.Struct:
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
				if err := marshal(v.Field(i), b); err != nil {
					return err
				}
			}
		}
		wb('}')
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			w("null")
		} else {
			return marshal(v.Elem(), b)
		}
	default:
		return fmt.Errorf("zon: unsupported type %s", v.Type())
	}

	return nil
}
