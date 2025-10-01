package zon

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func Marshal(v any) ([]byte, error) {
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
		s := v.String()
		if isHexLiteral(s) {
			w(s)
		} else {
			wb('"')
			w(s)
			wb('"')
		}
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

				tag := f.Tag.Get("zon")

				var name string
				var omitempty bool

				if tag == "" {
					name = f.Name
				} else {
					parts := strings.Split(tag, ",")
					if parts[0] != "" {
						name = strings.TrimSpace(parts[0])
					} else {
						name = f.Name
					}

					for _, instr := range parts[1:] {
						if strings.TrimSpace(instr) == "omitempty" {
							omitempty = true
						}
					}
				}

				fv := v.Field(i)

				if omitempty && isEmptyValue(fv) {
					continue // skip empty field
				}

				if !first {
					w(", ")
				}

				first = false

				wb('.')
				w(name)
				w(" = ")

				if err := marshal(fv, b); err != nil {
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

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isEmptyValue(v.Field(i)) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

func isHexLiteral(s string) bool {
	if len(s) < 3 {
		return false
	}

	if !(strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X")) {
		return false
	}

	for _, c := range s[2:] {
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {

			return false
		}
	}

	return true
}
