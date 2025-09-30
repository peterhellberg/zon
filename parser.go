package zon

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type parser struct {
	data []byte
	pos  int
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.data) && unicode.IsSpace(rune(p.data[p.pos])) {
		p.pos++
	}
}

func (p *parser) parseValue(v reflect.Value) error {
	p.skipWhitespace()

	if p.pos >= len(p.data) {
		return fmt.Errorf("zon: unexpected end of input")
	}

	if bytes.HasPrefix(p.data[p.pos:], []byte("null")) {
		p.pos += 4

		if v.CanSet() {
			v.Set(reflect.Zero(v.Type()))
		}

		return nil
	}

	for v.Kind() == reflect.Pointer {
		if v.IsNil() && v.CanSet() {
			v.Set(reflect.New(v.Type().Elem()))
		}

		v = v.Elem()
	}

	if v.Kind() == reflect.Interface {
		val, err := p.parseDynamic()
		if err != nil {
			return err
		}

		if v.CanSet() {
			v.Set(val)
		}

		return nil
	}

	switch v.Kind() {
	case reflect.Bool:
		return p.parseBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return p.parseInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return p.parseUint(v)
	case reflect.Float32, reflect.Float64:
		return p.parseFloat(v)
	case reflect.String:
		return p.parseString(v)
	case reflect.Slice:
		return p.parseSlice(v)
	case reflect.Map:
		return p.parseMap(v)
	case reflect.Struct:
		return p.parseStruct(v)
	default:
		return fmt.Errorf("zon: unsupported type %s", v.Type())
	}
}

func (p *parser) parseDynamic() (reflect.Value, error) {
	p.skipWhitespace()

	if bytes.HasPrefix(p.data[p.pos:], []byte("null")) {
		p.pos += 4

		return reflect.Zero(reflect.TypeOf(nil)), nil
	}

	if p.pos >= len(p.data) {
		return reflect.Value{}, fmt.Errorf("zon: unexpected end of input")
	}

	switch p.data[p.pos] {
	case '.':
		if p.pos+1 < len(p.data) && p.data[p.pos+1] == '{' {
			m := make(map[string]any)

			if err := p.parseMapDynamic(m); err != nil {
				return reflect.Value{}, err
			}

			return reflect.ValueOf(m), nil
		}

		return reflect.Value{}, fmt.Errorf("zon: unexpected '.' at pos %d", p.pos)
	case '[':
		var slice []any

		if err := p.parseSliceDynamic(&slice); err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(slice), nil
	case '"':
		var s string

		if err := p.parseStringDynamic(&s); err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(s), nil
	default:
		if bytes.HasPrefix(p.data[p.pos:], []byte("true")) {
			p.pos += 4

			return reflect.ValueOf(true), nil
		}

		if bytes.HasPrefix(p.data[p.pos:], []byte("false")) {
			p.pos += 5

			return reflect.ValueOf(false), nil
		}

		start := p.pos

		for p.pos < len(p.data) && (unicode.IsDigit(rune(p.data[p.pos])) || strings.ContainsRune(".-+eE", rune(p.data[p.pos]))) {
			p.pos++
		}

		numStr := string(p.data[start:p.pos])

		if strings.ContainsAny(numStr, ".eE") {
			f, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("zon: invalid float literal at pos %d: %w", start, err)
			}

			return reflect.ValueOf(f), nil
		}

		i, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("zon: invalid int literal at pos %d: %w", start, err)
		}

		return reflect.ValueOf(i), nil
	}
}

func (p *parser) parseBool(v reflect.Value) error {
	if bytes.HasPrefix(p.data[p.pos:], []byte("true")) {
		v.SetBool(true)
		p.pos += 4

		return nil
	} else if bytes.HasPrefix(p.data[p.pos:], []byte("false")) {
		v.SetBool(false)
		p.pos += 5

		return nil
	}

	return fmt.Errorf("zon: invalid boolean at pos %d", p.pos)
}

func (p *parser) parseInt(v reflect.Value) error {
	start := p.pos

	if p.data[p.pos] == '+' || p.data[p.pos] == '-' {
		p.pos++
	}

	for p.pos < len(p.data) && unicode.IsDigit(rune(p.data[p.pos])) {
		p.pos++
	}

	val, err := strconv.ParseInt(string(p.data[start:p.pos]), 10, v.Type().Bits())
	if err != nil {
		return fmt.Errorf("zon: invalid int literal at pos %d: %w", start, err)
	}

	v.SetInt(val)

	return nil
}

func (p *parser) parseUint(v reflect.Value) error {
	start := p.pos

	for p.pos < len(p.data) && unicode.IsDigit(rune(p.data[p.pos])) {
		p.pos++
	}

	val, err := strconv.ParseUint(string(p.data[start:p.pos]), 10, v.Type().Bits())
	if err != nil {
		return fmt.Errorf("zon: invalid uint literal at pos %d: %w", start, err)
	}

	v.SetUint(val)

	return nil
}

func (p *parser) parseFloat(v reflect.Value) error {
	start := p.pos

	for p.pos < len(p.data) && (unicode.IsDigit(rune(p.data[p.pos])) || strings.ContainsRune(".-+eE", rune(p.data[p.pos]))) {
		p.pos++
	}

	numStr := string(p.data[start:p.pos])
	if numStr == "" {
		return fmt.Errorf("zon: invalid float literal at pos %d", start)
	}

	f, err := strconv.ParseFloat(numStr, v.Type().Bits())
	if err != nil {
		return fmt.Errorf("zon: invalid float literal at pos %d: %w", start, err)
	}

	v.SetFloat(f)
	return nil
}

func (p *parser) parseString(v reflect.Value) error {
	if p.data[p.pos] != '"' {
		return fmt.Errorf("zon: expected '\"' at pos %d", p.pos)
	}

	p.pos++
	start := p.pos

	for p.pos < len(p.data) && p.data[p.pos] != '"' {
		p.pos++
	}

	if p.pos >= len(p.data) {
		return fmt.Errorf("zon: unterminated string")
	}

	v.SetString(string(p.data[start:p.pos]))

	p.pos++

	return nil
}

func (p *parser) parseSlice(v reflect.Value) error {
	if p.data[p.pos] != '[' {
		return fmt.Errorf("zon: expected '[' at pos %d", p.pos)
	}

	p.pos++

	slice := reflect.MakeSlice(v.Type(), 0, 0)

	for {
		p.skipWhitespace()

		if p.pos >= len(p.data) {
			return fmt.Errorf("zon: unexpected end of slice")
		}

		if p.data[p.pos] == ']' {
			p.pos++

			break
		}

		elem := reflect.New(v.Type().Elem()).Elem()

		if err := p.parseValue(elem); err != nil {
			return err
		}

		slice = reflect.Append(slice, elem)

		p.skipWhitespace()

		if p.pos < len(p.data) && p.data[p.pos] == ',' {
			p.pos++

			p.skipWhitespace()

			if p.pos < len(p.data) && p.data[p.pos] == ']' {
				p.pos++

				break
			}
		}
	}

	v.Set(slice)

	return nil
}

func (p *parser) parseSliceDynamic(out *[]any) error {
	if p.data[p.pos] != '[' {
		return fmt.Errorf("zon: expected '[' at pos %d", p.pos)
	}

	p.pos++

	var elems []any

	for {
		p.skipWhitespace()

		if p.pos >= len(p.data) {
			return fmt.Errorf("zon: unexpected end of slice")
		}

		if p.data[p.pos] == ']' {
			p.pos++

			break
		}

		elem, err := p.parseDynamic()
		if err != nil {
			return err
		}

		elems = append(elems, elem)

		p.skipWhitespace()

		if p.pos < len(p.data) && p.data[p.pos] == ',' {
			p.pos++

			p.skipWhitespace()

			if p.pos < len(p.data) && p.data[p.pos] == ']' {
				p.pos++

				break
			}
		}
	}

	*out = elems

	return nil
}

func (p *parser) parseMap(v reflect.Value) error {
	if !(p.data[p.pos] == '.' && p.pos+1 < len(p.data) && p.data[p.pos+1] == '{') {
		return fmt.Errorf("zon: expected '.{' at pos %d", p.pos)
	}

	p.pos += 2

	v.Set(reflect.MakeMap(v.Type()))

	for {
		p.skipWhitespace()

		if p.pos >= len(p.data) {
			return fmt.Errorf("zon: unexpected end of map")
		}

		if p.data[p.pos] == '}' {
			p.pos++

			break
		}

		if p.data[p.pos] != '.' {
			return fmt.Errorf("zon: expected '.' at pos %d", p.pos)
		}

		p.pos++

		start := p.pos

		for p.pos < len(p.data) && (unicode.IsLetter(rune(p.data[p.pos])) || unicode.IsDigit(rune(p.data[p.pos])) || p.data[p.pos] == '_') {
			p.pos++
		}

		key := string(p.data[start:p.pos])

		p.skipWhitespace()

		if p.pos >= len(p.data) || p.data[p.pos] != '=' {
			return fmt.Errorf("zon: expected '=' after key at pos %d", p.pos)
		}

		p.pos++

		p.skipWhitespace()

		val := reflect.New(v.Type().Elem()).Elem()

		if err := p.parseValue(val); err != nil {
			return err
		}

		v.SetMapIndex(reflect.ValueOf(key), val)

		p.skipWhitespace()

		if p.pos < len(p.data) && p.data[p.pos] == ',' {
			p.pos++

			p.skipWhitespace()

			if p.pos < len(p.data) && p.data[p.pos] == '}' {
				p.pos++

				break
			}
		}
	}

	return nil
}

func (p *parser) parseStruct(v reflect.Value) error {
	if !(p.data[p.pos] == '.' && p.pos+1 < len(p.data) && p.data[p.pos+1] == '{') {
		return fmt.Errorf("zon: expected '.{' at pos %d", p.pos)
	}

	p.pos += 2

	t := v.Type()

	for {
		p.skipWhitespace()

		if p.pos >= len(p.data) {
			return fmt.Errorf("zon: unexpected end of struct")
		}

		if p.data[p.pos] == '}' {
			p.pos++

			break
		}

		if p.data[p.pos] != '.' {
			return fmt.Errorf("zon: expected '.' at pos %d", p.pos)
		}

		p.pos++
		start := p.pos

		for p.pos < len(p.data) && (unicode.IsLetter(rune(p.data[p.pos])) || unicode.IsDigit(rune(p.data[p.pos])) || p.data[p.pos] == '_') {
			p.pos++
		}

		key := string(p.data[start:p.pos])

		p.skipWhitespace()

		if p.pos >= len(p.data) || p.data[p.pos] != '=' {
			return fmt.Errorf("zon: expected '=' after key at pos %d", p.pos)
		}

		p.pos++

		p.skipWhitespace()

		var field reflect.Value

		found := false

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)

			if f.PkgPath != "" {
				continue
			}

			name := f.Tag.Get("zon")

			if name == "" {
				name = f.Name
			}

			if name == key {
				field = v.Field(i)
				found = true

				break
			}
		}

		if !found {
			skip := reflect.New(reflect.TypeOf(new(any)).Elem()).Elem()

			_ = p.parseValue(skip)

			continue
		}

		if err := p.parseValue(field); err != nil {
			return err
		}

		p.skipWhitespace()

		if p.pos < len(p.data) && p.data[p.pos] == ',' {
			p.pos++

			p.skipWhitespace()

			if p.pos < len(p.data) && p.data[p.pos] == '}' {
				p.pos++

				break
			}
		}
	}

	return nil
}

func (p *parser) parseMapDynamic(out map[string]any) error {
	if !(p.data[p.pos] == '.' && p.pos+1 < len(p.data) && p.data[p.pos+1] == '{') {
		return fmt.Errorf("zon: expected '.{' at pos %d", p.pos)
	}

	p.pos += 2

	for {
		p.skipWhitespace()

		if p.pos >= len(p.data) {
			return fmt.Errorf("zon: unexpected end of map")
		}

		if p.data[p.pos] == '}' {
			p.pos++

			break
		}

		if p.data[p.pos] != '.' {
			return fmt.Errorf("zon: expected '.' for map key at pos %d", p.pos)
		}

		p.pos++

		start := p.pos

		for p.pos < len(p.data) && (unicode.IsLetter(rune(p.data[p.pos])) || unicode.IsDigit(rune(p.data[p.pos])) || p.data[p.pos] == '_') {
			p.pos++
		}

		key := string(p.data[start:p.pos])

		p.skipWhitespace()

		if p.pos >= len(p.data) || p.data[p.pos] != '=' {
			return fmt.Errorf("zon: expected '=' after key at pos %d", p.pos)
		}

		p.pos++

		p.skipWhitespace()

		val, err := p.parseDynamic()
		if err != nil {
			return err
		}

		out[key] = val.Interface()

		p.skipWhitespace()

		if p.pos < len(p.data) && p.data[p.pos] == ',' {
			p.pos++

			p.skipWhitespace()

			if p.pos < len(p.data) && p.data[p.pos] == '}' {
				p.pos++

				break
			}
		}
	}

	return nil
}

func (p *parser) parseStringDynamic(out *string) error {
	if p.data[p.pos] != '"' {
		return fmt.Errorf("zon: expected '\"' at pos %d", p.pos)
	}

	p.pos++

	start := p.pos

	for p.pos < len(p.data) && p.data[p.pos] != '"' {
		p.pos++
	}

	if p.pos >= len(p.data) {
		return fmt.Errorf("zon: unterminated string")
	}

	*out = string(p.data[start:p.pos])

	p.pos++

	return nil
}
