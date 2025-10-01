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

	if err := marshal(reflect.ValueOf(v), &b, 0); err != nil {
		return nil, err
	}

	_ = b.WriteByte('\n')

	return b.Bytes(), nil
}

func marshal(v reflect.Value, b *bytes.Buffer, indent int) error {
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

		if isDotLiteral(s) || isHexLiteral(s) {
			w(s)
		} else {
			wb('"')
			w(s)
			wb('"')
		}
	case reflect.Slice, reflect.Array:
		w(".{\n")

		for i := 0; i < v.Len(); i++ {
			writeIndent(b, indent+1)

			if err := marshal(v.Index(i), b, indent+1); err != nil {
				return err
			}

			w(",\n")
		}

		writeIndent(b, indent)

		wb('}')
	case reflect.Map:
		w(".{\n")

		keys := v.MapKeys()

		for i, k := range keys {
			writeIndent(b, indent+1)

			if k.Kind() == reflect.String {
				s := k.String()
				if !strings.HasPrefix(s, ".") {
					wb('.')
				}

				w(s)
			} else if err := marshal(k, b, indent+1); err != nil {
				return err
			}

			w(" = ")

			value := v.MapIndex(k)

			if err := marshal(value, b, indent+1); err != nil {
				return err
			}

			w(",")

			if i != len(keys)-1 {
				w("\n")
			} else {
				w("\n")
			}
		}

		writeIndent(b, indent)

		wb('}')
	case reflect.Struct:
		w(".{\n")

		first := true

		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)

			if f.PkgPath != "" {
				continue
			}

			var (
				fv        = v.Field(i)
				tag       = f.Tag.Get("zon")
				name      = f.Name
				omitempty = false
			)

			if tag != "" {
				parts := strings.Split(tag, ",")
				if parts[0] != "" {
					name = strings.TrimSpace(parts[0])
				}

				for _, instr := range parts[1:] {
					if strings.TrimSpace(instr) == "omitempty" {
						omitempty = true
					}
				}
			}

			if omitempty && isEmptyValue(fv) {
				continue
			}

			if !first {
				w("\n")
			}
			first = false

			writeIndent(b, indent+1)

			wb('.')
			w(name)
			w(" = ")

			if err := marshal(fv, b, indent+1); err != nil {
				return err
			}

			w(",")
		}

		w("\n")

		writeIndent(b, indent)

		wb('}')
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			w("null")
		} else {
			return marshal(v.Elem(), b, indent)
		}
	default:
		return fmt.Errorf("zon: unsupported type %s", v.Type())
	}

	return nil
}

func writeIndent(b *bytes.Buffer, indent int) {
	for i := 0; i < indent; i++ {
		b.WriteString("    ")
	}
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

func isDotLiteral(s string) bool {
	if len(s) < 2 || s[0] != '.' {
		return false
	}

	for _, c := range s[1:] {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '_') {
			return false
		}
	}

	return true
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
