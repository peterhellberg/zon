package zon

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"
)

type parser struct {
	data []byte
	pos  int
}

func (p *parser) parseValue(v reflect.Value) error {
	p.skipSpace()

	if p.pos >= len(p.data) {
		return fmt.Errorf("zon: unexpected end of input")
	}

	if hasPrefixAt(p.data, p.pos, "null") {
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

func (p *parser) parseBool(v reflect.Value) error {
	if hasPrefixAt(p.data, p.pos, "true") {
		v.SetBool(true)

		p.pos += 4

		return nil
	} else if hasPrefixAt(p.data, p.pos, "false") {
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

	if hasPrefixAt(p.data, p.pos, "0x") || hasPrefixAt(p.data, p.pos, "0X") {
		p.pos += 2

		for p.pos < len(p.data) && isHexDigit(p.data[p.pos]) {
			p.pos++
		}

		val, err := strconv.ParseInt(string(p.data[start:p.pos]), 0, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("zon: invalid hex int literal at pos %d: %w", start, err)
		}

		v.SetInt(val)

		return nil
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

	if hasPrefixAt(p.data, p.pos, "0x") || hasPrefixAt(p.data, p.pos, "0X") {
		p.pos += 2

		for p.pos < len(p.data) && isHexDigit(p.data[p.pos]) {
			p.pos++
		}

		val, err := strconv.ParseUint(string(p.data[start:p.pos]), 0, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("zon: invalid hex uint literal at pos %d: %w", start, err)
		}

		v.SetUint(val)

		return nil
	}

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

	for p.pos < len(p.data) && (unicode.IsDigit(rune(p.data[p.pos])) ||
		containsRune("+-eE.", rune(p.data[p.pos]))) {
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
	if !hasPrefixAt(p.data, p.pos, ".{") {
		return fmt.Errorf("zon: expected '.{' at pos %d", p.pos)
	}

	p.pos += 2

	// Always start with an empty slice (non-nil).
	slice := reflect.MakeSlice(v.Type(), 0, 0)

	for {
		p.skipSpace()
		if p.pos >= len(p.data) {
			return fmt.Errorf("zon: unexpected end of slice")
		}

		if p.data[p.pos] == '}' {
			p.pos++
			break
		}

		if p.data[p.pos] == ',' {
			p.pos++
			continue
		}

		elem := reflect.New(v.Type().Elem()).Elem()
		if err := p.parseValue(elem); err != nil {
			return err
		}

		slice = reflect.Append(slice, elem)
	}

	// Ensure non-nil slice is set
	v.Set(slice)

	return nil
}

func (p *parser) parseMap(v reflect.Value) error {
	if !hasPrefixAt(p.data, p.pos, ".{") {
		return fmt.Errorf("zon: expected '.{' at pos %d", p.pos)
	}

	p.pos += 2

	v.Set(reflect.MakeMap(v.Type()))

	for {
		p.skipSpace()

		if p.pos >= len(p.data) {
			return fmt.Errorf("zon: unexpected end of map")
		}

		if p.data[p.pos] == '}' {
			p.pos++

			break
		}

		if p.data[p.pos] == ',' {
			p.pos++

			continue
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

		p.skipSpace()

		if p.pos >= len(p.data) || p.data[p.pos] != '=' {
			return fmt.Errorf("zon: expected '=' after key at pos %d", p.pos)
		}

		p.pos++

		p.skipSpace()

		val := reflect.New(v.Type().Elem()).Elem()

		if err := p.parseValue(val); err != nil {
			return err
		}

		v.SetMapIndex(reflect.ValueOf(key), val)
	}

	return nil
}

func (p *parser) parseStruct(v reflect.Value) error {
	if !hasPrefixAt(p.data, p.pos, ".{") {
		return fmt.Errorf("zon: expected '.{' at pos %d", p.pos)
	}

	p.pos += 2

	t := v.Type()

	for {
		p.skipSpace()

		if p.pos >= len(p.data) {
			return fmt.Errorf("zon: unexpected end of struct")
		}

		if p.data[p.pos] == '}' {
			p.pos++

			break
		}

		if p.data[p.pos] == ',' {
			p.pos++

			continue
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

		p.skipSpace()

		if p.pos >= len(p.data) || p.data[p.pos] != '=' {
			return fmt.Errorf("zon: expected '=' after key at pos %d", p.pos)
		}

		p.pos++

		p.skipSpace()

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
	}

	return nil
}

func (p *parser) parseDynamic() (reflect.Value, error) {
	p.skipSpace()

	if p.pos >= len(p.data) {
		return reflect.Value{}, fmt.Errorf("zon: unexpected end of input")
	}

	c := p.data[p.pos]

	switch c {
	case '"':
		var s string

		if err := p.parseStringDynamic(&s); err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(s), nil
	case '.':
		if p.pos+1 < len(p.data) && p.data[p.pos+1] == '{' {
			return p.parseDynamicMapOrSlice()
		}

		start := p.pos + 1

		p.pos++

		for p.pos < len(p.data) && (unicode.IsLetter(rune(p.data[p.pos])) || unicode.IsDigit(rune(p.data[p.pos])) || p.data[p.pos] == '_') {
			p.pos++
		}

		ident := string(p.data[start:p.pos])
		if ident == "" {
			return reflect.Value{}, fmt.Errorf("zon: invalid identifier at pos %d", start)
		}

		return reflect.ValueOf(ident), nil
	default:
		if isDigit(c) || c == '+' || c == '-' {
			return p.parseNumberDynamic()
		} else if hasPrefixAt(p.data, p.pos, "true") {
			p.pos += 4

			return reflect.ValueOf(true), nil
		} else if hasPrefixAt(p.data, p.pos, "false") {
			p.pos += 5

			return reflect.ValueOf(false), nil
		} else {
			start := p.pos

			for p.pos < len(p.data) && (unicode.IsLetter(rune(p.data[p.pos])) || unicode.IsDigit(rune(p.data[p.pos])) || p.data[p.pos] == '_') {
				p.pos++
			}

			ident := string(p.data[start:p.pos])
			if ident == "" {
				return reflect.Value{}, fmt.Errorf("zon: unexpected token at pos %d", start)
			}

			return reflect.ValueOf(ident), nil
		}
	}
}

func (p *parser) parseNumberDynamic() (reflect.Value, error) {
	start := p.pos

	if p.data[p.pos] == '+' || p.data[p.pos] == '-' {
		p.pos++
	}

	if hasPrefixAt(p.data, p.pos, "0x") || hasPrefixAt(p.data, p.pos, "0X") {
		hexStart := p.pos
		p.pos += 2

		for p.pos < len(p.data) && isHexDigit(p.data[p.pos]) {
			p.pos++
		}

		hexStr := string(p.data[hexStart:p.pos])

		val, err := strconv.ParseUint(hexStr, 0, 64)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("zon: invalid hex literal at pos %d: %w", hexStart, err)
		}

		return reflect.ValueOf(val), nil
	}

	dotSeen := false

	for p.pos < len(p.data) && (unicode.IsDigit(rune(p.data[p.pos])) || p.data[p.pos] == '.' || p.data[p.pos] == 'e' || p.data[p.pos] == 'E' || p.data[p.pos] == '+' || p.data[p.pos] == '-') {
		if p.data[p.pos] == '.' {
			if dotSeen {
				break
			}

			dotSeen = true
		}

		p.pos++
	}

	numStr := string(p.data[start:p.pos])
	if numStr == "" {
		return reflect.Value{}, fmt.Errorf("zon: invalid number at pos %d", start)
	}

	if dotSeen || containsRune(numStr, 'e') || containsRune(numStr, 'E') {
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

func (p *parser) parseDynamicMapOrSlice() (reflect.Value, error) {
	p.pos += 2

	p.skipSpace()

	isMap := false

	if p.pos < len(p.data) && p.data[p.pos] == '.' {
		isMap = true
	}

	if isMap {
		m := make(map[string]any)
		for {
			p.skipSpace()

			if p.pos >= len(p.data) {
				return reflect.Value{}, fmt.Errorf("zon: unexpected end of map")
			}

			if p.data[p.pos] == '}' {
				p.pos++

				break
			}

			if p.data[p.pos] == ',' {
				p.pos++

				continue
			}

			if p.data[p.pos] != '.' {
				return reflect.Value{}, fmt.Errorf("zon: expected '.' for map key at pos %d", p.pos)
			}

			p.pos++

			start := p.pos

			for p.pos < len(p.data) && (unicode.IsLetter(rune(p.data[p.pos])) || unicode.IsDigit(rune(p.data[p.pos])) || p.data[p.pos] == '_') {
				p.pos++
			}

			key := string(p.data[start:p.pos])

			p.skipSpace()

			if p.pos >= len(p.data) || p.data[p.pos] != '=' {
				return reflect.Value{}, fmt.Errorf("zon: expected '=' after key at pos %d", p.pos)
			}

			p.pos++

			p.skipSpace()

			val, err := p.parseDynamic()
			if err != nil {
				return reflect.Value{}, err
			}

			m[key] = val.Interface()
		}

		return reflect.ValueOf(m), nil
	}

	var arr []any

	for {
		p.skipSpace()

		if p.pos >= len(p.data) {
			return reflect.Value{}, fmt.Errorf("zon: unexpected end of slice")
		}

		if p.data[p.pos] == '}' {
			p.pos++
			break
		}

		if p.data[p.pos] == ',' {
			p.pos++
			continue
		}

		elem, err := p.parseDynamic()
		if err != nil {
			return reflect.Value{}, err
		}

		arr = append(arr, elem.Interface())
	}

	if arr == nil {
		arr = []any{}
	}

	return reflect.ValueOf(arr), nil
}

func (p *parser) skipSpace() {
	for p.pos < len(p.data) {
		if unicode.IsSpace(rune(p.data[p.pos])) {
			p.pos++

			continue
		}

		if p.data[p.pos] == '/' && p.pos+1 < len(p.data) && p.data[p.pos+1] == '/' {
			p.pos += 2

			for p.pos < len(p.data) && p.data[p.pos] != '\n' {
				p.pos++
			}

			continue
		}

		break
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}

	return false
}

func hasPrefixAt(data []byte, pos int, prefix string) bool {
	if pos+len(prefix) > len(data) {
		return false
	}

	for i := range prefix {
		if data[pos+i] != prefix[i] {
			return false
		}
	}

	return true
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isHexDigit(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
